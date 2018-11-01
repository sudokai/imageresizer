package main

import (
	"context"
	"flag"
	"github.com/cloudflare/tableflip"
	"github.com/kxlt/imageresizer/api"
	"github.com/kxlt/imageresizer/config"
	"github.com/kxlt/imageresizer/imager"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	defer imager.ShutdownVIPS()

	configPath := flag.String("c", "config.properties", "configuration file path")
	flag.Parse()

	viper.SetConfigFile(*configPath)
	_, err := os.Stat(*configPath)
	if err == nil {
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatalln(err)
		}
	}
	config.RefreshConfig()

	upg, err := tableflip.New(tableflip.Options{})
	if err != nil {
		log.Fatalln(err)
	}
	defer upg.Stop()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			err := upg.Upgrade()
			if err != nil {
				log.Println("Upgrade failed", err)
				continue
			}
			log.Println("Upgrade succeeded")
		}
	}()

	ln, err := upg.Fds.Listen("tcp", config.C.ServerAddr)
	if err != nil {
		log.Fatalln("Can't listen:", err)
	}

	ready := make(chan bool, 1)
	server := &http.Server{Handler: api.NewApi(ready)}

	go server.Serve(ln)

	if !<-ready {
		log.Fatalln(err)
	}

	if err := upg.Ready(); err != nil {
		log.Fatalln(err)
	}
	log.Printf("Ready on %s", viper.GetString("server.addr"))
	<-upg.Exit()

	time.AfterFunc(30*time.Second, func() {
		os.Exit(1)
	})

	_ = server.Shutdown(context.Background())
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"github.com/cloudflare/tableflip"
	"github.com/kailt/imageresizer/api"
	"github.com/kailt/imageresizer/imager"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("server.addr", ":8080")
	viper.SetDefault("store.file.originals", "./images/originals")
	viper.SetDefault("store.file.cache", "./images/cache")
	viper.SetDefault("store.file.thumbnails", "./images/thumbnails")
	viper.SetConfigFile("config.properties")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("IMAGERESIZER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	_, err := os.Stat("config.properties")
	if err == nil {
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	defer imager.ShutdownVIPS()

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

	ln, err := upg.Fds.Listen("tcp", viper.GetString("server.addr"))
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

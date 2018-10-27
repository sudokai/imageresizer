package api

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func (api *Api) newFileWatcher() {
	var err error
	api.FileWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-api.FileWatcher.Events:
				if !ok {
					return
				}
				switch event.Op {
				case fsnotify.Write:
					buf, err := ioutil.ReadFile(event.Name)
					if err != nil {
						log.Fatalf(
							"error while reading file: %s with error: %s",
							event.Name, err)
					}
					_, fileName := filepath.Split(event.Name)
					err = api.Originals.Put(fileName, buf)
					if err != nil {
						log.Fatal(err)
					}
				case fsnotify.Remove:
					_, fileName := filepath.Split(event.Name)
					err := api.Originals.Remove(fileName)
					if err != nil {
						log.Fatal(err)
					}
					api.removeThumbnails(fileName)
				}
			case err, ok := <-api.FileWatcher.Errors:
				if !ok {
					return
				}
				log.Printf("fileWatcher error: %s", err)
			}
		}
	}()
}

func (api *Api) startWatchingFile(filePath string) {
	err := api.FileWatcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}
}

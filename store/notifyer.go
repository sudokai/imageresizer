package store

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

func (fs *FileStore) startSubDirectoriesFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {

			// iterate over all subdirectories under root path
			if err := filepath.Walk(fs.root,
				func(path string, fi os.FileInfo, err error) error {
					if fi.Mode().IsDir() {
						return watcher.Add(path)
					}
					return nil
				}); err != nil {

				log.Fatalf("ERROR: %s", err)
			}

			// gather events
			select {
			case event := <-watcher.Events:
				switch event.Op {

				case fsnotify.Write:
					buf, err := ioutil.ReadFile(event.Name)
					if err != nil {
						log.Fatalf(
							"error while reading file: %s with error: %s",
							event.Name, err)
					}
					size := int64(len(buf))

					// update metadata
					fs.metadata.Put(event.Name, file{filename: event.Name, size: size, atime: time.Now()})
					atomic.AddInt64(&fs.size, size)

				case fsnotify.Remove:
					p := fs.metadata.Get(event.Name)
					if p != nil {
						m := p.(file)
						atomic.AddInt64(&fs.size, -m.size)
						fs.metadata.Remove(event.Name)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Printf("fileWatcher error: %s", err)
			}
		}
	}()
}

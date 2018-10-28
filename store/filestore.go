package store

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/djherbis/atime"
	"github.com/fsnotify/fsnotify"
	"github.com/kailt/imageresizer/collections"
)

const (
	loadBatch         = 50
	loadSleepDuration = 50 * time.Millisecond
)

type FileStore struct {
	root     string
	metadata *collections.SyncMap
	size     int64
	maxSize  int64
}

type file struct {
	filename string
	atime    time.Time
	size     int64
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewFileStore(root string, maxSize int64) *FileStore {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		os.MkdirAll(root, 0755)
	}
	fs := &FileStore{
		root:     root,
		metadata: collections.NewSyncMap(),
		maxSize:  maxSize,
	}
	fs.startSubDirectoriesFileWatcher()
	return fs

}

func (fs *FileStore) startSubDirectoriesFileWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

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

	go func() {
		for {

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

				case fsnotify.Create:
					fi, _ := os.Stat(event.Name)
					if fi.Mode().IsDir() {
						watcher.Add(event.Name)
						fs.checkDirForChilds(watcher, event.Name)
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

func (fs *FileStore) checkDirForChilds(wacher *fsnotify.Watcher, path string) string {
	files, err := ioutil.ReadDir(path)
	if err != nil {

		// if err we've got file not a dir
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf(
				"error while reading file: %s with error: %s",
				path, err)
		}
		size := int64(len(buf))

		fs.metadata.Put(path, file{filename: path, size: size, atime: time.Now()})
		atomic.AddInt64(&fs.size, size)

		return ""
	}

	if len(files) == 0 {
		return ""
	}

	for _, file := range files {

		if file.IsDir() {
			wacher.Add(path + "/" + file.Name())
		}

		return fs.checkDirForChilds(wacher, path+"/"+file.Name())
	}

	return ""
}

func (s *FileStore) Get(filename string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path.Join(s.root, filename))
	if err != nil {
		if s.metadata.HasKey(filename) {
			s.metadata.Remove(filename)
		}
		return nil, err
	}

	if !s.metadata.HasKey(filename) {
		s.metadata.Put(filename, file{filename: filename, size: int64(len(buf)), atime: time.Now()})
	} else {
		// update timestamp
		file := s.metadata.Get(filename).(file)
		file.atime = time.Now()
		s.metadata.Put(filename, file)
	}

	return buf, nil
}

func (s *FileStore) Put(filename string, buf []byte) error {
	fullpath := path.Join(s.root, filename)
	err := os.MkdirAll(path.Dir(fullpath), 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fullpath, buf, 0644)
	if err != nil {
		return err
	}
	size := int64(len(buf))
	s.metadata.Put(filename, file{filename: filename, size: size, atime: time.Now()})
	atomic.AddInt64(&s.size, size)
	return nil
}

func (s *FileStore) Remove(filename string) error {
	err := os.Remove(path.Join(s.root, filename))
	if err != nil {
		return err
	}
	p := s.metadata.Get(filename)
	if p != nil {
		m := p.(file)
		atomic.AddInt64(&s.size, -m.size)
		s.metadata.Remove(filename)
	}
	return nil
}

func (s *FileStore) PruneCache() error {
	var oldest *file
	if s.maxSize <= 0 || atomic.LoadInt64(&s.size) <= s.maxSize {
		return nil
	}
	for i := 0; i < 5; i++ {
		f := s.metadata.GetRand().(file)
		if oldest == nil || f.atime.Before(oldest.atime) {
			oldest = &f
		}
	}
	return s.Remove(oldest.filename)
}

func (s *FileStore) LoadCache(walkFn func(item interface{}) error) error {
	count := 0
	return filepath.Walk(s.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := strings.Split(path, s.root+"/")[1]
		s.metadata.Put(filename, file{filename: filename, size: info.Size(), atime: atime.Get(info)})
		s.size += info.Size()
		if walkFn != nil {
			walkFn(filename)
		}
		count++
		if count%loadBatch == 0 {
			time.Sleep(loadSleepDuration)
			count = 0
		}
		return nil
	})
}

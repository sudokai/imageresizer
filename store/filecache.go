package store

import (
	"github.com/djherbis/atime"
	"github.com/kxlt/imageresizer/collections"
	"github.com/kxlt/imageresizer/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)
import "path"

type FileCache struct {
	root     string
	metadata collections.EvictableMap
	size     int64
	maxSize  int64
}

func NewFileCache(root string, maxSize int64) *FileCache {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return nil
		}
	}
	fc := &FileCache{
		root:     root,
		metadata: collections.NewShardedMap(256),
		maxSize:  maxSize,
	}
	go fc.loadCache()
	return fc
}

func (fc *FileCache) Get(filename string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path.Join(fc.root, filename))
	if err != nil {
		if fc.metadata.HasKey(filename) {
			fc.metadata.Remove(filename)
		}
		return nil, err
	}

	if !fc.metadata.HasKey(filename) {
		fc.metadata.Put(filename, &Metadata{filename: filename, size: int64(len(buf)), atime: time.Now()})
	} else {
		// update timestamp
		file := fc.metadata.Get(filename).(*Metadata)
		file.atime = time.Now()
		fc.metadata.Put(filename, file)
	}

	return buf, nil
}

func (fc *FileCache) Put(filename string, buf []byte) error {
	fullpath := path.Join(fc.root, filename)
	err := os.MkdirAll(path.Dir(fullpath), 0755)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fullpath, buf, 0644)
	if err != nil {
		return err
	}
	size := int64(len(buf))
	fc.metadata.Put(filename, &Metadata{filename: filename, size: size, atime: time.Now()})
	atomic.AddInt64(&fc.size, size)
	fc.pruneCache()
	return nil
}

func (fc *FileCache) Remove(filename string) error {
	err := os.Remove(path.Join(fc.root, filename))
	if err != nil {
		return err
	}
	p := fc.metadata.Get(filename)
	if p != nil {
		m := p.(*Metadata)
		atomic.AddInt64(&fc.size, -m.size)
		fc.metadata.Remove(filename)
	}
	return nil
}

func (fc *FileCache) pruneCache() error {
	var oldest *Metadata
	if fc.maxSize <= 0 || atomic.LoadInt64(&fc.size) <= fc.maxSize {
		return nil
	}
	for i := 0; i < 10; i++ {
		metadata := fc.metadata.GetEvictable().(*Metadata)
		if oldest == nil || metadata.atime.Before(oldest.atime) {
			oldest = metadata
		}
	}
	return fc.Remove(oldest.filename)
}

func (fc *FileCache) loadCache() error {
	count := 0
	t := time.Now()
	threshold := time.Duration(config.C.CacheLoaderThreshold) * time.Millisecond
	return filepath.Walk(fc.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		filename := strings.Split(path, fc.root+"/")[1]
		fc.metadata.Put(filename, &Metadata{filename: filename, size: info.Size(), atime: atime.Get(info)})
		atomic.AddInt64(&fc.size, info.Size())
		fc.pruneCache()
		count++
		if count%config.C.CacheLoaderFiles == 0 || time.Since(t) > threshold {
			time.Sleep(time.Duration(config.C.CacheLoaderSleep) * time.Millisecond)
			count = 0
			t = time.Now()
		}
		return nil
	})
}

package store

import (
	"io/ioutil"
	"os"
)
import "path"

type FileStore struct {
	root string
}

func NewFileStore(root string) *FileStore {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		os.MkdirAll(root, 0755)
	}
	return &FileStore{
		root: root,
	}
}

func (s *FileStore) Get(filename string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path.Join(s.root, filename))
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (s *FileStore) Put(filename string, buf []byte) error {
	fullpath := path.Join(s.root, filename)
	err := os.MkdirAll(path.Dir(fullpath), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fullpath, buf, 0644)
}

func (s *FileStore) Remove(filename string) error {
	return os.Remove(path.Join(s.root, filename))
}

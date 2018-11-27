package store

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
)

type TwoTier struct {
	Store Store
	Cache Cache
}

func (s *TwoTier) Get(aurl string) ([]byte, error) {
	var buf []byte
	var err error
	filename, isRemote := s.RemoteToLocalPath(aurl)

	if s.Cache != nil {
		buf, _ = s.Cache.Get(filename)
	}
	if buf == nil {
		buf, err = s.Store.Get(filename)
		if err != nil {
			if !isRemote {
				return nil, err
			} else {
				response, err := http.Get(aurl)
				if err != nil {
					return nil, err
				}
				if response.StatusCode >= 400 {
					return nil, errors.New("(" + response.Request.URL.String() + ") HTTP Error: " + response.Status)
				}
				defer response.Body.Close()
				buf, err = ioutil.ReadAll(response.Body)
				if err != nil {
					return nil, err
				}
				err = s.Store.Put(filename, buf)
			}
		}
		if s.Cache != nil {
			go s.Cache.Put(filename, buf)
		}
	}
	return buf, nil
}

func (s *TwoTier) Put(filename string, data []byte) error {
	err := s.Store.Put(filename, data)
	if err != nil {
		return err
	}
	if s.Cache != nil {
		go s.Cache.Put(filename, data)
	}
	return nil
}

func (s *TwoTier) Remove(filename string) error {
	if s.Cache != nil {
		go s.Cache.Remove(filename)
	}
	return s.Store.Remove(filename)
}

func (s *TwoTier) PruneCache() error {
	if s.Cache == nil {
		return nil
	}
	return s.Cache.PruneCache()
}

func (s *TwoTier) LoadCache(walkFn func(item interface{}) error) error {
	if s.Cache == nil {
		return nil
	}
	return s.Cache.LoadCache(walkFn)
}

func (s *TwoTier) RemoteToLocalPath(path string) (string, bool) {
	var localPath string
	var isRemote bool

	aurl, _ := url.Parse(path)
	if aurl.Scheme != "" {
		isRemote = true
		localPath = filepath.Join(aurl.Hostname(), aurl.EscapedPath())
	} else {
		isRemote = false
		localPath = path
	}
	return localPath, isRemote
}

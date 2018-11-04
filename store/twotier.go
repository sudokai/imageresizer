package store

type TwoTier struct {
	Store Store
	Cache Cache
}

func (s *TwoTier) Get(filename string) ([]byte, error) {
	var buf []byte
	var err error
	if s.Cache != nil {
		buf, _ = s.Cache.Get(filename)
	}
	if buf == nil {
		buf, err = s.Store.Get(filename)
		if err != nil {
			return nil, err
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

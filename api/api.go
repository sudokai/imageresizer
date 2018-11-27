package api

import (
	"github.com/gorilla/mux"
	"github.com/kxlt/imageresizer/collections"
	"github.com/kxlt/imageresizer/config"
	"github.com/kxlt/imageresizer/store"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"log"
	"path"
	"time"
)

func init() {
	exp.Exp(metrics.DefaultRegistry)
}

// Api type embeds a router
type Api struct {
	Originals  *store.TwoTier
	Thumbnails store.Cache
	Tiers      *collections.SyncStrSet
	Etags      *collections.SyncStrSet
	*mux.Router
}

func NewApi(ready chan<- bool) *Api {
	var origStore store.Store
	if config.C.S3Enable {
		var err error
		origStore, err = store.NewS3Store(&store.S3Config{
			Region: config.C.S3Region,
			Bucket: config.C.S3Bucket,
			Prefix: config.C.S3Prefix,
		})
		if err != nil {
			log.Fatalln("S3 store could not be initialized")
		}
	} else {
		origStore = store.NewFileStore(config.C.LocalPrefix)
	}
	var origCache store.Cache
	if config.C.CacheOrigEnable {
		origCache = store.NewFileCache(
			config.C.CacheOrigPath,
			config.C.CacheOrigMaxSize,
			config.C.CacheOrigShards)
	}
	var thumbCache store.Cache
	if config.C.CacheThumbEnable {
		thumbCache = store.NewFileCache(
			config.C.CacheThumbPath,
			config.C.CacheThumbMaxSize,
			config.C.CacheThumbShards)
	} else {
		thumbCache = &store.NoopCache{}
	}
	var etags *collections.SyncStrSet
	if config.C.EtagCacheEnable {
		etags = collections.NewSyncStrSet()
	}
	api := &Api{
		Originals: &store.TwoTier{
			Store: origStore,
			Cache: origCache,
		},
		Thumbnails: thumbCache,
		Tiers:      collections.NewSyncStrSet(),
		Etags:      etags,
		Router:     mux.NewRouter().StrictSlash(true).SkipClean(true),
	}
	go api.initCacheLoader(ready)
	api.initCacheManager()
	if config.C.EtagCacheEnable {
		api.initEtagManager()
	}
	api.routes()
	return api
}

func (api *Api) initCacheLoader(ready chan<- bool) {
	log.Println("Loading caches...")
	err := api.Originals.LoadCache(nil)
	if err != nil {
		ready <- false
		return
	}
	api.Thumbnails.LoadCache(func(item interface{}) error {
		filename := item.(string)
		api.Tiers.Add(path.Dir(filename))
		return nil
	})
	if err != nil {
		ready <- false
		return
	}
	ready <- true
	log.Println("Caches loaded")
}

func (api *Api) initCacheManager() {
	go func() {
		for range time.Tick(50 * time.Millisecond) {
			api.Originals.PruneCache()
			api.Thumbnails.PruneCache()
		}
	}()
}

func (api *Api) initEtagManager() {
	go func() {
		for range time.Tick(1 * time.Second) {
			numKeysToRemove := api.Etags.Size() - config.C.EtagCacheMaxSize
			if numKeysToRemove > 0 {
				for i := 0; i < numKeysToRemove; i++ {
					api.Etags.Remove(api.Etags.Get())
				}
			}
		}
	}()
}

func (api *Api) removeThumbnails(filePath string) {
	api.Tiers.Walk(func(item string) {
		api.Thumbnails.Remove(item + "/" + filePath)
	})
}

package api

import (
	"github.com/gorilla/mux"
	"github.com/kailt/imageresizer/collections"
	"github.com/kailt/imageresizer/store"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/spf13/viper"
	"log"
	"path"
	"time"
)

func init() {
	exp.Exp(metrics.DefaultRegistry)
}

// Api type embeds a router
type Api struct {
	Originals  *store.CachedStore
	Thumbnails store.Cache
	Tiers      *collections.SyncStrSet
	Etags      *collections.SyncStrSet
	*mux.Router
}

func NewApi(ready chan<- bool) *Api {
	api := &Api{
		Originals: &store.CachedStore{
			Store: store.NewFileStore(viper.GetString("store.file.originals"), 0),
			Cache: store.NewFileStore(viper.GetString("store.file.cache"), 0),
		},
		Thumbnails: store.NewFileStore(viper.GetString("store.file.thumbnails"), 0),
		Tiers:      collections.NewSyncStrSet(),
		Etags:      collections.NewSyncStrSet(),
		Router:     mux.NewRouter().StrictSlash(true),
	}
	go api.initCacheLoader(ready)
	api.initCacheManager()
	api.initEtagManager()
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
			numKeysToRemove := api.Etags.Size() - 200000
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

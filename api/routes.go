package api

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/kailt/imageresizer/imagine"
	"github.com/rcrowley/go-metrics"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (api *Api) routes() {
	api.Handle("/favicon.ico", api.handle404())
	api.Handle("/debug/metrics", http.DefaultServeMux)
	api.HandleFunc("/{width:[0-9]*}x{height:[0-9]*}/{gravity}/{path}", api.serveThumbs()).
		Methods("GET", "HEAD")
	api.HandleFunc("/{path}", api.serveOriginals()).
		Methods("GET", "HEAD")
	api.HandleFunc("/{path}", api.handleCreates()).Methods("POST")
	api.HandleFunc("/{path}", api.handleDeletes()).Methods("DELETE")
}

func (api *Api) serveOriginals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := metrics.GetOrRegisterTimer("api.originals.latency", nil)
		t.Time(func() {
			vars := mux.Vars(r)
			buf, err := api.Originals.Get(vars["path"])
			if err != nil {
				if os.IsNotExist(err) {
					respondWithErr(w, http.StatusNotFound)
				} else {
					respondWithErr(w, http.StatusInternalServerError)
				}
				return
			}
			respondWithImage(w, imagine.DetermineImageType(buf), buf)
		})
	}
}

func (api *Api) serveThumbs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := metrics.GetOrRegisterTimer("api.thumbs.latency", nil)
		t.Time(func() {
			vars := mux.Vars(r)
			resizeTier := vars["width"] + "x" + vars["height"] + "/" +
				vars["gravity"] + "/"
			path := vars["path"]
			thumbPath := resizeTier + path
			api.Tiers.Add(resizeTier)
			thumbBuf, err := api.Thumbnails.Get(thumbPath)
			if err != nil {
				srcBuf, err := api.Originals.Get(path)
				if err != nil {
					respondWithErr(w, http.StatusNotFound)
					return
				}
				options, err := parseParams(vars)
				if err != nil {
					respondWithErr(w, http.StatusBadRequest)
					return
				}
				thumbBuf, err = imagine.Resize(srcBuf, options)
				if err != nil {
					respondWithErr(w, http.StatusInternalServerError)
					return
				}
				api.Thumbnails.Put(thumbPath, thumbBuf)
			}
			respondWithImage(w, imagine.DetermineImageType(thumbBuf), thumbBuf)
		})
	}
}

func (api *Api) handleCreates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			reader io.Reader
			filename string
		)
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			file, _, err := r.FormFile("file")
			if err != nil {
				respondWithErr(w, http.StatusBadRequest)
				return
			}
			reader = file
		} else {
			reader = r.Body
		}
		filename = mux.Vars(r)["path"]
		limitReader := io.LimitReader(reader, 50*1024*1024)
		buf, err := ioutil.ReadAll(limitReader)
		if err != nil {
			respondWithErr(w, http.StatusBadRequest)
			return
		}
		_, err = limitReader.Read(nil)
		if err == io.EOF {
			respondWithErr(w, http.StatusRequestEntityTooLarge)
		}
		err = api.Originals.Put(filename, buf)
		if err != nil {
			respondWithErr(w, http.StatusInternalServerError)
			return
		}
		respondWithStatusCode(w, http.StatusCreated)
	}
}

func (api *Api) handleDeletes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := metrics.GetOrRegisterTimer("api.deletes.latency", nil)
		t.Time(func() {
			vars := mux.Vars(r)
			path := vars["path"]
			err := api.Originals.Remove(path)
			if err != nil {
				respondWithErr(w, http.StatusNotFound)
			}
			api.removeThumbnails(path)
			respondWithStatusCode(w, http.StatusNoContent)
		})
	}
}

func (api *Api) handle404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondWithErr(w, http.StatusNotFound)
	}
}

func parseParams(vars map[string]string) (imagine.Options, error) {
	width, err := strconv.Atoi(vars["width"])
	if err != nil {
		return imagine.Options{}, err
	}
	height, err := strconv.Atoi(vars["height"])
	if err != nil {
		return imagine.Options{}, err
	}
	gravity, ok := imagine.Gravity[vars["gravity"]]
	if !ok {
		return imagine.Options{}, errors.New("invalid gravity")
	}
	options := imagine.Options{
		Width:   width,
		Height:  height,
		Gravity: gravity,
	}
	return options, nil
}

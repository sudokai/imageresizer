package api

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/mux"
	"github.com/kxlt/imageresizer/etag"
	"github.com/kxlt/imageresizer/imager"
	"github.com/rcrowley/go-metrics"
)

const uploadSizeLimit = 50 * 1024 * 1024
const pathMatch = "{path:[a-zA-Z/\\.]+}"

func (api *Api) routes() {
	api.Handle("/favicon.ico", api.handle404())
	api.Handle("/debug/metrics", http.DefaultServeMux)
	// shortcut
	api.HandleFunc("/{width:[1-9][0-9]*}/{resizeOp}/{options}/" + pathMatch,
		api.etagMiddleware(api.serveThumbs())).Methods("GET", "HEAD")
	api.HandleFunc("/{width:[1-9][0-9]*}x{height:[1-9][0-9]*}/{resizeOp}/{options}/" + pathMatch,
		api.etagMiddleware(api.serveThumbs())).Methods("GET", "HEAD")
	api.HandleFunc("/" + pathMatch, api.etagMiddleware(api.serveOriginals())).
		Methods("GET", "HEAD")
	api.HandleFunc("/" + pathMatch, api.handleCreates()).Methods("POST")
	api.HandleFunc("/" + pathMatch, api.handleDeletes()).Methods("DELETE")
}

func (api *Api) etagMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ifNoneMatch := r.Header.Get("If-None-Match")
		if ifNoneMatch != "" && api.Etags.Contains(ifNoneMatch) {
			respondWithStatusCode(w, http.StatusNotModified)
			return
		}
		h(w, r)
	}
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
			et := etag.Generate(buf, true)
			api.Etags.Add(et)
			if r.Header.Get("If-None-Match") == et {
				respondWithStatusCode(w, http.StatusNotModified)
				return
			}
			respondWithImage(w, imager.GetImageType(buf), buf, et)
		})
	}
}

func (api *Api) serveThumbs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := metrics.GetOrRegisterTimer("api.thumbs.latency", nil)
		t.Time(func() {
			vars := mux.Vars(r)
			if _, ok := vars["height"]; !ok {
				vars["height"] = vars["width"]
			}
			resizeTier := fmt.Sprintf("%sx%s/%s/%s",
				vars["width"],
				vars["height"],
				vars["resizeOp"],
				vars["options"])
			path := vars["path"]
			thumbPath := resizeTier + "/" + path
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
				thumbBuf, err = imager.Resize(srcBuf, options)
				if err != nil {
					respondWithErr(w, http.StatusInternalServerError)
					return
				}
				go api.Thumbnails.Put(thumbPath, thumbBuf)
			}
			et := etag.Generate(thumbBuf, true)
			api.Etags.Add(et)
			if r.Header.Get("If-None-Match") == et {
				respondWithStatusCode(w, http.StatusNotModified)
				return
			}
			respondWithImage(w, imager.GetImageType(thumbBuf), thumbBuf, et)
		})
	}
}

func (api *Api) handleCreates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			reader   io.Reader
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
		buf, err := ioutil.ReadAll(io.LimitReader(reader, uploadSizeLimit))
		if len(buf) == 0 || err != nil {
			respondWithErr(w, http.StatusBadRequest)
			return
		}
		if len(buf) == uploadSizeLimit {
			respondWithErr(w, http.StatusRequestEntityTooLarge)
			return
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

func parseParams(vars map[string]string) (imager.Options, error) {
	width, err := strconv.Atoi(vars["width"])
	if err != nil {
		return imager.Options{}, err
	}
	height, err := strconv.Atoi(vars["height"])
	if err != nil {
		return imager.Options{}, err
	}
	resizeOp, ok := imager.ResizeOp[vars["resizeOp"]]
	if !ok {
		return imager.Options{}, errors.New("invalid resizeOp")
	}
	options := imager.Options{
		Width:    width,
		Height:   height,
		ResizeOp: resizeOp,
	}
	switch resizeOp {
	case imager.CROP:
		gravity, ok := imager.Gravity[vars["options"]]
		if !ok {
			return imager.Options{}, errors.New("invalid gravity")
		}
		options.Gravity = gravity
	case imager.FIT:
		extend := vars["options"]
		if utf8.RuneCountInString(extend) == 6 { // hex rgb
			rgb, err := decodeHexRGB(extend)
			if err != nil {
				return imager.Options{}, err
			}
			options.ExtendBackground = rgb
		}
	}

	return options, nil
}

func decodeHexRGB(hexRGB string) ([]float64, error) {
	runes := []rune(hexRGB)
	var (
		rgb []float64
		buf []byte
		err error
	)
	for i := 0; i < 3; i++ {
		buf, err = hex.DecodeString(string(runes[i*2 : i*2+2]))
		if err != nil {
			return nil, errors.New("invalid color")
		}
		rgb = append(rgb, float64(buf[0]))
	}
	return rgb, nil
}

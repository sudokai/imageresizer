package api

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/kxlt/imageresizer/collections"
	"github.com/kxlt/imageresizer/config"
	"github.com/kxlt/imageresizer/etag"
	"github.com/kxlt/imageresizer/imager"
	"github.com/kxlt/imageresizer/store"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func setupApi() *Api {
	api := &Api{
		Originals:  &store.TwoTier{Store: store.NewFileStore("../testdata")},
		Thumbnails: &store.NoopCache{},
		Router:     mux.NewRouter().StrictSlash(true),
		Etags:      collections.NewSyncStrSet(),
		Tiers:      collections.NewSyncStrSet(),
	}
	api.routes()
	return api
}

func TestServeOriginals(t *testing.T) {
	api := setupApi()
	config.C.EtagCacheEnable = false
	buf, err := ioutil.ReadFile("../testdata/samuel-clara-69657-unsplash.jpg")
	if err != nil {
		t.Fatal(err)
	}
	err = api.Originals.Put("samuel-clara-69657-unsplash.jpg", buf)
	if err != nil {
		t.Fatal(err)
	}

	// test 200 OK
	req, err := http.NewRequest("GET", "/samuel-clara-69657-unsplash.jpg", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %d", rr.Code)
	}
	if bytes.Compare(rr.Body.Bytes(), buf) != 0 {
		t.Fatalf("Response body doesn't match with image")
	}
	if rr.Header().Get("ETag") != etag.Generate(buf, true) {
		t.Fatalf("ETag doesn't match")
	}
	if rr.Header().Get("Content-Type") != mimeTypes[imager.JPEG] {
		t.Fatalf("Content-Type doesn't match: %s", rr.Header().Get("Content-Type"))
	}
	if contentLength, _ := strconv.Atoi(rr.Header().Get("Content-Length")); contentLength != len(buf) {
		t.Fatalf("Content-Length doesn't match")
	}
	if api.Etags.Contains(etag.Generate(buf, true)) {
		t.Fatalf("ETag shouldn't be cached")
	}

	// test 404 Not Found
	req, err = http.NewRequest("GET", "/does-not-exist", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	api.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Wrong status code: %d", rr.Code)
	}

	// test ETag cache
	config.C.EtagCacheEnable = true
	req, err = http.NewRequest("GET", "/samuel-clara-69657-unsplash.jpg", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	api.ServeHTTP(rr, req)
	if !api.Etags.Contains(etag.Generate(buf, true)) {
		t.Fatalf("Generated ETag wasn't cached")
	}
}

func TestServeThumbs(t *testing.T) {
	api := setupApi()
	config.C.CacheThumbEnable = false
	config.C.EtagCacheEnable = false

	buf, err := ioutil.ReadFile("../testdata/300x300/crop/s/natasha-kasim-708827-unsplash.jpg")
	if err != nil {
		t.Fatal(err)
	}

	// test 200 Ok
	req, err := http.NewRequest("GET", "/300x300/crop/s/natasha-kasim-708827-unsplash.jpg", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	api.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %d", rr.Code)
	}
	if bytes.Compare(rr.Body.Bytes(), buf) != 0 {
		t.Fatalf("Response body doesn't match with image")
	}
	if rr.Header().Get("ETag") != etag.Generate(buf, true) {
		t.Fatalf("ETag doesn't match")
	}
	if rr.Header().Get("Content-Type") != mimeTypes[imager.JPEG] {
		t.Fatalf("Content-Type doesn't match: %s", rr.Header().Get("Content-Type"))
	}
	if contentLength, _ := strconv.Atoi(rr.Header().Get("Content-Length")); contentLength != len(buf) {
		t.Fatalf("Content-Length doesn't match")
	}
	if api.Etags.Contains(etag.Generate(buf, true)) {
		t.Fatalf("ETag shouldn't be cached")
	}

	// test 404 Not Found
	req, err = http.NewRequest("GET", "/does-not-exist", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	api.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("Wrong status code: %d", rr.Code)
	}

	// test ETag cache
	config.C.EtagCacheEnable = true
	req, err = http.NewRequest("GET", "/300x300/crop/s/natasha-kasim-708827-unsplash.jpg", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	api.ServeHTTP(rr, req)
	if !api.Etags.Contains(etag.Generate(buf, true)) {
		t.Fatalf("Generated ETag wasn't cached")
	}
}

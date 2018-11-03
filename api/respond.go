package api

import (
	"encoding/json"
	"github.com/kxlt/imageresizer/imager"
	"net/http"
	"strconv"
)

type ImageResponse struct {
	format imager.ImageType
	buf    []byte
	etag   string
}

var mimeTypes = map[imager.ImageType]string{
	imager.JPEG: "image/jpeg",
	imager.PNG:  "image/png",
}

func respondWithImage(w http.ResponseWriter, imgResponse *ImageResponse) {
	w.Header().Set("Content-Type", mimeTypes[imgResponse.format])
	w.Header().Set("Content-Length", strconv.Itoa(len(imgResponse.buf)))
	w.Header().Set("ETag", imgResponse.etag)
	w.WriteHeader(http.StatusOK)
	w.Write(imgResponse.buf)
}

func respondWithErr(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": http.StatusText(statusCode),
	})
}

func respondWithStatusCode(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

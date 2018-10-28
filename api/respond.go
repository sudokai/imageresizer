package api

import (
	"encoding/json"
	"github.com/kxlt/imageresizer/imager"
	"net/http"
	"strconv"
)

var mimeTypes = map[imager.ImageType]string{
	imager.JPEG: "image/jpeg",
	imager.PNG:  "image/png",
}

func respondWithImage(w http.ResponseWriter, format imager.ImageType, data []byte, etag string) {
	w.Header().Set("Content-Type", mimeTypes[format])
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
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

package api

import (
	"encoding/json"
	"github.com/kailt/imageresizer/imagine"
	"net/http"
	"strconv"
)

var mimeTypes = map[imagine.ImageType]string{
	imagine.JPEG: "image/jpeg",
	imagine.PNG:  "image/png",
}

func respondWithImage(w http.ResponseWriter, format imagine.ImageType, data []byte) {
	w.Header().Set("Content-Type", mimeTypes[format])
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
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
package config

import (
	"github.com/spf13/viper"
	"log"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Config struct {
	ServerAddr string

	LocalPrefix string

	S3Enable bool
	S3Region string
	S3Bucket string
	S3Prefix string

	CacheOrigEnable      bool
	CacheOrigPath        string
	CacheOrigMaxSize     int64
	CacheOrigShards      int
	CacheThumbEnable     bool
	CacheThumbPath       string
	CacheThumbMaxSize    int64
	CacheThumbShards     int
	CacheLoaderFiles     int
	CacheLoaderSleep     int
	CacheLoaderThreshold int

	UploadMaxSize int64

	EtagCacheEnable  bool
	EtagCacheMaxSize int
}

var C Config

func init() {
	viper.SetEnvPrefix("IMAGERESIZER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("server.addr", ":8080")
	viper.SetDefault("local.prefix", "./images/originals")
	viper.SetDefault("s3.enable", false)
	viper.SetDefault("s3.prefix", "")
	viper.SetDefault("cache.orig.enable", true)
	viper.SetDefault("cache.orig.path", "./images/cache")
	viper.SetDefault("cache.orig.maxsize", "1G")
	viper.SetDefault("cache.orig.shards", 256)
	viper.SetDefault("cache.thumb.enable", true)
	viper.SetDefault("cache.thumb.path", "./images/thumbnails")
	viper.SetDefault("cache.thumb.maxsize", "1G")
	viper.SetDefault("cache.thumb.shards", 256)
	viper.SetDefault("cache.loader.files", 100)
	viper.SetDefault("cache.loader.sleep", 50)
	viper.SetDefault("cache.loader.threshold", 200)
	viper.SetDefault("upload.maxsize", "50M")
	viper.SetDefault("etag.cache.enable", true)
	viper.SetDefault("etag.cache.maxsize", 50000)
}

func RefreshConfig() {
	C.ServerAddr = viper.GetString("server.addr")
	C.LocalPrefix = viper.GetString("local.prefix")
	C.S3Enable = viper.GetBool("s3.enable")
	C.S3Region = viper.GetString("s3.region")
	C.S3Bucket = viper.GetString("s3.bucket")
	C.S3Prefix = viper.GetString("s3.prefix")
	C.CacheOrigEnable = viper.GetBool("cache.orig.enable")
	C.CacheOrigPath = viper.GetString("cache.orig.path")
	C.CacheOrigMaxSize = parseSize(viper.GetString("cache.orig.maxsize"))
	C.CacheOrigShards = viper.GetInt("cache.orig.shards")
	if C.CacheOrigShards < 1 {
		log.Fatalln("Minimum 1 shard required")
	}
	C.CacheThumbEnable = viper.GetBool("cache.thumb.enable")
	C.CacheThumbPath = viper.GetString("cache.thumb.path")
	C.CacheThumbMaxSize = parseSize(viper.GetString("cache.thumb.maxsize"))
	C.CacheThumbShards = viper.GetInt("cache.thumb.shards")
	if C.CacheThumbShards < 1 {
		log.Fatalln("Minimum 1 shard required")
	}
	C.CacheLoaderFiles = viper.GetInt("cache.loader.files")
	C.CacheLoaderSleep = viper.GetInt("cache.loader.sleep")
	C.CacheLoaderThreshold = viper.GetInt("cache.loader.threshold")
	C.UploadMaxSize = parseSize(viper.GetString("upload.maxsize"))
	C.EtagCacheEnable = viper.GetBool("etag.cache.enable")
	C.EtagCacheMaxSize = viper.GetInt("etag.cache.maxsize")
}

func parseSize(sizeStr string) int64 {
	runes := []rune(sizeStr)
	length := utf8.RuneCountInString(sizeStr)
	unit := string(runes[length-1:])
	number, err := strconv.Atoi(string(runes[:length-1]))
	if err != nil {
		log.Fatalln("Could not parse config size")
	}
	var factor int
	switch unit {
	case "k":
		fallthrough
	case "K":
		factor = 1024
	case "m":
		fallthrough
	case "M":
		factor = 1024 * 1024
	case "g":
		fallthrough
	case "G":
		factor = 1024 * 1024 * 1024
	case "t":
		fallthrough
	case "T":
		factor = 1024 * 1024 * 1024 * 1024
	default:
		factor = 1
	}
	return int64(number * factor)
}

package etag

import (
	"crypto/sha1"
	"fmt"
)

func getHash(buf []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(buf))
}

// Generate an Etag for given sring. Allows specifying whether to generate weak
// Etag or not as second parameter
func Generate(buf []byte, weak bool) string {
	tag := fmt.Sprintf("\"%d-%s\"", len(buf), getHash(buf))
	if weak {
		tag = "W/" + tag
	}

	return tag
}

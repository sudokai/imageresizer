# imageresizer

***Warning**: at an early stage of development, NOT suitable for production use.*

Prototype of a image resizing server written in Go. 

I decided to write this software after dealing with image thumbnailing at
scale (15 million images, multiple thumbnail sizes, 500k new images
per month) at work.

`imageresizer` attempts to address the image thumbnailing needs of a company 
such as the one I work for:

- High volume of basic resize operations. Fast resizing is a must.
- Caching of thumbnails and originals, to avoid hitting S3 continuously ($$$).
  Caches should use LRU or LFU with a maximum on-disk size.
- Ease of deployment and maintenance. No supervisors or http servers.
- New deployments should incur near-zero downtime.


## Getting started

Building and running the code:

Install `libvips`.
Then: 
```bash
make
./imageresizer
```

URL format:

```
/{width:[0-9]+}x{height:[0-9]+}/{gravity}/{path}
```

At the moment, only two `gravity` settings are supported:
- `sm`: smart
- `ce`: center

## Features

- Fast resizes using libvips through a cgo bridge (JPEG and PNG)
- Local caching of originals and thumbnails with approximate LRU eviction based on file atimes.
- Smart cropping.
- Image uploads and deletions.
- S3 storage support.
- Graceful zero-downtime upgrades/restarts.
- 304 Not Modified responses.

## Roadmap

In order of priority:

- More resize operations. Eg: fit without cropping.
- Older libvips (<8.7) compatibility.
- WEBP and GIF support?
- Security controls for uploads and deletions?
- Secure links?
- Cache sharding.
- LFU instead of LRU.

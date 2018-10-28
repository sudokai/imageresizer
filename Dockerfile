FROM golang:1.11-alpine3.8 as builder

RUN apk add --no-cache gcc musl-dev make git
RUN apk add --no-cache \
            --repository https://dl-3.alpinelinux.org/alpine/edge/testing/ \
            --repository https://dl-3.alpinelinux.org/alpine/edge/main \
            vips-dev

ADD . /src
WORKDIR /src
RUN make

#################
FROM alpine:3.8
ARG SERVER_ADDR=:8080
ARG LOCAL_PREFIX=/images/originals
ARG CACHE_ORIG_PATH=/images/cache
ARG CACHE_THUMB_PATH=/images/thumbnails

LABEL default.server.addr=${SERVER_ADDR} \
      default.local.prefix=${LOCAL_PREFIX} \
      default.cache.orig.path=${CACHE_ORIG_PATH} \
      default.cache.thumb.path=${CACHE_THUMB_PATH}

ENV IMAGERESIZER_SERVER_ADDR=${SERVER_ADDR}
ENV IMAGERESIZER_LOCAL_PREFIX=${LOCAL_PREFIX}
ENV IMAGERESIZER_CACHE_ORIG_PATH=${CACHE_ORIG_PATH}
ENV IMAGERESIZER_CACHE_THUMB_PATH=${CACHE_THUMB_PATH}

WORKDIR /app

# Ensure expat version is not coming from edge
RUN apk add --no-cache expat ca-certificates
RUN apk add --no-cache \
            --repository https://dl-3.alpinelinux.org/alpine/edge/testing/ \
            --repository https://dl-3.alpinelinux.org/alpine/edge/main \
            vips
COPY --from=builder /src/imageresizer /app/imageresizer

VOLUME ["${IMAGERESIZER_CACHE_ORIG_PATH}", "${IMAGERESIZER_LOCAL_PREFIX}", "${IMAGERESIZER_CACHE_THUMB_PATH}"]
EXPOSE "${IMAGERESIZER_SERVER_ADDR}"

CMD ["/app/imageresizer"]


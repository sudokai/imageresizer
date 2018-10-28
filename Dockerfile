FROM alpine:3.7

RUN apk add --no-cache --virtual .build-deps gcc musl-dev
RUN apk add vips-dev --verbose --update-cache --repository https://dl-3.alpinelinux.org/alpine/edge/testing/ --repository https://dl-3.alpinelinux.org/alpine/edge/main
RUN make dep
RUN make
RUN ./imageresizer
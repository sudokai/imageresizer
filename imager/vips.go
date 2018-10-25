package imager

/*
#cgo pkg-config: vips
#include "vips.h"
*/
import "C"
import (
	"errors"
	"log"
	"runtime"
	"unsafe"
)

type ImageType int

const (
	UNKNOWN ImageType = iota
	JPEG
	PNG
)

type GravityType int

const (
	C GravityType = iota + 1
	S
)

var Gravity = map[string]GravityType{
	"c": C,
	"s": S,
}

type ResizeOpType int

const (
	CROP ResizeOpType = iota
	FIT
)

var ResizeOp = map[string]ResizeOpType{
	"crop": CROP,
	"fit":  FIT,
}

type ExtendType int

const (
	NOOP ExtendType = iota
	NEAREST
	BACKGROUND
)

type Options struct {
	Width            int
	Height           int
	ResizeOp         ResizeOpType
	Gravity          GravityType
	Quality          int
	Extend           ExtendType
	ExtendBackground []float64
}

func init() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	err := C.vips_init(C.CString("imagine"))
	if err != 0 {
		C.vips_shutdown()
		log.Fatalf("vips_init failed\n")
	}
}

func ShutdownVIPS() {
	C.vips_shutdown()
}

func GetImageType(buf []byte) ImageType {
	if len(buf) < 12 {
		return UNKNOWN
	}
	if buf[0] == 0xFF && buf[1] == 0xD8 && buf[2] == 0xFF {
		return JPEG
	}
	if buf[0] == 0x89 && buf[1] == 0x50 && buf[2] == 0x4E && buf[3] == 0x47 {
		return PNG
	}
	return UNKNOWN
}

func Resize(buf []byte, options Options) ([]byte, error) {
	defer C.vips_thread_shutdown()

	var iWidth, iHeight, origOWidth, origOHeight int
	if options.ResizeOp == FIT {
		image, err := vipsImageNew(buf) // this is efficient because vips only reads bytes as needed
		if err != nil {
			return nil, err
		}
		iWidth = int(C.vips_image_get_width(image))
		iHeight = int(C.vips_image_get_height(image))
		origOWidth = options.Width
		origOHeight = options.Height
		if iWidth*options.Height > options.Width*iHeight {
			// aspect ratio of original image is bigger than target aspect ratio
			// shrink height
			options.Height = options.Width * iHeight / iWidth
		} else {
			options.Width = iWidth * options.Height / iHeight
		}
		C.g_object_unref(C.gpointer(image))
	}

	image, err := vipsThumbnail(buf, options.Width, options.Height, options.Gravity)
	if err != nil {
		return nil, err
	}
	defer C.g_object_unref(C.gpointer(image))

	if options.Extend != NOOP {
		x := (origOWidth - options.Width) / 2
		y := (origOHeight - options.Height) / 2
		image, err = vipsEmbed(image, x, y, origOWidth, origOHeight, options.Extend, options.ExtendBackground)
		if err != nil {
			return nil, err
		}
		defer C.g_object_unref(C.gpointer(image))
	}

	thumbBuf, err := vipsSave(GetImageType(buf), image)
	if err != nil {
		return nil, err
	}

	return thumbBuf, nil
}

func vipsEmbed(
	in *C.VipsImage,
	x int,
	y int,
	width int,
	height int,
	extend ExtendType,
	bg []float64) (*C.VipsImage, error) {

	var image *C.VipsImage
	var err C.int = 1
	if extend == NEAREST {
		err = C.vips_embed_copy_cgo(
			in,
			&image,
			C.int(x),
			C.int(y),
			C.int(width),
			C.int(height))
	} else if extend == BACKGROUND {
		err = C.vips_embed_background_cgo(
			in,
			&image,
			C.int(x),
			C.int(y),
			C.int(width),
			C.int(height),
			(*C.double)(&bg[0]))
	}

	if err != 0 {
		return nil, vipsError()
	}
	return image, nil
}

func vipsImageNew(buf []byte) (*C.VipsImage, error) {
	var image *C.VipsImage
	err := C.vips_image_new_cgo(
		C.int(GetImageType(buf)),
		unsafe.Pointer(&buf[0]),
		C.size_t(len(buf)),
		&image)
	if err != 0 {
		return nil, vipsError()
	}
	return image, nil
}

func vipsThumbnail(buf []byte, width int, height int, gravity GravityType) (*C.VipsImage, error) {
	smart := gravity == S
	cSmart := C.int(0)
	if smart {
		cSmart = C.int(1)
	}

	var image *C.VipsImage
	// cgo doesn't allow calling functions with variadic arguments directly
	err := C.vips_thumbnail_cgo(
		unsafe.Pointer(&buf[0]),
		C.size_t(len(buf)),
		&image,
		C.int(width),
		C.int(height),
		cSmart)
	if err != 0 {
		return nil, vipsError()
	}
	return image, nil
}

func vipsSave(imageType ImageType, image *C.VipsImage) ([]byte, error) {
	var ptr unsafe.Pointer
	length := C.size_t(0)
	err := C.vips_save_buffer_cgo(C.int(imageType), image, &ptr, &length)
	if err != 0 {
		return nil, vipsError()
	}
	buf := C.GoBytes(ptr, C.int(length))
	C.g_free(C.gpointer(ptr))
	return buf, nil
}

func vipsError() error {
	s := C.GoString(C.vips_error_buffer())
	C.vips_error_clear()
	return errors.New(s)
}

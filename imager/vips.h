#include "vips/vips.h"

enum imageTypes {
    UNKNOWN = 0,
    JPEG,
    PNG
};

int vips_save_buffer_cgo(int imageType, VipsImage *in, void **buf, size_t *len) {
    int err = 1;
    switch (imageType) {
    case JPEG:
        err = vips_jpegsave_buffer(in, buf, len,
            "optimize_coding", TRUE,
            "strip", TRUE,
            NULL);
        break;
    case PNG:
         err = vips_pngsave_buffer(in, buf, len, NULL);
         break;
    }
    return err;
}

int vips_thumbnail_cgo(void *buf, size_t len, VipsImage **out, int width, int height, int smart) {
    VipsInteresting crop = VIPS_INTERESTING_CENTRE;
    if (smart > 0) {
        crop = VIPS_INTERESTING_ENTROPY;
    }
    return vips_thumbnail_buffer(
        buf,
        len,
        out,
        width,
        "height", height,
        "crop", crop,
        "intent", VIPS_INTENT_PERCEPTUAL,
        NULL);
}

int vips_image_new_cgo(int imageType, void *buf, size_t len, VipsImage **out) {
    int err = 1;
    switch (imageType) {
    case JPEG:
        err = vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
        break;
    case PNG:
         err = vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
         break;
    }
    return err;
}

int vips_embed_background_cgo(VipsImage *in, VipsImage **out, int x, int y, int width, int height, double *bg) {
    int err = 1;
    VipsArrayDouble *background = vips_array_double_new(bg, 3);
    err = vips_embed(in, out, x, y, width, height, "extend", VIPS_EXTEND_BACKGROUND, "background", background, NULL);
    vips_area_unref(VIPS_AREA(background));
    return err;
}
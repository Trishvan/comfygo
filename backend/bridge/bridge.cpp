#include "bridge.h"
#include <string>
#include <unordered_map>
#include <cstring>

// stable-diffusion.cpp headers — adjust include path as needed
// #include "sd.h"

static std::unordered_map<int, void*> g_models;
static int g_next_handle = 1;
static std::string g_last_error;

static void set_error(const std::string& err) {
    g_last_error = err;
}

int load_model(const char* model_path) {
    (void)model_path;
    // TODO: call sd_model_new(model_path, ...)
    // For now: return a mock handle
    int handle = g_next_handle++;
    g_models[handle] = (void*)(intptr_t)handle;
    return handle;
}

int load_vae(int model_handle, const char* vae_path) {
    (void)model_handle;
    (void)vae_path;
    // TODO: sd_model_set_vae(model, vae_path)
    return 0;
}

image_result_t txt2img_c(int model_handle, sd_config_t config, progress_callback_t progress_cb, void* user_data) {
    image_result_t result = {0, 0, 0, nullptr};
    (void)model_handle;
    (void)config;
    (void)progress_cb;
    (void)user_data;

    // TODO: call sd_image_generate() with config params
    // Invoke progress_cb(step, total, user_data) during generation

    // Mock: return a small checkerboard
    int w = config.width > 0 ? config.width : 512;
    int h = config.height > 0 ? config.height : 512;
    int ch = 3;
    size_t sz = (size_t)w * h * ch;
    uint8_t* buf = (uint8_t*)malloc(sz);
    if (!buf) return result;

    for (int y = 0; y < h; y++) {
        for (int x = 0; x < w; x++) {
            int idx = (y * w + x) * ch;
            uint8_t val = ((x / 32) + (y / 32)) % 2 ? 0x80 : 0x40;
            buf[idx + 0] = val;
            buf[idx + 1] = val;
            buf[idx + 2] = val;
        }
    }
    result.width = w;
    result.height = h;
    result.channels = ch;
    result.data = buf;
    return result;
}

void free_image(image_result_t* img) {
    if (img && img->data) {
        free(img->data);
        img->data = nullptr;
        img->width = 0;
        img->height = 0;
        img->channels = 0;
    }
}

void free_model_c(int model_handle) {
    auto it = g_models.find(model_handle);
    if (it != g_models.end()) {
        // TODO: sd_model_free(it->second)
        g_models.erase(it);
    }
}

const char* get_last_error(void) {
    return g_last_error.c_str();
}

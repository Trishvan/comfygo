#ifndef BRIDGE_H
#define BRIDGE_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    int width;
    int height;
    int channels;
    uint8_t* data;
} image_result_t;

typedef struct {
    const char* prompt;
    const char* negative_prompt;
    const char* model_path;
    const char* vae_path;
    int steps;
    float cfg_scale;
    int seed;
    int width;
    int height;
    const char* sampler_name;
} sd_config_t;

typedef void (*progress_callback_t)(int step, int total, void* user_data);

int load_model(const char* model_path);
int load_vae(int model_handle, const char* vae_path);
image_result_t txt2img_c(int model_handle, sd_config_t config, progress_callback_t progress_cb, void* user_data);
void free_image(image_result_t* img);
void free_model_c(int model_handle);
const char* get_last_error(void);

#ifdef __cplusplus
}
#endif

#endif

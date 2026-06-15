#ifndef BRIDGE_H
#define BRIDGE_H

#include "sd.h"

#ifdef __cplusplus
extern "C" {
#endif

// Simplified config for the bridge layer
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
    const char** lora_paths;
    float* lora_scales;
    int lora_count;
} sd_config_t;

// Model management
int load_model(const char* model_path, const char* vae_path);
void free_model_c(int model_handle);

// Image generation    
sd_image_t txt2img_c(int model_handle, sd_config_t config);

// The Go-exported progress callback, callable from C
extern void goProgressCb(int step, int steps, float time, void* data);

#ifdef __cplusplus
}
#endif

#endif

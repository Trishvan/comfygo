#include "bridge.h"
#include <cstdlib>
#include <cstring>
#include <string>
#include <unordered_map>
#include <vector>

static std::unordered_map<int, sd_ctx_t*> g_models;
static int g_next_handle = 1;
static std::string g_last_error;

static void set_error(const std::string& err) {
    g_last_error = err;
}

int load_model(const char* model_path, const char* vae_path) {
    sd_ctx_params_t ctx_params;
    sd_ctx_params_init(&ctx_params);

    ctx_params.model_path = model_path;
    ctx_params.vae_path = vae_path;
    ctx_params.n_threads = sd_get_num_physical_cores();
    ctx_params.wtype = SD_TYPE_F32;
    ctx_params.rng_type = CUDA_RNG;
    ctx_params.enable_mmap = true;

    sd_ctx_t* ctx = new_sd_ctx(&ctx_params);
    if (ctx == nullptr) {
        set_error("new_sd_ctx failed");
        return 0;
    }

    int handle = g_next_handle++;
    g_models[handle] = ctx;
    return handle;
}

void free_model_c(int model_handle) {
    auto it = g_models.find(model_handle);
    if (it != g_models.end()) {
        free_sd_ctx(it->second);
        g_models.erase(it);
    }
}

static int sample_method_from_name(const char* name) {
    if (name == nullptr || name[0] == '\0') return EULER_A_SAMPLE_METHOD;
    enum sample_method_t m = str_to_sample_method(name);
    return m < SAMPLE_METHOD_COUNT ? m : EULER_A_SAMPLE_METHOD;
}

sd_image_t txt2img_c(int model_handle, sd_config_t config) {
    sd_image_t result = {0, 0, 0, nullptr};

    auto it = g_models.find(model_handle);
    if (it == g_models.end()) {
        set_error("invalid model handle");
        return result;
    }
    sd_ctx_t* ctx = it->second;

    sd_sample_params_t sample_params;
    sd_sample_params_init(&sample_params);
    sample_params.sample_method = (enum sample_method_t)sample_method_from_name(config.sampler_name);
    sample_params.sample_steps = config.steps;
    sample_params.guidance.txt_cfg = config.cfg_scale;

    sd_img_gen_params_t gen_params;
    sd_img_gen_params_init(&gen_params);

    gen_params.prompt = config.prompt ? config.prompt : "";
    gen_params.negative_prompt = config.negative_prompt ? config.negative_prompt : "";
    gen_params.width = config.width > 0 ? config.width : 512;
    gen_params.height = config.height > 0 ? config.height : 512;
    gen_params.seed = config.seed >= 0 ? config.seed : -1;
    gen_params.batch_count = 1;
    gen_params.sample_params = sample_params;
    gen_params.strength = 1.0f;

    // Enable VAE tiling for resolutions above 512 to avoid memory/artifacts
    if (config.width > 512 || config.height > 512) {
        gen_params.vae_tiling_params.enabled = true;
    }

    // Apply LoRAs
    std::vector<sd_lora_t> loras;
    if (config.lora_count > 0 && config.lora_paths != nullptr) {
        loras.reserve(config.lora_count);
        for (int i = 0; i < config.lora_count; i++) {
            sd_lora_t lora{};
            lora.path = config.lora_paths[i];
            lora.multiplier = config.lora_scales ? config.lora_scales[i] : 1.0f;
            lora.is_high_noise = false;
            loras.push_back(lora);
        }
        gen_params.loras = loras.data();
        gen_params.lora_count = static_cast<uint32_t>(config.lora_count);
    }

    // Register the Go-exported progress callback
    sd_set_progress_callback(goProgressCb, nullptr);

    int image_count = 0;
    sd_image_t* images = generate_image(ctx, &gen_params);
    if (images == nullptr) {
        set_error("generate_image returned null");
        return result;
    }

    result = images[0];
    // Free the image array but keep the first image — caller must free it later
    // Actually, we can't free images[0] and keep it at the same time.
    // We need to copy the data or rely on the caller to free.
    // For zero-copy, we just pass the pointer ownership to the caller.
    // The caller will call free_sd_images. But then we need the original pointer.
    // Approach: allocate a single image, copy data, free the array.

    sd_image_t copy;
    copy.width = result.width;
    copy.height = result.height;
    copy.channel = result.channel;
    size_t sz = (size_t)result.width * result.height * result.channel;
    copy.data = (uint8_t*)malloc(sz);
    if (copy.data && result.data) {
        memcpy(copy.data, result.data, sz);
    }
    free_sd_images(images, image_count);

    return copy;
}

const char* get_last_error(void) {
    return g_last_error.c_str();
}

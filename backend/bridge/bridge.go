package bridge

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -L${SRCDIR}/../../Sdcpp/sd-master-c2df4e1-bin-Linux-Ubuntu-24.04-x86_64-vulkan -Wl,-rpath,${SRCDIR}/../../Sdcpp/sd-master-c2df4e1-bin-Linux-Ubuntu-24.04-x86_64-vulkan -lstable-diffusion -lstdc++ -lm -ldl

#include <stdlib.h>
#include "bridge.h"

extern void goProgressCb(int step, int steps, float time, void* data);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type ImageResult struct {
	Width    int
	Height   int
	Channels int
	Data     []byte
	cData    *C.uint8_t
}

type GenerationConfig struct {
	Prompt         string
	NegativePrompt string
	ModelPath      string
	VaePath        string
	Steps          int
	CfgScale       float64
	Seed           int
	Width          int
	Height         int
	SamplerName    string
}

type ProgressCallback func(step, total int)

var currentProgressCb ProgressCallback

//export goProgressCb
func goProgressCb(step C.int, steps C.int, time C.float, data unsafe.Pointer) {
	if currentProgressCb != nil {
		currentProgressCb(int(step), int(steps))
	}
}

func LoadModel(modelPath, vaePath string) (int, error) {
	cModelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cModelPath))

	var cVaePath *C.char
	if vaePath != "" {
		cVaePath = C.CString(vaePath)
		defer C.free(unsafe.Pointer(cVaePath))
	}

	handle := int(C.load_model(cModelPath, cVaePath))
	if handle == 0 {
		return 0, fmt.Errorf("load_model failed")
	}
	return handle, nil
}

func FreeModel(handle int) {
	C.free_model_c(C.int(handle))
}

func Txt2Img(handle int, cfg GenerationConfig, cb ProgressCallback) (ImageResult, error) {
	cCfg := C.sd_config_t{
		prompt:          C.CString(cfg.Prompt),
		negative_prompt: C.CString(cfg.NegativePrompt),
		model_path:      C.CString(cfg.ModelPath),
		vae_path:        C.CString(cfg.VaePath),
		steps:           C.int(cfg.Steps),
		cfg_scale:       C.float(cfg.CfgScale),
		seed:            C.int(cfg.Seed),
		width:           C.int(cfg.Width),
		height:          C.int(cfg.Height),
		sampler_name:    C.CString(cfg.SamplerName),
	}
	defer C.free(unsafe.Pointer(cCfg.prompt))
	defer C.free(unsafe.Pointer(cCfg.negative_prompt))
	defer C.free(unsafe.Pointer(cCfg.model_path))
	defer C.free(unsafe.Pointer(cCfg.vae_path))
	defer C.free(unsafe.Pointer(cCfg.sampler_name))

	currentProgressCb = cb
	defer func() { currentProgressCb = nil }()

	cResult := C.txt2img_c(C.int(handle), cCfg)
	if cResult.data == nil {
		return ImageResult{}, fmt.Errorf("txt2img_c failed")
	}

	totalBytes := int(cResult.width) * int(cResult.height) * int(cResult.channel)
	goSlice := unsafe.Slice((*byte)(unsafe.Pointer(cResult.data)), totalBytes)

	return ImageResult{
		Width:    int(cResult.width),
		Height:   int(cResult.height),
		Channels: int(cResult.channel),
		Data:     goSlice,
		cData:    cResult.data,
	}, nil
}

func FreeImage(img *ImageResult) {
	if img == nil || img.cData == nil {
		return
	}
	C.free(unsafe.Pointer(img.cData))
	img.Data = nil
	img.cData = nil
}

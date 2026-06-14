package bridge

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -lstdc++ -lm

#include <stdlib.h>
#include "bridge.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type ImageResult struct {
	Width  int
	Height int
	Channels int
	Data   []byte
	cPtr   *C.uchar
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

func LoadModel(modelPath string) (int, error) {
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	handle := int(C.load_model(cPath))
	if handle == 0 {
		return 0, fmt.Errorf("load_model failed: %s", lastError())
	}
	return handle, nil
}

func LoadVae(handle int, vaePath string) error {
	cPath := C.CString(vaePath)
	defer C.free(unsafe.Pointer(cPath))

	ret := int(C.load_vae(C.int(handle), cPath))
	if ret != 0 {
		return fmt.Errorf("load_vae failed: %s", lastError())
	}
	return nil
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

	var cbData progressCallbackData
	cbData.fn = cb

	cResult := C.txt2img_c(C.int(handle), cCfg, nil, nil)
	if cResult.data == nil {
		return ImageResult{}, fmt.Errorf("txt2img_c failed: %s", lastError())
	}

	totalBytes := int(cResult.width) * int(cResult.height) * int(cResult.channels)
	goSlice := unsafe.Slice((*byte)(unsafe.Pointer(cResult.data)), totalBytes)

	return ImageResult{
		Width:    int(cResult.width),
		Height:   int(cResult.height),
		Channels: int(cResult.channels),
		Data:     goSlice,
		cPtr:     cResult.data,
	}, nil
}

func FreeImage(img *ImageResult) {
	if img == nil || img.cPtr == nil {
		return
	}
	cResult := C.image_result_t{
		data:     img.cPtr,
		width:    C.int(img.Width),
		height:   C.int(img.Height),
		channels: C.int(img.Channels),
	}
	C.free_image(&cResult)
	img.Data = nil
	img.cPtr = nil
}

func FreeModel(handle int) {
	C.free_model_c(C.int(handle))
}

func lastError() string {
	cErr := C.get_last_error()
	if cErr == nil {
		return "unknown error"
	}
	return C.GoString(cErr)
}

type progressCallbackData struct {
	fn ProgressCallback
}

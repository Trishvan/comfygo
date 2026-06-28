.PHONY: setup dev build clean bindings lint package

setup:
	scripts/setup.sh

dev:
	wails dev -tags webkit2_41

build:
	wails build -tags webkit2_41 -o comfygo

package:
	scripts/package.sh

clean:
	rm -rf build/bin

bindings:
	wails generate module

lint:
	cd frontend && npx svelte-check

help:
	@echo "Targets:"
	@echo "  make setup    — run setup.sh (one-time install)"
	@echo "  make dev      — start Wails dev server"
	@echo "  make build    — production build"
	@echo "  make package  — create distributable packages (.deb/.rpm/.AppImage)"
	@echo "  make clean    — remove build artifacts"
	@echo "  make bindings — regenerate Wails bindings"
	@echo "  make lint     — run Svelte type-check"

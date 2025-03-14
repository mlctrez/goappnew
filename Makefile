
APP_NAME=goapp

VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "v0.0.0")
COMMIT=$(shell git rev-parse --short HEAD || echo "HEAD")
MODULE=$(shell grep ^module go.mod | awk '{print $$2;}')
LD_FLAGS="-w -X $(MODULE)/goapp.Version=$(VERSION) -X $(MODULE)/goapp.Commit=$(COMMIT)"
MAIN="goapp/service/main/main.go"

LD_FLAGS_TINY_GO="-X $(MODULE)/goapp.Version=$(VERSION) -X $(MODULE)/goapp.Commit=$(COMMIT)"
# USE_TINY_GO=1

run: binary
	@DEV=1 PORT=8080 GOAPP_USE_COMPRESSION=1 ./temp/$(APP_NAME)

binary: wasm
	@mkdir -p temp
	@echo "ldflags=$(LD_FLAGS)"
	@CGO_ENABLED=0 go build -o temp/$(APP_NAME) -ldflags $(LD_FLAGS) $(MAIN)

wasm:
	@rm -f goapp/web/app.wasm
ifdef USE_TINY_GO
	@cp $(shell tinygo env TINYGOROOT)/targets/wasm_exec.js goapp/web/wasm_exec.js
	@GOARCH=wasm GOOS=js tinygo build -o goapp/web/app.wasm -ldflags $(LD_FLAGS_TINY_GO) \
		-panic=trap -no-debug -target=wasm $(MAIN)
else
	@rm -f goapp/web/wasm_exec.js
	@GOARCH=wasm GOOS=js go build -o goapp/web/app.wasm -ldflags $(LD_FLAGS) $(MAIN)
endif

clean:
	@rm -rf temp
	@rm -f goapp/web/app.wasm

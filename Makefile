BUILD_TAGS_LINUX   = production,webkit2_41
BUILD_TAGS_WINDOWS = production
BINARY_LINUX       = linkit-price-updater
BINARY_WINDOWS     = linkit-price-updater.exe

.PHONY: build linux windows run clean

build: linux

linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags "$(BUILD_TAGS_LINUX)" -o $(BINARY_LINUX) ./cmd/

windows:
	@if command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1; then \
		CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -tags "$(BUILD_TAGS_WINDOWS)" -ldflags "-H windowsgui" -o $(BINARY_WINDOWS) ./cmd/ ; \
	else \
		echo "Building Windows binary without CGO (x86_64-w64-mingw32-gcc not found)..." ; \
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags "$(BUILD_TAGS_WINDOWS)" -ldflags "-H windowsgui" -o $(BINARY_WINDOWS) ./cmd/ ; \
	fi

build-debug:
	CGO_ENABLED=1 go build -tags "dev,webkit2_41" -o $(BINARY_LINUX) ./cmd/

run: linux
	DISPLAY=:0 ./$(BINARY_LINUX)

clean:
	rm -f $(BINARY_LINUX) $(BINARY_WINDOWS)


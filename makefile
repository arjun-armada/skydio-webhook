# Check to see if we can use ash, in Alpine images, or default to BASH.Add commentMore actions
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

run:
	go run apis/services/web-hook/main.go | go run apis/tooling/logfmt/main.go

# ==============================================================================
# Modules support

tidy:
	go mod tidy
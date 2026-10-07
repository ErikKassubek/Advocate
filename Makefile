.PHONY: all build-patch patch clean-cache build-gocct build-bench

PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

all: build

build: patch build-gocct build-bench

patch: clean-dir build-patch clean-cache

build-patch:
	cd ./goPatch/src && ./make.bash

clean-dir:
	rm -rf ./goPatch/bin ./goPatch/pkg

clean-cache:
	./goPatch/bin/go clean -cache

build-gocct: 
	cd goCCT && go build

build-bench:
	cd benchmark/runner/ && go build

benchmark:
	cd benchmark/runner/ && ./runner

benchmark-gfuzz:
	cd benchmark/runner/ && ./runner -mode fuzzing -fuzzingMode gfuzz 

benchmark-gopie:
	cd benchmark/runner/ && ./runner -mode fuzzing -fuzzingMode gopie

paper-gfuzz:
	cd goCCT && ./gocct fuzzing -path $(PROJECT_ROOT)/benchmark/paper/ -mode GFuzz -exec TestBlkBug -stats -time

paper-gopie:
	cd goCCT && ./gocct fuzzing -path $(PROJECT_ROOT)/benchmark/paper/ -mode GoPie -exec TestPanicBug -stats -time
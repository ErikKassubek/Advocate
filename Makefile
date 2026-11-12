.PHONY: all build-patch patch clean-cache gosc

all: build

gosc: 
	cd gosc && go build

build: patch gosc

patch: clean-dir build-patch clean-cache

build-patch:
	cd ./goPatch/src && ./make.bash

clean-dir:
	rm -rf ./goPatch/bin ./goPatch/pkg

clean-cache:
	./goPatch/bin/go clean -cache

adv: gosc
	cd gosc && ./gosc static -path /home/erik/Arbeit/examples/blocking/main.go -main -output -panic

mkbuiltin:
	cd ./goPatch/src/cmd/compile/internal/typecheck && go run mkbuiltin.go
 
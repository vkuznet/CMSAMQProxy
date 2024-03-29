VERSION=`git rev-parse --short HEAD`
flags=-ldflags="-s -w -X main.version=${VERSION}"

all: build

build:
	go clean; rm -rf pkg; go build -o cmsamqproxy ${flags}

build_debug:
	go clean; rm -rf pkg; go build -o cmsamqproxy ${flags} -gcflags="-m -m"

build_all: build_darwin build_amd64 build

build_darwin:
	go clean; rm -rf pkg cmsamqproxy_darwin; GOOS=darwin go build -o cmsamqproxy ${flags}

build_amd64:
	go clean; rm -rf pkg cmsamqproxy_amd64; GOOS=linux go build -o cmsamqproxy ${flags}

build_power8:
	go clean; rm -rf pkg cmsamqproxy_power8; GOARCH=ppc64le GOOS=linux go build -o cmsamqproxy ${flags}

build_arm64:
	go clean; rm -rf pkg cmsamqproxy_arm64; GOARCH=arm64 GOOS=linux go build -o cmsamqproxy ${flags}

build_windows:
	go clean; rm -rf pkg cmsamqproxy.exe; GOARCH=amd64 GOOS=windows go build -o cmsamqproxy ${flags}

install:
	go install

clean:
	go clean; rm -rf pkg

test : test1

test1:
	go test

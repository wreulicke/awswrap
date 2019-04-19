ARCH := 386 amd64
OS := linux darwin windows

VERSION=0.0.1

VERSION_FLAG=\"main.version=$VERSION\"

.PHONY: test

setup:
	# for build
	GO111MODULE=off go get github.com/mitchellh/gox

format:
	GO111MODULE=on go fmt ./...

test:
	go test ./... 

build:
	GO111MODULE=on go build -ldflags "-X $(VERSION_FLAG)" -o ./dist/awswrap .

clean:
	rm -rf dist/
	rm -rf build/

build-all: 
	GO111MODULE=on gox -os="$(OS)" -arch="$(ARCH)" -ldflags "-X $(VERSION_FLAG)" -output "./dist/{{.Dir}}_{{.OS}}_{{.Arch}}" .

release:
	GO111MODULE=off go get github.com/tcnksm/ghr
	@ghr -t $(GITHUB_API_TOKEN_CM_PRODUCTDEV) -u $(CIRCLE_PROJECT_USERNAME) -r $(CIRCLE_PROJECT_REPONAME) $(VERSION) dist/

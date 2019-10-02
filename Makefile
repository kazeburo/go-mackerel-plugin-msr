VERSION=0.0.2
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
GO111MODULE=on

all: mackerel-plugin-msr

.PHONY: mackerel-plugin-msr

mackerel-plugin-msr: mackerel-plugin-msr.go
	go build $(LDFLAGS) -o mackerel-plugin-msr

linux: mackerel-plugin-msr.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o mackerel-plugin-msr

clean:
	rm -rf mackerel-plugin-msr

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
	goreleaser --rm-dist

VERSION=0.0.1

all: mackerel-plugin-msr

.PHONY: mackerel-plugin-msr

gom:
	go get -u github.com/mattn/gom

bundle:
	gom install

mackerel-plugin-msr: mackerel-plugin-msr.go
	gom build -o mackerel-plugin-msr

linux: mackerel-plugin-msr.go
	GOOS=linux GOARCH=amd64 gom build -o mackerel-plugin-msr

fmt:
	go fmt ./...

dist:
	git archive --format tgz HEAD -o mackerel-plugin-msr-$(VERSION).tar.gz --prefix mackerel-plugin-msr-$(VERSION)/

clean:
	rm -rf mackerel-plugin-msr mackerel-plugin-msr-*.tar.gz


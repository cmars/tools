
GB_BINS=gocov godef godeps http pt represent tmuxg

all: $(GB_BINS:%=bin/%)

$(GB_BINS:%=bin/%): bin/gb
	GOROOT=$(shell go env GOROOT) bin/gb build && [ -x "$@" ]

bin/gb:
	GOBIN=$(shell pwd)/bin GOPATH=$(shell mktemp -d) \
		go get github.com/constabulary/gb/cmd/gb

sysdeps:
	for step in sysdeps/*; do $$step; done

clean:
	$(RM) -f $(GB_BINS:%=bin/%) bin/gb

.PHONY: all sysdeps clean

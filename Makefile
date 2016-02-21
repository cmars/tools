
GB_BINS=gocov godef godeps http pt represent tmuxg

all: $(GB_BINS:%=bin/%)

$(GB_BINS:%=bin/%): bin/gb
	gb build && [ -x "$@" ]

bin/gb:
	GOBIN=$(shell pwd)/bin go get github.com/constabulary/gb/cmd/gb

clean:
	$(RM) -f $(GB_BINS:%=bin/%) bin/gb

SHELL=/bin/sh

demo_%:
	(cd examples/demo_$* && go build && ./demo_$(*) && diff out.wav ../../testdata/demo_$*/out.wav)

examples: demo_basic demo_chip demo_fixed demo_stereo
	cd examples/demo_sdl && go build -o demo_sdl

.PHONY: examples

tests: examples
	go test -v ./...

.PHONY: tests

clean:
	find examples -name out.wav -delete
	find examples -executable -type f -delete

.PHONY: clean

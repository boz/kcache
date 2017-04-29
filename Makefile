build:
	govendor build +local

install-libs:
	govendor install +vendor,^program

example:
	go build -o _example/example ./_example

.PHONY: build install-libs example

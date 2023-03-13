help:
	@echo help

init-example:
	mkdir -p examples/filedrop
	echo "this file will not match" examples/filedrop/area5.txt
	fortune >examples/filedrop/area001.txt
	date >examples/filedrop/area314.txt

examples: init-example
	envexec examples/*.env -- \
		go run ./src -f examples/filem.yaml

dry-run: init-example
	envexec examples/*.env -- \
		go run ./src -f examples/filem.yaml --dry-run

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/filem ./src

release:
	gh release create $(TAG) -t $(TAG)

check:
	@echo "Checking...\n"
	gocyclo -over 15 . || echo -n ""
	@echo ""
	golint -min_confidence 0.21 -set_exit_status ./...
	@echo "\nAll ok!"
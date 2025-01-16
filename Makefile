build:
	@go build -o ./bin/omega ./cmd
run: build
	@./bin/omega

build-once:
	@go build -o ./bin/once ./programs/once
run-once: build-once
	@./bin/once
hot: run-once

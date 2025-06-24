hello:
	echo "Hello, World!"

build:
	@go build -o myapp main.go

test:
	@go test -v ./...
s
run : build
	@./myapp
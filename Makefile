build: test compile

run:
	go run cmd/main.go

test:
	go test ./...

compile:
	echo "Compiling for multiple platforms"
	GOOS=freebsd GOARCH=amd64 go build -o bin/freebsd/todo cmd/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/linux/todo cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/win/todo.exe cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/macos/todo cmd/main.go
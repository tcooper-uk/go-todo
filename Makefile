run:
	go run cmd/main.go

build: 
	go build -o cmd/todo cmd/main.go

compile:
	echo "Compiling for multiple platforms"
	GOOS=freebsd GOARCH=amd64 go build -o bin/freebsd/todo cmd/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/linux/todo cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/win/todo.exe cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/macos/todo cmd/main.go
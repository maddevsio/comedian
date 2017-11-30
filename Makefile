TARGET=comedian

all: fmt clean build

clean:
	rm -rf $(TARGET)

fmt:
	go fmt ./...

build:
	go build -o $(TARGET) main.go

build_linux:
	GOOS=linux GOARCH=amd64 go build -o $(TARGET) main.go

build_docker:
	docker build -t comedian .

docker: build_linux build_docker

migrate:
	goose -dir migrations mysql "root:root@tcp(localhost:3306)/comedian"  up

c_migration:
	goose -dir migrations create migration_name sql
	
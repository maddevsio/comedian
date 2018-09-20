TARGET=comedian

all: fmt clean build docker

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
	goose -dir migrations mysql "comedian:comedian@tcp(172.18.0.3:3306)/comedian"  up

c_migration:
	goose -dir migrations create migration_name sql
	
db_clean:
	goose -dir migrations mysql "comedian:comedian@/comedian"  reset
	goose -dir migrations mysql "comedian:comedian@/comedian"  up

run_tests:
	go test ./storage/ ./chat/ ./notifier/ ./reporting/ ./config/ ./api/ ./teammonitoring/ ./utils/ -cover

test: db_clean run_tests

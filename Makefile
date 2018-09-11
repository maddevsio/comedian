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

tag_docker: 
	docker tag comedian:latest anatoliyfedorenko/comedian

deploy_docker:
	docker push anatoliyfedorenko/comedian

docker: build_linux build_docker tag_docker deploy_docker

migrate:
	goose -dir migrations mysql "comedian:comedian@tcp(172.18.0.3:3306)/comedian"  up

c_migration:
	goose -dir migrations create migration_name sql
	
ct:
	docker-compose run --rm comedian bash -c 'go test -cover -race ./api/ ./chat/ ./config/ ./notifier/ ./reporting/ ./storage/'
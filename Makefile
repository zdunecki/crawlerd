grpc-gen:
	protoc --go_out=. --go-grpc_out=require_unimplemented_servers=false:. crawlerd.proto

compose-infra:
	docker-compose -f ./build/docker-compose.infra.yml up

compose-up:
	docker-compose -f docker-compose.yml up --build

tests:
	go test ./...

docker-clean:
	docker rmi $(docker images -f "dangling=true" -q) -f

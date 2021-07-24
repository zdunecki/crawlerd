grpc-gen:
	protoc --go_out=. --go-grpc_out=require_unimplemented_servers=false:. crawlerd.proto

tests:
	go test ./...

docker-build-api:
	docker build -f build/api/Dockerfile -t localhost:5000/crawlerd/api:latest .

docker-push-api:
	docker push localhost:5000/crawlerd/api:latest

docker-build-scheduler:
	docker build -f build/scheduler/Dockerfile -t localhost:5000/crawlerd/scheduler:latest .

docker-push-scheduler:
	docker push localhost:5000/crawlerd/scheduler:latest

docker-build-worker:
	docker build -f build/worker/Dockerfile -t localhost:5000/crawlerd/worker:latest .

docker-push-worker:
	docker push localhost:5000/crawlerd/worker:latest

docker-run-local-registry:
	docker run -d -p 5000:5000 --restart=always --name registry registry:2

docker-clean:
	docker rmi $(docker images -f "dangling=true" -q) -f

compose-infra:
	docker-compose -f ./build/docker-compose.infra.yml up

compose-up:
	docker-compose -f docker-compose.yml up --build

# TODO: not working via make
debug-bin:
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./$1
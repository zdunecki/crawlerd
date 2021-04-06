grpc-gen:
	protoc --go_out=. --go-grpc_out=require_unimplemented_servers=false:. crawlerd.proto

compose-up:
	docker-compose -f docker-compose.yml up --build

run-tests:
	make run-e2e-tests
	make run-integration-tests

run-e2e-tests:
	docker-compose -f ./test/e2e/docker-compose.yml up --build --abort-on-container-exit --renew-anon-volumes

run-worker-integration-tests:
	docker-compose -f ./test/integration/worker/docker-compose.yml up --build --abort-on-container-exit --renew-anon-volumes

docker-clean:
	docker rmi $(docker images -f "dangling=true" -q) -f

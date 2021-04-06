# crawlerd

**crawlerd** is a crawler daemon for small crawling jobs. 

## commands
* `make grpc-gen` - generate grpc code
* `make compose-up` - run all crawlerd services in docker
* `make run-e2e-tests` - run e2e tests inside docker
* `make run-worker-integration-tests` - run integration tests inside docker

## TODO
**crawlerd** is still not production ready, below is the list needed for first production release:
- [ ] Rotating IP
- [ ] HTTP/Socks Proxy
- [ ] Fix bugs

build:
	go build -o bin/etcd-dashboard

test:
	go test -v ./...

release:
	CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w' -o bin/etcd-dashboard .

docker: release
	docker build -t etcd-dashboard .

#gox allows for cross-compiling for linux bits on a mac/windows go implementation
#this would be used by mac or windows developer who wants to use docker
#run this once - no need to run it again after that
cross-compile-setup:
	go get github.com/mitchellh/gox
	gox -build-toolchain -osarch="linux/amd64"


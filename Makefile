clean:
	go clean
deps:
	make clean
	-rm -rf vendor
	-rm -f go.mod
	-rm -f go.sum
	env GO111MODULE=on go mod init
	env GO111MODULE=on go mod vendor
test:
	go clean
	go test ./...
run:
	go run main.go
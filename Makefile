deps:
	-rm Gopkg.toml
	-rm Gopkg.lock
	-rm -r vendor
	dep init
test:
	go clean
	go test ./...
run:
	go run main.go
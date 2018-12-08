deps:
	-rm -r vendor
	dep ensure
test:
	go clean
	go test ./...
run:
	go run main.go
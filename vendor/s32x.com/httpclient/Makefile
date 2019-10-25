clean:
	go clean
deps:
	make clean
	-rm -rf vendor
	-rm -r glide.yaml
	-rm -f glide.lock
	glide init --non-interactive
	glide install
test:
	go clean
	go test ./...
run:
	go run main.go
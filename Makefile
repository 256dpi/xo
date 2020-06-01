all: fmt vet lint

fmt:
	go fmt .

vet:
	go vet .

lint:
	golint .

test:
	go test -cover .

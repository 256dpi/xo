all:
	go fmt ./...
	go vet ./...
	golint ./...

jaeger:
	# access UI on http://localhost:16686
	# send traces via OTLP to http://localhost:4318
	docker run -it --rm \
      -p 16686:16686 \
      -p 4317:4317 \
      -p 4318:4318 \
      -p 9411:9411 \
      jaegertracing/all-in-one:latest

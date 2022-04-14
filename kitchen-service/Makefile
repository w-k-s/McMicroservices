SOURCE=$(wildcard pkg/*.go) $(wildcard internal/*.go) $(wildcard test/*.go)

run: $(SOURCE)
	go run github.com/w-k-s/McMicroservices/kitchen-service/cmd/server
	
test: $(SOURCE)
	mkdir -p ~/.kitchen/migrations/
	cp migrations/*.sql ~/.kitchen/migrations/
	go test -coverprofile=coverage.txt -coverpkg=test/...,./... ./...
	
fmt: $(SOURCE)
	gofmt -w */**
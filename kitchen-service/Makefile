run:
	go run github.com/w-k-s/McMicroservices/kitchen-service/cmd/server
	
test:
	mkdir -p ~/.kitchen/migrations.d/
	cp migrations/*.sql ~/.kitchen/migrations.d/
	go test -coverprofile=coverage.txt -coverpkg=test/...,./... ./...
fmt:
	gofmt -w */**
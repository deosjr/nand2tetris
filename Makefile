run:
	@go run ./...
test:
	@go test -count=1 ./...
testrace:
	@go test -race -count=1 ./...

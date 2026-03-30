run:
	@go run . -tags gui
headless:
	@go run . -tags !gui
test:
	@go test -count=1 .
testrace:
	@go test -race -count=1 .

REPO = github.com/imega/snake-game
CWD = /go/src/$(REPO)
GO_IMG = golang:1.15.8-alpine3.13

test: lint unit

lint:
	@-docker run --rm -t -v $(CURDIR):$(CWD) -w $(CWD) \
		golangci/golangci-lint golangci-lint run

unit:
	@docker run --rm -w $(CWD) -v $(CURDIR):$(CWD) \
		$(GO_IMG) sh -c "go list ./... | xargs go test -vet=off"

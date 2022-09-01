lint:
	golangci-lint run --fix

test:
	go test -mod=mod --race .

test-ci:
	go test -mod=mod -count=1 -timeout 60s  -coverprofile=coverage.txt -covermode=atomic .
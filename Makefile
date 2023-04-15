lint:
	golangci-lint run --fix

test:
	go test -mod=mod -shuffle=on -race .

test-ci:
	go test -mod=mod -shuffle=on -race -timeout 60s -coverprofile=coverage.txt -covermode=atomic .

benchmark:
	go test -bench=.  -benchmem
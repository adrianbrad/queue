lint:
	golangci-lint run --fix

test:
	go test -mod=mod -shuffle=on -race .

test-ci:
	go test -mod=mod -shuffle=on -race -timeout 60s -coverprofile=coverage.txt -covermode=atomic .

test-coverage: test-ci
	@total=$$(go tool cover -func=coverage.txt | awk '/^total:/ {print $$3}'); \
	if [ "$$total" != "100.0%" ]; then \
		echo "coverage $$total is below required 100.0%"; \
		go tool cover -func=coverage.txt | awk '$$3 != "100.0%" && NF == 3'; \
		exit 1; \
	fi; \
	echo "coverage: $$total"

benchmark:
	go test -bench=.  -benchmem
.PHONY: test
test:
	go test -race

.PHONY: bench
bench:
	go test -race -bench=.

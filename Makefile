BINARY      := bin/chesssy
PKGS        := ./...
STATICCHECK := honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: all build run test test-race perft perft-deep bench bench-smp bench-go suite agreement measure pgo eval fmt vet lint check clean

all: check build

build:
	go build -o $(BINARY) ./cmd/chesssy

run: build
	$(BINARY)

# Unit tests, including perft to depth 4 for every position in the suite.
test:
	go test $(PKGS)

test-race:
	go test -race $(PKGS)

# Move generation verification against published node counts, shallow then deep.
perft:
	go test ./engine -run 'TestPerft$$|TestMakeUnmake' -count=1 -v

perft-deep:
	go test ./engine -run TestPerftDeep -count=1 -timeout 30m -v

# Search benchmark over a fixed position set; comparable between builds.
bench: build
	$(BINARY) bench -depth 10

# The same benchmark on every core, to show what the extra threads buy.
bench-smp: build
	$(BINARY) bench -depth 12 -threads 1
	$(BINARY) bench -depth 12 -threads $(shell sysctl -n hw.ncpu 2>/dev/null || nproc)

# Best-of-N timings for the workloads that matter. Use it to judge an optimisation:
# a change that only moves the node count has changed the search, not its speed.
measure: build
	scripts/measure 5

bench-go:
	go test ./engine ./search -run XXX -bench . -benchtime 2s

# How many nodes the search needs over forty positions from self-play. Single threaded
# the count is exactly reproducible, so it is the number to compare when changing move
# ordering or pruning.
suite:
	go test ./search -run XXX -bench BenchmarkSuite -benchtime 1x

# How often a shallow search picks the move a deep search chose. Read it together with
# suite: fewer nodes at lower agreement means the search is pruning too much.
agreement:
	go test ./search -run TestMoveAgreement -count=1 -v

eval: build
	$(BINARY) eval

# Regenerate the profile-guided optimisation input from the real binary, covering
# single threaded search, threaded search and move generation. Go picks up
# cmd/chesssy/default.pgo automatically, so this only needs rerunning when the hot
# paths change shape.
pgo: build
	$(BINARY) bench -depth 12 -cpuprofile /tmp/chesssy-bench.prof > /dev/null
	$(BINARY) bench -depth 12 -threads $(shell sysctl -n hw.ncpu 2>/dev/null || nproc) -cpuprofile /tmp/chesssy-smp.prof > /dev/null
	$(BINARY) perft -depth 6 -cpuprofile /tmp/chesssy-perft.prof > /dev/null
	go tool pprof -proto /tmp/chesssy-bench.prof /tmp/chesssy-smp.prof /tmp/chesssy-perft.prof > cmd/chesssy/default.pgo
	go build -o $(BINARY) ./cmd/chesssy

fmt:
	gofmt -l -w .

vet:
	go vet $(PKGS)

lint: vet
	go run $(STATICCHECK) $(PKGS)

check: fmt lint test

clean:
	rm -rf bin
	go clean -cache -testcache

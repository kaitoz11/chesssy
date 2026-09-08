# Contributing to chesssy

Thanks for your interest in chesssy. This document covers how to build the
project, what the tests expect, and how a change to the search or the evaluation
is judged — which is the part that differs most from an ordinary Go project.

By participating you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting set up

You need Go 1.23 or newer and `make`. There are no dependencies outside the
standard library, and none should be added without a strong reason: the engine is
meant to be readable and buildable with nothing but a Go toolchain.

```sh
git clone https://github.com/kaitoz11/chesssy
cd chesssy
make build          # writes bin/chesssy
make check          # format, lint and run every test
```

`make check` runs the same formatting check, vet, staticcheck and tests that CI
does, and reports formatting rather than fixing it, so that a passing `make check`
means CI will pass too. Run it before opening a pull request. CI additionally runs
the race detector, perft and a binary smoke test.

## Making a change

1. Fork the repository and create a branch off `main`.
2. Make the change, with tests beside their subject: `engine/magic.go` is tested
   by `engine/magic_test.go`.
3. Run `make check`. For anything touching move generation, also run
   `make perft` — and `make perft-deep` if the change could affect node counts,
   which takes several minutes and about 465M nodes.
4. Open a pull request describing what changed and, for search or evaluation
   changes, the measurements described below.

Keep pull requests focused on one thing. A move ordering change and a new
evaluation term are two pull requests, because their effects have to be measured
separately to mean anything.

## Code style

- `gofmt` decides formatting; `make fmt` applies it and CI fails if it would
  change anything.
- `go vet` and [staticcheck](https://staticcheck.dev/) must be clean. `make lint`
  runs both with the version CI uses.
- One topic per file, with a header comment explaining the idea before the code
  that implements it. `doc.go` in `engine` and `search` lists the files in
  reading order; add new files to that list.
- Exported identifiers get doc comments. Unexported ones get a comment when the
  reason for the code is not obvious from the code — which, in a chess engine,
  is often.
- Prefer clarity over cleverness except on the hot path, and when the hot path
  wins, say so in a comment and bring a measurement.

## Judging a change

Two numbers matter and they answer different questions.

- **Speed.** `make measure` (best of several runs) and `make bench` say how fast
  the engine gets through a node. `bench` searches a fixed nine-position set, so
  its node counts are reproducible and comparable between builds.
- **Search quality.** `make suite` reports the node count over forty positions
  from self-play, exactly reproducible single threaded, and says how many nodes
  the search needed in the first place. A move ordering or pruning change shows
  up here, not in the speed numbers.

Node count on its own is not enough, because pruning more always searches fewer
nodes and past some point does it by missing the best move. `make agreement`
reports how often a nine ply search picks the move a three second search on every
core chose — currently 62% of forty positions. A change that cuts nodes while
agreement falls is pruning too much, and will not be merged on the node count
alone. Razoring, for one, cut 5% of nodes and lost twenty points of agreement,
which is why it is not in the engine.

So a pull request that touches search or evaluation should report before/after
numbers for the parts it could plausibly move:

```sh
make suite        # node count over forty positions
make agreement    # move agreement with a deep search
make measure      # wall clock, best of five
```

State the machine you measured on. Comparisons are only meaningful between runs
on the same hardware, and nodes per second is not comparable between engines at
all.

If a change is a pure speed optimisation, the node counts must not move. If they
do, the change also changed the search, and both effects need reporting.

## Regenerating the PGO profile

The build uses [profile-guided optimisation](https://go.dev/doc/pgo) with the
profile in `cmd/chesssy/default.pgo`, collected from the binary running its own
bench and perft workloads. It only needs regenerating when the hot paths change
shape:

```sh
make pgo
```

Include the regenerated profile in the pull request when you do, and say why it
was needed.

## Reporting bugs

Open an issue using the bug report template. For anything involving a position,
include the FEN, the exact UCI command sequence, and what you expected instead.
A failing test case is the most useful bug report there is.

Illegal moves, crashes and protocol violations are the most serious class of bug:
they make the engine unusable in a GUI or a tournament. Playing strength
complaints are welcome, but they are feature requests, not bugs.

## Security

Do not open a public issue for a security problem. See [SECURITY.md](SECURITY.md).

## Licence

chesssy is distributed under the GNU General Public License v3.0 or later.
Contributions are accepted under those same terms: by opening a pull request you
license your work under the GPL-3.0-or-later, and it will be distributed with the
rest of the project under that licence. You keep the copyright to what you write;
there is no CLA.

Because the project is GPL licensed, code copied from a project under an
incompatible licence cannot be merged. If a change is derived from another engine
or from published work, say so in the pull request with a link, so the licence can
be checked before review.

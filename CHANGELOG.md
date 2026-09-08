# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html) — where a major version
bump means a breaking change to the command line, the UCI options or the exported
Go packages, not a change in playing strength.

## [Unreleased]

Nothing has been tagged yet. Everything below is the first release, which will be
`v1.0.0`; pushing that tag builds the binaries and publishes it.

### Added

A UCI chess engine in Go with no dependencies outside the standard library.

- **Board representation** — bitboards with a piece-per-square array, magic
  bitboard sliding attacks with constants searched at startup from a fixed seed,
  16-bit packed moves, and incremental Zobrist hashing with a make/unmake stack.
- **Move generation** — strictly legal moves generated directly, with checkers,
  pinned pieces, king moves against a king-removed occupancy, and en passant
  handled by simulated occupancy.
- **Search** — fail-soft principal variation search under iterative deepening with
  aspiration windows, a 16-byte-entry transposition table, quiescence search with
  SEE and delta pruning, null move pruning, reverse futility, futility and late
  move pruning, late move reductions, move ordering by TT move, SEE and MVV-LVA,
  killers, counter moves and a history table with gravity, mate distance pruning,
  draw detection, and soft/hard time management.
- **Evaluation** — tapered material and piece-square tables, mobility, pawn
  structure, rooks on open files, bishop pair, king safety and a tempo bonus, with
  a pawn structure cache.
- **Lazy SMP** — optional multi-threaded search via the `Threads` option, off by
  default so that a single thread searches deterministically.
- **UCI** — the protocol plus interactive `d`, `eval`, `fen`, `perft` and `bench`
  commands, and `perft`, `bench`, `eval` and `version` subcommands.
- **Profile-guided optimisation** — a profile collected from the engine's own
  bench and perft workloads, worth about 4% of search time.
- **Verification** — perft against published counts for the six standard positions
  plus a promotion-heavy one, make/unmake round trips with incremental Zobrist
  checks, brute-force-verified forced mates, evaluation colour symmetry, and a
  self-play game checked move by move for legality.
- **Project setup** — licence (GNU GPL v3.0 or later) with per-file SPDX
  identifiers, contribution guide, code of conduct and security policy;
  continuous integration across Linux, macOS and Windows with vet, staticcheck,
  the race detector, perft, govulncheck and CodeQL; a weekly deep verification
  run; and release automation producing binaries with checksums on tag. Actions
  are pinned by commit SHA, workflows default to a read-only token, and no
  workflow runs fork code with secrets.
- Startup notice reporting the copyright and the absence of warranty, as the GPL
  asks of an interactive program.

[Unreleased]: https://github.com/kaitoz11/chesssy/commits/main

# chesssy

```
   _|_  chesssy
  (o o)  a UCI chess engine
 --|-|--
```

chesssy is a [UCI](https://www.chessprogramming.org/UCI)-compatible chess engine
written in Go, with no dependencies outside the standard library.

## Quick start

```sh
make build          # writes bin/chesssy
./bin/chesssy       # speak UCI on stdin
```

Or install it straight from the module path:

```sh
go install github.com/kaitoz11/chesssy/cmd/chesssy@latest
```

To play against it, point a UCI GUI such as [Cute Chess](https://github.com/cutechess/cutechess),
[Arena](http://www.playwitharena.de/) or [BanksiaGUI](https://banksiagui.com/) at
the `bin/chesssy` binary.

You can also drive it by hand. Besides the UCI commands it accepts `d` to draw the
board, `eval`, `fen`, `perft <depth>` and `bench [depth]`:

```
$ ./bin/chesssy
position startpos moves e2e4 e7e5
d

8: r n b q k b n r
7: p p p p . p p p
6: . . . . . . . .
5: . . . . p . . .
4: . . . . P . . .
3: . . . . . . . .
2: P P P P . P P P
1: R N B Q K B N R
   a b c d e f g h

   w KQkq e6 0 2

go depth 12
info depth 1 seldepth 2 score cp 62 nodes 32 nps 372093 hashfull 0 time 0 pv g1f3
...
bestmove g1f3
```

Piping a script works as expected: end of input waits for a running search to
finish, while `quit` stops it immediately, as the protocol requires.

```sh
printf 'position startpos\ngo depth 12\n' | ./bin/chesssy
```

Subcommands are available for one-shot use:

```sh
./bin/chesssy perft -depth 6                      # verify move generation
./bin/chesssy perft -fen "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1" -depth 5
./bin/chesssy bench -depth 10                      # search a fixed position set
./bin/chesssy eval -fen "<fen>"                    # static evaluation
```

## Development

| Command | Description |
| --- | --- |
| `make check` | format, vet, lint and run every test |
| `make test` | unit tests, including perft to depth 4 |
| `make test-race` | same under the race detector |
| `make lint` | `go vet` plus staticcheck |
| `make perft` | move generation verification and make/unmake round trips |
| `make perft-deep` | the expensive perft depths (about 430M nodes) |
| `make bench` | search benchmark over a fixed position set |
| `make bench-go` | Go benchmarks for attacks, move generation, eval and search |

## Layout

| Package | Contents |
| --- | --- |
| `engine` | board representation, attack tables, move generation, FEN, Zobrist hashing, SEE, perft |
| `search` | evaluation, transposition table, move ordering, alpha-beta search |
| `uci` | the UCI protocol and the interactive commands |
| `cmd/chesssy` | the `chesssy` binary |

Each package keeps one topic per file, and each file has a header comment
explaining the idea it implements before the code that implements it. `doc.go` in
`engine` and `search` lists the files in reading order. Tests sit beside their
subject: `engine/magic.go` is tested by `engine/magic_test.go`.

## Design

### Board representation

Positions are bitboards: one `uint64` per piece type and one per color, plus a
piece-per-square array for fast lookups. Squares are indexed little-endian
rank-file (`A1 = 0`, `H8 = 63`), which makes "north" a shift left by 8.

Sliding attacks use plain [magic bitboards](https://www.chessprogramming.org/Magic_Bitboards):
the occupancy bits that matter for a square are multiplied by a per-square
constant, and the top bits index a dense attack table. The magic constants are
searched for at startup with a fixed-seed generator, so the tables are
deterministic without shipping a hard-coded list. `TestMagicAttacksMatchReference`
checks every lookup against a slow ray walker over 128,000 random occupancies.

Moves are packed into 16 bits: origin, destination, promotion piece and one of
four kinds (normal, promotion, en passant, castling). `MakeMove` updates the
Zobrist key incrementally and pushes the irreversible state onto a stack, so
`UnmakeMove` is cheap and the same `Position` is reused throughout a search.

### Move generation

The generator produces strictly legal moves directly rather than filtering
pseudo-legal ones by playing them:

- **Checkers** decide the shape of the move list. Under double check only the king
  can move; under single check every other move must capture the checker or block
  the line to the king.
- **Pinned pieces** are found once per node by looking for enemy sliders whose ray
  to our king is blocked by exactly one piece. A pinned piece may only move along
  that line.
- **King moves** are tested against an occupancy with the king removed, so sliders
  are not blocked by the piece that is about to move.
- **En passant** is the one case that needs a simulated occupancy, because removing
  the captured pawn can expose the king along a rank.

### Search

Fail-soft principal variation search under iterative deepening:

- aspiration windows around the previous score, widening on failure
- transposition table with 16-byte entries, four to a cache line, replacement by
  depth and age
- quiescence search with stand-pat, SEE pruning of losing captures and delta
  pruning
- null move pruning, reverse futility pruning, futility pruning, late move
  pruning and SEE pruning of bad captures — none of which apply to checking moves,
  which is what keeps forced mates visible
- late move reductions scaled by depth and move number, adjusted for the node
  type, the improving flag and history
- move ordering by transposition table move, then winning captures and queen
  promotions by SEE and MVV-LVA, killers, counter moves and a history table with
  gravity. The table move is tried before anything else is scored, since it often
  causes an immediate cutoff
- mate distance pruning, and draw detection by repetition, the fifty-move rule and
  insufficient material
- time management with a soft limit that prevents starting an iteration that
  cannot finish and a hard limit that aborts mid-search

### Evaluation

Tapered evaluation, blending middlegame and endgame terms by the material left on
the board. Terms are material and piece-square tables, mobility weighted by piece
type, pawn structure (doubled, isolated, backward and passed pawns), rooks on open
and semi-open files, the bishop pair, king shelter, pressure on the enemy king
zone, and a tempo bonus. The pawn-only part of the score is memoised in a cache
owned by the searcher, keyed by the pawn placement of both sides.

## Verification

Move generation is validated against published
[perft](https://www.chessprogramming.org/Perft_Results) node counts for the six
standard positions plus a promotion-heavy position. `make perft-deep` checks the
deep counts, for example:

| Position | Depth | Nodes |
| --- | --- | --- |
| startpos | 6 | 119,060,324 |
| kiwipete | 5 | 193,690,690 |
| position 3 | 6 | 11,030,083 |
| position 4 | 5 | 15,833,292 |
| position 5 | 5 | 89,941,194 |

Beyond perft, the test suite covers make/unmake round trips with incremental
Zobrist verification over the whole tree, castling and en passant edge cases, SEE,
colour symmetry of the evaluation, the UCI protocol, and a full self-play game
whose every move is checked for legality. Forced-mate expectations are not taken
from a puzzle collection: a brute-force solver in the test computes the true mate
distance, and the search must match it.

## Performance

Measured on an Apple M2 (4 performance + 4 efficiency cores) with `scripts/measure`,
which reports the best of several runs, since a run can only be made slower by
interference. `bench` searches a fixed nine-position set, so its node counts are
reproducible and comparable between builds.

| Measurement | Result |
| --- | --- |
| perft, bulk counted | ~145M nodes/s |
| search, 1 thread | ~2.5M nodes/s |
| `bench -depth 13`, 1 thread | 474 ms, 1.23M nodes |
| `bench -depth 13`, 4 threads | 269 ms |
| startpos to depth 15 | ~180 ms |

More threads help until the performance cores run out. Beyond that, lazy SMP puts work
on the efficiency cores, which run the same search two to three times slower while still
writing to the shared table, and the total gets worse — so the useful thread count on
this chip is four, not eight:

| Threads | Time to depth 15 | Speedup |
| --- | --- | --- |
| 1 | 1170 ms | 1.00x |
| 2 | 1129 ms | 1.04x |
| 4 | 732 ms | 1.60x |
| 8 | slower than four | — |

Threads are off by default, since a single thread searches deterministically; set the
`Threads` option to turn lazy SMP on.

The build uses [profile-guided optimisation](https://go.dev/doc/pgo): the profile in
`cmd/chesssy/default.pgo` is collected from the binary running its own bench and perft
workloads, and is worth about 4% of search time. `make pgo` regenerates it.

### Judging a change

Two numbers matter and they answer different questions. `scripts/measure` and the
nine-position `bench` say how fast the engine gets through a node. `BenchmarkSuite` in the
search package searches forty positions taken from self-play and reports the node count,
which is exactly reproducible single threaded, and says how many nodes the search needed
in the first place. A change to move ordering or pruning shows up in the second, not the
first.

Node count on its own is not enough, because pruning more always searches fewer nodes and
past some point does it by missing the best move. `TestMoveAgreement` is the counterweight:
it reports how often a nine ply search picks the move a three second search on every core
chose, currently 62% of forty positions. A change that cuts nodes while agreement falls is
pruning too much: razoring, for one, cut 5% of nodes and lost twenty points of agreement,
which is why it is not here.

Two notes on reading these numbers. Nodes per second is not comparable between engines: an
engine with a cheaper evaluation searches more nodes per second and may still play worse.
And chesssy is not competitive with the strongest engines — Stockfish evaluates positions
with a neural network, searches with a decade of tuning behind its pruning, and would beat
this engine from any starting position. What is claimed here is only that these numbers are
what chesssy does, measured the same way before and after each change.

## Options

| Option | Default | Meaning |
| --- | --- | --- |
| `Hash` | 64 | transposition table size in MB |
| `Threads` | 1 | search threads; more is faster up to the number of performance cores, and makes the search nondeterministic |



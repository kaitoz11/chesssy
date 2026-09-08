package engine

import "testing"

// perftSuite holds published node counts from the Chess Programming Wiki.
var perftSuite = []struct {
	name  string
	fen   string
	nodes []uint64 // nodes[i] is the count at depth i+1
}{
	{
		name:  "startpos",
		fen:   StartFEN,
		nodes: []uint64{20, 400, 8902, 197281, 4865609, 119060324},
	},
	{
		name:  "kiwipete",
		fen:   "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		nodes: []uint64{48, 2039, 97862, 4085603, 193690690},
	},
	{
		name:  "position3",
		fen:   "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		nodes: []uint64{14, 191, 2812, 43238, 674624, 11030083},
	},
	{
		name:  "position4",
		fen:   "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		nodes: []uint64{6, 264, 9467, 422333, 15833292},
	},
	{
		name:  "position4-mirrored",
		fen:   "r2q1rk1/pP1p2pp/Q4n2/bbp1p3/Np6/1B3NBn/pPPP1PPP/R3K2R b KQ - 0 1",
		nodes: []uint64{6, 264, 9467, 422333, 15833292},
	},
	{
		name:  "position5",
		fen:   "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		nodes: []uint64{44, 1486, 62379, 2103487, 89941194},
	},
	{
		name:  "position6",
		fen:   "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		nodes: []uint64{46, 2079, 89890, 3894594},
	},
	{
		name:  "promotion-bugs",
		fen:   "n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1",
		nodes: []uint64{24, 496, 9483, 182838, 3605103},
	},
}

// maxPerftDepth keeps the default test run fast. Deeper counts run with -perft.
const maxPerftDepth = 4

func TestPerft(t *testing.T) {
	for _, tc := range perftSuite {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatalf("parse FEN: %v", err)
			}
			for depth, want := range tc.nodes {
				if depth+1 > maxPerftDepth {
					break
				}
				if got := p.Perft(depth + 1); got != want {
					_, total := p.PerftDivide(depth + 1)
					t.Fatalf("perft(%d) = %d, want %d (divide total %d)", depth+1, got, want, total)
				}
			}
		})
	}
}

// TestPerftDeep covers the expensive depths. Run it with:
//
//	go test ./engine -run TestPerftDeep -timeout 30m
func TestPerftDeep(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping deep perft in short mode")
	}
	for _, tc := range perftSuite {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				t.Fatalf("parse FEN: %v", err)
			}
			for depth, want := range tc.nodes {
				if depth+1 <= maxPerftDepth {
					continue
				}
				if got := p.Perft(depth + 1); got != want {
					t.Fatalf("perft(%d) = %d, want %d", depth+1, got, want)
				}
			}
		})
	}
}

func BenchmarkPerft(b *testing.B) {
	for _, tc := range []struct {
		name  string
		fen   string
		depth int
	}{
		{"startpos-5", StartFEN, 5},
		{"kiwipete-4", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 4},
	} {
		b.Run(tc.name, func(b *testing.B) {
			p, err := NewPositionFromFEN(tc.fen)
			if err != nil {
				b.Fatal(err)
			}
			var nodes uint64
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				nodes = p.Perft(tc.depth)
			}
			b.StopTimer()
			b.ReportMetric(float64(nodes)*float64(b.N)/b.Elapsed().Seconds(), "nodes/s")
		})
	}
}

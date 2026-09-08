package search

import "math"

// Search tuning lives here, in one place, because these numbers are the knobs a
// engine is tuned on: every one of them trades safety for speed, and each is a
// guess that only self-play can confirm. Keeping them together makes a change
// reviewable and stops magic numbers from spreading through the search.

// MaxPly bounds the search stack: the deepest ply the search may visit.
const MaxPly = 128

// DefaultHashSizeMB is the transposition table size used unless the host asks for
// another one.
const DefaultHashSizeMB = 64

// Iterative deepening and aspiration windows.
const (
	// aspirationMinDepth is the first depth searched with a narrow window. Below
	// it the score is too unstable for a guess to pay off.
	aspirationMinDepth = 4

	// aspirationDelta is the initial half-width of that window, in centipawns.
	aspirationDelta = 25
)

// nodesBetweenLimitChecks is how often the clock is consulted. Reading the clock is
// relatively slow, and a thousand nodes is a small fraction of any sensible time
// limit. It must stay a power of two so that the modulo compiles to a mask.
const nodesBetweenLimitChecks = 1024

// Pruning margins. Each is the amount of score the search is willing to assume it
// can concede before it stops looking.
const (
	// reverseFutilityMaxDepth bounds where "our score is so far above beta that
	// the opponent cannot claw it back" is trusted.
	reverseFutilityMaxDepth = 7
	// reverseFutilityMargin is charged per ply of remaining depth, and
	// reverseFutilityImprovingBonus relaxes it when the score is rising.
	reverseFutilityMargin         = 85
	reverseFutilityImprovingBonus = 40

	// nullMoveMinDepth is the shallowest depth at which giving the opponent a
	// free move is informative.
	nullMoveMinDepth = 3
	// nullMoveBaseReduction and nullMoveDepthDivisor set how much shallower the
	// null-move search is: base + depth/divisor, plus up to
	// nullMoveMaxEvalBonus more when the static score is far above beta.
	nullMoveBaseReduction = 3
	nullMoveDepthDivisor  = 4
	nullMoveEvalDivisor   = 200
	nullMoveMaxEvalBonus  = 3

	// futilityMaxDepth, futilityBase and futilityPerDepth prune quiet moves that
	// cannot reasonably reach alpha.
	futilityMaxDepth = 6
	futilityBase     = 110
	futilityPerDepth = 120

	// lateMoveBase and lateMoveDepthDivisor set how many quiet moves are tried
	// before the rest are assumed to be no better, and
	// lateMoveImprovingScale relaxes that when the score is rising.
	lateMoveBase         = 3
	lateMoveDepthDivisor = 2
	lateMoveImprovingNum = 3
	lateMoveImprovingDen = 2

	// captureSeeMaxDepth and captureSeeMargin discard captures that lose too
	// much material, scaled so that deeper searches are more forgiving.
	captureSeeMaxDepth = 6
	captureSeeMargin   = 25

	// quiescenceDeltaMargin is the slack allowed when deciding that even winning
	// a capture outright cannot reach alpha.
	quiescenceDeltaMargin = 150
)

// Internal iterative reduction. A node with no transposition table move has nothing to
// order its moves by, so a full depth search of it is largely wasted: better to search it
// a ply shallower, and let the move that search finds order the re-search.
const iirMinDepth = 4

// lazyEvalMargin is how far outside the search window the material count has to be before
// the positional terms are taken on trust. It is generous: the terms it skips are worth a
// few hundred centipawns between them, and being wrong here means pruning a node that
// deserved a look.
const lazyEvalMargin = 600

// Late move reductions.
const (
	// lmrMinDepth and lmrMinMoveCount keep the first moves of shallow nodes at
	// full depth.
	lmrMinDepth     = 3
	lmrMinMoveCount = 2

	// lmrBase and lmrDivisor shape the reduction curve, which grows with the logarithm
	// of both depth and move number. They were swept against the measurement suite:
	// this pair searches a fifth fewer nodes than a shallower curve and agrees with a
	// deep search slightly more often, so it is not trading accuracy for speed.
	lmrBase    = 1.00
	lmrDivisor = 2.10
)

// lmrReductions[depth][moveCount] is the base late move reduction. It is indexed
// with saturating bounds, since the curve is flat by the time either axis runs
// out.
var lmrReductions [64][64]int

func init() {
	for depth := 1; depth < len(lmrReductions); depth++ {
		for moveCount := 1; moveCount < len(lmrReductions[depth]); moveCount++ {
			r := lmrBase + math.Log(float64(depth))*math.Log(float64(moveCount))/lmrDivisor
			lmrReductions[depth][moveCount] = int(r)
		}
	}
}

// baseReduction returns the tabulated reduction for a depth and move number.
func baseReduction(depth, moveCount int) int {
	return lmrReductions[min(depth, len(lmrReductions)-1)][min(moveCount, len(lmrReductions)-1)]
}

// lateMoveThreshold returns how many quiet moves a node tries before pruning the
// rest.
func lateMoveThreshold(depth int, improving bool) int {
	threshold := lateMoveBase + depth*depth/lateMoveDepthDivisor
	if improving {
		threshold = threshold * lateMoveImprovingNum / lateMoveImprovingDen
	}
	return threshold
}

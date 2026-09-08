// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package search

import "github.com/kaitoz11/chesssy/engine"

// Evaluation answers "who is better, and by how much?" for a position the search
// has decided not to look past. It is tapered: every term has a middlegame value
// and an endgame value, and the two are blended by how much material is left. A
// king that belongs in a corner during the middlegame belongs in the centre once
// the queens are gone, and one set of numbers cannot say both.

// Piece term weights, {middlegame, endgame}.
var (
	// Two bishops cover both square colors, which is worth more than the sum of
	// the parts, especially in an open endgame.
	bishopPairBonus = [2]int{22, 52}
	// A rook on a file with no pawns at all, or with only enemy pawns.
	rookOpenFile     = [2]int{28, 14}
	rookSemiOpenFile = [2]int{12, 8}
	// Mobility, per square a piece can safely move to. Rooks and queens gain more
	// in the endgame, when the lines open.
	mobilityBonus = [engine.PieceTypeCount][2]int{
		engine.Knight: {4, 3},
		engine.Bishop: {4, 4},
		engine.Rook:   {3, 5},
		engine.Queen:  {2, 4},
	}
	// What an attack on a square near the enemy king is worth, per piece type. It
	// is scaled down in the endgame, where an attack with few pieces left rarely
	// becomes an attack at all.
	kingAttackWeight = [engine.PieceTypeCount]int{
		engine.Knight: 12,
		engine.Bishop: 10,
		engine.Rook:   16,
		engine.Queen:  24,
	}
	// Charged per missing pawn in front of our own king, in the middlegame only:
	// in the endgame the king wants to walk, not hide.
	kingShieldPenalty = [2]int{-14, 0}
	// A small bonus for having the move, which keeps the evaluation from calling
	// otherwise equal positions exactly equal.
	tempoBonus = 12

	// endgameKingPressureDivisor scales king pressure down for the endgame.
	endgameKingPressureDivisor = 4
	// expectedShieldPawns is how many pawns a castled king normally has in front
	// of it; the penalty counts how many are missing.
	expectedShieldPawns = 3
)

// King zone masks.
var (
	// kingZone[s] holds the squares near a king on s: its own square, everything it
	// attacks, and one rank either side of that. Attacks landing in the zone are
	// what "pressure on the king" counts.
	kingZone = computeKingZones()
	// kingShield[c][s] holds the two ranks in front of a king on s, on its own and
	// adjacent files, where its pawn cover should be.
	kingShield = computeKingShields()
)

func computeKingZones() (zones [engine.SquareCount]engine.Bitboard) {
	for s := engine.A1; s < engine.SquareCount; s++ {
		around := engine.KingAttacks(s).With(s)
		zones[s] = around | around.Shift(engine.North) | around.Shift(engine.South)
	}
	return zones
}

func computeKingShields() (shields [engine.ColorCount][engine.SquareCount]engine.Bitboard) {
	for s := engine.A1; s < engine.SquareCount; s++ {
		region := engine.FileBB(s.File()) | adjacentFiles[s.File()]
		for c := engine.White; c < engine.ColorCount; c++ {
			shields[c][s] = region & shieldRanks(c, s.Rank())
		}
	}
	return shields
}

// shieldRanks returns the two ranks in front of r from c's point of view.
func shieldRanks(c engine.Color, r engine.Rank) engine.Bitboard {
	var bb engine.Bitboard
	for step := engine.Rank(1); step <= 2; step++ {
		switch {
		case c == engine.White && r+step <= engine.Rank8:
			bb |= engine.RankBB(r + step)
		case c == engine.Black && r >= step:
			bb |= engine.RankBB(r - step)
		}
	}
	return bb
}

// Evaluate scores the position in centipawns from the point of view of the side to move.
// A positive score means the side to move is better.
func Evaluate(p *engine.Position) int { return evaluate(p, nil) }

// evaluateLazy is Evaluate with a shortcut. Counting material is a fraction of the cost
// of the full evaluation, and when the material count alone puts the position far outside
// the window the search is asking about, the positional terms cannot bridge the gap and
// there is no point computing them.
//
// The second result says whether the score is the full evaluation. A shortcut score is
// good enough to prune on but is not stored as though it were exact.
func evaluateLazy(p *engine.Position, cache *pawnCache, alpha, beta int) (score int, exact bool) {
	coarse := materialBalance(p)
	if coarse-lazyEvalMargin >= beta || coarse+lazyEvalMargin <= alpha {
		return coarse, false
	}
	return evaluate(p, cache), true
}

// materialBalance scores the material on the board, from the point of view of the side to
// move, blended by phase like the full evaluation so that the two are on one scale.
func materialBalance(p *engine.Position) int {
	mg, eg, phase := 0, 0, 0
	for pt := engine.Pawn; pt < engine.King; pt++ {
		count := p.Count(engine.White, pt) - p.Count(engine.Black, pt)
		mg += count * mgPieceValue[pt]
		eg += count * egPieceValue[pt]
		phase += (p.Count(engine.White, pt) + p.Count(engine.Black, pt)) * phaseWeight[pt]
	}

	score := taper(mg, eg, phase)
	if p.SideToMove() == engine.Black {
		score = -score
	}
	return score + tempoBonus
}

// evaluate is Evaluate with an optional pawn cache, which a Searcher supplies.
func evaluate(p *engine.Position, cache *pawnCache) int {
	pawns := [engine.ColorCount]engine.Bitboard{
		p.PieceBB(engine.White, engine.Pawn),
		p.PieceBB(engine.Black, engine.Pawn),
	}

	// Terms are summed as white minus black throughout, and flipped at the end for
	// the side to move.
	mg, eg := pawnScore(pawns, cache)
	phase := 0
	for c := engine.White; c < engine.ColorCount; c++ {
		cmg, ceg, cphase := evaluatePieces(p, c, pawns)
		sign := signFor(c)
		mg += sign * cmg
		eg += sign * ceg
		phase += cphase
	}

	score := taper(mg, eg, phase)
	if p.SideToMove() == engine.Black {
		score = -score
	}
	return score + tempoBonus
}

// taper blends a middlegame and an endgame score by how much material is left.
func taper(mg, eg, phase int) int {
	phase = min(phase, maxPhase)
	return (mg*phase + eg*(maxPhase-phase)) / maxPhase
}

// evaluatePieces scores everything about c's pieces other than its pawn structure,
// and reports c's share of the game phase.
func evaluatePieces(p *engine.Position, c engine.Color, pawns [engine.ColorCount]engine.Bitboard) (mg, eg, phase int) {
	them := c.Flip()
	occupied := p.Occupied()
	// Squares guarded by an enemy pawn are not somewhere a piece wants to go, so
	// they do not count towards mobility.
	enemyPawnAttacks := engine.PawnAttacksBB(them, pawns[them])
	theirKingZone := kingZone[p.KingSquare(them)]
	kingPressure := 0

	for pt := engine.Knight; pt < engine.PieceTypeCount; pt++ {
		piece := engine.NewPiece(c, pt)

		for bb := p.PieceBB(c, pt); bb != 0; {
			s := bb.PopLSB()
			mg += mgTable[piece][s]
			eg += egTable[piece][s]
			phase += phaseWeight[pt]

			if pt == engine.Rook {
				fileMg, fileEg := rookFileBonus(s, c, them, pawns)
				mg += fileMg
				eg += fileEg
			}
			if pt == engine.King {
				continue // the king neither roams nor attacks in this evaluation
			}

			attacks := engine.PieceAttacks(pt, s, occupied)
			mobility := (attacks & ^p.ColorBB(c) & ^enemyPawnAttacks).Count()
			mg += mobility * mobilityBonus[pt][0]
			eg += mobility * mobilityBonus[pt][1]

			kingPressure += (attacks & theirKingZone).Count() * kingAttackWeight[pt]
		}
	}

	if p.Count(c, engine.Bishop) >= 2 {
		mg += bishopPairBonus[0]
		eg += bishopPairBonus[1]
	}

	missing := expectedShieldPawns - (kingShield[c][p.KingSquare(c)] & pawns[c]).Count()
	if missing > 0 {
		mg += missing * kingShieldPenalty[0]
		eg += missing * kingShieldPenalty[1]
	}

	mg += kingPressure
	eg += kingPressure / endgameKingPressureDivisor
	return mg, eg, phase
}

// rookFileBonus rewards a rook on a file its own pawns have left, which is where a
// rook does its work.
func rookFileBonus(s engine.Square, c, them engine.Color, pawns [engine.ColorCount]engine.Bitboard) (mg, eg int) {
	file := engine.FileBB(s.File())
	if pawns[c]&file != 0 {
		return 0, 0
	}
	bonus := rookSemiOpenFile
	if pawns[them]&file == 0 {
		bonus = rookOpenFile
	}
	return bonus[0], bonus[1]
}

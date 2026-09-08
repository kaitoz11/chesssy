package uci

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kaitoz11/chesssy/engine"
	"github.com/kaitoz11/chesssy/search"
)

// The go command starts a search:
//
//	go depth 12
//	go movetime 1000
//	go wtime 300000 btime 300000 winc 3000 binc 3000
//	go infinite                      (until stop)
//	go perft 5                       (count moves instead of searching)
//
// The search runs on its own goroutine, because the protocol requires the engine to
// keep reading commands while it thinks: stop has to arrive somehow. It reports
// progress with info lines and finishes with exactly one bestmove, which is what the
// host waits for.

func (e *Engine) handleGo(args []string) bool {
	e.stopSearch()

	limits, perftDepth := parseGoArgs(args)
	if perftDepth > 0 {
		e.runPerft(perftDepth)
		return false
	}

	// The search works on a copy, so that the next command can change the engine's
	// position while it runs.
	position := e.position.Clone()
	e.searcher.OnInfo = e.printInfo

	e.searches.Add(1)
	e.searcher.SearchAsync(position, limits, func(info search.Info) {
		defer e.searches.Done()
		e.printf("bestmove %s\n", bestMoveOf(info, position))
	})
	return false
}

// bestMoveOf returns the move to report. A search stopped before it finished an
// iteration may have no move, and the protocol still expects one, so fall back on
// any legal move and finally on the "no move" token.
func bestMoveOf(info search.Info, position *engine.Position) engine.Move {
	if best := info.BestMove(); best != engine.NoMove {
		return best
	}
	var list engine.MoveList
	position.GenerateMoves(&list)
	if list.Len() > 0 {
		return list.At(0)
	}
	return engine.NoMove // renders as "0000": checkmate or stalemate
}

// parseGoArgs reads the search limits from a go command. Unknown or malformed
// arguments are skipped rather than rejected, since a missing limit still leaves a
// searchable position.
func parseGoArgs(args []string) (limits search.Limits, perftDepth int) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "infinite":
			limits.Infinite = true
			continue
		case "ponder":
			continue // pondering is not implemented; search normally
		}

		// Everything else is a name followed by a number.
		i++
		if i >= len(args) {
			break
		}
		n, err := strconv.Atoi(args[i])
		if err != nil {
			continue
		}

		switch args[i-1] {
		case "depth":
			limits.Depth = n
		case "nodes":
			limits.Nodes = uint64(n)
		case "movetime":
			limits.MoveTime = milliseconds(n)
		case "wtime":
			limits.Time[engine.White] = milliseconds(n)
		case "btime":
			limits.Time[engine.Black] = milliseconds(n)
		case "winc":
			limits.Increment[engine.White] = milliseconds(n)
		case "binc":
			limits.Increment[engine.Black] = milliseconds(n)
		case "movestogo":
			limits.MovesToGo = n
		case "perft":
			perftDepth = n
		}
	}
	return limits, perftDepth
}

func milliseconds(n int) time.Duration { return time.Duration(n) * time.Millisecond }

// printInfo reports one completed iteration in the format hosts parse.
func (e *Engine) printInfo(info search.Info) {
	score := fmt.Sprintf("cp %d", info.Score)
	if info.Mate != 0 {
		score = fmt.Sprintf("mate %d", info.Mate)
	}

	moves := make([]string, 0, len(info.PV))
	for _, m := range info.PV {
		moves = append(moves, m.String())
	}

	e.printf("info depth %d seldepth %d score %s nodes %d nps %d hashfull %d time %d pv %s\n",
		info.Depth, info.SelDepth, score, info.Nodes, info.NodesPerSecond(),
		e.searcher.HashFull(), info.Elapsed.Milliseconds(), strings.Join(moves, " "))
}

// stopSearch aborts any running search and waits for it to report its bestmove, so
// that the engine is settled before a command changes its state.
func (e *Engine) stopSearch() {
	e.searcher.Stop()
	e.searches.Wait()
}

// Package uci speaks the Universal Chess Interface, the line-based protocol chess
// GUIs use to drive engines.
//
// A host sends commands on one stream and reads replies on another:
//
//	uci                                  -> id, options, uciok
//	isready                              -> readyok
//	position startpos moves e2e4 e7e5
//	go depth 12                          -> info lines, then bestmove
//
// [Engine.Run] reads commands until end of input. Searches run on their own
// goroutine so that stop arrives while one is in progress; see search.go.
package uci

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/kaitoz11/chesssy/engine"
	"github.com/kaitoz11/chesssy/search"
)

// Identification reported in reply to the uci command.
const (
	Name    = "chesssy"
	Version = "1.0"
	Author  = "cuongluu"
)

// Engine holds the protocol state: the current position, the searcher, and the
// goroutine running any search in progress.
type Engine struct {
	position *engine.Position
	searcher *search.Searcher

	// searches counts running searches, so that a command needing a settled
	// engine can wait for one to finish.
	searches sync.WaitGroup

	// outMu guards writes, since the search goroutine reports progress while the
	// command loop may be replying to something else.
	outMu sync.Mutex
	out   *bufio.Writer
}

// New returns an engine that writes its replies to out.
func New(out io.Writer) *Engine {
	return &Engine{
		position: engine.NewPosition(),
		searcher: search.NewSearcher(search.DefaultHashSizeMB),
		out:      bufio.NewWriter(out),
	}
}

// Run reads commands from in until end of input or a quit command.
//
// A quit command stops a running search immediately, as the protocol requires.
// Plain end of input instead waits for it to finish, so that piping a script such
// as "position startpos\ngo depth 12\n" prints the whole search rather than an
// aborted one.
func (e *Engine) Run(in io.Reader) error {
	scanner := bufio.NewScanner(in)
	// "position ... moves" grows with the game and can outgrow the default limit.
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	quit := false
	for scanner.Scan() && !quit {
		quit = e.Execute(scanner.Text())
	}
	if quit {
		e.searcher.Stop()
	}
	e.searches.Wait()
	return scanner.Err()
}

// Execute runs one command line and reports whether the engine should quit.
// Unknown commands are reported and ignored, as the protocol requires.
func (e *Engine) Execute(line string) (quit bool) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	name, args := fields[0], fields[1:]

	cmd, ok := commandsByName[name]
	if !ok {
		e.printf("info string unknown command %q\n", name)
		return false
	}
	return cmd.run(e, args)
}

// command is one line-oriented command the engine understands.
type command struct {
	name string
	// summary is what the help command prints. An empty summary hides the
	// command, which is what the standard protocol commands do: a GUI knows them
	// already and a human does not need them listed.
	summary string
	run     func(e *Engine, args []string) (quit bool)
}

// commands lists every command. The protocol ones come first, then the extras that
// make the engine pleasant to drive by hand.
//
// It is a function rather than a variable because handleHelp reads the list, and a
// variable holding handlers that read it would be an initialisation cycle.
func commands() []command {
	return []command{
		{name: "uci", run: (*Engine).handleUCI},
		{name: "isready", run: (*Engine).handleIsReady},
		{name: "ucinewgame", run: (*Engine).handleNewGame},
		{name: "setoption", run: (*Engine).handleSetOption},
		{name: "position", run: (*Engine).handlePosition},
		{name: "go", run: (*Engine).handleGo},
		{name: "stop", run: (*Engine).handleStop},
		{name: "ponderhit", run: (*Engine).handlePonderHit},
		{name: "quit", run: (*Engine).handleQuit},

		{name: "d", summary: "draw the current board", run: (*Engine).handleDisplay},
		{name: "display", run: (*Engine).handleDisplay},
		{name: "board", run: (*Engine).handleDisplay},
		{name: "eval", summary: "print the static evaluation", run: (*Engine).handleEval},
		{name: "fen", summary: "print the current position as FEN", run: (*Engine).handleFEN},
		{name: "perft", summary: "perft <depth>: count moves per legal move", run: (*Engine).handlePerft},
		{name: "bench", summary: "bench [depth]: search a fixed position set", run: (*Engine).handleBench},
		{name: "help", summary: "list these commands", run: (*Engine).handleHelp},
	}
}

// commandsByName indexes the commands for dispatch.
var commandsByName = func() map[string]command {
	list := commands()
	byName := make(map[string]command, len(list))
	for _, c := range list {
		byName[c.name] = c
	}
	return byName
}()

func (e *Engine) handleUCI([]string) bool {
	e.printf("id name %s %s\n", Name, Version)
	e.printf("id author %s\n", Author)
	e.printf("option name Hash type spin default %d min %d max %d\n",
		search.DefaultHashSizeMB, minHashMB, maxHashMB)
	e.printf("option name Threads type spin default 1 min %d max %d\n", minThreads, maxThreads)
	e.printf("uciok\n")
	return false
}

// handleIsReady answers immediately, even mid-search: the host uses it to check
// that the engine is alive, not that it is idle.
func (e *Engine) handleIsReady([]string) bool {
	e.printf("readyok\n")
	return false
}

func (e *Engine) handleNewGame([]string) bool {
	e.stopSearch()
	e.searcher.NewGame()
	e.position = engine.NewPosition()
	return false
}

func (e *Engine) handleStop([]string) bool {
	e.stopSearch()
	return false
}

// handlePonderHit would confirm that the opponent played the move being pondered.
// Pondering is not implemented, so the search already running is a normal one and
// needs no adjustment.
func (e *Engine) handlePonderHit([]string) bool { return false }

func (e *Engine) handleQuit([]string) bool {
	e.stopSearch()
	return true
}

func (e *Engine) handleDisplay([]string) bool {
	e.printf("\n%s\n", e.position)
	return false
}

func (e *Engine) handleEval([]string) bool {
	e.printf("info string static eval %d cp (side to move)\n", search.Evaluate(e.position))
	return false
}

func (e *Engine) handleFEN([]string) bool {
	e.printf("%s\n", e.position.FEN())
	return false
}

func (e *Engine) handleHelp([]string) bool {
	var sb strings.Builder
	sb.WriteString("commands:\n")
	for _, c := range commands() {
		if c.summary != "" {
			fmt.Fprintf(&sb, "  %-10s %s\n", c.name, c.summary)
		}
	}
	sb.WriteString("  plus the UCI commands: uci, isready, ucinewgame, setoption, position, go, stop, quit\n")
	e.printf("%s", sb.String())
	return false
}

// printf writes one reply. Every write goes through here, which is what keeps the
// search goroutine's info lines from interleaving with command replies.
func (e *Engine) printf(format string, args ...any) {
	e.outMu.Lock()
	defer e.outMu.Unlock()
	fmt.Fprintf(e.out, format, args...)
	// Hosts read line by line and will wait forever for a buffered reply.
	e.out.Flush()
}

package uci

import (
	"runtime"
	"strconv"
	"strings"
)

// Options are set with lines of the form
//
//	setoption name Hash value 128
//
// where the name may contain spaces, which is why it has to be collected up to the
// value keyword rather than simply split.

// Option bounds advertised to the host.
const (
	minHashMB = 1
	maxHashMB = 4096

	minThreads = 1
)

// maxThreads is the number of cores, since lazy SMP scales with them and beyond them
// the threads only compete for the same cache.
var maxThreads = runtime.NumCPU()

func (e *Engine) handleSetOption(args []string) bool {
	name, value := parseOption(args)

	switch strings.ToLower(name) {
	case "hash":
		mb, err := strconv.Atoi(value)
		if err != nil || mb < minHashMB || mb > maxHashMB {
			e.printf("info string Hash needs a value between %d and %d\n", minHashMB, maxHashMB)
			return false
		}
		e.stopSearch()
		e.searcher.SetHashSize(mb)

	case "clear hash":
		e.stopSearch()
		e.searcher.NewGame()

	case "threads":
		n, err := strconv.Atoi(value)
		if err != nil || n < minThreads || n > maxThreads {
			e.printf("info string Threads needs a value between %d and %d\n", minThreads, maxThreads)
			return false
		}
		e.stopSearch()
		e.searcher.SetThreads(n)

	default:
		e.printf("info string unknown option %q\n", name)
	}
	return false
}

// parseOption splits a setoption argument list into its name and value.
func parseOption(args []string) (name, value string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "name":
			start := i + 1
			for i+1 < len(args) && args[i+1] != "value" {
				i++
			}
			name = strings.Join(args[start:min(i+1, len(args))], " ")
		case "value":
			value = strings.Join(args[i+1:], " ")
			return name, value
		}
	}
	return name, value
}

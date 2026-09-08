// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

// Command chesssy is a UCI chess engine.
//
// Run it without arguments to speak UCI on standard input, which is what a chess
// GUI expects:
//
//	chesssy
//
// The same binary carries a few development subcommands:
//
//	chesssy perft [-depth N] [-fen FEN]   count moves, to verify move generation
//	chesssy bench [-depth N]              search a fixed position set
//	chesssy eval  [-fen FEN]              print the static evaluation
//	chesssy version
//
// Any subcommand accepts -cpuprofile, which writes a CPU profile of the run. Merging
// such profiles into cmd/chesssy/default.pgo is how the build gets its
// profile-guided optimisation input; see the Makefile's pgo target.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/kaitoz11/chesssy/engine"
	"github.com/kaitoz11/chesssy/uci"
)

const banner = `
   _|_  chesssy %s
  (o o)  a UCI chess engine
 --|-|--
%s
`

const usage = "usage: chesssy [uci|perft|bench|eval|version] [flags]"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "chesssy:", err)
		os.Exit(1)
	}
}

// run dispatches a command line. Taking the streams as arguments rather than
// reaching for the process globals is what lets the tests drive it.
func run(args []string, out io.Writer, in io.Reader) error {
	// A leading word that is not a flag selects a subcommand; anything else means
	// the default, which is to speak UCI.
	name := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		name, args = args[0], args[1:]
	}

	switch name {
	case "", "uci":
		return runUCI(out, in)
	case "perft":
		return runPerft(args, out)
	case "bench":
		return runBench(args, out)
	case "eval":
		return runEval(args, out)
	case "version":
		fmt.Fprintf(out, "%s %s\n%s\n", uci.Name, uci.Version, uci.Notice)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n%s", name, usage)
	}
}

// runUCI speaks the protocol until end of input or a quit command.
func runUCI(out io.Writer, in io.Reader) error {
	fmt.Fprintf(out, banner, uci.Version, uci.Notice)
	return uci.New(out).Run(in)
}

func runPerft(args []string, out io.Writer) error {
	flags, profile := newFlagSet("perft", out)
	depth := flags.Int("depth", 5, "depth to count to")
	fen := flags.String("fen", engine.StartFEN, "position to count from")
	if err := flags.Parse(args); err != nil {
		return err
	}
	stop, err := startProfile(*profile)
	if err != nil {
		return err
	}
	defer stop()
	return runCommands(out, *fen, fmt.Sprintf("perft %d", *depth))
}

func runBench(args []string, out io.Writer) error {
	flags, profile := newFlagSet("bench", out)
	depth := flags.Int("depth", 8, "depth to search each position to")
	threads := flags.Int("threads", 1, "search threads")
	if err := flags.Parse(args); err != nil {
		return err
	}
	stop, err := startProfile(*profile)
	if err != nil {
		return err
	}
	defer stop()
	return runCommands(out, "", fmt.Sprintf("bench %d %d", *depth, *threads))
}

func runEval(args []string, out io.Writer) error {
	flags, _ := newFlagSet("eval", out)
	fen := flags.String("fen", engine.StartFEN, "position to evaluate")
	if err := flags.Parse(args); err != nil {
		return err
	}
	return runCommands(out, *fen, "d", "eval")
}

// newFlagSet returns a flag set with the flags every subcommand shares.
func newFlagSet(name string, out io.Writer) (*flag.FlagSet, *string) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(out)
	profile := flags.String("cpuprofile", "", "write a CPU profile of this run to a file")
	return flags, profile
}

// startProfile begins CPU profiling when a path was given. The returned function
// stops it, and does nothing when profiling was not requested.
func startProfile(path string) (stop func(), err error) {
	if path == "" {
		return func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return nil, err
	}
	return func() {
		pprof.StopCPUProfile()
		f.Close()
	}, nil
}

// runCommands drives the UCI engine programmatically, so that the subcommands and
// the interactive protocol share one implementation. An empty fen leaves the engine
// on its default position.
func runCommands(out io.Writer, fen string, commands ...string) error {
	e := uci.New(out)
	if fen != "" {
		commands = append([]string{"position fen " + fen}, commands...)
	}
	for _, c := range commands {
		if quit := e.Execute(c); quit {
			return nil
		}
	}
	return nil
}

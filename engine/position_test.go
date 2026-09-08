// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2026 kaitoz11

package engine

import (
	"testing"
)

func TestCloneIsIndependent(t *testing.T) {
	p := NewPosition()
	m, _ := p.ParseMove("e2e4")
	p.MakeMove(m)

	c := p.Clone()
	m, _ = c.ParseMove("e7e5")
	c.MakeMove(m)

	if p.FEN() == c.FEN() {
		t.Error("clone shares state with the original")
	}
	c.UnmakeMove()
	if p.FEN() != c.FEN() {
		t.Errorf("clone diverged: %q vs %q", p.FEN(), c.FEN())
	}
}

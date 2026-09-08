package search

import (
	"math/bits"
	"sync/atomic"

	"github.com/kaitoz11/chesssy/engine"
)

// The transposition table is a lossy cache of searched positions. Transpositions are
// common in chess, so remembering "this position was already searched to depth n and
// scored s" saves whole subtrees; the stored move is worth keeping even when the
// score is not usable, because trying it first tends to cut off immediately.
//
// # Sharing it between threads
//
// Every search thread reads and writes the same table, with no locks: locking a
// shared structure this hot would cost more than the table saves. Instead each entry
// is two atomic words, the payload and the payload exclusive-ored with the position
// key. A reader recovers the key by exclusive-oring them back together, so an entry
// half-written by another thread fails that check and is simply treated as a miss.
// The technique is due to Robert Hyatt; the cost of a rare lost entry is nothing,
// while acting on a half-written one would mean searching an illegal move.

// Bound describes how a stored score relates to the true score of a position.
type Bound uint8

const (
	// BoundNone marks an empty slot.
	BoundNone Bound = iota
	// BoundUpper means the true score is at most the stored score, the result of a
	// node where no move reached alpha.
	BoundUpper
	// BoundLower means the true score is at least the stored score, the result of a
	// beta cutoff.
	BoundLower
	// BoundExact means the stored score is the true score.
	BoundExact
)

// Entry is one transposition table record, as callers see it.
type Entry struct {
	// Move is the best move found in the position, or engine.NoMove.
	Move engine.Move
	// Score is the search score. Mate scores are stored relative to the mating
	// position; scoreToTT and scoreFromTT convert.
	Score int
	// StaticEval is the static evaluation of the position, or noEval.
	StaticEval int
	// Depth is the remaining depth the score was searched to.
	Depth int
	// Bound says how Score relates to the true score.
	Bound Bound
}

// slot is an entry as stored: the payload packed into 64 bits, and the same payload
// exclusive-ored with the position key so that readers can both identify and validate
// it. Sixteen bytes, so a bucket of four is exactly one cache line.
//
// The fields are plain uint64 and are read and written through sync/atomic while a
// search is running. Keeping them plain is what lets Clear zero the whole table with a
// single memset, which matters: a table is tens of megabytes, and storing to it one
// atomic word at a time takes tens of milliseconds.
type slot struct {
	payload uint64
	checked uint64
}

func (s *slot) loadPayload() uint64 { return atomic.LoadUint64(&s.payload) }
func (s *slot) loadChecked() uint64 { return atomic.LoadUint64(&s.checked) }

// store writes an entry. The payload goes first: a reader that catches the pair
// mid-write recovers a key that matches neither the old entry nor the new one, and
// treats it as a miss.
func (s *slot) store(payload, key uint64) {
	atomic.StoreUint64(&s.payload, payload)
	atomic.StoreUint64(&s.checked, payload^key)
}

// Payload bit layout.
const (
	payloadMoveShift  = 0
	payloadScoreShift = 16
	payloadEvalShift  = 32
	payloadDepthShift = 48
	payloadAgeShift   = 56

	payloadMoveMask  = 0xFFFF
	payloadScoreMask = 0xFFFF
	payloadEvalMask  = 0xFFFF
	payloadDepthMask = 0xFF
	payloadAgeMask   = 0xFF
)

const (
	// bucketSize entries share one index, which turns a probe into a single cache
	// line read and gives a cheap choice of what to evict.
	bucketSize = 4
	slotBytes  = 16

	ageShift  = 2
	ageMask   = 0x3F
	boundMask = 0x3

	// staleDepthPenalty is how much depth an entry from an earlier search is
	// discounted by when choosing what to evict.
	staleDepthPenalty = 8
	// keepDeeperMargin protects a deeper exact entry from being overwritten by a
	// shallower inexact one.
	keepDeeperMargin = 2
)

type bucket [bucketSize]slot

// Table is a transposition table of fixed size. Its zero value is not usable; build
// one with NewTable. All methods are safe for concurrent use.
type Table struct {
	buckets []bucket
	mask    uint64
	age     atomic.Uint32
}

// NewTable allocates a table of about sizeMB megabytes, rounded down to a power of
// two number of buckets so that indexing is a mask rather than a division.
func NewTable(sizeMB int) *Table {
	sizeMB = max(sizeMB, 1)
	wanted := uint64(sizeMB) * 1024 * 1024 / uint64(bucketSize*slotBytes)
	count := uint64(1) << (bits.Len64(wanted) - 1)
	return &Table{buckets: make([]bucket, count), mask: count - 1}
}

// Clear empties the table. It must not be called while a search is running.
//
// It is called between games, not between moves: zeroing tens of megabytes costs
// milliseconds, and reallocating instead is worse, because the fresh pages then have to
// fault in one by one as the search touches them.
func (t *Table) Clear() {
	clear(t.buckets)
	t.age.Store(0)
}

// NewSearch marks the start of a search, so that entries from earlier ones are
// evicted before fresh ones.
func (t *Table) NewSearch() { t.age.Store((t.age.Load() + 1) & ageMask) }

// Probe looks up key. The returned Slot records where a matching entry was found, or
// else where one should go, so that a later Store needs no second pass over the
// bucket: probing and storing are the two halves of visiting a node, and the bucket is
// the same one both times.
func (t *Table) Probe(key uint64) (Entry, bool, Slot) {
	b := &t.buckets[key&t.mask]
	age := uint8(t.age.Load())

	victim, victimValue := &b[0], int(1<<30)
	victimMove := engine.NoMove

	for i := range b {
		payload := b[i].loadPayload()
		if payload == 0 {
			// Empty, and therefore also the cheapest slot to reuse. No second atomic
			// load is needed to know that.
			return Entry{Move: engine.NoMove}, false, Slot{slot: &b[i], move: engine.NoMove}
		}

		if payload^b[i].loadChecked() == key {
			entry := decodeEntry(payload)
			return entry, entry.Bound != BoundNone, Slot{
				slot:         &b[i],
				move:         entry.Move,
				depth:        int16(entry.Depth),
				samePosition: true,
			}
		}

		// Prefer to discard shallow entries, and entries from an earlier search. Only
		// the depth and the age are needed for that, so the rest stays packed.
		value := entryDepth(payload)
		if entryAge(payload) != age {
			value -= staleDepthPenalty
		}
		if value < victimValue {
			victim, victimValue, victimMove = &b[i], value, entryMove(payload)
		}
	}
	return Entry{Move: engine.NoMove}, false, Slot{slot: victim, move: victimMove}
}

// Slot is where an entry belongs, as worked out by Probe. It is deliberately small: it
// lives in a local variable across the whole of a search node, so every extra word in
// it is stack traffic in the hottest function there is.
type Slot struct {
	slot *slot
	// move and depth describe what the slot already holds, which is what Store needs
	// to decide whether replacing it is an improvement.
	move         engine.Move
	depth        int16
	samePosition bool
}

// Store records e under key in the slot a previous Probe of the same key picked out.
func (t *Table) Store(at Slot, key uint64, e Entry) {
	if at.slot == nil {
		return
	}
	// Keep a deeper exact entry rather than replacing it with a shallow guess.
	if at.samePosition && e.Bound != BoundExact && e.Depth < int(at.depth)-keepDeeperMargin {
		return
	}
	// Keep the previous best move when this search did not produce one.
	if e.Move == engine.NoMove && at.samePosition {
		e.Move = at.move
	}

	at.slot.store(encodeEntry(e, uint8(t.age.Load())), key)
}

// Hashfull reports occupancy in per mille, the unit UCI expects. It samples the first
// thousand slots rather than walking the whole table.
func (t *Table) Hashfull() int {
	sampled, used := 0, 0
	for i := 0; i < len(t.buckets) && sampled < 1000; i++ {
		for j := range t.buckets[i] {
			sampled++
			if decodeEntry(t.buckets[i][j].loadPayload()).Bound != BoundNone {
				used++
			}
		}
	}
	if sampled == 0 {
		return 0
	}
	return used * 1000 / sampled
}

func encodeEntry(e Entry, age uint8) uint64 {
	return uint64(uint16(e.Move))<<payloadMoveShift |
		uint64(uint16(int16(e.Score)))<<payloadScoreShift |
		uint64(uint16(int16(e.StaticEval)))<<payloadEvalShift |
		uint64(uint8(int8(e.Depth)))<<payloadDepthShift |
		uint64(age<<ageShift|uint8(e.Bound))<<payloadAgeShift
}

func decodeEntry(payload uint64) Entry {
	return Entry{
		Move:       engine.Move(payload >> payloadMoveShift & payloadMoveMask),
		Score:      int(int16(payload >> payloadScoreShift & payloadScoreMask)),
		StaticEval: int(int16(payload >> payloadEvalShift & payloadEvalMask)),
		Depth:      int(int8(payload >> payloadDepthShift & payloadDepthMask)),
		Bound:      Bound(payload >> payloadAgeShift & boundMask),
	}
}

func entryDepth(payload uint64) int {
	return int(int8(payload >> payloadDepthShift & payloadDepthMask))
}

func entryMove(payload uint64) engine.Move {
	return engine.Move(payload >> payloadMoveShift & payloadMoveMask)
}

func entryAge(payload uint64) uint8 {
	return uint8(payload>>payloadAgeShift&payloadAgeMask) >> ageShift
}

// scoreToTT rewrites a mate score so that it counts from the mating position rather
// than from the current ply, which is what makes it reusable at another ply.
func scoreToTT(score, ply int) int {
	switch {
	case score >= mateInMaxPly:
		return score + ply
	case score <= -mateInMaxPly:
		return score - ply
	default:
		return score
	}
}

// scoreFromTT is the inverse of scoreToTT.
func scoreFromTT(score, ply int) int {
	switch {
	case score >= mateInMaxPly:
		return score - ply
	case score <= -mateInMaxPly:
		return score + ply
	default:
		return score
	}
}

// allowsCutoff reports whether a stored score can be returned immediately for the
// window alpha..beta.
func (e Entry) allowsCutoff(alpha, beta int) bool {
	switch e.Bound {
	case BoundExact:
		return true
	case BoundLower:
		return e.Score >= beta
	case BoundUpper:
		return e.Score <= alpha
	default:
		return false
	}
}

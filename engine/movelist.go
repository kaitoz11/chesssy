package engine

// MaxMoves bounds the number of legal moves in any reachable position. The known
// maximum is 218, so a 256 entry buffer can never overflow.
const MaxMoves = 256

// MoveList is the fixed-capacity buffer move generation writes into. Because it
// is sized for the worst case, generating moves never allocates, and a caller can
// keep one list per search ply and reuse it.
//
// The zero value is an empty list ready for use.
type MoveList struct {
	moves [MaxMoves]Move
	count int
}

// Add appends m. Adding more than MaxMoves moves is a programming error and
// panics with an index out of range.
func (l *MoveList) Add(m Move) {
	l.moves[l.count] = m
	l.count++
}

// Reset empties the list, keeping its capacity.
func (l *MoveList) Reset() { l.count = 0 }

// Len returns the number of moves in the list.
func (l *MoveList) Len() int { return l.count }

// At returns the move at index i, which must be less than Len.
func (l *MoveList) At(i int) Move { return l.moves[i] }

// Swap exchanges the moves at i and j. Move ordering is built on it.
func (l *MoveList) Swap(i, j int) { l.moves[i], l.moves[j] = l.moves[j], l.moves[i] }

// Moves returns the moves as a slice, for ranging over or copying. The slice
// aliases the list's storage and is only valid until the next Add or Reset.
func (l *MoveList) Moves() []Move { return l.moves[:l.count] }

// Contains reports whether the list holds m.
func (l *MoveList) Contains(m Move) bool {
	for _, candidate := range l.Moves() {
		if candidate == m {
			return true
		}
	}
	return false
}

// IndexOf returns the position of m in the list, or -1 when it is absent.
func (l *MoveList) IndexOf(m Move) int {
	for i, candidate := range l.Moves() {
		if candidate == m {
			return i
		}
	}
	return -1
}

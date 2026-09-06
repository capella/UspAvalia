package models

import "fmt"

// Movement is how an entry's position in a top-rated list changed since the
// end of the previous semester.
type Movement struct {
	// Known is false when there is no baseline to compare with.
	Known bool
	// New is true when the entry was not in the baseline list.
	New bool
	// Delta is the number of positions gained (negative when it dropped).
	Delta int
}

// Label is the short text shown next to the entry: "novo", "▲2", "▼1" or
// empty when unchanged or unknown.
func (m Movement) Label() string {
	switch {
	case !m.Known:
		return ""
	case m.New:
		return "novo"
	case m.Delta > 0:
		return fmt.Sprintf("▲%d", m.Delta)
	case m.Delta < 0:
		return fmt.Sprintf("▼%d", -m.Delta)
	default:
		return ""
	}
}

// Class is the Bootstrap text colour class matching Label.
func (m Movement) Class() string {
	switch {
	case !m.Known:
		return ""
	case m.New:
		return "text-primary"
	case m.Delta > 0:
		return "text-success"
	case m.Delta < 0:
		return "text-danger"
	default:
		return "text-muted"
	}
}

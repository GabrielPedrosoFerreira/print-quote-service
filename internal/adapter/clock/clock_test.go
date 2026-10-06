package clock

import (
	"regexp"
	"testing"
	"time"
)

func TestRandomIDsAreWellFormedAndUnique(t *testing.T) {
	format := regexp.MustCompile(`^qt_[0-9a-f]{32}$`)
	seen := make(map[string]struct{})

	for range 1_000 {
		id := RandomIDs{}.NewID()
		if !format.MatchString(id) {
			t.Fatalf("id %q does not match %s", id, format)
		}
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestSystemClockIsUTC(t *testing.T) {
	if loc := (System{}).Now().Location(); loc != time.UTC {
		t.Errorf("location = %v, want UTC", loc)
	}
}

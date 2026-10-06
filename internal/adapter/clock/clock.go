// Package clock provides the production Clock and IDGenerator.
package clock

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// System is the wall clock, in UTC so stored timestamps never depend on server locale.
type System struct{}

func (System) Now() time.Time { return time.Now().UTC() }

// RandomIDs generates IDs like "qt_3f9a1c...". 128 random bits make
// collisions negligible without a coordination service.
type RandomIDs struct{}

func (RandomIDs) NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand only fails if the OS entropy source is broken;
		// continuing would risk duplicate IDs, so stop loudly.
		panic("crypto/rand unavailable: " + err.Error())
	}
	return "qt_" + hex.EncodeToString(b[:])
}

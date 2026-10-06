package domain

import "fmt"

// Money is an amount in US cents. Integer cents keep pricing exact:
// float64 cannot represent values like 0.1 and drifts when multiplied by quantities.
type Money int64

// String formats the amount as dollars, e.g. 1234 -> "$12.34".
func (m Money) String() string {
	sign := ""
	if m < 0 {
		sign = "-"
		m = -m
	}
	return fmt.Sprintf("%s$%d.%02d", sign, m/100, m%100)
}

// percentOf returns m * basisPoints / 10_000, rounded half up.
// One basis point is 0.01%, so 2_500 basis points is 25%.
func (m Money) percentOf(basisPoints int64) Money {
	return Money((int64(m)*basisPoints + 5_000) / 10_000)
}

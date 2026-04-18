package calc

import (
	"math"
	"time"
)

func ABV(initialGravity, finalGravity float64) float64 {
	return (initialGravity - finalGravity) * 131.25
}

func FermentationDays(startedAt string) (int, error) {
	t, err := time.Parse("2006-01-02", startedAt)
	if err != nil {
		return 0, err
	}
	return int(math.Round(time.Since(t).Hours() / 24)), nil
}

// SugarsRemaining estimates grams of sugar left given initial amount and observed gravity drop.
// Assumes a typical kombucha OG drop of 0.040 for full fermentation.
func SugarsRemaining(initialSugarG, gravityDelta float64) float64 {
	const typicalDrop = 0.040
	remaining := initialSugarG * (1 - gravityDelta/typicalDrop)
	if remaining < 0 {
		return 0
	}
	return remaining
}

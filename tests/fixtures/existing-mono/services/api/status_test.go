package api

import "testing"

func TestCanReviewSeller(t *testing.T) {
	if !CanReviewSeller("operator") {
		t.Fatal("operator should be able to review a seller")
	}
}

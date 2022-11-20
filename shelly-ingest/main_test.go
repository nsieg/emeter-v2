package main

import (
	"testing"
)

func TestCalcSends(t *testing.T) {
	// Given
	resp := response{[]float64{1.3, 2.4, 3.5}, true, 1.0, 1.0, 1654274493, 1.0}

	// When
	dp := calcSends(resp)

	// Then
	if dp[0].Time != 1654274460 || dp[0].Wh != 1.3/60 ||
		dp[1].Time != 1654274400 || dp[1].Wh != 2.4/60 ||
		dp[2].Time != 1654274340 || dp[2].Wh != 3.5/60 {
		t.Error("Wrong measurements!")
	}
}

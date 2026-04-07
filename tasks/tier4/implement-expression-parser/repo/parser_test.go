package exprparser

import "testing"

func TestCalcStub(t *testing.T) {
	// This test currently fails because Calc is not implemented.
	// It's just here to verify the package compiles.
	_, err := Calc("1+1")
	if err == nil {
		// Once implemented, this should succeed.
	}
	_ = err
}

package strkit

import "testing"

func TestAbs(t *testing.T) {
	if Abs(-5) != 5 {
		t.Fatal("Abs(-5) should be 5")
	}
	if Abs(5) != 5 {
		t.Fatal("Abs(5) should be 5")
	}
	if Abs(0) != 0 {
		t.Fatal("Abs(0) should be 0")
	}
}

func TestClamp(t *testing.T) {
	if Clamp(5, 0, 10) != 5 {
		t.Fatal("Clamp(5, 0, 10) should be 5")
	}
	if Clamp(-5, 0, 10) != 0 {
		t.Fatal("Clamp(-5, 0, 10) should be 0")
	}
	if Clamp(15, 0, 10) != 10 {
		t.Fatal("Clamp(15, 0, 10) should be 10")
	}
}

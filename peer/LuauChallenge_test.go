package peer

import "testing"

func TestLuauChallenge(t *testing.T) {
	var expected int32 = 5553188
	var result = ResolveLuaChallenge(1215435312, 250190180)
	if result != expected {
		t.Fatalf("%08X != %08X", result, expected)
	}
}

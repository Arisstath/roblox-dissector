package peer

import "testing"

func TestLuauChallenge(t *testing.T) {
	var expected int32 = 1890828
	var result = ResolveLuaChallenge(-0x1558389F, -0x75320513)
	if result != expected {
		t.Fatalf("%08X != %08X", result, expected)
	}
}

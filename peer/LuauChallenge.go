package peer

import "strconv"

var luauOps = []byte{8, 5, 7, 1, 6, 3, 4, 2}

func intAbs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// ResolveLuaChallenge implements the Luau callback challenge algorithm
func ResolveLuaChallenge(a1, a2 int32) int32 {
	var a int32 = 0x189
	var c int32 = 1

	for i := 0; i < 16; i++ {
		e := intAbs(a2+a+c) % 8
		var f int32

		op1 := a + a1
		op2 := a2
		switch luauOps[e] - 1 {
		case 0:
			f = op1 + op2
		case 1:
			f = op1 - op2
		case 2:
			f = op2 - op1
		case 3:
			t, _ := strconv.Atoi(strconv.Itoa(int(op2%1000)) + strconv.Itoa(int(intAbs(op1%1000))))
			f = int32(t)
		case 4:
			t, _ := strconv.Atoi(strconv.Itoa(int(op2%1337)) + strconv.Itoa(int(intAbs(op1%1337))))
			f = int32(t)
		case 5:
			f = (op1 % 1000) * (op2 % 1000)
		case 6:
			f = (op1 % 1337) * (op2 % 1337)
		case 7:
			oldOp1 := op1
			op1 = op1 + op2
			op2 = oldOp1 - op2

			f = (op1 % 1000) * (op2 % 1000)
		default:
			panic("bad luauchallenge op")
		}

		a = (a + f) % 10000000
		c++
	}

	return a
}

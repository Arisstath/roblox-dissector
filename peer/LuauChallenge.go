package peer

import (
	"math"
	"strconv"
)

var luauOps = []byte{8, 5, 7, 1, 6, 3, 4, 2}

func luaLikeMod(x, y float64) float64 {
	if x >= 0 {
		return math.Mod(x, y)
	}

	q := math.Floor(-x / y)
	m := y*q + y
	return math.Mod(x+m, y)
}

func floatToString(f float64) string {
	s := strconv.FormatFloat(f, 'f', 0, 64)
	return s
}

// ResolveLuaChallenge implements the Luau callback challenge algorithm
func ResolveLuaChallenge(arg1, arg2 int32) int32 {
	a1 := float64(arg1)
	a2 := float64(arg2)
	var a float64 = 0x189
	var c float64 = 1
	var err error

	for i := 0; i < 16; i++ {
		e := luaLikeMod(math.Abs(a2+a+c), 8)
		var f float64

		op1 := a + a1
		op2 := a2
		switch luauOps[int(e)] - 1 {
		case 0:
			f = op1 + op2
		case 1:
			f = op1 - op2
		case 2:
			f = op2 - op1
		case 3:
			f, err = strconv.ParseFloat(
				floatToString(luaLikeMod(op2, 1000))+
					floatToString(
						math.Abs(luaLikeMod(op1, 1000))), 64)
			if err != nil {
				panic(err)
			}
		case 4:
			f, err = strconv.ParseFloat(
				floatToString(luaLikeMod(op2, 1337))+
					floatToString(
						math.Abs(luaLikeMod(op1, 1337))), 64)
			if err != nil {
				panic(err)
			}
		case 5:
			f = luaLikeMod(op1, 1000) * luaLikeMod(op2, 1000)
		case 6:
			f = luaLikeMod(op1, 1337) * luaLikeMod(op2, 1337)
		case 7:
			oldOp1 := op1
			op1 = op1 + op2
			op2 = oldOp1 - op2

			f = luaLikeMod(op1, 1000) * luaLikeMod(op2, 1000)
		default:
			panic("bad luauchallenge op")
		}

		a = luaLikeMod((a + f), 10000000)
		c++
	}

	return int32(a)
}

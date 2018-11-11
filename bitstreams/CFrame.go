package bitstreams
import "github.com/gskartwii/rbxfile"
import "errors"

func (b *BitstreamReader) ReadCFrameSimple() (rbxfile.ValueCFrame, error) {
	return rbxfile.ValueCFrame{}, nil // nop for now, since nothing uses this
}

func quaternionToRotMatrix(q [4]float32) [9]float32 {
	X := q[0]
	Y := q[1]
	Z := q[2]
	W := q[3]
	XS := X * 2
	YS := Y * 2
	ZS := Z * 2
	WX := W * XS
	WY := W * YS
	WZ := W * ZS

	XX := XS * X
	XY := YS * X
	XZ := ZS * X
	YY := YS * Y
	YZ := ZS * Y
	ZZ := ZS * Z

	return [9]float32{
		1 - (YY + ZZ),
		XY - WZ,
		XZ + WY,
		XY + WZ,
		1 - (XX + ZZ),
		YZ - WX,
		XZ - WY,
		YZ + WX,
		1 - (XX + YY),
	}
}

func transformQuaternionToMatrix(q [4]float32) [9]float32 {
	midresult := quaternionToRotMatrix(q)

	xScaleFactor := float32(1.0 / math.Sqrt(float64(midresult[0]*midresult[0]+midresult[3]*midresult[3]+midresult[6]*midresult[6])))

	result := [9]float32{
		xScaleFactor * midresult[0], 0, 0,
		xScaleFactor * midresult[3], 0, 0,
		xScaleFactor * midresult[6], 0, 0,
	} // X has been normalized

	sXYSum := result[0]*midresult[1] + result[3]*midresult[4] + result[6]*midresult[7]
	trueR01 := midresult[1] - result[0]*sXYSum
	trueR11 := midresult[4] - result[3]*sXYSum
	trueR21 := midresult[7] - result[6]*sXYSum

	yScaleFactor := float32(1.0 / math.Sqrt(float64(trueR01*trueR01+trueR11*trueR11+trueR21*trueR21)))
	result[1] = trueR01 * yScaleFactor
	result[4] = trueR11 * yScaleFactor
	result[7] = trueR21 * yScaleFactor

	sYZSum := result[1]*midresult[2] + result[4]*midresult[5] + result[7]*midresult[8]
	sXZSum := result[0]*midresult[2] + result[3]*midresult[5] + result[6]*midresult[8]
	trueR02 := midresult[2] - result[0]*sXZSum - result[1]*sYZSum
	trueR12 := midresult[5] - result[3]*sXZSum - result[4]*sYZSum
	trueR22 := midresult[8] - result[6]*sXZSum - result[7]*sYZSum

	zScaleFactor := float32(1.0 / math.Sqrt(float64(trueR02*trueR02+trueR12*trueR12+trueR22*trueR22)))
	result[2] = trueR02 * zScaleFactor
	result[5] = trueR12 * zScaleFactor
	result[8] = trueR22 * zScaleFactor

	return result
}

var specialRows = [6][3]float32{
	[3]float32{1, 0, 0},
	[3]float32{0, 1, 0},
	[3]float32{0, 0, 1},
	[3]float32{-1, 0, 0},
	[3]float32{0, -1, 0},
	[3]float32{0, 0, -1},
}

func lookupRotMatrix(special uint64) [9]float32 {
	specialRowDiv6 := specialRows[special/6]
	specialRowMod6 := specialRows[special%6]

	ret := [9]float32{
		0, 0, 0,
		specialRowMod6[0], specialRowMod6[1], specialRowMod6[2],
		specialRowDiv6[0], specialRowDiv6[1], specialRowDiv6[2],
	}
	ret[0] = ret[2*3+1]*ret[1*3+2] - ret[2*3+2]*ret[1*3+1]
	ret[1] = ret[1*3+0]*ret[2*3+2] - ret[2*3+0]*ret[1*3+2]
	ret[2] = ret[2*3+0]*ret[1*3+1] - ret[1*3+0]*ret[2*3+1]

	trueRet := [9]float32{
		ret[6], ret[7], ret[8],
		ret[3], ret[4], ret[5],
		ret[0], ret[1], ret[2],
	}

	return trueRet
}

func (b *BitstreamReader) ReadCFrame() (rbxfile.ValueCFrame, error) {
	var err error
	val := rbxfile.ValueCFrame{}
	val.Position, err = b.ReadVector3Simple()
	if err != nil {
		return val, err
	}

	special, err := b.ReadUint8()
	if err != nil {
		return val, err
	}
	if special > 0 {
		if err != nil {
			return val, err
		}
		if special > 36 {
			println("oob, special", special)
			return val, errors.New("special rotmatrix oob")
		}
		val.Rotation = lookupRotMatrix(uint64(special - 1))
	} else {
		val.Rotation, err = b.ReadPhysicsMatrix()
	}

	return val, err
}

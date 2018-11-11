package bitstreams
import "github.com/gskartwii/rbxfile"
import "errors"
import "math"

// PhysicsMotor is an alias type for rbxfile.ValueCFrames. They are used to
// describe motors in physics packets
type PhysicsMotor rbxfile.ValueCFrame

// Returns the stringified version of the motor
func (m PhysicsMotor) String() string {
	return rbxfile.ValueCFrame(m).String()
}

func (b *BitstreamReader) readCoordsMode0() (rbxfile.ValueVector3, error) {
	return b.readVector3Simple()
}
func (b *BitstreamReader) readCoordsMode1() (rbxfile.ValueVector3, error) {
	value := rbxfile.ValueVector3{}
	cRange, err := b.readFloat32BE()
	if err != nil {
		return value, err
	}
	if cRange <= 0.0000099999997 { // Has to be precise
		return rbxfile.ValueVector3{0, 0, 0}, nil
	}
	x, err := b.readUint16BE()
	if err != nil {
		return value, err
	}
	y, err := b.readUint16BE()
	if err != nil {
		return value, err
	}
	z, err := b.readUint16BE()
	if err != nil {
		return value, err
	}
	value.X = (float32(x)/32767.0 - 1.0) * cRange
	value.Y = (float32(y)/32767.0 - 1.0) * cRange
	value.Z = (float32(z)/32767.0 - 1.0) * cRange
	return value, nil
}
func (b *BitstreamReader) readCoordsMode2() (rbxfile.ValueVector3, error) {
	val := rbxfile.ValueVector3{}
	x, err := b.bits(15)
	if err != nil {
		return val, err
	}
	xShort := uint16((x >> 7) | ((x & 0x7F) << 8))
	y, err := b.bits(14)
	if err != nil {
		return val, err
	}
	yShort := uint16((y >> 6) | ((y & 0x3F) >> 8))
	z, err := b.bits(15)
	if err != nil {
		return val, err
	}
	zShort := uint16((z >> 7) | ((z & 0x7F) << 8))

	val.X = float32(xShort)*0.0625 - 2048.0
	val.Y = float32(yShort)*0.0625 - 2048.0
	val.Z = float32(zShort)*0.0625 - 2048.0
	return val, nil
}

func (b *BitstreamReader) readPhysicsCoords() (rbxfile.ValueVector3, error) {
	var val rbxfile.ValueVector3
	flags, err := b.readUint8()
	if err != nil {
		return val, err
	}
	exp := flags >> 3 // void signs
	scale := float32(math.Exp2(float64(exp)))

	// nice bitpacking!
	if exp > 4 {
		if exp > 10 { // when scale is > 0x400
			// minimum precision is 2^11/2^21 = 2^-10
			// 21 bits per value
			val1, err := b.readUint32BE()
			if err != nil {
				return val, err
			}
			val2, err := b.readUint32BE()
			if err != nil {
				return val, err
			}
			xSign := uint32((flags >> 2) & 1)
			ySign := uint32((flags >> 1) & 1)
			zSign := uint32((flags >> 0) & 1)

			v22 := (xSign << 21) | (val1 >> 11)
			v25 := int32(v22<<10) >> 10
			v26 := int32(((val2>>21)|(((ySign<<11)|val1&0x7FF)<<10))<<10) >> 10
			v27 := val2&0x1FFFFF | (zSign << 31 >> 10)
			val.X = float32(float32(v25)*0.00000047683739) * scale
			val.Y = float32(float32(v26)*0.00000047683739) * scale
			val.Y = float32(float32(v27)*0.00000047683739) * scale
		} else { // when scale is in range 0x10 < scale <= 0x400
			// 16 bits per value (and sign), precision ranges from 1/4096 to 1/128
			// 2^16 = 65536, so values range from -16~16 to -512~512
			x, err := b.readUint16BE()
			if err != nil {
				return val, err
			}
			y, err := b.readUint16BE()
			if err != nil {
				return val, err
			}
			z, err := b.readUint16BE()
			if err != nil {
				return val, err
			}
			xSign := uint32((flags >> 2) & 1)
			ySign := uint32((flags >> 1) & 1)
			zSign := uint32((flags >> 0) & 1)
			rx := uint32(x)
			ry := uint32(y)
			rz := uint32(z)

			v20 := int32((ry|(ySign<<16))<<15) >> 15
			v21 := int32(((rz)|(zSign<<16))<<15) >> 15 // for some reason IDA doesn't show this as signed, despite that fact that it's created by a sar instruction!
			val.X = float32(float32(int32((rx|(xSign<<16))<<15)>>15) * 0.000015259022 * scale)
			val.Y = float32(float32(v20) * 0.000015259022 * scale)
			val.Z = float32(float32(v21) * 0.000015259022 * scale)
		}
	} else { // when scale is in range 0 <= scale <= 0x10
		val1, err := b.readUint32BE()
		if err != nil {
			return val, err
		}
		xSign := uint32((flags >> 2) & 1)
		ySign := uint32((flags >> 1) & 1)
		zSign := uint32((flags >> 0) & 1)

		res := (int32(ySign<<31) >> 21) | int32((val1>>10)&0x3FF)
		v17 := int32(val1&0x3FF) | (int32(zSign<<31) >> 21)
		val.X = float32(float32(int32(((xSign<<10)|(val1>>20))<<21)>>21)*0.00097751711) * scale
		val.Y = float32(float32(res)*0.00097751711) * scale
		val.Z = float32(float32(v17)*0.00097751711) * scale
	}
	return val, err
}

func (b *BitstreamReader) readMatrixMode0() ([9]float32, error) {
	var val [9]float32
	var err error

	q := [4]float32{}
	q[3], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	q[0], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	q[1], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}
	q[2], err = b.readFloat32BE()
	if err != nil {
		return val, err
	}

	return quaternionToRotMatrix(q), nil
}

func (b *BitstreamReader) readMatrixMode1() ([9]float32, error) {
	q := [4]float32{}
	var val [9]float32
	invertW, err := b.readBool()
	if err != nil {
		return val, err
	}
	invertX, err := b.readBool()
	if err != nil {
		return val, err
	}
	invertY, err := b.readBool()
	if err != nil {
		return val, err
	}
	invertZ, err := b.readBool()
	if err != nil {
		return val, err
	}
	x, err := b.readUint16LE()
	if err != nil {
		return val, err
	}
	y, err := b.readUint16LE()
	if err != nil {
		return val, err
	}
	z, err := b.readUint16LE()
	if err != nil {
		return val, err
	}
	xs := float32(x) / 65535.0
	ys := float32(y) / 65535.0
	zs := float32(z) / 65535.0
	if invertX {
		xs = -xs
	}
	if invertY {
		ys = -ys
	}
	if invertZ {
		zs = -zs
	}
	w := float32(math.Sqrt(math.Max(0.0, float64(1.0-xs-ys-zs))))
	if invertW {
		w = -w
	}
	q = [4]float32{xs, ys, zs, w}
	return quaternionToRotMatrix(q), nil
}
func (b *BitstreamReader) readMatrixMode2() ([9]float32, error) {
	return b.readMatrixMode1()
}

var quaternionIndices = [4][3]int{
	// the index is the number that is omitted
	[3]int{1, 2, 3}, // index 0
	[3]int{0, 2, 3}, // index 1
	[3]int{0, 1, 3}, // index 2
	[3]int{0, 1, 2}, // index 3
}

func (b *BitstreamReader) readPhysicsMatrix() ([9]float32, error) {
	var quat [4]float32
	var err error

	val1, err := b.readUint16BE()
	if err != nil {
		return [9]float32{}, err
	}
	val2, err := b.readUint32BE()
	if err != nil {
		return [9]float32{}, err
	}
	mode := val2 >> 0x1E
	indexSet := quaternionIndices[mode]

	//		16 bits for val 1
	// +	2 bits for mode from val 2
	// +	17 bits from val 2
	// +	17 bits from val 2
	//		52 bits
	xmm4 := float32(int32(uint32(val1)<<17) >> 17)
	xmm4 = xmm4 / (16383.0 * math.Sqrt2)
	xmm3 := float32(int32(val2<<2) >> 17)
	xmm3 = xmm3 / (16383.0 * math.Sqrt2)
	xmm2 := float32(int32(val2<<17) >> 17)
	xmm2 = xmm2 / (16383.0 * math.Sqrt2)

	quat[indexSet[0]] = xmm4
	quat[indexSet[1]] = xmm3
	quat[indexSet[2]] = xmm2
	quat[mode] = float32(math.Sqrt(math.Max(0.0, float64(1.0-xmm4*xmm4-xmm3*xmm3-xmm2*xmm2))))

	return quaternionToRotMatrix(quat), err
}

func (b *BitstreamReader) readPhysicsCFrame() (rbxfile.ValueCFrame, error) {
	var val rbxfile.ValueCFrame
	coords, err := b.readPhysicsCoords()
	if err != nil {
		return val, err
	}
	matrix, err := b.readPhysicsMatrix()
	if err != nil {
		return val, err
	}
	return rbxfile.ValueCFrame{coords, matrix}, nil
}

func (b *BitstreamReader) readPhysicsMotor() (PhysicsMotor, error) {
	var motor PhysicsMotor
	flags, err := b.readUint8()
	if err != nil {
		return motor, err
	}

	if flags&1 == 1 {
		motor.Position, err = b.readPhysicsCoords()
		if err != nil {
			return motor, err
		}
	}

	if flags&2 == 2 {
		var quat [4]float32

		val1, err := b.readUint32BE()
		if err != nil {
			return motor, err
		}
		mode := val1 >> 0x1E
		indexSet := quaternionIndices[mode]

		xmm4 := float32(int32(uint32(val1)<<2) >> 22)
		xmm4 = xmm4 / (511.0 * math.Sqrt2)
		xmm3 := float32(int32(val1<<12) >> 22)
		xmm3 = xmm3 / (511.0 * math.Sqrt2)
		xmm2 := float32(int32(val1<<22) >> 22)
		xmm2 = xmm2 / (511.0 * math.Sqrt2)

		quat[indexSet[0]] = xmm4
		quat[indexSet[1]] = xmm3
		quat[indexSet[2]] = xmm2
		quat[mode] = float32(math.Sqrt(math.Max(0.0, float64(1.0-xmm4*xmm4-xmm3*xmm3-xmm2*xmm2))))

		motor.Rotation = quaternionToRotMatrix(quat)
	}

	return motor, nil
}

func (b *BitstreamReader) readMotors() ([]PhysicsMotor, error) {
	countMotors, err := b.readVarint64()
	if err != nil {
		return nil, err
	}
	//println("reading", countMotors, "motors")
	if countMotors > 0x10000 {
		return nil, errors.New("numMotors is excessive")
	}

	motors := make([]PhysicsMotor, countMotors)
	var i uint64
	for i = 0; i < countMotors; i++ {
		motors[i], err = b.readPhysicsMotor()
		if err != nil {
			return motors, err
		}
	}
	return motors, nil
}

func (b *BitstreamReader) readPhysicsVelocity() (rbxfile.ValueVector3, error) {
	var val rbxfile.ValueVector3
	flags, err := b.readUint8()
	if err != nil || flags == 0 {
		return val, err
	}
	exp := flags>>4 - 1
	scale := float32(math.Exp2(float64(exp)))
	val1 := flags & 0xF
	val2, err := b.readUint32BE()
	if err != nil {
		return val, err
	}
	res := int32(val2<<12) >> 20
	v9 := int32(val1) | (int32(val2<<24) >> 20)
	val.X = float32(int32(val2)>>20) * 0.00048851978 * scale
	val.Y = float32(res) * 0.00048851978 * scale
	val.Z = float32(v9) * 0.00048851978 * scale

	return val, err
}

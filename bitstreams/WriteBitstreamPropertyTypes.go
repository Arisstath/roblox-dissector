package bitstreams

import "math"
import "github.com/gskartwii/rbxfile"
import "github.com/gskartwii/roblox-dissector/util"
import "github.com/gskartwii/roblox-dissector/schema"
import "errors"
import "strconv"
import "net"

func (b *BitstreamWriter) WriteUDim(val rbxfile.ValueUDim) error {
	err := b.WriteFloat32BE(val.Scale)
	if err != nil {
		return err
	}
	return b.WriteUint32BE(uint32(val.Offset))
}
func (b *BitstreamWriter) WriteUDim2(val rbxfile.ValueUDim2) error {
	err := b.WriteUDim(val.X)
	if err != nil {
		return err
	}
	return b.WriteUDim(val.Y)
}

func (b *BitstreamWriter) WriteRay(val rbxfile.ValueRay) error {
	err := b.WriteVector3Simple(val.Origin)
	if err != nil {
		return err
	}
	return b.WriteVector3Simple(val.Direction)
}

func (b *BitstreamWriter) WriteRegion3(val rbxfile.ValueRegion3) error {
	err := b.WriteVector3Simple(val.Start)
	if err != nil {
		return err
	}
	return b.WriteVector3Simple(val.End)
}

func (b *BitstreamWriter) WriteRegion3int16(val rbxfile.ValueRegion3int16) error {
	err := b.WriteVector3int16(val.Start)
	if err != nil {
		return err
	}
	return b.WriteVector3int16(val.End)
}

func (b *BitstreamWriter) WriteColor3(val rbxfile.ValueColor3) error {
	err := b.WriteFloat32BE(val.R)
	if err != nil {
		return err
	}
	err = b.WriteFloat32BE(val.G)
	if err != nil {
		return err
	}
	return b.WriteFloat32BE(val.B)
}
func (b *BitstreamWriter) WriteColor3uint8(val rbxfile.ValueColor3uint8) error {
	err := b.WriteByte(val.R)
	if err != nil {
		return err
	}
	err = b.WriteByte(val.G)
	if err != nil {
		return err
	}
	return b.WriteByte(val.B)
}
func (b *BitstreamWriter) WriteVector2(val rbxfile.ValueVector2) error {
	err := b.WriteFloat32BE(val.X)
	if err != nil {
		return err
	}
	return b.WriteFloat32BE(val.Y)
}
func (b *BitstreamWriter) WriteVector3Simple(val rbxfile.ValueVector3) error {
	err := b.WriteFloat32BE(val.X)
	if err != nil {
		return err
	}
	err = b.WriteFloat32BE(val.Y)
	if err != nil {
		return err
	}
	return b.WriteFloat32BE(val.Z)
}
func (b *BitstreamWriter) WriteVector3(val rbxfile.ValueVector3) error {
	var err error
	if math.Mod(float64(val.X), 0.5) != 0 ||
		math.Mod(float64(val.Y), 0.1) != 0 ||
		math.Mod(float64(val.Z), 0.5) != 0 ||
		val.X > 511.5 ||
		val.X < -511.5 ||
		val.Y > 204.7 ||
		val.Y < 0 ||
		val.Z > 511.5 ||
		val.Z < -511.5 {
		err = b.WriteBoolByte(false)
		if err != nil {
			return err
		}
		err = b.WriteVector3Simple(val)
		return err
	}
	err = b.WriteBoolByte(true)
	if err != nil {
		return err
	}
	xScaled := uint16(val.X * 2)
	yScaled := uint16(val.Y * 10)
	zScaled := uint16(val.Z * 2)
	xScaled = (xScaled >> 8 & 7) | ((xScaled & 0xFF) << 3)
	yScaled = (yScaled >> 8 & 7) | ((yScaled & 0xFF) << 3)
	zScaled = (zScaled >> 8 & 7) | ((zScaled & 0xFF) << 3)
	err = b.WriteBits(uint64(xScaled), 11)
	if err != nil {
		return err
	}
	err = b.WriteBits(uint64(yScaled), 11)
	if err != nil {
		return err
	}
	err = b.WriteBits(uint64(zScaled), 11)
	return err
}

func (b *BitstreamWriter) WriteVector2int16(val rbxfile.ValueVector2int16) error {
	err := b.WriteUint16BE(uint16(val.X))
	if err != nil {
		return err
	}
	return b.WriteUint16BE(uint16(val.Y))
}
func (b *BitstreamWriter) WriteVector3int16(val rbxfile.ValueVector3int16) error {
	err := b.WriteUint16BE(uint16(val.X))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE(uint16(val.Y))
	if err != nil {
		return err
	}
	return b.WriteUint16BE(uint16(val.Z))
}

func (b *BitstreamWriter) WriteAxes(val rbxfile.ValueAxes) error {
	write := 0
	if val.X {
		write |= 4
	}
	if val.Y {
		write |= 2
	}
	if val.Z {
		write |= 1
	}
	return b.WriteUint32BE(uint32(write))
}
func (b *BitstreamWriter) WriteFaces(val rbxfile.ValueFaces) error {
	write := 0
	if val.Right {
		write |= 32
	}
	if val.Top {
		write |= 16
	}
	if val.Back {
		write |= 8
	}
	if val.Left {
		write |= 4
	}
	if val.Bottom {
		write |= 2
	}
	if val.Front {
		write |= 1
	}
	return b.WriteUint32BE(uint32(write))
}

func (b *BitstreamWriter) WriteBrickColor(val rbxfile.ValueBrickColor) error {
	return b.WriteUint16BE(uint16(val))
}

func (b *BitstreamWriter) WriteVarLengthString(val string) error {
	err := b.WriteUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	return b.WriteASCII(val)
}

func rotMatrixToQuaternion(r [9]float32) [4]float32 {
	q := float32(math.Sqrt(float64(1+r[0*3+0]+r[1*3+1]+r[2*3+2])) / 2)
	return [4]float32{
		(r[2*3+1] - r[1*3+2]) / (4 * q),
		(r[0*3+2] - r[2*3+0]) / (4 * q),
		(r[1*3+0] - r[0*3+1]) / (4 * q),
		q,
	}
} // So nice to not have to worry about normalization on this side!
func (b *BitstreamWriter) WriteCFrame(val rbxfile.ValueCFrame) error {
	err := b.WriteVector3Simple(val.Position)
	if err != nil {
		return err
	}
	err = b.WriteBoolByte(false) // Not going to bother with lookup stuff
	if err != nil {
		return err
	}

	quat := rotMatrixToQuaternion(val.Rotation)
	b.WriteFloat16BE(quat[3], -1.0, 1.0)
	for i := 0; i < 3; i++ {
		err = b.WriteFloat16BE(quat[i], -1.0, 1.0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BitstreamWriter) WriteSintUTF8(val int32) error {
	return b.WriteUintUTF8(uint32(val)<<1 ^ -(uint32(val) >> 31))
}
func (b *BitstreamWriter) WriteNewPSint(val rbxfile.ValueInt) error {
	return b.WriteSintUTF8(int32(val))
}
func (b *BitstreamWriter) WriteVarint64(value uint64) error {
	if value == 0 {
		return b.WriteByte(0)
	}
	for value != 0 {
		nextValue := value >> 7
		if nextValue != 0 {
			err := b.WriteByte(byte(value&0x7F | 0x80))
			if err != nil {
				return err
			}
		} else {
			err := b.WriteByte(byte(value & 0x7F))
			if err != nil {
				return err
			}
		}
		value = nextValue
	}
	return nil
}
func (b *BitstreamWriter) WriteVarsint64(val int64) error {
	return b.WriteVarint64(uint64(val)<<1 ^ -(uint64(val) >> 63))
}

var typeToNetworkConvTable = map[rbxfile.Type]uint8{
	rbxfile.TypeString:                 schema.PROP_TYPE_STRING,
	rbxfile.TypeBinaryString:           schema.PROP_TYPE_BINARYSTRING,
	rbxfile.TypeProtectedString:        schema.PROP_TYPE_PROTECTEDSTRING_0,
	rbxfile.TypeContent:                schema.PROP_TYPE_CONTENT,
	rbxfile.TypeBool:                   schema.PROP_TYPE_PBOOL,
	rbxfile.TypeInt:                    schema.PROP_TYPE_PSINT,
	rbxfile.TypeFloat:                  schema.PROP_TYPE_PFLOAT,
	rbxfile.TypeDouble:                 schema.PROP_TYPE_PDOUBLE,
	rbxfile.TypeUDim:                   schema.PROP_TYPE_UDIM,
	rbxfile.TypeUDim2:                  schema.PROP_TYPE_UDIM2,
	rbxfile.TypeRay:                    schema.PROP_TYPE_RAY,
	rbxfile.TypeFaces:                  schema.PROP_TYPE_FACES,
	rbxfile.TypeAxes:                   schema.PROP_TYPE_AXES,
	rbxfile.TypeBrickColor:             schema.PROP_TYPE_BRICKCOLOR,
	rbxfile.TypeColor3:                 schema.PROP_TYPE_COLOR3,
	rbxfile.TypeVector2:                schema.PROP_TYPE_VECTOR2,
	rbxfile.TypeVector3:                schema.PROP_TYPE_VECTOR3_COMPLICATED,
	rbxfile.TypeCFrame:                 schema.PROP_TYPE_CFRAME_COMPLICATED,
	rbxfile.TypeToken:                  schema.PROP_TYPE_ENUM,
	rbxfile.TypeReference:              schema.PROP_TYPE_INSTANCE,
	rbxfile.TypeVector3int16:           schema.PROP_TYPE_VECTOR3UINT16,
	rbxfile.TypeVector2int16:           schema.PROP_TYPE_VECTOR2UINT16,
	rbxfile.TypeNumberSequence:         schema.PROP_TYPE_NUMBERSEQUENCE,
	rbxfile.TypeColorSequence:          schema.PROP_TYPE_COLORSEQUENCE,
	rbxfile.TypeNumberRange:            schema.PROP_TYPE_NUMBERRANGE,
	rbxfile.TypeRect2D:                 schema.PROP_TYPE_RECT2D,
	rbxfile.TypePhysicalProperties:     schema.PROP_TYPE_PHYSICALPROPERTIES,
	rbxfile.TypeColor3uint8:            schema.PROP_TYPE_COLOR3UINT8,
	rbxfile.TypeNumberSequenceKeypoint: schema.PROP_TYPE_NUMBERSEQUENCEKEYPOINT,
	rbxfile.TypeColorSequenceKeypoint:  schema.PROP_TYPE_COLORSEQUENCEKEYPOINT,
	rbxfile.TypeSystemAddress:          schema.PROP_TYPE_SYSTEMADDRESS,
	rbxfile.TypeMap:                    schema.PROP_TYPE_MAP,
	rbxfile.TypeDictionary:             schema.PROP_TYPE_DICTIONARY,
	rbxfile.TypeArray:                  schema.PROP_TYPE_ARRAY,
	rbxfile.TypeTuple:                  schema.PROP_TYPE_TUPLE,
}

func typeToNetwork(val rbxfile.Value) uint8 {
	if val == nil {
		return 0
	}
	return typeToNetworkConvTable[val.Type()]
}
func (b *BitstreamWriter) WriteSerializedValueGeneric(val rbxfile.Value, valueType uint8) error {
	if val == nil {
		return nil
	}
	var err error
	switch valueType {
	case schema.PROP_TYPE_ENUM:
		err = b.WriteNewEnumValue(val.(rbxfile.ValueToken))
	case schema.PROP_TYPE_BINARYSTRING:
		err = b.WriteNewBinaryString(val.(rbxfile.ValueBinaryString))
	case schema.PROP_TYPE_PBOOL:
		err = b.WritePBool(val.(rbxfile.ValueBool))
	case schema.PROP_TYPE_PSINT:
		err = b.WriteNewPSint(val.(rbxfile.ValueInt))
	case schema.PROP_TYPE_PFLOAT:
		err = b.WritePFloat(val.(rbxfile.ValueFloat))
	case schema.PROP_TYPE_PDOUBLE:
		err = b.WritePDouble(val.(rbxfile.ValueDouble))
	case schema.PROP_TYPE_UDIM:
		err = b.WriteUDim(val.(rbxfile.ValueUDim))
	case schema.PROP_TYPE_UDIM2:
		err = b.WriteUDim2(val.(rbxfile.ValueUDim2))
	case schema.PROP_TYPE_RAY:
		err = b.WriteRay(val.(rbxfile.ValueRay))
	case schema.PROP_TYPE_FACES:
		err = b.WriteFaces(val.(rbxfile.ValueFaces))
	case schema.PROP_TYPE_AXES:
		err = b.WriteAxes(val.(rbxfile.ValueAxes))
	case schema.PROP_TYPE_BRICKCOLOR:
		err = b.WriteBrickColor(val.(rbxfile.ValueBrickColor))
	case schema.PROP_TYPE_COLOR3:
		err = b.WriteColor3(val.(rbxfile.ValueColor3))
	case schema.PROP_TYPE_COLOR3UINT8:
		err = b.WriteColor3uint8(val.(rbxfile.ValueColor3uint8))
	case schema.PROP_TYPE_VECTOR2:
		err = b.WriteVector2(val.(rbxfile.ValueVector2))
	case schema.PROP_TYPE_VECTOR3_SIMPLE:
		err = b.WriteVector3Simple(val.(rbxfile.ValueVector3))
	case schema.PROP_TYPE_VECTOR3_COMPLICATED:
		err = b.WriteVector3(val.(rbxfile.ValueVector3))
	case schema.PROP_TYPE_VECTOR2UINT16:
		err = b.WriteVector2int16(val.(rbxfile.ValueVector2int16))
	case schema.PROP_TYPE_VECTOR3UINT16:
		err = b.WriteVector3int16(val.(rbxfile.ValueVector3int16))
	case schema.PROP_TYPE_CFRAME_SIMPLE:
		err = b.WriteCFrameSimple(val.(rbxfile.ValueCFrame))
	case schema.PROP_TYPE_CFRAME_COMPLICATED:
		err = b.WriteCFrame(val.(rbxfile.ValueCFrame))
	case schema.PROP_TYPE_NUMBERSEQUENCE:
		err = b.WriteNumberSequence(val.(rbxfile.ValueNumberSequence))
	case schema.PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		err = b.WriteNumberSequenceKeypoint(val.(rbxfile.ValueNumberSequenceKeypoint))
	case schema.PROP_TYPE_NUMBERRANGE:
		err = b.WriteNumberRange(val.(rbxfile.ValueNumberRange))
	case schema.PROP_TYPE_COLORSEQUENCE:
		err = b.WriteColorSequence(val.(rbxfile.ValueColorSequence))
	case schema.PROP_TYPE_COLORSEQUENCEKEYPOINT:
		err = b.WriteColorSequenceKeypoint(val.(rbxfile.ValueColorSequenceKeypoint))
	case schema.PROP_TYPE_RECT2D:
		err = b.WriteRect2D(val.(rbxfile.ValueRect2D))
	case schema.PROP_TYPE_PHYSICALPROPERTIES:
		err = b.WritePhysicalProperties(val.(rbxfile.ValuePhysicalProperties))
	case schema.PROP_TYPE_INT64:
		err = b.WriteVarsint64(int64(val.(rbxfile.ValueInt64)))
	case schema.PROP_TYPE_STRING_NO_CACHE:
		err = b.WritePStringNoCache(val.(rbxfile.ValueString))
	default:
		return errors.New("Unsupported property type: " + strconv.Itoa(int(valueType)))
	}
	return err
}

func (b *BitstreamWriter) WriteNewTypeAndValue(val rbxfile.Value, writer util.PacketWriter) error {
	var err error
	valueType := typeToNetwork(val)
	//println("Writing typeandvalue", valueType)
	err = b.WriteByte(uint8(valueType))
	if valueType == 7 {
		err = b.WriteUint16BE(val.(rbxfile.ValueToken).ID)
		if err != nil {
			return err
		}
	}
	return b.WriteSerializedValue(val, writer, valueType)
}

func (b *BitstreamWriter) WriteNewTuple(val rbxfile.ValueTuple, writer util.PacketWriter) error {
	err := b.WriteUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.WriteNewTypeAndValue(val[i], writer)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BitstreamWriter) WriteNewArray(val rbxfile.ValueArray, writer util.PacketWriter) error {
	return b.WriteNewTuple(rbxfile.ValueTuple(val), writer)
}

func (b *BitstreamWriter) WriteNewDictionary(val rbxfile.ValueDictionary, writer util.PacketWriter) error {
	err := b.WriteUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for key, value := range val {
		err = b.WriteUintUTF8(uint32(len(key)))
		if err != nil {
			return err
		}
		err = b.WriteASCII(key)
		if err != nil {
			return err
		}
		err = b.WriteNewTypeAndValue(value, writer)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BitstreamWriter) WriteNewMap(val rbxfile.ValueMap, writer util.PacketWriter) error {
	return b.WriteNewDictionary(rbxfile.ValueDictionary(val), writer)
}

func (b *BitstreamWriter) WriteNumberSequenceKeypoint(val rbxfile.ValueNumberSequenceKeypoint) error {
	err := b.WriteFloat32BE(val.Time)
	if err != nil {
		return err
	}
	err = b.WriteFloat32BE(val.Value)
	if err != nil {
		return err
	}
	err = b.WriteFloat32BE(val.Envelope)
	return err
}
func (b *BitstreamWriter) WriteNumberSequence(val rbxfile.ValueNumberSequence) error {
	err := b.WriteUint32BE(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.WriteNumberSequenceKeypoint(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BitstreamWriter) WriteNumberRange(val rbxfile.ValueNumberRange) error {
	err := b.WriteFloat32BE(val.Min)
	if err != nil {
		return err
	}
	return b.WriteFloat32BE(val.Max)
}

func (b *BitstreamWriter) WriteColorSequenceKeypoint(val rbxfile.ValueColorSequenceKeypoint) error {
	err := b.WriteFloat32BE(val.Time)
	if err != nil {
		return err
	}
	err = b.WriteColor3(val.Value)
	if err != nil {
		return err
	}
	return b.WriteFloat32BE(val.Envelope)
}
func (b *BitstreamWriter) WriteColorSequence(val rbxfile.ValueColorSequence) error {
	err := b.WriteUint32BE(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.WriteColorSequenceKeypoint(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BitstreamWriter) WriteNewEnumValue(val rbxfile.ValueToken) error {
	return b.WriteUintUTF8(val.Value)
}

func (b *BitstreamWriter) WriteSystemAddressRaw(val rbxfile.ValueSystemAddress) error {
	addr, err := net.ResolveUDPAddr("udp", string(val))
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		addr.IP[i] = addr.IP[i] ^ 0xFF // bitwise NOT
	}

	err = b.WriteN(addr.IP, 4)
	if err != nil {
		return err
	}

	// in case the value will be used again
	for i := 0; i < 4; i++ {
		addr.IP[i] = addr.IP[i] ^ 0xFF
	}

	return b.WriteUint16BE(uint16(addr.Port))
}

func (b *BitstreamWriter) WriteSystemAddress(val rbxfile.ValueSystemAddress, caches *util.Caches) error {
	return b.WriteCachedSystemAddress(val, caches)
}
func (b *JoinSerializeWriter) WriteSystemAddress(val rbxfile.ValueSystemAddress) error {
	return b.WriteSystemAddressRaw(val)
}

func (b *BitstreamWriter) WriteRect2D(val rbxfile.ValueRect2D) error {
	err := b.WriteFloat32BE(val.Min.X)
	if err != nil {
		return err
	}
	err = b.WriteFloat32BE(val.Min.Y)
	if err != nil {
		return err
	}
	err = b.WriteFloat32BE(val.Max.X)
	if err != nil {
		return err
	}
	err = b.WriteFloat32BE(val.Max.Y)
	return err
}

func (b *BitstreamWriter) WritePhysicalProperties(val rbxfile.ValuePhysicalProperties) error {
	err := b.WriteBoolByte(val.CustomPhysics)
	if err != nil {
		return err
	}
	if val.CustomPhysics {
		err := b.WriteFloat32BE(val.Density)
		if err != nil {
			return err
		}
		err = b.WriteFloat32BE(val.Friction)
		if err != nil {
			return err
		}
		err = b.WriteFloat32BE(val.Elasticity)
		if err != nil {
			return err
		}
		err = b.WriteFloat32BE(val.FrictionWeight)
		if err != nil {
			return err
		}
		err = b.WriteFloat32BE(val.ElasticityWeight)
	}
	return err
}

func (b *BitstreamWriter) WriteCoordsMode0(val rbxfile.ValueVector3) error {
	return b.WriteVector3Simple(val)
}
func (b *BitstreamWriter) WriteCoordsMode1(val rbxfile.ValueVector3) error {
	valRange := float32(math.Max(math.Max(math.Abs(float64(val.X)), math.Abs(float64(val.Y))), math.Abs(float64(val.Z))))
	err := b.WriteFloat32BE(valRange)
	if err != nil {
		return err
	}
	if valRange <= 0.0000099999997 {
		return nil
	}
	err = b.WriteUint16BE(uint16(val.X/valRange*32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE(uint16(val.Y/valRange*32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE(uint16(val.Z/valRange*32767.0 + 32767.0))
	return err
}
func (b *BitstreamWriter) WriteCoordsMode2(val rbxfile.ValueVector3) error {
	xShort := uint16((val.X + 1024.0) * 16.0)
	yShort := uint16((val.Y + 1024.0) * 16.0)
	zShort := uint16((val.Z + 1024.0) * 16.0)

	err := b.WriteUint16BE((xShort&0x7F)<<7 | (xShort >> 8))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE((yShort&0x7F)<<7 | (yShort >> 8))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE((zShort&0x7F)<<7 | (zShort >> 8))
	return err
}

func (b *BitstreamWriter) WritePhysicsCoords(val rbxfile.ValueVector3) error {
	var xModifier, yModifier, zModifier float32
	var err error
	xAbs := math.Abs(float64(val.X))
	yAbs := math.Abs(float64(val.Y))
	zAbs := math.Abs(float64(val.Z))
	largest := xAbs
	if yAbs > xAbs {
		largest = yAbs
	}
	if zAbs > largest {
		largest = zAbs
	}

	_, exp := math.Frexp(largest)
	if exp < 0 {
		exp = 0
	} else if exp > 31 {
		exp = 31
	}

	scale := float32(math.Exp2(float64(exp)))

	xScale := float32(-1.0)
	yScale := float32(-1.0)
	zScale := float32(-1.0)

	if val.X/scale > -1.0 {
		xScale = val.X / scale
	}
	if val.Y/scale > -1.0 {
		yScale = val.Y / scale
	}
	if val.Y/scale > -1.0 {
		zScale = val.Z / scale
	}

	if xScale > 1.0 {
		xScale = 1.0
	}
	if yScale > 1.0 {
		yScale = 1.0
	}
	if zScale > 1.0 {
		zScale = 1.0
	}

	xModifier = -0.5
	yModifier = -0.5
	zModifier = -0.5
	if val.X > 0.0 {
		xModifier = 0.5
	}
	if val.Y > 0.0 {
		yModifier = 0.5
	}
	if val.Z > 0.0 {
		zModifier = 0.5
	}

	if exp <= 4.0 {
		xScale *= 1023.0
		yScale *= 1023.0
		zScale *= 1023.0

		xScale += xModifier
		yScale += yModifier
		zScale += zModifier

		xScaleInt := int32(xScale)
		yScaleInt := int32(yScale)
		zScaleInt := int32(zScale)

		xSign := xScaleInt >> 10 & 1
		ySign := yScaleInt >> 10 & 1
		zSign := zScaleInt >> 10 & 1

		var header = uint8(exp << 3)
		header |= uint8(xSign << 2)
		header |= uint8(ySign << 1)
		header |= uint8(zSign << 0)
		err = b.WriteByte(header)
		if err != nil {
			return err
		}

		var val1 uint32
		val1 |= uint32(xScaleInt<<20) & 0x3FF00000
		val1 |= uint32(yScaleInt&0x3FF) << 10
		val1 |= uint32(zScaleInt) & 0x3FF

		err = b.WriteUint32BE(val1)
		if err != nil {
			return err
		}
	} else if exp > 10.0 {
		xScale *= 2097200.0
		yScale *= 2097200.0
		zScale *= 2097200.0

		xScale += xModifier
		yScale += yModifier
		zScale += zModifier

		xScaleInt := int32(xScale)
		yScaleInt := int32(yScale)
		zScaleInt := int32(zScale)

		xSign := xScaleInt >> 21 & 1
		ySign := yScaleInt >> 21 & 1
		zSign := zScaleInt >> 21 & 1

		var header = uint8(exp << 3)
		header |= uint8(xSign << 2)
		header |= uint8(ySign << 1)
		header |= uint8(zSign << 0)
		err = b.WriteByte(header)
		if err != nil {
			return err
		}

		var val1, val2 uint32
		val1 |= uint32(xScaleInt << 11)
		val1 |= uint32((yScaleInt >> 10) & 0x7FF)

		val2 |= uint32((yScaleInt << 21) & 0x7FE00000)
		val2 |= uint32(zScaleInt & 0x1FFFFF)

		err = b.WriteUint32BE(val1)
		if err != nil {
			return err
		}
		err = b.WriteUint32BE(val2)
		if err != nil {
			return err
		}
	} else {
		xScale *= 65535.0
		yScale *= 65535.0
		zScale *= 65535.0

		xScale += xModifier
		yScale += yModifier
		zScale += zModifier

		xScaleInt := int32(xScale)
		yScaleInt := int32(yScale)
		zScaleInt := int32(zScale)

		xSign := xScaleInt >> 16 & 1
		ySign := yScaleInt >> 16 & 1
		zSign := zScaleInt >> 16 & 1

		var header = uint8(exp << 3)
		header |= uint8(xSign << 2)
		header |= uint8(ySign << 1)
		header |= uint8(zSign << 0)
		err = b.WriteByte(header)
		if err != nil {
			return err
		}

		err = b.WriteUint16BE(uint16(xScaleInt))
		if err != nil {
			return err
		}
		err = b.WriteUint16BE(uint16(yScaleInt))
		if err != nil {
			return err
		}
		err = b.WriteUint16BE(uint16(zScaleInt))
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BitstreamWriter) WriteMatrixMode0(val [9]float32) error {
	var err error
	q := rotMatrixToQuaternion(val)
	b.WriteFloat32BE(q[3])
	for i := 0; i < 3; i++ {
		err = b.WriteFloat32BE(q[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BitstreamWriter) WriteMatrixMode1(val [9]float32) error {
	q := rotMatrixToQuaternion(val)
	err := b.WriteBoolByte(q[3] < 0) // sqrt doesn't return negative numbers
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		err = b.WriteBoolByte(q[i] < 0)
		if err != nil {
			return err
		}
	}
	for i := 0; i < 3; i++ {
		err = b.WriteUint16LE(uint16(math.Abs(float64(q[i]))))
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *BitstreamWriter) WriteMatrixMode2(val [9]float32) error {
	return b.WriteMatrixMode1(val)
}
func (b *BitstreamWriter) WritePhysicsMatrix(val [9]float32) error {
	var err error
	quat := rotMatrixToQuaternion(val)
	largestIndex := 0
	largest := math.Abs(float64(quat[0]))
	for i := 1; i < 4; i++ {
		if math.Abs(float64(quat[i])) > largest {
			largest = math.Abs(float64(quat[i]))
			largestIndex = i
		}
	}
	indexSet := quaternionIndices[largestIndex]
	norm := float32(math.Sqrt(float64(quat[0]*quat[0] + quat[1]*quat[1] + quat[2]*quat[2] + quat[3]*quat[3])))
	for i := 0; i < 4; i++ {
		quat[i] /= norm
	}
	if quat[largestIndex] < 0.0 {
		for i := 0; i < 4; i++ {
			quat[i] = -quat[i]
		}
	}

	val1 := quat[indexSet[0]] * math.Sqrt2 * 16383.0
	val2 := quat[indexSet[1]] * math.Sqrt2 * 16383.0
	val3 := quat[indexSet[2]] * math.Sqrt2 * 16383.0

	if quat[indexSet[0]] < 0.0 {
		val1 -= 0.5
	} else {
		val1 += 0.5
	}
	if quat[indexSet[1]] < 0.0 {
		val2 -= 0.5
	} else {
		val2 += 0.5
	}
	if quat[indexSet[2]] < 0.0 {
		val3 -= 0.5
	} else {
		val3 += 0.5
	}

	val1Int := int32(val1) & 0x7FFF
	val2Int := int32(val2) & 0x7FFF
	val3Int := int32(val3) & 0x7FFF

	err = b.WriteUint16BE(uint16(val1Int))
	if err != nil {
		return err
	}

	var val2Encoded uint32
	val2Encoded |= uint32(largestIndex << 30)
	val2Encoded |= uint32(val2Int << 15)
	val2Encoded |= uint32(val3Int << 0)

	err = b.WriteUint32BE(uint32(val2Encoded))
	return err
}
func (b *BitstreamWriter) WritePhysicsCFrame(val rbxfile.ValueCFrame) error {
	err := b.WritePhysicsCoords(val.Position)
	if err != nil {
		return err
	}
	return b.WritePhysicsMatrix(val.Rotation)
}

func (b *BitstreamWriter) WritePhysicsVelocity(val rbxfile.ValueVector3) error {
	var err error
	var xModifier, yModifier, zModifier, xScale, yScale, zScale float32
	xAbs := math.Abs(float64(val.X))
	yAbs := math.Abs(float64(val.Y))
	zAbs := math.Abs(float64(val.Z))
	largest := xAbs
	if yAbs > xAbs {
		largest = yAbs
	}
	if zAbs > largest {
		largest = zAbs
	}

	if largest < 0.001 {
		return b.WriteByte(0) // no velocity!
	}

	_, exp := math.Frexp(largest)
	if exp < 0 {
		exp = 0
	} else if exp > 14 {
		exp = 14
	}

	scale := float32(math.Exp2(float64(exp)))

	xScale = -1.0
	yScale = -1.0
	zScale = -1.0

	if val.X/scale > -1.0 {
		xScale = val.X / scale
	}
	if val.Y/scale > -1.0 {
		yScale = val.Y / scale
	}
	if val.Z/scale > -1.0 {
		zScale = val.Z / scale
	}

	if xScale > 1.0 {
		xScale = 1.0
	}
	if yScale > 1.0 {
		yScale = 1.0
	}
	if zScale > 1.0 {
		zScale = 1.0
	}

	xModifier = -0.5
	yModifier = -0.5
	zModifier = -0.5
	if val.X > 0.0 {
		xModifier = 0.5
	}
	if val.Y > 0.0 {
		yModifier = 0.5
	}
	if val.Z > 0.0 {
		zModifier = 0.5
	}

	xScale *= 2047.0
	yScale *= 2047.0
	zScale *= 2047.0

	xScale += xModifier
	yScale += yModifier
	zScale += zModifier

	xScaleInt := int32(xScale)
	yScaleInt := int32(yScale)
	zScaleInt := int32(zScale)

	var header uint8
	header |= uint8((exp + 1) << 4)
	header |= uint8(zScaleInt & 0xF)
	err = b.WriteByte(header)
	if err != nil {
		return err
	}

	var val1 uint32
	val1 |= uint32((zScaleInt >> 4) & 0xFF)
	val1 |= uint32(yScaleInt << 8)
	val1 |= uint32(xScaleInt << 20)

	err = b.WriteUint32BE(val1)
	return err
}

func (b *BitstreamWriter) WriteMotor(motor PhysicsMotor) error {
	hasCoords := false
	hasRotation := false
	norm := motor.Position.X*motor.Position.X + motor.Position.Y*motor.Position.Y + motor.Position.Z*motor.Position.Z
	// I don't understand the point of the following code, other than
	// norm != 0.0. Why do we need to check if v4 is less than normAbs?
	if norm != 0.0 {
		normAbs := math.Abs(float64(norm))
		normPlus1 := normAbs + 1.0
		v4 := 1.0 / 100000.0
		if !math.IsInf(normPlus1, 0) {
			v4 = normPlus1 / 100000.0
		}
		if v4 < normAbs {
			hasCoords = true
		}
	}

	motorRot := motor.Rotation
	trace := motorRot[0] + motorRot[4] + motorRot[8]
	traceCos := 0.5 * (trace - 1.0)
	angle := math.Acos(float64(traceCos))
	if angle != 0.0 {
		angleAbs := math.Abs(float64(angle))
		anglePlus1 := angleAbs + 1.0
		v7 := 1.0 / 100000.0
		if !math.IsInf(anglePlus1, 0) {
			v7 = anglePlus1 / 100000.0
		}
		if v7 < angleAbs {
			hasRotation = true
		}
	}

	var header uint8
	if hasCoords {
		header |= 1 << 0
	}
	if hasRotation {
		header |= 1 << 1
	}
	err := b.WriteByte(header)
	if err != nil {
		return err
	}

	if hasCoords {
		err = b.WritePhysicsCoords(motor.Position)
		if err != nil {
			return err
		}
	}

	if hasRotation {
		quat := rotMatrixToQuaternion(motor.Rotation)
		largestIndex := 0
		largest := math.Abs(float64(quat[0]))
		for i := 1; i < 4; i++ {
			if math.Abs(float64(quat[i])) > largest {
				largest = math.Abs(float64(quat[i]))
				largestIndex = i
			}
		}
		indexSet := quaternionIndices[largestIndex]
		rotationNorm := float32(math.Sqrt(float64(quat[0]*quat[0] + quat[1]*quat[1] + quat[2]*quat[2] + quat[3]*quat[3])))
		for i := 0; i < 4; i++ {
			quat[i] /= rotationNorm
		}
		if quat[largestIndex] < 0.0 {
			for i := 0; i < 4; i++ {
				quat[i] = -quat[i]
			}
		}

		val1 := quat[indexSet[0]] * math.Sqrt2 * 511.0
		val2 := quat[indexSet[1]] * math.Sqrt2 * 511.0
		val3 := quat[indexSet[2]] * math.Sqrt2 * 511.0

		if quat[indexSet[0]] < 0.0 {
			val1 -= 0.5
		} else {
			val1 += 0.5
		}
		if quat[indexSet[1]] < 0.0 {
			val2 -= 0.5
		} else {
			val2 += 0.5
		}
		if quat[indexSet[2]] < 0.0 {
			val3 -= 0.5
		} else {
			val3 += 0.5
		}

		val1Int := int32(val1) & 0x3FF
		val2Int := int32(val2) & 0x3FF
		val3Int := int32(val3) & 0x3FF

		var val1Encoded uint32
		val1Encoded |= uint32(largestIndex << 30)
		val1Encoded |= uint32(val1Int << 20)
		val1Encoded |= uint32(val2Int << 10)
		val1Encoded |= uint32(val3Int << 0)

		err = b.WriteUint32BE(uint32(val1Encoded))
		return err
	}
	return nil
}

func (b *BitstreamWriter) WriteMotors(val []PhysicsMotor) error {
	err := b.WriteUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.WriteMotor(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}

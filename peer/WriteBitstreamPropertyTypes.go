package peer
import "math"
import "github.com/gskartwii/rbxfile"
import "errors"
import "strconv"
import "net"

func (b *ExtendedWriter) WriteUDim(val rbxfile.ValueUDim) error {
	err := b.WriteFloat32BE(val.Scale)
	if err != nil {
		return err
	}
	return b.WriteUint32BE(uint32(val.Offset))
}
func (b *ExtendedWriter) WriteUDim2(val rbxfile.ValueUDim2) error {
	err := b.WriteUDim(val.X)
	if err != nil {
		return err
	}
	return b.WriteUDim(val.Y)
}

func (b *ExtendedWriter) WriteRay(val rbxfile.ValueRay) error {
	err := b.WriteVector3Simple(val.Origin)
	if err != nil {
		return err
	}
	return b.WriteVector3Simple(val.Direction)
}

func (b *ExtendedWriter) WriteRegion3(val rbxfile.ValueRegion3) error {
	err := b.WriteVector3Simple(val.Start)
	if err != nil {
		return err
	}
	return b.WriteVector3Simple(val.End)
}

func (b *ExtendedWriter) WriteRegion3int16(val rbxfile.ValueRegion3int16) error {
	err := b.WriteVector3int16(val.Start)
	if err != nil {
		return err
	}
	return b.WriteVector3int16(val.End)
}

func (b *ExtendedWriter) WriteColor3(val rbxfile.ValueColor3) error {
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
func (b *ExtendedWriter) WriteColor3uint8(val rbxfile.ValueColor3uint8) error {
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
func (b *ExtendedWriter) WriteVector2(val rbxfile.ValueVector2) error {
	err := b.WriteFloat32BE(val.X)
	if err != nil {
		return err
	}
	return b.WriteFloat32BE(val.Y)
}
func (b *ExtendedWriter) WriteVector3Simple(val rbxfile.ValueVector3) error {
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
func (b *ExtendedWriter) WriteVector3(val rbxfile.ValueVector3) error {
	if math.Mod(float64(val.X), 0.5) != 0 ||
	   math.Mod(float64(val.Y), 0.1) != 0 ||
	   math.Mod(float64(val.Z), 0.5) != 0 ||
	   val.X >  511.5	||
	   val.X < -511.5	||
	   val.Y >  204.7	||
	   val.Y <      0	||
	   val.Z >  511.5	||
	   val.Z < -511.5	{
		err := b.WriteBool(false)
		if err != nil {
			return err
		}
		err = b.WriteVector3Simple(val)
		return err
	} else {
		err := b.WriteBool(true)
		if err != nil {
			return err
		}
		x_scaled := uint16(val.X * 2)
		y_scaled := uint16(val.Y * 10)
		z_scaled := uint16(val.Z * 2)
		x_scaled = (x_scaled >> 8 & 7) | ((x_scaled & 0xFF) << 3)
		y_scaled = (y_scaled >> 8 & 7) | ((y_scaled & 0xFF) << 3)
		z_scaled = (z_scaled >> 8 & 7) | ((z_scaled & 0xFF) << 3)
		err = b.Bits(11, uint64(x_scaled))
		if err != nil {
			return err
		}
		err = b.Bits(11, uint64(y_scaled))
		if err != nil {
			return err
		}
		err = b.Bits(11, uint64(z_scaled))
		return err
	}
}

func (b *ExtendedWriter) WriteVector2int16(val rbxfile.ValueVector2int16) error {
	err := b.WriteUint16BE(uint16(val.X))
	if err != nil {
		return err
	}
	return b.WriteUint16BE(uint16(val.Y))
}
func (b *ExtendedWriter) WriteVector3int16(val rbxfile.ValueVector3int16) error {
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

func (b *ExtendedWriter) WritePBool(val rbxfile.ValueBool) error {
	return b.WriteBool(bool(val))
}
func (b *ExtendedWriter) WritePSint(val rbxfile.ValueInt) error {
	return b.WriteUint32BE(uint32(val))
}
func (b *ExtendedWriter) WritePFloat(val rbxfile.ValueFloat) error {
	return b.WriteFloat32BE(float32(val))
}
func (b *ExtendedWriter) WritePDouble(val rbxfile.ValueDouble) error {
	return b.WriteFloat64BE(float64(val))
}

func (b *ExtendedWriter) WriteAxes(val rbxfile.ValueAxes) error {
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
func (b *ExtendedWriter) WriteFaces(val rbxfile.ValueFaces) error {
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

func (b *ExtendedWriter) WriteBrickColor(val rbxfile.ValueBrickColor) error {
	return b.Bits(7, uint64(val))
}

func (b *ExtendedWriter) WriteNewPString(isClient bool, val rbxfile.ValueString, isJoinData bool, context *CommunicationContext) (error) {
	if !isJoinData {
		return b.WriteCached(isClient, string(val), context)
	}
	err := b.WriteUintUTF8(len(val))
	if err != nil {
		return err
	}
	return b.WriteASCII(string(val))
}

func (b *ExtendedWriter) WriteNewProtectedString(isClient bool, val rbxfile.ValueProtectedString, isJoinData bool, context *CommunicationContext) error {
	if !isJoinData {
		return b.WriteNewCachedProtectedString(isClient, string(val), context)
	}
	return b.WriteNewPString(isClient, rbxfile.ValueString(val), true, context)
}
func (b *ExtendedWriter) WriteNewBinaryString(val rbxfile.ValueBinaryString) error {
	return b.WriteNewPString(false, rbxfile.ValueString(val), true, nil) // hack: isClient doesn't matter because cache isn't used
}
func (b *ExtendedWriter) WriteNewContent(isClient bool, val rbxfile.ValueContent, isJoinData bool, context *CommunicationContext) error {
	return b.WriteNewPString(isClient, rbxfile.ValueString(val), isJoinData, context)
}

func (b *ExtendedWriter) WriteCFrameSimple(val rbxfile.ValueCFrame) error {
	return nil
}

func rotMatrixToQuaternion(r [9]float32) [4]float32 {
	q := float32(math.Sqrt(float64(1 + r[0*3+0] + r[1*3+1] + r[2*3+2]))/2)
	return [4]float32{
		(r[2*3+1]-r[1*3+2])/(4*q),
		(r[0*3+2]-r[2*3+0])/(4*q),
		(r[1*3+0]-r[0*3+1])/(4*q),
		q,
	}
} // So nice to not have to worry about normalization on this side!
func (b *ExtendedWriter) WriteCFrame(val rbxfile.ValueCFrame) error {
	err := b.WriteVector3(val.Position)
	if err != nil {
		return err
	}
	err = b.WriteBool(false) // Not going to bother with lookup stuff
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

func (b *ExtendedWriter) WriteSintUTF8(val int32) error {
	return b.WriteUintUTF8(int(uint32(val) << 1 ^ -(uint32(val) >> 31)))
}
func (b *ExtendedWriter) WriteNewPSint(val rbxfile.ValueInt) error {
	return b.WriteSintUTF8(int32(val))
}

var typeToNetworkConvTable = map[rbxfile.Type]uint8{
	rbxfile.TypeString: PROP_TYPE_STRING,
	rbxfile.TypeBinaryString: PROP_TYPE_BINARYSTRING,
	rbxfile.TypeProtectedString: PROP_TYPE_PROTECTEDSTRING_0,
	rbxfile.TypeContent: PROP_TYPE_CONTENT,
	rbxfile.TypeBool: PROP_TYPE_PBOOL,
	rbxfile.TypeInt: PROP_TYPE_PSINT,
	rbxfile.TypeFloat: PROP_TYPE_PFLOAT,
	rbxfile.TypeDouble: PROP_TYPE_PDOUBLE,
	rbxfile.TypeUDim: PROP_TYPE_UDIM,
	rbxfile.TypeUDim2: PROP_TYPE_UDIM2,
	rbxfile.TypeRay: PROP_TYPE_RAY,
	rbxfile.TypeFaces: PROP_TYPE_FACES,
	rbxfile.TypeAxes: PROP_TYPE_AXES,
	rbxfile.TypeBrickColor: PROP_TYPE_BRICKCOLOR,
	rbxfile.TypeColor3: PROP_TYPE_COLOR3,
	rbxfile.TypeVector2: PROP_TYPE_VECTOR2,
	rbxfile.TypeVector3: PROP_TYPE_VECTOR3_COMPLICATED,
	rbxfile.TypeCFrame: PROP_TYPE_CFRAME_COMPLICATED,
	rbxfile.TypeToken: PROP_TYPE_ENUM,
	rbxfile.TypeReference: PROP_TYPE_INSTANCE,
	rbxfile.TypeVector3int16: PROP_TYPE_VECTOR3UINT16,
	rbxfile.TypeVector2int16: PROP_TYPE_VECTOR2UINT16,
	rbxfile.TypeNumberSequence: PROP_TYPE_NUMBERSEQUENCE,
	rbxfile.TypeColorSequence: PROP_TYPE_COLORSEQUENCE,
	rbxfile.TypeNumberRange: PROP_TYPE_NUMBERRANGE,
	rbxfile.TypeRect2D: PROP_TYPE_RECT2D,
	rbxfile.TypePhysicalProperties: PROP_TYPE_PHYSICALPROPERTIES,
	rbxfile.TypeColor3uint8: PROP_TYPE_COLOR3UINT8,
	rbxfile.TypeNumberSequenceKeypoint: PROP_TYPE_NUMBERSEQUENCEKEYPOINT,
	rbxfile.TypeColorSequenceKeypoint: PROP_TYPE_COLORSEQUENCEKEYPOINT,
	rbxfile.TypeSystemAddress: PROP_TYPE_SYSTEMADDRESS,
	rbxfile.TypeMap: PROP_TYPE_MAP,
	rbxfile.TypeDictionary: PROP_TYPE_DICTIONARY,
	rbxfile.TypeArray: PROP_TYPE_ARRAY,
	rbxfile.TypeTuple: PROP_TYPE_TUPLE,
}
func typeToNetwork(val rbxfile.Value) uint8 {
	return typeToNetworkConvTable[val.Type()]
}
func (b *ExtendedWriter) writeSerializedValue(isClient bool, val rbxfile.Value, isJoinData bool, valueType uint8, context *CommunicationContext) error {
	var err error
	switch valueType {
	case PROP_TYPE_STRING:
		err = b.WriteNewPString(isClient, val.(rbxfile.ValueString), isJoinData, context)
	case PROP_TYPE_STRING_NO_CACHE:
		err = b.WriteNewPString(isClient, val.(rbxfile.ValueString), true, context)
	case PROP_TYPE_PROTECTEDSTRING_0:
		err = b.WriteNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_1:
		err = b.WriteNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_2:
		err = b.WriteNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_3:
		err = b.WriteNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_ENUM:
		err = b.WriteNewEnumValue(val.(rbxfile.ValueToken))
	case PROP_TYPE_BINARYSTRING:
		err = b.WriteNewBinaryString(val.(rbxfile.ValueBinaryString))
	case PROP_TYPE_PBOOL:
		err = b.WritePBool(val.(rbxfile.ValueBool))
	case PROP_TYPE_PSINT:
		err = b.WriteNewPSint(val.(rbxfile.ValueInt))
	case PROP_TYPE_PFLOAT:
		err = b.WritePFloat(val.(rbxfile.ValueFloat))
	case PROP_TYPE_PDOUBLE:
		err = b.WritePDouble(val.(rbxfile.ValueDouble))
	case PROP_TYPE_UDIM:
		err = b.WriteUDim(val.(rbxfile.ValueUDim))
	case PROP_TYPE_UDIM2:
		err = b.WriteUDim2(val.(rbxfile.ValueUDim2))
	case PROP_TYPE_RAY:
		err = b.WriteRay(val.(rbxfile.ValueRay))
	case PROP_TYPE_FACES:
		err = b.WriteFaces(val.(rbxfile.ValueFaces))
	case PROP_TYPE_AXES:
		err = b.WriteAxes(val.(rbxfile.ValueAxes))
	case PROP_TYPE_BRICKCOLOR:
		err = b.WriteBrickColor(val.(rbxfile.ValueBrickColor))
	case PROP_TYPE_COLOR3:
		err = b.WriteColor3(val.(rbxfile.ValueColor3))
	case PROP_TYPE_COLOR3UINT8:
		err = b.WriteColor3uint8(val.(rbxfile.ValueColor3uint8))
	case PROP_TYPE_VECTOR2:
		err = b.WriteVector2(val.(rbxfile.ValueVector2))
	case PROP_TYPE_VECTOR3_SIMPLE:
		err = b.WriteVector3Simple(val.(rbxfile.ValueVector3))
	case PROP_TYPE_VECTOR3_COMPLICATED:
		err = b.WriteVector3(val.(rbxfile.ValueVector3))
	case PROP_TYPE_VECTOR2UINT16:
		err = b.WriteVector2int16(val.(rbxfile.ValueVector2int16))
	case PROP_TYPE_VECTOR3UINT16:
		err = b.WriteVector3int16(val.(rbxfile.ValueVector3int16))
	case PROP_TYPE_CFRAME_SIMPLE:
		err = b.WriteCFrameSimple(val.(rbxfile.ValueCFrame))
	case PROP_TYPE_CFRAME_COMPLICATED:
		err = b.WriteCFrame(val.(rbxfile.ValueCFrame))
	case PROP_TYPE_INSTANCE:
        err = b.WriteObject(isClient, val.(rbxfile.ValueReference).Instance, isJoinData, context)
	case PROP_TYPE_CONTENT:
		err = b.WriteNewContent(isClient, val.(rbxfile.ValueContent), isJoinData, context)
	case PROP_TYPE_SYSTEMADDRESS:
		err = b.WriteSystemAddress(isClient, val.(rbxfile.ValueSystemAddress), isJoinData, context)
	case PROP_TYPE_TUPLE:
		err = b.WriteNewTuple(isClient, val.(rbxfile.ValueTuple), isJoinData, context)
	case PROP_TYPE_ARRAY:
		err = b.WriteNewArray(isClient, val.(rbxfile.ValueArray), isJoinData, context)
	case PROP_TYPE_DICTIONARY:
		err = b.WriteNewDictionary(isClient, val.(rbxfile.ValueDictionary), isJoinData, context)
	case PROP_TYPE_MAP:
		err = b.WriteNewMap(isClient, val.(rbxfile.ValueMap), isJoinData, context)
	case PROP_TYPE_NUMBERSEQUENCE:
		err = b.WriteNumberSequence(val.(rbxfile.ValueNumberSequence))
	case PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		err = b.WriteNumberSequenceKeypoint(val.(rbxfile.ValueNumberSequenceKeypoint))
	case PROP_TYPE_NUMBERRANGE:
		err = b.WriteNumberRange(val.(rbxfile.ValueNumberRange))
	case PROP_TYPE_COLORSEQUENCE:
		err = b.WriteColorSequence(val.(rbxfile.ValueColorSequence))
	case PROP_TYPE_COLORSEQUENCEKEYPOINT:
		err = b.WriteColorSequenceKeypoint(val.(rbxfile.ValueColorSequenceKeypoint))
	case PROP_TYPE_RECT2D:
		err = b.WriteRect2D(val.(rbxfile.ValueRect2D))
	case PROP_TYPE_PHYSICALPROPERTIES:
		err = b.WritePhysicalProperties(val.(rbxfile.ValuePhysicalProperties))
	default:
		return errors.New("Unsupported property type: " + strconv.Itoa(int(valueType)))
	}
	return err
}
func (b *ExtendedWriter) WriteNewTypeAndValue(isClient bool, val rbxfile.Value, isJoinData bool, context *CommunicationContext) error {
	var err error
	valueType := typeToNetwork(val)
	println("Writing typeandvalue", valueType)
	err = b.WriteByte(uint8(valueType))
	if valueType == 7 {
		err = b.WriteUint16BE(val.(rbxfile.ValueToken).ID)
		if err != nil {
			return err
		}
	}
	return b.writeSerializedValue(isClient, val, isJoinData, valueType, context)
}

func (b *ExtendedWriter) WriteNewTuple(isClient bool, val rbxfile.ValueTuple, isJoinData bool, context *CommunicationContext) error {
	err := b.WriteUintUTF8(len(val))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.WriteNewTypeAndValue(isClient, val[i], isJoinData, context)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *ExtendedWriter) WriteNewArray(isClient bool, val rbxfile.ValueArray, isJoinData bool, context *CommunicationContext) error {
	return b.WriteNewTuple(isClient, rbxfile.ValueTuple(val), isJoinData, context)
}

func (b *ExtendedWriter) WriteNewDictionary(isClient bool, val rbxfile.ValueDictionary, isJoinData bool, context *CommunicationContext) error {
	err := b.WriteUintUTF8(len(val))
	if err != nil {
		return err
	}
	for key, value := range val {
		err = b.WriteUintUTF8(len(key))
		if err != nil {
			return err
		}
		err = b.WriteASCII(key)
		if err != nil {
			return err
		}
		err = b.WriteNewTypeAndValue(isClient, value, isJoinData, context)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *ExtendedWriter) WriteNewMap(isClient bool, val rbxfile.ValueMap, isJoinData bool, context *CommunicationContext) error {
	return b.WriteNewDictionary(isClient, rbxfile.ValueDictionary(val), isJoinData, context)
}

func (b *ExtendedWriter) WriteNumberSequenceKeypoint(val rbxfile.ValueNumberSequenceKeypoint) error {
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
func (b *ExtendedWriter) WriteNumberSequence(val rbxfile.ValueNumberSequence) error {
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
func (b *ExtendedWriter) WriteNumberRange(val rbxfile.ValueNumberRange) error {
	err := b.WriteFloat32BE(val.Min)
	if err != nil {
		return err
	}
	return b.WriteFloat32BE(val.Max)
}

func (b *ExtendedWriter) WriteColorSequenceKeypoint(val rbxfile.ValueColorSequenceKeypoint) error {
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
func (b *ExtendedWriter) WriteColorSequence(val rbxfile.ValueColorSequence) error {
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

func (b *ExtendedWriter) WriteNewEnumValue(val rbxfile.ValueToken) error {
	return b.WriteUintUTF8(int(val.Value))
}

func (b *ExtendedWriter) WriteSystemAddress(isClient bool, val rbxfile.ValueSystemAddress, isJoinData bool, context *CommunicationContext) error {
	if !isJoinData {
		return b.WriteCachedSystemAddress(isClient, val, context)
	}
	addr, err := net.ResolveUDPAddr("udp", string(val))
	if err != nil {
		return err
	}
	err = b.Bytes(4, addr.IP)
	if err != nil {
		return err
	}
	return b.WriteUint16BE(uint16(addr.Port))
}

func (b *ExtendedWriter) WriteRect2D(val rbxfile.ValueRect2D) error {
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

func (b *ExtendedWriter) WritePhysicalProperties(val rbxfile.ValuePhysicalProperties) error {
	err := b.WriteBool(val.CustomPhysics)
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

func (b *ExtendedWriter) WriteCoordsMode0(val rbxfile.ValueVector3) error {
	return b.WriteVector3Simple(val)
}
func (b *ExtendedWriter) WriteCoordsMode1(val rbxfile.ValueVector3) error {
	valRange := float32(math.Max(math.Max(math.Abs(float64(val.X)), math.Abs(float64(val.Y))), math.Abs(float64(val.Z))))
	err := b.WriteFloat32BE(valRange)
	if err != nil {
		return err
	}
	if valRange <= 0.0000099999997 {
		return nil
	}
	err = b.WriteUint16BE(uint16(val.X / valRange * 32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE(uint16(val.Y / valRange * 32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE(uint16(val.Z / valRange * 32767.0 + 32767.0))
	return err
}
func (b *ExtendedWriter) WriteCoordsMode2(val rbxfile.ValueVector3) error {
	x_short := uint16((val.X + 1024.0) * 16.0)
	y_short := uint16((val.Y + 1024.0) * 16.0)
	z_short := uint16((val.Z + 1024.0) * 16.0)

	err := b.WriteUint16BE((x_short & 0x7F) << 7 | (x_short >> 8))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE((y_short & 0x7F) << 7 | (y_short >> 8))
	if err != nil {
		return err
	}
	err = b.WriteUint16BE((z_short & 0x7F) << 7 | (z_short >> 8))
	return err
}
func (b *ExtendedWriter) WritePhysicsCoords(val rbxfile.ValueVector3) error {
	mode := 0
	if	val.X < 1024.0 &&
		val.X > -1024.0 &&
		val.Y < 512.0 &&
		val.Y > -512.0 &&
		val.X < 1024.0 &&
		val.Z > -1024.0 &&
		math.Mod(float64(val.X), 0.0625) == 0 &&
		math.Mod(float64(val.Y), 0.0625) == 0 &&
		math.Mod(float64(val.Z), 0.0625) == 0 {
		mode = 2
	} else if val.X < 256.0 && val.X > -256.0 &&
			  val.Y < 256.0 && val.Y > -256.0 &&
			  val.Z < 256.0 && val.Z > -256.0 {
		mode = 1
	}
	err := b.Bits(2, uint64(mode))
	if err != nil {
		return err
	}
	switch mode {
	case 0:
		return b.WriteCoordsMode0(val)
	case 1:
		return b.WriteCoordsMode1(val)
	case 2:
		return b.WriteCoordsMode2(val)
	}
	return nil
}

func (b *ExtendedWriter) WriteMatrixMode0(val [9]float32) error {
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
func (b *ExtendedWriter) WriteMatrixMode1(val [9]float32) error {
	q := rotMatrixToQuaternion(val)
	err := b.WriteBool(q[3] < 0) // sqrt doesn't return negative numbers
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		err = b.WriteBool(q[i] < 0)
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
func (b *ExtendedWriter) WriteMatrixMode2(val [9]float32) error {
	return b.WriteMatrixMode1(val)
}
func (b *ExtendedWriter) WritePhysicsMatrix(val [9]float32) error {
	mode := 0
	// Let's just never use modes 1/2. It will be okay.
	err := b.Bits(2, uint64(mode))
	if err != nil {
		return err
	}
	switch mode {
	case 0:
		return b.WriteMatrixMode0(val)
	case 1:
		return b.WriteMatrixMode1(val)
	case 2:
		return b.WriteMatrixMode2(val)
	}
	return nil
}
func (b *ExtendedWriter) WritePhysicsCFrame(val rbxfile.ValueCFrame) error {
	err := b.WritePhysicsCoords(val.Position)
	if err != nil {
		return err
	}
	return b.WritePhysicsMatrix(val.Rotation)
}

func (b *ExtendedWriter) WriteMotor(motor PhysicsMotor) error {
	err := b.WriteBool(!motor.HasCoords1 && motor.HasCoords2)
	if err != nil {
		return err
	}
	if motor.HasCoords1 || motor.HasCoords2 {
		err = b.WriteBool(motor.HasCoords1)
		if err != nil {
			return err
		}
		err = b.WriteBool(motor.HasCoords2)
		if err != nil {
			return err
		}

		if motor.HasCoords1 {
			err = b.WritePhysicsCoords(motor.Coords1)
			if err != nil {
				return err
			}
		}
		if motor.HasCoords2 {
			err = b.WriteCoordsMode1(motor.Coords2)
			if err != nil {
				return err
			}
		}
	}
	return b.WriteByte(motor.Angle)
}

func (b *ExtendedWriter) WriteMotors(val []PhysicsMotor) error {
	err := b.WriteByte(uint8(len(val))) // causes issues. Roblox plsfix
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

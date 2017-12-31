package peer
import "math"
import "github.com/gskartwii/rbxfile"
import "errors"
import "strconv"
import "net"

func (b *extendedWriter) writeUDim(val rbxfile.ValueUDim) error {
	err := b.writeFloat32BE(val.Scale)
	if err != nil {
		return err
	}
	return b.writeUint32BE(uint32(val.Offset))
}
func (b *extendedWriter) writeUDim2(val rbxfile.ValueUDim2) error {
	err := b.writeUDim(val.X)
	if err != nil {
		return err
	}
	return b.writeUDim(val.Y)
}

func (b *extendedWriter) writeRay(val rbxfile.ValueRay) error {
	err := b.writeVector3Simple(val.Origin)
	if err != nil {
		return err
	}
	return b.writeVector3Simple(val.Direction)
}

func (b *extendedWriter) writeRegion3(val rbxfile.ValueRegion3) error {
	err := b.writeVector3Simple(val.Start)
	if err != nil {
		return err
	}
	return b.writeVector3Simple(val.End)
}

func (b *extendedWriter) writeRegion3int16(val rbxfile.ValueRegion3int16) error {
	err := b.writeVector3int16(val.Start)
	if err != nil {
		return err
	}
	return b.writeVector3int16(val.End)
}

func (b *extendedWriter) writeColor3(val rbxfile.ValueColor3) error {
	err := b.writeFloat32BE(val.R)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.G)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.B)
}
func (b *extendedWriter) writeColor3uint8(val rbxfile.ValueColor3uint8) error {
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
func (b *extendedWriter) writeVector2(val rbxfile.ValueVector2) error {
	err := b.writeFloat32BE(val.X)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Y)
}
func (b *extendedWriter) writeVector3Simple(val rbxfile.ValueVector3) error {
	err := b.writeFloat32BE(val.X)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Y)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Z)
}
func (b *extendedWriter) writeVector3(val rbxfile.ValueVector3) error {
	if math.Mod(float64(val.X), 0.5) != 0 ||
	   math.Mod(float64(val.Y), 0.1) != 0 ||
	   math.Mod(float64(val.Z), 0.5) != 0 ||
	   val.X >  511.5	||
	   val.X < -511.5	||
	   val.Y >  204.7	||
	   val.Y <      0	||
	   val.Z >  511.5	||
	   val.Z < -511.5	{
		err := b.writeBool(false)
		if err != nil {
			return err
		}
		err = b.writeVector3Simple(val)
		return err
	} else {
		err := b.writeBool(true)
		if err != nil {
			return err
		}
		x_scaled := uint16(val.X * 2)
		y_scaled := uint16(val.Y * 10)
		z_scaled := uint16(val.Z * 2)
		x_scaled = (x_scaled >> 8 & 7) | ((x_scaled & 0xFF) << 3)
		y_scaled = (y_scaled >> 8 & 7) | ((y_scaled & 0xFF) << 3)
		z_scaled = (z_scaled >> 8 & 7) | ((z_scaled & 0xFF) << 3)
		err = b.bits(11, uint64(x_scaled))
		if err != nil {
			return err
		}
		err = b.bits(11, uint64(y_scaled))
		if err != nil {
			return err
		}
		err = b.bits(11, uint64(z_scaled))
		return err
	}
}

func (b *extendedWriter) writeVector2int16(val rbxfile.ValueVector2int16) error {
	err := b.writeUint16BE(uint16(val.X))
	if err != nil {
		return err
	}
	return b.writeUint16BE(uint16(val.Y))
}
func (b *extendedWriter) writeVector3int16(val rbxfile.ValueVector3int16) error {
	err := b.writeUint16BE(uint16(val.X))
	if err != nil {
		return err
	}
	err = b.writeUint16BE(uint16(val.Y))
	if err != nil {
		return err
	}
	return b.writeUint16BE(uint16(val.Z))
}

func (b *extendedWriter) writePBool(val rbxfile.ValueBool) error {
	return b.writeBool(bool(val))
}
func (b *extendedWriter) writePSint(val rbxfile.ValueInt) error {
	return b.writeUint32BE(uint32(val))
}
func (b *extendedWriter) writePFloat(val rbxfile.ValueFloat) error {
	return b.writeFloat32BE(float32(val))
}
func (b *extendedWriter) writePDouble(val rbxfile.ValueDouble) error {
	return b.writeFloat64BE(float64(val))
}

func (b *extendedWriter) writeAxes(val rbxfile.ValueAxes) error {
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
	return b.writeUint32BE(uint32(write))
}
func (b *extendedWriter) writeFaces(val rbxfile.ValueFaces) error {
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
	return b.writeUint32BE(uint32(write))
}

func (b *extendedWriter) writeBrickColor(val rbxfile.ValueBrickColor) error {
	return b.bits(7, uint64(val))
}

func (b *extendedWriter) writeNewPString(isClient bool, val rbxfile.ValueString, isJoinData bool, context *CommunicationContext) (error) {
	if !isJoinData {
		return b.writeCached(isClient, string(val), context)
	}
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	return b.writeASCII(string(val))
}

func (b *extendedWriter) writeNewProtectedString(isClient bool, val rbxfile.ValueProtectedString, isJoinData bool, context *CommunicationContext) error {
	if !isJoinData {
		return b.writeNewCachedProtectedString(isClient, string(val), context)
	}
	return b.writeNewPString(isClient, rbxfile.ValueString(val), true, context)
}
func (b *extendedWriter) writeNewBinaryString(val rbxfile.ValueBinaryString) error {
	return b.writeNewPString(false, rbxfile.ValueString(val), true, nil) // hack: isClient doesn't matter because cache isn't used
}
func (b *extendedWriter) writeNewContent(isClient bool, val rbxfile.ValueContent, isJoinData bool, context *CommunicationContext) error {
	return b.writeNewPString(isClient, rbxfile.ValueString(val), isJoinData, context)
}

func (b *extendedWriter) writeCFrameSimple(val rbxfile.ValueCFrame) error {
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
func (b *extendedWriter) writeCFrame(val rbxfile.ValueCFrame) error {
	err := b.writeVector3(val.Position)
	if err != nil {
		return err
	}
	err = b.writeBool(false) // Not going to bother with lookup stuff
	if err != nil {
		return err
	}

	quat := rotMatrixToQuaternion(val.Rotation)
	b.writeFloat16BE(quat[3], -1.0, 1.0)
	for i := 0; i < 3; i++ {
		err = b.writeFloat16BE(quat[i], -1.0, 1.0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *extendedWriter) writeSintUTF8(val int32) error {
	return b.writeUintUTF8(uint32(val) << 1 ^ -(uint32(val) >> 31))
}
func (b *extendedWriter) writeNewPSint(val rbxfile.ValueInt) error {
	return b.writeSintUTF8(int32(val))
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
func (b *extendedWriter) writeSerializedValue(isClient bool, val rbxfile.Value, isJoinData bool, valueType uint8, context *CommunicationContext) error {
	var err error
	switch valueType {
	case PROP_TYPE_STRING:
		err = b.writeNewPString(isClient, val.(rbxfile.ValueString), isJoinData, context)
	case PROP_TYPE_STRING_NO_CACHE:
		err = b.writeNewPString(isClient, val.(rbxfile.ValueString), true, context)
	case PROP_TYPE_PROTECTEDSTRING_0:
		err = b.writeNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_1:
		err = b.writeNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_2:
		err = b.writeNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_3:
		err = b.writeNewProtectedString(isClient, val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_ENUM:
		err = b.writeNewEnumValue(val.(rbxfile.ValueToken))
	case PROP_TYPE_BINARYSTRING:
		err = b.writeNewBinaryString(val.(rbxfile.ValueBinaryString))
	case PROP_TYPE_PBOOL:
		err = b.writePBool(val.(rbxfile.ValueBool))
	case PROP_TYPE_PSINT:
		err = b.writeNewPSint(val.(rbxfile.ValueInt))
	case PROP_TYPE_PFLOAT:
		err = b.writePFloat(val.(rbxfile.ValueFloat))
	case PROP_TYPE_PDOUBLE:
		err = b.writePDouble(val.(rbxfile.ValueDouble))
	case PROP_TYPE_UDIM:
		err = b.writeUDim(val.(rbxfile.ValueUDim))
	case PROP_TYPE_UDIM2:
		err = b.writeUDim2(val.(rbxfile.ValueUDim2))
	case PROP_TYPE_RAY:
		err = b.writeRay(val.(rbxfile.ValueRay))
	case PROP_TYPE_FACES:
		err = b.writeFaces(val.(rbxfile.ValueFaces))
	case PROP_TYPE_AXES:
		err = b.writeAxes(val.(rbxfile.ValueAxes))
	case PROP_TYPE_BRICKCOLOR:
		err = b.writeBrickColor(val.(rbxfile.ValueBrickColor))
	case PROP_TYPE_COLOR3:
		err = b.writeColor3(val.(rbxfile.ValueColor3))
	case PROP_TYPE_COLOR3UINT8:
		err = b.writeColor3uint8(val.(rbxfile.ValueColor3uint8))
	case PROP_TYPE_VECTOR2:
		err = b.writeVector2(val.(rbxfile.ValueVector2))
	case PROP_TYPE_VECTOR3_SIMPLE:
		err = b.writeVector3Simple(val.(rbxfile.ValueVector3))
	case PROP_TYPE_VECTOR3_COMPLICATED:
		err = b.writeVector3(val.(rbxfile.ValueVector3))
	case PROP_TYPE_VECTOR2UINT16:
		err = b.writeVector2int16(val.(rbxfile.ValueVector2int16))
	case PROP_TYPE_VECTOR3UINT16:
		err = b.writeVector3int16(val.(rbxfile.ValueVector3int16))
	case PROP_TYPE_CFRAME_SIMPLE:
		err = b.writeCFrameSimple(val.(rbxfile.ValueCFrame))
	case PROP_TYPE_CFRAME_COMPLICATED:
		err = b.writeCFrame(val.(rbxfile.ValueCFrame))
	case PROP_TYPE_INSTANCE:
        err = b.writeObject(isClient, val.(rbxfile.ValueReference).Instance, isJoinData, context)
	case PROP_TYPE_CONTENT:
		err = b.writeNewContent(isClient, val.(rbxfile.ValueContent), isJoinData, context)
	case PROP_TYPE_SYSTEMADDRESS:
		err = b.writeSystemAddress(isClient, val.(rbxfile.ValueSystemAddress), isJoinData, context)
	case PROP_TYPE_TUPLE:
		err = b.writeNewTuple(isClient, val.(rbxfile.ValueTuple), isJoinData, context)
	case PROP_TYPE_ARRAY:
		err = b.writeNewArray(isClient, val.(rbxfile.ValueArray), isJoinData, context)
	case PROP_TYPE_DICTIONARY:
		err = b.writeNewDictionary(isClient, val.(rbxfile.ValueDictionary), isJoinData, context)
	case PROP_TYPE_MAP:
		err = b.writeNewMap(isClient, val.(rbxfile.ValueMap), isJoinData, context)
	case PROP_TYPE_NUMBERSEQUENCE:
		err = b.writeNumberSequence(val.(rbxfile.ValueNumberSequence))
	case PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		err = b.writeNumberSequenceKeypoint(val.(rbxfile.ValueNumberSequenceKeypoint))
	case PROP_TYPE_NUMBERRANGE:
		err = b.writeNumberRange(val.(rbxfile.ValueNumberRange))
	case PROP_TYPE_COLORSEQUENCE:
		err = b.writeColorSequence(val.(rbxfile.ValueColorSequence))
	case PROP_TYPE_COLORSEQUENCEKEYPOINT:
		err = b.writeColorSequenceKeypoint(val.(rbxfile.ValueColorSequenceKeypoint))
	case PROP_TYPE_RECT2D:
		err = b.writeRect2D(val.(rbxfile.ValueRect2D))
	case PROP_TYPE_PHYSICALPROPERTIES:
		err = b.writePhysicalProperties(val.(rbxfile.ValuePhysicalProperties))
	default:
		return errors.New("Unsupported property type: " + strconv.Itoa(int(valueType)))
	}
	return err
}
func (b *extendedWriter) writeNewTypeAndValue(isClient bool, val rbxfile.Value, isJoinData bool, context *CommunicationContext) error {
	var err error
	valueType := typeToNetwork(val)
	println("Writing typeandvalue", valueType)
	err = b.WriteByte(uint8(valueType))
	if valueType == 7 {
		err = b.writeUint16BE(val.(rbxfile.ValueToken).ID)
		if err != nil {
			return err
		}
	}
	return b.writeSerializedValue(isClient, val, isJoinData, valueType, context)
}

func (b *extendedWriter) writeNewTuple(isClient bool, val rbxfile.ValueTuple, isJoinData bool, context *CommunicationContext) error {
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeNewTypeAndValue(isClient, val[i], isJoinData, context)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeNewArray(isClient bool, val rbxfile.ValueArray, isJoinData bool, context *CommunicationContext) error {
	return b.writeNewTuple(isClient, rbxfile.ValueTuple(val), isJoinData, context)
}

func (b *extendedWriter) writeNewDictionary(isClient bool, val rbxfile.ValueDictionary, isJoinData bool, context *CommunicationContext) error {
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for key, value := range val {
		err = b.writeUintUTF8(uint32(len(key)))
		if err != nil {
			return err
		}
		err = b.writeASCII(key)
		if err != nil {
			return err
		}
		err = b.writeNewTypeAndValue(isClient, value, isJoinData, context)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeNewMap(isClient bool, val rbxfile.ValueMap, isJoinData bool, context *CommunicationContext) error {
	return b.writeNewDictionary(isClient, rbxfile.ValueDictionary(val), isJoinData, context)
}

func (b *extendedWriter) writeNumberSequenceKeypoint(val rbxfile.ValueNumberSequenceKeypoint) error {
	err := b.writeFloat32BE(val.Time)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Value)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Envelope)
	return err
}
func (b *extendedWriter) writeNumberSequence(val rbxfile.ValueNumberSequence) error {
	err := b.writeUint32BE(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeNumberSequenceKeypoint(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeNumberRange(val rbxfile.ValueNumberRange) error {
	err := b.writeFloat32BE(val.Min)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Max)
}

func (b *extendedWriter) writeColorSequenceKeypoint(val rbxfile.ValueColorSequenceKeypoint) error {
	err := b.writeFloat32BE(val.Time)
	if err != nil {
		return err
	}
	err = b.writeColor3(val.Value)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Envelope)
}
func (b *extendedWriter) writeColorSequence(val rbxfile.ValueColorSequence) error {
	err := b.writeUint32BE(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeColorSequenceKeypoint(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *extendedWriter) writeNewEnumValue(val rbxfile.ValueToken) error {
	return b.writeUintUTF8(val.Value)
}

func (b *extendedWriter) writeSystemAddress(isClient bool, val rbxfile.ValueSystemAddress, isJoinData bool, context *CommunicationContext) error {
	if !isJoinData {
		return b.writeCachedSystemAddress(isClient, val, context)
	}
	addr, err := net.ResolveUDPAddr("udp", string(val))
	if err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		addr.IP[i] = addr.IP[i] ^ 0xFF // bitwise NOT
	}

	err = b.bytes(4, addr.IP)
	if err != nil {
		return err
	}

	// in case the value will be used again
	for i := 0; i < 4; i++ {
		addr.IP[i] = addr.IP[i] ^ 0xFF
	}

	return b.writeUint16BE(uint16(addr.Port))
}

func (b *extendedWriter) writeRect2D(val rbxfile.ValueRect2D) error {
	err := b.writeFloat32BE(val.Min.X)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Min.Y)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Max.X)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Max.Y)
	return err
}

func (b *extendedWriter) writePhysicalProperties(val rbxfile.ValuePhysicalProperties) error {
	err := b.writeBool(val.CustomPhysics)
	if err != nil {
		return err
	}
	if val.CustomPhysics {
		err := b.writeFloat32BE(val.Density)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.Friction)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.Elasticity)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.FrictionWeight)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.ElasticityWeight)
	}
	return err
}

func (b *extendedWriter) writeCoordsMode0(val rbxfile.ValueVector3) error {
	return b.writeVector3Simple(val)
}
func (b *extendedWriter) writeCoordsMode1(val rbxfile.ValueVector3) error {
	valRange := float32(math.Max(math.Max(math.Abs(float64(val.X)), math.Abs(float64(val.Y))), math.Abs(float64(val.Z))))
	err := b.writeFloat32BE(valRange)
	if err != nil {
		return err
	}
	if valRange <= 0.0000099999997 {
		return nil
	}
	err = b.writeUint16BE(uint16(val.X / valRange * 32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.writeUint16BE(uint16(val.Y / valRange * 32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.writeUint16BE(uint16(val.Z / valRange * 32767.0 + 32767.0))
	return err
}
func (b *extendedWriter) writeCoordsMode2(val rbxfile.ValueVector3) error {
	x_short := uint16((val.X + 1024.0) * 16.0)
	y_short := uint16((val.Y + 1024.0) * 16.0)
	z_short := uint16((val.Z + 1024.0) * 16.0)

	err := b.writeUint16BE((x_short & 0x7F) << 7 | (x_short >> 8))
	if err != nil {
		return err
	}
	err = b.writeUint16BE((y_short & 0x7F) << 7 | (y_short >> 8))
	if err != nil {
		return err
	}
	err = b.writeUint16BE((z_short & 0x7F) << 7 | (z_short >> 8))
	return err
}
func (b *extendedWriter) writePhysicsCoords(val rbxfile.ValueVector3) error {
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
	err := b.bits(2, uint64(mode))
	if err != nil {
		return err
	}
	switch mode {
	case 0:
		return b.writeCoordsMode0(val)
	case 1:
		return b.writeCoordsMode1(val)
	case 2:
		return b.writeCoordsMode2(val)
	}
	return nil
}

func (b *extendedWriter) writeMatrixMode0(val [9]float32) error {
	var err error
	q := rotMatrixToQuaternion(val)
	b.writeFloat32BE(q[3])
	for i := 0; i < 3; i++ {
		err = b.writeFloat32BE(q[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeMatrixMode1(val [9]float32) error {
	q := rotMatrixToQuaternion(val)
	err := b.writeBool(q[3] < 0) // sqrt doesn't return negative numbers
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		err = b.writeBool(q[i] < 0)
		if err != nil {
			return err
		}
	}
	for i := 0; i < 3; i++ {
		err = b.writeUint16LE(uint16(math.Abs(float64(q[i]))))
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeMatrixMode2(val [9]float32) error {
	return b.writeMatrixMode1(val)
}
func (b *extendedWriter) writePhysicsMatrix(val [9]float32) error {
	mode := 0
	// Let's just never use modes 1/2. It will be okay.
	err := b.bits(2, uint64(mode))
	if err != nil {
		return err
	}
	switch mode {
	case 0:
		return b.writeMatrixMode0(val)
	case 1:
		return b.writeMatrixMode1(val)
	case 2:
		return b.writeMatrixMode2(val)
	}
	return nil
}
func (b *extendedWriter) writePhysicsCFrame(val rbxfile.ValueCFrame) error {
	err := b.writePhysicsCoords(val.Position)
	if err != nil {
		return err
	}
	return b.writePhysicsMatrix(val.Rotation)
}

func (b *extendedWriter) writeMotor(motor PhysicsMotor) error {
	return errors.New("sorry, not implemented") // TODO
}

func (b *extendedWriter) writeMotors(val []PhysicsMotor) error {
	err := b.WriteByte(uint8(len(val))) // causes issues. Roblox plsfix
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeMotor(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}

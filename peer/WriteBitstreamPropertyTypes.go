package peer
import "math"
import "github.com/gskartwii/rbxfile"

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

func (b *ExtendedWriter) WriteNewPString(val rbxfile.ValueString, isJoinData bool, context *CommunicationContext) (error) {
	if !isJoinData {
		return b.WriteCached(string(val), context)
	}
	err := b.WriteUintUTF8(len(val))
	if err != nil {
		return err
	}
	return b.WriteASCII(string(val))
}

func (b *ExtendedWriter) WriteNewProtectedString(val rbxfile.ValueProtectedString, isJoinData bool, context *CommunicationContext) error {
	if !isJoinData {
		return b.WriteNewCachedProtectedString(string(val), context)
	}
	return b.WriteNewPString(rbxfile.ValueString(val), true, context)
}
func (b *ExtendedWriter) WriteNewBinaryString(val rbxfile.ValueBinaryString) error {
	return b.WriteNewPString(rbxfile.ValueString(val), true, nil)
}
func (b *ExtendedWriter) WriteNewContent(val rbxfile.ValueContent) error {
	return b.WriteNewPString(rbxfile.ValueString(val), true, nil)
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
func (b *ExtendedWriter) writeSerializedValue(val rbxfile.Value, isJoinData bool, valueType uint8, context *CommunicationContext) error {
	var err error
	var result rbxfile.Value
	switch valueType {
	case PROP_TYPE_STRING:
		err = b.WriteNewPString(val.(rbxfile.ValueString), isJoinData, context)
	case PROP_TYPE_STRING_NO_CACHE:
		err = b.WriteNewPString(val.(rbxfile.ValueString), true, context)
	case PROP_TYPE_PROTECTEDSTRING_0:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_1:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_2:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_3:
		err = b.WriteNewProtectedString(val.(rbxfile.ValueProtectedString), isJoinData, context)
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
        err = b.WriteObject(val.(rbxfile.ValueReference), isJoinData, context)
	case PROP_TYPE_CONTENT:
		err = b.WriteNewContent(val.(rbxfile.ValueContent), isJoinData, context)
	case PROP_TYPE_SYSTEMADDRESS:
		err = b.WriteSystemAddress(val.(rbxfile.ValueSystemAddress), isJoinData, context)
	case PROP_TYPE_TUPLE:
		err = b.WriteNewTuple(val.(rbxfile.ValueTuple), isJoinData, context)
	case PROP_TYPE_ARRAY:
		err = b.WriteNewArray(val.(rbxfile.ValueArray), isJoinData, context)
	case PROP_TYPE_DICTIONARY:
		err = b.WriteNewDictionary(val.(rbxfile.ValueDictionary), isJoinData, context)
	case PROP_TYPE_MAP:
		err = b.WriteNewMap(val.(rbxfile.ValueMap), isJoinData, context)
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
		return nil, errors.New("Unsupported property type: " + strconv.Itoa(int(valueType)))
	}
	return result, err
}
func (b *ExtendedWriter) WriteNewTypeAndValue(val rbxfile.Value, isJoinData bool, context *CommunicationContext) error {
	var err error
	valueType := typeToNetwork(val)
	if valueType == 7 {
		err = b.WriteUint16BE(val.(rbxfile.ValueToken).ID)
		if err != nil {
			return err
		}
	}
	return b.writeSerializedValue(val, isJoinData, valueType, context)
}

func (b *ExtendedWriter) WriteNewTuple(val rbxfile.ValueTuple, isJoinData bool, context *CommunicationContext) error {
	err := b.WriteUintUTF8(len(val))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.WriteNewTypeAndValue(val[i], isJoinData, context)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *ExtendedWriter) WriteNewArray(val rbxfile.ValueArray, isJoinData bool, context *CommunicationContext) error {
	return b.WriteNewTuple(rbxfile.ValueTuple(val), isJoinData, context)
}

func (b *ExtendedWriter) WriteNewDictionary(val rbxfile.ValueDictionary, isJoinData bool, context *CommunicationContext) error {
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
		err = b.WriteNewTypeAndValue(value, isJoinData, context)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *ExtendedWriter) WriteNewMap(val rbxfile.ValueMap, isJoinData bool, context *CommunicationContext) error {
	return b.WriteNewDictionary(rbxfile.ValueDictionary(val), isJoinData, context)
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

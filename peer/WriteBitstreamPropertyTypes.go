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
	if val.X % 0.5 != 0 ||
	   val.Y % 0.1 != 0 ||
	   val.Z % 0.5 != 0 ||
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
		err = b.Bits(11, x_scaled)
		if err != nil {
			return err
		}
		err = b.Bits(11, y_scaled)
		if err != nil {
			return err
		}
		err = b.Bits(11, z_scaled)
		return err
	}
}

func (b *ExtendedWriter) WriteVector2int16(val rbxfile.ValueVector2int16) error {
	err := b.WriteUint16BE(val.X)
	if err != nil {
		return err
	}
	return b.WriteUint16BE(val.Y)
}
func (b *ExtendedWriter) WriteVector3int16(val rbxfile.ValueVector3int16) error {
	err := b.WriteUint16BE(val.X)
	if err != nil {
		return err
	}
	err = b.WriteUint16BE(val.Y)
	if err != nil {
		return err
	}
	return b.WriteUint16BE(val.Z)
}

func (b *ExtendedWriter) WritePBool(val rbxfile.ValueBool) error {
	return b.WriteBool(val)
}
func (b *ExtendedWriter) WritePSint(val rbxfile.ValueInt) error {
	return b.WriteUint32BE(val)
}
func (b *ExtendedWriter) WritePFloat(val rbxfile.ValueFloat) error {
	return b.WriteFloat32BE(val)
}
func (b *ExtendedWriter) WritePDouble(val rbxfile.ValueDouble) error {
	return b.WriteFloat64BE(val)
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
	return b.Bits(7, val)
}

func (b *ExtendedWriter) WriteNewPString(val rbxfile.ValueString, isJoinData bool, context *CommunicationContext) (error) {
	if !isJoinData {
		return b.WriteCached(val, context)
	}
	err := b.WriteUintUTF8(len(val))
	if err != nil {
		return err
	}
	return b.WriteASCII(val)
}

func rotMatrixToQuaternion(r [9]float32) [4]float32 {
	q := math.Sqrt(1 + r[0*3+0] + r[1*3+1] + r[2*3+2])/2
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
	return b.WriteUintUTF8(val << 1 ^ -(val >> 31))
}
func (b *ExtendedWriter) WriteNewPSint(val rbxfile.ValueInt) error {
	return b.WriteSintUTF8(val)
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
		err = thisBitstream.WriteNewPString(val, isJoinData, context)
	case PROP_TYPE_STRING_NO_CACHE:
		err = thisBitstream.WriteNewPString(valtrue, context)
	case PROP_TYPE_PROTECTEDSTRING_0:
		err = thisBitstream.WriteNewProtectedString(val, isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_1:
		err = thisBitstream.WriteNewProtectedString(val, isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_2:
		err = thisBitstream.WriteNewProtectedString(val, isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_3:
		err = thisBitstream.WriteNewProtectedString(val, isJoinData, context)
	case PROP_TYPE_ENUM:
		err = thisBitstream.WriteNewEnumValue(val)
	case PROP_TYPE_BINARYSTRING:
		err = thisBitstream.WriteNewBinaryString(val)
	case PROP_TYPE_PBOOL:
		err = thisBitstream.WritePBool(val)
	case PROP_TYPE_PSINT:
		err = thisBitstream.WriteNewPSint(val)
	case PROP_TYPE_PFLOAT:
		err = thisBitstream.WritePFloat(val)
	case PROP_TYPE_PDOUBLE:
		err = thisBitstream.WritePDouble(val)
	case PROP_TYPE_UDIM:
		err = thisBitstream.WriteUDim(val)
	case PROP_TYPE_UDIM2:
		err = thisBitstream.WriteUDim2(val)
	case PROP_TYPE_RAY:
		err = thisBitstream.WriteRay(val)
	case PROP_TYPE_FACES:
		err = thisBitstream.WriteFaces(val)
	case PROP_TYPE_AXES:
		err = thisBitstream.WriteAxes(val)
	case PROP_TYPE_BRICKCOLOR:
		err = thisBitstream.WriteBrickColor(val)
	case PROP_TYPE_COLOR3:
		err = thisBitstream.WriteColor3(val)
	case PROP_TYPE_COLOR3UINT8:
		err = thisBitstream.WriteColor3uint8(val)
	case PROP_TYPE_VECTOR2:
		err = thisBitstream.WriteVector2(val)
	case PROP_TYPE_VECTOR3_SIMPLE:
		err = thisBitstream.WriteVector3Simple(val)
	case PROP_TYPE_VECTOR3_COMPLICATED:
		err = thisBitstream.WriteVector3(val)
	case PROP_TYPE_VECTOR2UINT16:
		err = thisBitstream.WriteVector2int16(val)
	case PROP_TYPE_VECTOR3UINT16:
		err = thisBitstream.WriteVector3int16(val)
	case PROP_TYPE_CFRAME_SIMPLE:
		err = thisBitstream.WriteCFrameSimple(val)
	case PROP_TYPE_CFRAME_COMPLICATED:
		err = thisBitstream.WriteCFrame(val)
	case PROP_TYPE_INSTANCE:
        err = thisBitstream.WriteObject(val, isJoinData, context)
	case PROP_TYPE_CONTENT:
		err = thisBitstream.WriteNewContent(val, isJoinData, context)
	case PROP_TYPE_SYSTEMADDRESS:
		err = thisBitstream.WriteSystemAddress(val, isJoinData, context)
	case PROP_TYPE_TUPLE:
		err = thisBitstream.WriteNewTuple(val, isJoinData, context)
	case PROP_TYPE_ARRAY:
		err = thisBitstream.WriteNewArray(val, isJoinData, context)
	case PROP_TYPE_DICTIONARY:
		err = thisBitstream.WriteNewDictionary(val, isJoinData, context)
	case PROP_TYPE_MAP:
		err = thisBitstream.WriteNewMap(val, isJoinData, context)
	case PROP_TYPE_NUMBERSEQUENCE:
		err = thisBitstream.WriteNumberSequence(val)
	case PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		err = thisBitstream.WriteNumberSequenceKeypoint(val)
	case PROP_TYPE_NUMBERRANGE:
		err = thisBitstream.WriteNumberRange(val)
	case PROP_TYPE_COLORSEQUENCE:
		err = thisBitstream.WriteColorSequence(val)
	case PROP_TYPE_COLORSEQUENCEKEYPOINT:
		err = thisBitstream.WriteColorSequenceKeypoint(val)
	case PROP_TYPE_RECT2D:
		err = thisBitstream.WriteRect2D(val)
	case PROP_TYPE_PHYSICALPROPERTIES:
		err = thisBitstream.WritePhysicalProperties(val)
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

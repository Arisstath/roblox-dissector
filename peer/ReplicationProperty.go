package peer
import "errors"
import "strconv"
import "github.com/gskartwii/rbxfile"

const (
	ROUND_JOINDATA	= iota
	ROUND_STRINGS	= iota
	ROUND_OTHER		= iota
	ROUND_UPDATE	= iota
)

func readSerializedValue(isJoinData bool, valueType uint8, thisBitstream *ExtendedReader, context *CommunicationContext) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case PROP_TYPE_STRING:
		result, err = thisBitstream.ReadNewPString(isJoinData, context)
	case PROP_TYPE_STRING_NO_CACHE:
		result, err = thisBitstream.ReadNewPString(true, context)
	case PROP_TYPE_PROTECTEDSTRING_0:
		result, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_1:
		result, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_2:
		result, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_3:
		result, err = thisBitstream.ReadNewProtectedString(isJoinData, context)
	case PROP_TYPE_ENUM:
		result, err = thisBitstream.ReadNewEnumValue()
	case PROP_TYPE_BINARYSTRING:
		result, err = thisBitstream.ReadNewBinaryString()
	case PROP_TYPE_PBOOL:
		result, err = thisBitstream.ReadPBool()
	case PROP_TYPE_PSINT:
		result, err = thisBitstream.ReadNewPSint()
	case PROP_TYPE_PFLOAT:
		result, err = thisBitstream.ReadPFloat()
	case PROP_TYPE_PDOUBLE:
		result, err = thisBitstream.ReadPDouble()
	case PROP_TYPE_UDIM:
		result, err = thisBitstream.ReadUDim()
	case PROP_TYPE_UDIM2:
		result, err = thisBitstream.ReadUDim2()
	case PROP_TYPE_RAY:
		result, err = thisBitstream.ReadRay()
	case PROP_TYPE_FACES:
		result, err = thisBitstream.ReadFaces()
	case PROP_TYPE_AXES:
		result, err = thisBitstream.ReadAxes()
	case PROP_TYPE_BRICKCOLOR:
		result, err = thisBitstream.ReadBrickColor()
	case PROP_TYPE_COLOR3:
		result, err = thisBitstream.ReadColor3()
	case PROP_TYPE_COLOR3UINT8:
		result, err = thisBitstream.ReadColor3uint8()
	case PROP_TYPE_VECTOR2:
		result, err = thisBitstream.ReadVector2()
	case PROP_TYPE_VECTOR3_SIMPLE:
		result, err = thisBitstream.ReadVector3Simple()
	case PROP_TYPE_VECTOR3_COMPLICATED:
		result, err = thisBitstream.ReadVector3()
	case PROP_TYPE_VECTOR2UINT16:
		result, err = thisBitstream.ReadVector2int16()
	case PROP_TYPE_VECTOR3UINT16:
		result, err = thisBitstream.ReadVector3int16()
	case PROP_TYPE_CFRAME_SIMPLE:
		result, err = thisBitstream.ReadCFrameSimple()
	case PROP_TYPE_CFRAME_COMPLICATED:
		result, err = thisBitstream.ReadCFrame()
	case PROP_TYPE_INSTANCE:
        var referent Referent
        referent, err = thisBitstream.ReadObject(isJoinData, context)
        instance := context.InstancesByReferent.TryGetInstance(referent)
        result = rbxfile.ValueReference{instance}
	case PROP_TYPE_CONTENT:
		result, err = thisBitstream.ReadNewContent(isJoinData, context)
	case PROP_TYPE_SYSTEMADDRESS:
		result, err = thisBitstream.ReadSystemAddress(isJoinData, context)
	case PROP_TYPE_TUPLE:
		result, err = thisBitstream.ReadNewTuple(isJoinData, context)
	case PROP_TYPE_ARRAY:
		result, err = thisBitstream.ReadNewArray(isJoinData, context)
	case PROP_TYPE_DICTIONARY:
		result, err = thisBitstream.ReadNewDictionary(isJoinData, context)
	case PROP_TYPE_MAP:
		result, err = thisBitstream.ReadNewMap(isJoinData, context)
	case PROP_TYPE_NUMBERSEQUENCE:
		result, err = thisBitstream.ReadNumberSequence()
	case PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		result, err = thisBitstream.ReadNumberSequenceKeypoint()
	case PROP_TYPE_NUMBERRANGE:
		result, err = thisBitstream.ReadNumberRange()
	case PROP_TYPE_COLORSEQUENCE:
		result, err = thisBitstream.ReadColorSequence()
	case PROP_TYPE_COLORSEQUENCEKEYPOINT:
		result, err = thisBitstream.ReadColorSequenceKeypoint()
	case PROP_TYPE_RECT2D:
		result, err = thisBitstream.ReadRect2D()
	case PROP_TYPE_PHYSICALPROPERTIES:
		result, err = thisBitstream.ReadPhysicalProperties()
	default:
		return nil, errors.New("Unsupported property type: " + strconv.Itoa(int(valueType)))
	}
	return result, err
}

func (schema StaticPropertySchema) Decode(round int, packet *UDPPacket, context *CommunicationContext) (rbxfile.Value, error) {
	var err error
	thisBitstream := packet.Stream
	isJoinData := round == ROUND_JOINDATA
	if round != ROUND_UPDATE {
        var isDefault bool
        isDefault, err = thisBitstream.ReadBool()
		if isDefault || err != nil {
			if DEBUG && round != ROUND_JOINDATA {
				println("Read", schema.Name, "default")
			}
			return nil, err
		}
	}

    val, err := readSerializedValue(isJoinData, schema.Type, thisBitstream, context)
    if val.Type().String() != "ProtectedString" && round != ROUND_JOINDATA && DEBUG {
        println("Read", schema.Name, val.String())
    }
    if err != nil {
        return val, errors.New("while parsing " + schema.Name + ": " + err.Error())
    }
    return val, nil
}

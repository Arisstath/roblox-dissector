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

func readSerializedValue(isClient bool, isJoinData bool, enumId uint16, valueType uint8, thisBitstream *extendedReader, context *CommunicationContext) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	switch valueType {
	case PROP_TYPE_STRING:
		result, err = thisBitstream.readNewPString(isClient, isJoinData, context)
	case PROP_TYPE_STRING_NO_CACHE:
		result, err = thisBitstream.readNewPString(isClient, true, context)
	case PROP_TYPE_PROTECTEDSTRING_0:
		result, err = thisBitstream.readNewProtectedString(isClient, isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_1:
		result, err = thisBitstream.readNewProtectedString(isClient, isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_2:
		result, err = thisBitstream.readNewProtectedString(isClient, isJoinData, context)
	case PROP_TYPE_PROTECTEDSTRING_3:
		result, err = thisBitstream.readNewProtectedString(isClient, isJoinData, context)
	case PROP_TYPE_ENUM:
		result, err = thisBitstream.readNewEnumValue(enumId, context)
	case PROP_TYPE_BINARYSTRING:
		result, err = thisBitstream.readNewBinaryString()
	case PROP_TYPE_PBOOL:
		result, err = thisBitstream.readPBool()
	case PROP_TYPE_PSINT:
		result, err = thisBitstream.readNewPSint()
	case PROP_TYPE_PFLOAT:
		result, err = thisBitstream.readPFloat()
	case PROP_TYPE_PDOUBLE:
		result, err = thisBitstream.readPDouble()
	case PROP_TYPE_UDIM:
		result, err = thisBitstream.readUDim()
	case PROP_TYPE_UDIM2:
		result, err = thisBitstream.readUDim2()
	case PROP_TYPE_RAY:
		result, err = thisBitstream.readRay()
	case PROP_TYPE_FACES:
		result, err = thisBitstream.readFaces()
	case PROP_TYPE_AXES:
		result, err = thisBitstream.readAxes()
	case PROP_TYPE_BRICKCOLOR:
		result, err = thisBitstream.readBrickColor()
	case PROP_TYPE_COLOR3:
		result, err = thisBitstream.readColor3()
	case PROP_TYPE_COLOR3UINT8:
		result, err = thisBitstream.readColor3uint8()
	case PROP_TYPE_VECTOR2:
		result, err = thisBitstream.readVector2()
	case PROP_TYPE_VECTOR3_SIMPLE:
		result, err = thisBitstream.readVector3Simple()
	case PROP_TYPE_VECTOR3_COMPLICATED:
		result, err = thisBitstream.readVector3()
	case PROP_TYPE_VECTOR2UINT16:
		result, err = thisBitstream.readVector2int16()
	case PROP_TYPE_VECTOR3UINT16:
		result, err = thisBitstream.readVector3int16()
	case PROP_TYPE_CFRAME_SIMPLE:
		result, err = thisBitstream.readCFrameSimple()
	case PROP_TYPE_CFRAME_COMPLICATED:
		result, err = thisBitstream.readCFrame()
	case PROP_TYPE_INSTANCE:
        var referent Referent
        referent, err = thisBitstream.readObject(isClient, isJoinData, context)
        instance := context.InstancesByReferent.TryGetInstance(referent)
        result = rbxfile.ValueReference{instance}
	case PROP_TYPE_CONTENT:
		result, err = thisBitstream.readNewContent(isClient, isJoinData, context)
	case PROP_TYPE_SYSTEMADDRESS:
		result, err = thisBitstream.readSystemAddress(isClient, isJoinData, context)
	case PROP_TYPE_TUPLE:
		result, err = thisBitstream.readNewTuple(isClient, isJoinData, context)
	case PROP_TYPE_ARRAY:
		result, err = thisBitstream.readNewArray(isClient, isJoinData, context)
	case PROP_TYPE_DICTIONARY:
		result, err = thisBitstream.readNewDictionary(isClient, isJoinData, context)
	case PROP_TYPE_MAP:
		result, err = thisBitstream.readNewMap(isClient, isJoinData, context)
	case PROP_TYPE_NUMBERSEQUENCE:
		result, err = thisBitstream.readNumberSequence()
	case PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		result, err = thisBitstream.readNumberSequenceKeypoint()
	case PROP_TYPE_NUMBERRANGE:
		result, err = thisBitstream.readNumberRange()
	case PROP_TYPE_COLORSEQUENCE:
		result, err = thisBitstream.readColorSequence()
	case PROP_TYPE_COLORSEQUENCEKEYPOINT:
		result, err = thisBitstream.readColorSequenceKeypoint()
	case PROP_TYPE_RECT2D:
		result, err = thisBitstream.readRect2D()
	case PROP_TYPE_PHYSICALPROPERTIES:
		result, err = thisBitstream.readPhysicalProperties()
	default:
		return nil, errors.New("Unsupported property type: " + strconv.Itoa(int(valueType)))
	}
	return result, err
}

func (schema StaticPropertySchema) Decode(isClient bool, round int, packet *UDPPacket, context *CommunicationContext) (rbxfile.Value, error) {
	var err error
	thisBitstream := packet.stream
	isJoinData := round == ROUND_JOINDATA
	if round != ROUND_UPDATE {
        var isDefault bool
        isDefault, err = thisBitstream.readBool()
		if isDefault || err != nil {
			if DEBUG && round == ROUND_JOINDATA && isClient {
				println("read", schema.Name, "default")
			}
			return rbxfile.DefaultValue, err
		}
	}

    val, err := readSerializedValue(isClient, isJoinData, schema.EnumID, schema.Type, thisBitstream, context)
    if val.Type().String() != "ProtectedString" && round == ROUND_JOINDATA && DEBUG && isClient {
        println("read", schema.Name, val.String())
    }
    if err != nil {
        return val, errors.New("while parsing " + schema.Name + ": " + err.Error())
    }
    return val, nil
}

func (schema StaticPropertySchema) serialize(isClient bool, value rbxfile.Value, round int, context *CommunicationContext, stream *extendedWriter) error {
    if round != ROUND_UPDATE {
        if value == rbxfile.DefaultValue {
            return stream.writeBool(true)
        }
    }
	err := stream.writeBool(false) // not default
	if err != nil {
		return err
	}
    return stream.writeSerializedValue(isClient, value, round == ROUND_JOINDATA, schema.Type, context) // let's just pray that this works
}

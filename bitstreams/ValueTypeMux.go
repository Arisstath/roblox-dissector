package bitstreams
import "github.com/gskartwii/roblox-dissector/util"
import "github.com/gskartwii/roblox-dissector/schema"
import "github.com/gskartwii/rbxfile"

// readNewTypeAndValue is never used by join data!
func (b *BitstreamReader) ReadNewTypeAndValue(reader util.PacketReader) (rbxfile.Value, error) {
	var val rbxfile.Value
	thisType, err := b.ReadUint8()
	if err != nil {
		return val, err
	}

	var enumID uint16
	if thisType == schema.PROP_TYPE_ENUM {
		enumID, err = b.ReadUint16BE()
		if err != nil {
			return val, err
		}
	}

	val, err = b.ReadSerializedValue(reader, thisType, enumID)
	return val, err
}

func (b *BitstreamReader) ReadSerializedValueGeneric(reader util.PacketReader, valueType uint8, enumId uint16) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	var temp string
	switch valueType {
	case schema.PROP_TYPE_INVALID: // I assume this is how it works, anyway
		result = nil
		err = nil
	case schema.PROP_TYPE_STRING_NO_CACHE:
		temp, err = b.ReadVarLengthString()
		result = rbxfile.ValueString(temp)
	case schema.PROP_TYPE_ENUM:
		result, err = b.ReadNewEnumValue(enumId)
	case schema.PROP_TYPE_BINARYSTRING:
		result, err = b.ReadNewBinaryString()
	case schema.PROP_TYPE_PBOOL:
		result, err = b.ReadPBool()
	case schema.PROP_TYPE_PSINT:
		result, err = b.ReadNewPSint()
	case schema.PROP_TYPE_PFLOAT:
		result, err = b.ReadPFloat()
	case schema.PROP_TYPE_PDOUBLE:
		result, err = b.ReadPDouble()
	case schema.PROP_TYPE_UDIM:
		result, err = b.ReadUDim()
	case schema.PROP_TYPE_UDIM2:
		result, err = b.ReadUDim2()
	case schema.PROP_TYPE_RAY:
		result, err = b.ReadRay()
	case schema.PROP_TYPE_FACES:
		result, err = b.ReadFaces()
	case schema.PROP_TYPE_AXES:
		result, err = b.ReadAxes()
	case schema.PROP_TYPE_BRICKCOLOR:
		result, err = b.ReadBrickColor()
	case schema.PROP_TYPE_COLOR3:
		result, err = b.ReadColor3()
	case schema.PROP_TYPE_COLOR3UINT8:
		result, err = b.ReadColor3uint8()
	case schema.PROP_TYPE_VECTOR2:
		result, err = b.ReadVector2()
	case schema.PROP_TYPE_VECTOR3_SIMPLE:
		result, err = b.ReadVector3Simple()
	case schema.PROP_TYPE_VECTOR3_COMPLICATED:
		result, err = b.ReadVector3()
	case schema.PROP_TYPE_VECTOR2UINT16:
		result, err = b.ReadVector2int16()
	case schema.PROP_TYPE_VECTOR3UINT16:
		result, err = b.ReadVector3int16()
	case schema.PROP_TYPE_CFRAME_SIMPLE:
		result, err = b.ReadCFrameSimple()
	case schema.PROP_TYPE_CFRAME_COMPLICATED:
		result, err = b.ReadCFrame()
	case schema.PROP_TYPE_NUMBERSEQUENCE:
		result, err = b.ReadNumberSequence()
	case schema.PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		result, err = b.ReadNumberSequenceKeypoint()
	case schema.PROP_TYPE_NUMBERRANGE:
		result, err = b.ReadNumberRange()
	case schema.PROP_TYPE_COLORSEQUENCE:
		result, err = b.ReadColorSequence()
	case schema.PROP_TYPE_COLORSEQUENCEKEYPOINT:
		result, err = b.ReadColorSequenceKeypoint()
	case schema.PROP_TYPE_RECT2D:
		result, err = b.ReadRect2D()
	case schema.PROP_TYPE_PHYSICALPROPERTIES:
		result, err = b.ReadPhysicalProperties()
	case schema.PROP_TYPE_REGION3:
		result, err = b.ReadRegion3()
	case schema.PROP_TYPE_REGION3INT16:
		result, err = b.ReadRegion3int16()
	case schema.PROP_TYPE_INT64:
		result, err = b.ReadInt64()
	}
	return result, err
}

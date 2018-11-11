package bitstreams
// readNewTypeAndValue is never used by join data!
func (b *BitstreamReader) readNewTypeAndValue(reader PacketReader) (rbxfile.Value, error) {
	var val rbxfile.Value
	thisType, err := b.readUint8()
	if err != nil {
		return val, err
	}

	var enumID uint16
	if thisType == PROP_TYPE_ENUM {
		enumID, err = b.readUint16BE()
		if err != nil {
			return val, err
		}
	}

	val, err = b.ReadSerializedValue(reader, thisType, enumID)
	return val, err
}

func (b *BitstreamReader) readSerializedValueGeneric(reader PacketReader, valueType uint8, enumId uint16) (rbxfile.Value, error) {
	var err error
	var result rbxfile.Value
	var temp string
	switch valueType {
	case PROP_TYPE_INVALID: // I assume this is how it works, anyway
		result = nil
		err = nil
	case PROP_TYPE_STRING_NO_CACHE:
		temp, err = b.readVarLengthString()
		result = rbxfile.ValueString(temp)
	case PROP_TYPE_ENUM:
		result, err = b.readNewEnumValue(enumId, reader.Context())
	case PROP_TYPE_BINARYSTRING:
		result, err = b.readNewBinaryString()
	case PROP_TYPE_PBOOL:
		result, err = b.readPBool()
	case PROP_TYPE_PSINT:
		result, err = b.readNewPSint()
	case PROP_TYPE_PFLOAT:
		result, err = b.readPFloat()
	case PROP_TYPE_PDOUBLE:
		result, err = b.readPDouble()
	case PROP_TYPE_UDIM:
		result, err = b.readUDim()
	case PROP_TYPE_UDIM2:
		result, err = b.readUDim2()
	case PROP_TYPE_RAY:
		result, err = b.readRay()
	case PROP_TYPE_FACES:
		result, err = b.readFaces()
	case PROP_TYPE_AXES:
		result, err = b.readAxes()
	case PROP_TYPE_BRICKCOLOR:
		result, err = b.readBrickColor()
	case PROP_TYPE_COLOR3:
		result, err = b.readColor3()
	case PROP_TYPE_COLOR3UINT8:
		result, err = b.readColor3uint8()
	case PROP_TYPE_VECTOR2:
		result, err = b.readVector2()
	case PROP_TYPE_VECTOR3_SIMPLE:
		result, err = b.readVector3Simple()
	case PROP_TYPE_VECTOR3_COMPLICATED:
		result, err = b.readVector3()
	case PROP_TYPE_VECTOR2UINT16:
		result, err = b.readVector2int16()
	case PROP_TYPE_VECTOR3UINT16:
		result, err = b.readVector3int16()
	case PROP_TYPE_CFRAME_SIMPLE:
		result, err = b.readCFrameSimple()
	case PROP_TYPE_CFRAME_COMPLICATED:
		result, err = b.readCFrame()
	case PROP_TYPE_NUMBERSEQUENCE:
		result, err = b.readNumberSequence()
	case PROP_TYPE_NUMBERSEQUENCEKEYPOINT:
		result, err = b.readNumberSequenceKeypoint()
	case PROP_TYPE_NUMBERRANGE:
		result, err = b.readNumberRange()
	case PROP_TYPE_COLORSEQUENCE:
		result, err = b.readColorSequence()
	case PROP_TYPE_COLORSEQUENCEKEYPOINT:
		result, err = b.readColorSequenceKeypoint()
	case PROP_TYPE_RECT2D:
		result, err = b.readRect2D()
	case PROP_TYPE_PHYSICALPROPERTIES:
		result, err = b.readPhysicalProperties()
	case PROP_TYPE_REGION3:
		result, err = b.readRegion3()
	case PROP_TYPE_REGION3INT16:
		result, err = b.readRegion3int16()
	case PROP_TYPE_INT64:
		result, err = b.readInt64()
	}
	return result, err
}

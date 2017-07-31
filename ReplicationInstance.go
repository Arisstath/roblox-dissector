package main

type PropertyValue interface {
	//Show() *widgets.QWidget_ITF
}

type ReplicationProperty struct {
	Schema *PropertySchemaItem
	Value PropertyValue
	IsDefault bool
}

type ReplicationInstance struct {
	Referent string
	ReferentInt1 uint32
	Int1 uint32
	ClassName string
	Bool1 bool
	Referent2 string
	ReferentInt2 uint32
	Properties []*ReplicationProperty
}

func DecodeReplicationInstance(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, instanceSchema []*InstanceSchemaItem, classDescriptor Descriptor) (ReplicationInstance*, error) {
	thisInstance := &ReplicationInstance{}
	thisInstance.Referent, thisInstance.ReferentInt1, err = gzipStream.ReadJoinReferent()
	if err != nil {
		return layer, err
	}

	classIDx, err := gzipStream.Bits(9)
	if err != nil {
		return layer, err
	}
	realIDx := (classIDx & 1 << 8) | classIDx >> 1
	if int(realIDx) > int(len(instanceSchema)) {
		return layer, errors.New(fmt.Sprintf("idx %d is higher than %d", realIDx, len(context.InstanceSchema)))
	}
	thisInstance.ClassName = classDescriptor[uint32(realIDx)]
	println(DebugInfo(context, packet), "Our class: ", thisInstance.ClassName)

	thisPropertySchema := instanceSchema[realIDx].PropertySchema

	thisInstance.Bool1, err = gzipStream.ReadBool()
	if err != nil {
		return layer, err
	}

	for _, property := range thisPropertySchema {
		if !property.Bool1 {
			continue
		}
		var val PropertyValue = nil
		if property.Type == "bool" {
			if err != nil {
				return layer, err
			}
			val, err = gzipStream.ReadPBool()
			if err != nil {
				return layer, err
			}
		} else {
			isDefault, err := gzipStream.ReadBool()
			if err != nil {
				return layer, err
			}
			if isDefault {
				println(DebugInfo(context, packet), "Read", property.Name, "1 bit: default")
			} else {
				switch property.Type {
				case "string":
					val, err = gzipStream.ReadPString()
					break
				case "ProtectedString":
					val, err = gzipStream.ReadProtectedString()
					break
				case "BinaryString":
					val, err = gzipStream.ReadBinaryString()
					break
				case "int":
					val, err = gzipStream.ReadPInt()
					break
				case "float":
					val, err = gzipStream.ReadPFloat()
					break
				case "double":
					val, err = gzipStream.ReadPDouble()
					break
				case "Axes":
					val, err = gzipStream.ReadAxes()
					break
				case "Faces":
					val, err = gzipStream.ReadFaces()
					break
				case "BrickColor":
					val, err = gzipStream.ReadBrickColor()
					break
				case "Object":
					val, err = gzipStream.ReadObject()
					break
				case "UDim":
					val, err = gzipStream.ReadUDim()
					break
				case "UDim2":
					val, err = gzipStream.ReadUDim2()
					break
				case "Vector2":
					val, err = gzipStream.ReadVector2()
					break
				case "Vector3":
					val, err = gzipStream.ReadVector3()
					break
				case "Vector2uint16":
					val, err = gzipStream.ReadVector2uint16()
					break
				case "Vector3uint16":
					val, err = gzipStream.ReadVector3uint16()
					break
				case "Ray":
					val, err = gzipStream.ReadRay()
					break
				case "Color3":
					val, err = gzipStream.ReadColor3()
					break
				case "Color3uint8":
					val, err = gzipStream.ReadColor3uint8()
					break
				case "CoordinateFrame":
					val, err = gzipStream.ReadCFrame()
					break
				case "Content":
					val, err = gzipStream.ReadContent()
					break
				default:
					if property.IsEnum {
						val, err = gzipStream.ReadEnumValue(property.BitSize)
					} else {
						return layer, errors.New("joindata parser encountered unknown type")
					}
				}
				if val == nil {
					break
				}
				println(DebugInfo(context, packet), "Read", property.Name, spew.Sdump(val))
			}
		}
	}
	thisInstance.Referent2, thisInstance.ReferentInt2, err = gzipStream.ReadJoinReferent()
	if err != nil {
		return layer, err
	}

	println(DebugInfo(context, packet), "Read instance", thisInstance.Referent, ",", thisInstance.ReferentInt1, ",",thisInstance.Int1, ",", thisInstance.ClassName, ",", thisInstance.Referent2, ",", thisInstance.ReferentInt2)
}

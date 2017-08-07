package main
import "fmt"
import "github.com/google/gopacket"
import "errors"

type PropertyValue interface {
	//Show() *widgets.QWidget_ITF
}

type ReplicationInstance struct {
	Object1 Object
	Int1 uint32
	ClassName string
	Bool1 bool
	Object2 Object
	Properties []*ReplicationProperty
}

func DecodeReplicationInstance(isJoinData bool, thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, instanceSchema []*InstanceSchemaItem) (*ReplicationInstance, error) {
	var err error
	thisInstance := &ReplicationInstance{}
	thisInstance.Object1, err = thisBitstream.ReadObject(isJoinData, context)
	if err != nil {
		return thisInstance, err
	}

	classIDx, err := thisBitstream.Bits(9)
	if err != nil {
		return thisInstance, err
	}
	realIDx := (classIDx & 1 << 8) | classIDx >> 1
	if int(realIDx) > int(len(instanceSchema)) {
		return thisInstance, errors.New(fmt.Sprintf("idx %d is higher than %d", realIDx, len(context.InstanceSchema)))
	}
	thisInstance.ClassName = instanceSchema[realIDx].Name
	println(DebugInfo2(context, packet, isJoinData), "Read referent", thisInstance.Object1.Referent, thisInstance.ClassName)

	thisPropertySchema := instanceSchema[realIDx].PropertySchema

	thisInstance.Bool1, err = thisBitstream.ReadBool()
	if err != nil {
		return thisInstance, err
	}

	if isJoinData {
		for _, schema := range thisPropertySchema {
			property, err := schema.Decode(ROUND_JOINDATA, thisBitstream, context, packet)
			if err != nil {
				return thisInstance, err
			}
			if property != nil {
				thisInstance.Properties = append(thisInstance.Properties, property)
			}
		}
	} else {
		for _, schema := range thisPropertySchema {
			property, err := schema.Decode(ROUND_STRINGS, thisBitstream, context, packet)
			if err != nil {
				return thisInstance, err
			}
			if property != nil {
				thisInstance.Properties = append(thisInstance.Properties, property)
			}
		}
		for _, schema := range thisPropertySchema {
			property, err := schema.Decode(ROUND_OTHER, thisBitstream, context, packet)
			if err != nil {
				return thisInstance, err
			}
			if property != nil {
				thisInstance.Properties = append(thisInstance.Properties, property)
			}
		}
	}

	thisInstance.Object2, err = thisBitstream.ReadObject(isJoinData, context)
	if err != nil {
		return thisInstance, err
	}
	println(DebugInfo2(context, packet, isJoinData), "Parent referent", thisInstance.Object2.Referent)

	return thisInstance, nil
}

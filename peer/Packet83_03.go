package peer
import "errors"
import "fmt"

type Packet83_03 struct {
	Object1 Object
	Bool1 bool
	PropertyName string
	Value *ReplicationProperty
}

func DecodePacket83_03(packet *UDPPacket, context *CommunicationContext, propertySchema []*PropertySchemaItem) (interface{}, error) {
	var err error
	layer := &Packet83_03{}
	thisBitstream := packet.Stream
	layer.Object1, err = thisBitstream.ReadObject(false, context)
	if err != nil {
		return layer, err
	}

	if !context.UseStaticSchema {
		propertyIDx, err := thisBitstream.Bits(0xB)
		if err != nil {
			return layer, err
		}
		realIDx := (propertyIDx & 0x7 << 8) | propertyIDx >> 3

		if int(realIDx) > int(len(propertySchema)) {
			return layer, errors.New(fmt.Sprintf("prop idx %d is higher than %d", realIDx, len(propertySchema)))
		}

		schema := propertySchema[realIDx]
		layer.PropertyName = schema.Name

		layer.Bool1, err = thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}

		layer.Value, err = schema.Decode(ROUND_UPDATE, packet, context)
		return layer, err
	} else {
		propertyIDx, err := thisBitstream.ReadUint16BE()
		if err != nil {
			return layer, err
		}

		if int(propertyIDx) >= int(len(context.StaticSchema.Properties)) {
			return layer, errors.New(fmt.Sprintf("prop idx %d is higher than %d", propertyIDx, len(context.StaticSchema.Properties)))
		}
		schema := context.StaticSchema.Properties[propertyIDx]
		layer.PropertyName = schema.Name

		layer.Bool1, err = thisBitstream.ReadBool()
		if err != nil {
			return layer, err
		}

		layer.Value, err = schema.Decode(ROUND_UPDATE, packet, context, false)
		return layer, err
	}
}

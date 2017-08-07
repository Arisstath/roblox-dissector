package main
import "github.com/google/gopacket"

type Packet83_02 ReplicationInstance

func DecodePacket83_02(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet, instanceSchema []*InstanceSchemaItem) (interface{}, error) {
	return DecodeReplicationInstance(false, thisBitstream, context, packet, instanceSchema)
}

package peer
import "errors"
import "fmt"

type DecoderFunc func(*UDPPacket, *CommunicationContext) (interface{}, error)
var PacketDecoders = map[byte]DecoderFunc{
	0x05: DecodePacket05Layer,
	0x06: DecodePacket06Layer,
	0x07: DecodePacket07Layer,
	0x08: DecodePacket08Layer,
	0x00: DecodePacket00Layer,
	0x03: DecodePacket03Layer,
	0x09: DecodePacket09Layer,
	0x10: DecodePacket10Layer,
	0x13: DecodePacket13Layer,

	//0x8A: DecodePacket8ALayer,
	0x82: DecodePacket82Layer,
	0x93: DecodePacket93Layer,
	0x91: DecodePacket91Layer,
	0x92: DecodePacket92Layer,
	0x90: DecodePacket90Layer,
	0x8F: DecodePacket8FLayer,
	0x81: DecodePacket81Layer,
	0x83: DecodePacket83Layer,
	0x97: DecodePacket97Layer,
}

type ReceiveHandler func(byte, *UDPPacket, *PacketLayers)
type ErrorHandler func(error)

type PacketReader struct {
	SimpleHandler ReceiveHandler
	ReliableHandler ReceiveHandler
	FullReliableHandler ReceiveHandler
	ErrorHandler ErrorHandler
	Context *CommunicationContext
}

func (this *PacketReader) ReadPacket(payload []byte, packet *UDPPacket) {
	context := this.Context

	packet.Stream = BufferToStream(payload)
	rakNetLayer, err := DecodeRakNetLayer(payload[0], packet, context)
	if err != nil {
		this.ErrorHandler(err)
		return
	}
	if rakNetLayer.IsDuplicate {
		return
	}

	layers := &PacketLayers{RakNet: rakNetLayer}
	if rakNetLayer.IsSimple {
		var err error
		packetType := rakNetLayer.SimpleLayerID
		decoder := PacketDecoders[packetType]
		if decoder != nil {
			layers.Main, err = decoder(packet, context)
			if err != nil {
				this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode simple packet %02X: %s", packetType, err.Error())))
				return
			}
		}

		this.SimpleHandler(payload[0], packet, layers)
		return
	} else if !rakNetLayer.IsValid {
		this.ErrorHandler(errors.New(fmt.Sprintf("Sent invalid packet (packet header %x)", payload[0])))
		return
	} else if rakNetLayer.IsACK {
		// nop
	} else if !rakNetLayer.IsNAK {
		packet.Stream = rakNetLayer.Payload
		reliabilityLayer, err := DecodeReliabilityLayer(packet, context, rakNetLayer)
		if err != nil {
			this.ErrorHandler(errors.New("Failed to decode reliable packet: " + err.Error()))
			return
		}

		for _, subPacket := range reliabilityLayer.Packets {
			reliablePacketLayers := &PacketLayers{RakNet: layers.RakNet, Reliability: subPacket}

			this.ReliableHandler(subPacket.PacketType, packet, reliablePacketLayers)

			if subPacket.HasPacketType && !subPacket.HasBeenDecoded && subPacket.IsFinal {
				subPacket.HasBeenDecoded = true
				go func(subPacket *ReliablePacket, reliablePacketLayers *PacketLayers) {
					packetType := subPacket.PacketType
					_, err = subPacket.FullDataReader.ReadByte()
					if err != nil {
						this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliablePacket %02X: %s", packetType, err.Error())))
						return
					}
					newPacket := &UDPPacket{subPacket.FullDataReader, packet.Source, packet.Destination}

					decoder := PacketDecoders[packetType]
					if decoder != nil {
						reliablePacketLayers.Main, err = decoder(newPacket, context)

						if err != nil {
							this.ErrorHandler(errors.New(fmt.Sprintf("Failed to decode reliable packet %02X: %s", packetType, err.Error())))

							return
						}
					}
					this.FullReliableHandler(subPacket.PacketType, newPacket, reliablePacketLayers)
				}(subPacket, reliablePacketLayers)
			}
		}
	}
}

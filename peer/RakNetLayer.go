package peer
import "bytes"
import "io/ioutil"
import "errors"

const DEBUG bool = true
type RakNetPacket interface {
	Serialize(bool, *CommunicationContext, *ExtendedWriter) error
}

type PacketLayers struct {
	RakNet *RakNetLayer
	Reliability *ReliablePacket
	Timestamp *Packet1BLayer
	Main interface{}
}

type ACKRange struct {
	Min uint32
	Max uint32
}

type RakNetLayer struct {
	Payload *ExtendedReader
	IsSimple bool
	IsDuplicate bool
	SimpleLayerID uint8
	IsValid bool
	IsACK bool
	IsNAK bool
	HasBAndAS bool
	ACKs []ACKRange
	IsPacketPair bool
	IsContinuousSend bool
	NeedsBAndAS bool
	DatagramNumber uint32
}

func NewRakNetLayer() *RakNetLayer {
	return &RakNetLayer{}
}

var OfflineMessageID = []byte{0x00,0xFF,0xFF,0x00,0xFE,0xFE,0xFE,0xFE,0xFD,0xFD,0xFD,0xFD,0x12,0x34,0x56,0x78}

func DecodeRakNetLayer(packetType byte, packet *UDPPacket, context *CommunicationContext) (*RakNetLayer, error) {
	layer := NewRakNetLayer()
	bitstream := packet.Stream

	var err error
	if packetType == 0x5 {
		_, err = bitstream.ReadByte()
		if err != nil {
			return layer, err
		}
		thisOfflineMessage := make([]byte, 0x10)
		err = bitstream.Bytes(thisOfflineMessage, 0x10)
		if err != nil {
			return layer, err
		}

		if bytes.Compare(thisOfflineMessage, OfflineMessageID) != 0 {
			return layer, errors.New("offline message didn't match in packet 5!")
		}

		client := packet.Source
		server := packet.Destination
		println("Automatically detected Roblox peers using 0x5 packet:")
		println("Client:", client.String())
		println("Server:", server.String())
		context.SetClient(client.String())
		context.SetServer(server.String())
		layer.SimpleLayerID = packetType
		layer.Payload = bitstream
		layer.IsSimple = true
		return layer, nil
	} else if packetType >= 0x6 && packetType <= 0x8 {
		_, err = bitstream.ReadByte()
		if err != nil {
			return layer, err
		}
		layer.IsSimple = true
		layer.Payload = bitstream
		layer.SimpleLayerID = packetType
		return layer, nil
	}

	layer.IsValid, err = bitstream.ReadBool()
	if !layer.IsValid {
		return layer, nil
	}
	if err != nil {
		return layer, err
	}
	layer.IsACK, err = bitstream.ReadBool()
	if err != nil {
		return layer, err
	}
	if !layer.IsACK {
		layer.IsNAK, err = bitstream.ReadBool()
		if err != nil {
			return layer, err
		}
	}

	if layer.IsACK || layer.IsNAK {
		layer.HasBAndAS, err = bitstream.ReadBool()
		bitstream.Align()

		ackCount, err := bitstream.ReadUint16BE()
		if err != nil {
			return layer, err
		}
		var i uint16
		for i = 0; i < ackCount; i++ {
			var min, max uint32

			minEqualToMax, err := bitstream.ReadBoolByte()
			if err != nil {
				return layer, err
			}
			min, err = bitstream.ReadUint24LE()
			if err != nil {
				return layer, err
			}
			if minEqualToMax {
				max = min
			} else {
				max, err = bitstream.ReadUint24LE()
			}

			layer.ACKs = append(layer.ACKs, ACKRange{min, max})
		}
		return layer, nil
	} else {
		layer.IsPacketPair, err = bitstream.ReadBool()
		if err != nil {
			return layer, err
		}
		layer.IsContinuousSend, err = bitstream.ReadBool()
		if err != nil {
			return layer, err
		}
		layer.NeedsBAndAS, err = bitstream.ReadBool()
		if err != nil {
			return layer, err
		}
		bitstream.Align()

		layer.DatagramNumber, err = bitstream.ReadUint24LE()
		if err != nil {
			return layer, err
		}
		context.MUniques.Lock()
		if context.IsClient(packet.Source) {
			_, layer.IsDuplicate = context.UniqueDatagramsClient[layer.DatagramNumber]
			context.UniqueDatagramsClient[layer.DatagramNumber] = struct{}{}
		} else if context.IsServer(packet.Source) {
			_, layer.IsDuplicate = context.UniqueDatagramsServer[layer.DatagramNumber]
			context.UniqueDatagramsServer[layer.DatagramNumber] = struct{}{}
		}
		context.MUniques.Unlock()

		layer.Payload = bitstream
		return layer, nil
	}
}

func (layer *RakNetLayer) Serialize(isClient bool, context *CommunicationContext, outStream *ExtendedWriter) (error) {
	var err error
	err = outStream.WriteBool(layer.IsValid)
	if err != nil {
		return err
	}
	err = outStream.WriteBool(layer.IsACK)
	if err != nil {
		return err
	}
	if !layer.IsACK {
		err = outStream.WriteBool(layer.IsNAK)
		if err != nil {
			return err
		}
	}

	if layer.IsACK || layer.IsNAK {
		err = outStream.WriteBool(layer.HasBAndAS)
		if err != nil {
			return err
		}
		err = outStream.Align()
		if err != nil {
			return err
		}

		err = outStream.WriteUint16BE(uint16(len(layer.ACKs)))
		if err != nil {
			return err
		}

		for _, ack := range layer.ACKs {
			if ack.Min == ack.Max {
				err = outStream.WriteBoolByte(true)
				if err != nil {
					return err
				}
				err = outStream.WriteUint24LE(ack.Min)
				if err != nil {
					return err
				}
			} else {
				err = outStream.WriteBoolByte(false)
				if err != nil {
					return err
				}
				err = outStream.WriteUint24LE(ack.Min)
				if err != nil {
					return err
				}
				err = outStream.WriteUint24LE(ack.Max)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = outStream.WriteBool(layer.IsPacketPair)
		if err != nil {
			return err
		}
		err = outStream.WriteBool(layer.IsContinuousSend)
		if err != nil {
			return err
		}
		err = outStream.WriteBool(layer.NeedsBAndAS)
		if err != nil {
			return err
		}
		err = outStream.Align()
		if err != nil {
			return err
		}

		err = outStream.WriteUint24LE(layer.DatagramNumber)
		if err != nil {
			return err
		}
		
		content, err := ioutil.ReadAll(layer.Payload.GetReader())
		if err != nil {
			return err
		}
		err = outStream.AllBytes(content)
		if err != nil {
			return err
		}
	}
	return nil
}

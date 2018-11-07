// The peer package can be used for communication with Roblox servers, as well as
// parsing packets captured from Roblox network traffic.
package peer

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net"
)

// DEBUG decides whether debug mode should be on or not.
const DEBUG bool = true

// RakNetPacket describes any packet that can be serialized and written to UDP
type RakNetPacket interface {
	Serialize(writer PacketWriter, stream *extendedWriter) error
}

type RootLayer struct {
	logBuffer   bytes.Buffer
	Logger      *log.Logger
	Source      *net.UDPAddr
	Destination *net.UDPAddr
	FromClient  bool
	FromServer  bool
}

func (layer *RootLayer) GetLog() string {
	return layer.logBuffer.String()
}

// PacketLayers contains the different layers a packet can have.
type PacketLayers struct {
	// Root is the a basic layer containg information about a packet's source and destination
	Root RootLayer
	// RakNetLayer is the outermost layer. All packets have a RakNetLayer.
	RakNet *RakNetLayer
	// Most packets have a ReliabilityLayer. The exceptions to this are ACKs, NAKs and
	// pre-connection packets.
	Reliability *ReliablePacket
	// Contains data about the split packets this packet has.
	SplitPacket *SplitPacketBuffer
	// Timestamped packets (i.e. physics packets) may have a Timestamp layer.
	Timestamp *Packet1BLayer
	// Almost all packets have a Main layer. The exceptions to this are ACKs and NAKs.
	Main interface{}
	// Possible parsing error?
	Error error
}

// ACKRange describes the range of an ACK or an NAK.
type ACKRange struct {
	Min uint32
	Max uint32
}

// RakNetLayer is the outermost layer of all packets. It contains basic information
// about every packet.
type RakNetLayer struct {
	payload *extendedReader
	// Is the packet a simple pre-connection packet?
	IsSimple bool
	// If IsSimple is true, this is the packet type.
	SimpleLayerID uint8
	// Drop any non-simple packets which don't have IsValid set.
	IsValid          bool
	IsACK            bool
	IsNAK            bool
	HasBAndAS        bool
	ACKs             []ACKRange
	IsPacketPair     bool
	IsContinuousSend bool
	NeedsBAndAS      bool
	// A datagram number that is used to keep the packets in order.
	DatagramNumber uint32
}

func NewRakNetLayer() *RakNetLayer {
	return &RakNetLayer{}
}

// The offline message contained in pre-connection packets.
var OfflineMessageID = []byte{0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE, 0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78}

func IsOfflineMessage(data []byte) bool {
	if len(data) < 1+len(OfflineMessageID) {
		return false
	}
	return bytes.Compare(data[1:1+len(OfflineMessageID)], OfflineMessageID) == 0
}

func (bitstream *extendedReader) DecodeRakNetLayer(reader PacketReader, packetType byte, layers *PacketLayers) (*RakNetLayer, error) {
	layer := NewRakNetLayer()

	var err error
	if packetType == 0x5 {
		_, err = bitstream.ReadByte()
		if err != nil {
			return layer, err
		}
		thisOfflineMessage := make([]byte, 0x10)
		err = bitstream.bytes(thisOfflineMessage, 0x10)
		if err != nil {
			return layer, err
		}

		if bytes.Compare(thisOfflineMessage, OfflineMessageID) != 0 {
			return layer, errors.New("offline message didn't match in packet 5")
		}

		layer.SimpleLayerID = packetType
		layer.payload = bitstream
		layer.IsSimple = true
		return layer, nil
	} else if packetType >= 0x6 && packetType <= 0x8 {
		_, err = bitstream.ReadByte()
		if err != nil {
			return layer, err
		}
		layer.IsSimple = true
		layer.payload = bitstream
		layer.SimpleLayerID = packetType
		return layer, nil
	}

	layer.IsValid, err = bitstream.readBool()
	if !layer.IsValid {
		return layer, nil
	}
	if err != nil {
		return layer, err
	}
	layer.IsACK, err = bitstream.readBool()
	if err != nil {
		return layer, err
	}
	if !layer.IsACK {
		layer.IsNAK, err = bitstream.readBool()
		if err != nil {
			return layer, err
		}
	}

	if layer.IsACK || layer.IsNAK {
		layer.HasBAndAS, err = bitstream.readBool()
		bitstream.Align()

		ackCount, err := bitstream.readUint16BE()
		if err != nil {
			return layer, err
		}
		var i uint16
		for i = 0; i < ackCount; i++ {
			var min, max uint32

			minEqualToMax, err := bitstream.readBoolByte()
			if err != nil {
				return layer, err
			}
			min, err = bitstream.readUint24LE()
			if err != nil {
				return layer, err
			}
			if minEqualToMax {
				max = min
			} else {
				max, err = bitstream.readUint24LE()
			}

			layer.ACKs = append(layer.ACKs, ACKRange{min, max})
		}
		return layer, nil
	} else {
		layer.IsPacketPair, err = bitstream.readBool()
		if err != nil {
			return layer, err
		}
		layer.IsContinuousSend, err = bitstream.readBool()
		if err != nil {
			return layer, err
		}
		layer.NeedsBAndAS, err = bitstream.readBool()
		if err != nil {
			return layer, err
		}
		bitstream.Align()

		layer.DatagramNumber, err = bitstream.readUint24LE()
		if err != nil {
			return layer, err
		}

		layer.payload = bitstream
		return layer, nil
	}
}

func (layer *RakNetLayer) Serialize(writer PacketWriter, outStream *extendedWriter) error {
	var err error
	err = outStream.writeBool(layer.IsValid)
	if err != nil {
		return err
	}
	err = outStream.writeBool(layer.IsACK)
	if err != nil {
		return err
	}
	if !layer.IsACK {
		err = outStream.writeBool(layer.IsNAK)
		if err != nil {
			return err
		}
	}

	if layer.IsACK || layer.IsNAK {
		err = outStream.writeBool(layer.HasBAndAS)
		if err != nil {
			return err
		}
		err = outStream.Align()
		if err != nil {
			return err
		}

		err = outStream.writeUint16BE(uint16(len(layer.ACKs)))
		if err != nil {
			return err
		}

		for _, ack := range layer.ACKs {
			if ack.Min == ack.Max {
				err = outStream.writeBoolByte(true)
				if err != nil {
					return err
				}
				err = outStream.writeUint24LE(ack.Min)
				if err != nil {
					return err
				}
			} else {
				err = outStream.writeBoolByte(false)
				if err != nil {
					return err
				}
				err = outStream.writeUint24LE(ack.Min)
				if err != nil {
					return err
				}
				err = outStream.writeUint24LE(ack.Max)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = outStream.writeBool(layer.IsPacketPair)
		if err != nil {
			return err
		}
		err = outStream.writeBool(layer.IsContinuousSend)
		if err != nil {
			return err
		}
		err = outStream.writeBool(layer.NeedsBAndAS)
		if err != nil {
			return err
		}
		err = outStream.Align()
		if err != nil {
			return err
		}

		err = outStream.writeUint24LE(layer.DatagramNumber)
		if err != nil {
			return err
		}

		content, err := ioutil.ReadAll(layer.payload.GetReader())
		if err != nil {
			return err
		}
		err = outStream.allBytes(content)
		if err != nil {
			return err
		}
	}
	return nil
}

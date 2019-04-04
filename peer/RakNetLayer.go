// The peer package can be used for communication with Roblox servers, as well as
// parsing packets captured from Roblox network traffic.
package peer

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

func bufferToStream(buffer []byte) *extendedReader {
	return &extendedReader{bytes.NewReader(buffer)}
}

func bitsToBytes(bits uint) uint {
	return (bits + 7) >> 3
}

func bytesToBits(bytes uint) uint {
	return bytes << 3
}

// RakNetPacket describes any packet that can be serialized and written to UDP
type RakNetPacket interface {
	fmt.Stringer
	Serialize(writer PacketWriter, stream *extendedWriter) error
	TypeString() string
	Type() byte
}

type RootLayer struct {
	logBuffer   *strings.Builder
	Logger      *log.Logger
	Source      *net.UDPAddr
	Destination *net.UDPAddr
	FromClient  bool
	FromServer  bool
}

func (layer *RootLayer) GetLog() string {
	if layer.logBuffer != nil {
		return layer.logBuffer.String()
	}
	return ""
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
	Main RakNetPacket
	// Possible parsing error?
	Error error

	PacketType byte
	Subpacket  Packet83Subpacket
}

var PacketNames = map[byte]string{
	0xFF: "???",
	0x00: "ID_CONNECTED_PING",
	0x01: "ID_UNCONNECTED_PING",
	0x03: "ID_CONNECTED_PONG",
	0x04: "ID_DETECT_LOST_CONNECTIONS",
	0x05: "ID_OPEN_CONNECTION_REQUEST_1",
	0x06: "ID_OPEN_CONNECTION_REPLY_1",
	0x07: "ID_OPEN_CONNECTION_REQUEST_2",
	0x08: "ID_OPEN_CONNECTION_REPLY_2",
	0x09: "ID_CONNECTION_REQUEST",
	0x10: "ID_CONNECTION_REQUEST_ACCEPTED",
	0x11: "ID_CONNECTION_ATTEMPT_FAILED",
	0x13: "ID_NEW_INCOMING_CONNECTION",
	0x15: "ID_DISCONNECTION_NOTIFICATION",
	0x18: "ID_INVALID_PASSWORD",
	0x1B: "ID_TIMESTAMP",
	0x1C: "ID_UNCONNECTED_PONG",
	0x81: "ID_SET_GLOBALS",
	0x82: "ID_TEACH_DESCRIPTOR_DICTIONARIES",
	0x83: "ID_DATA",
	0x84: "ID_MARKER",
	0x85: "ID_PHYSICS",
	0x86: "ID_TOUCHES",
	0x87: "ID_CHAT_ALL",
	0x88: "ID_CHAT_TEAM",
	0x89: "ID_REPORT_ABUSE",
	0x8A: "ID_SUBMIT_TICKET",
	0x8B: "ID_CHAT_GAME",
	0x8C: "ID_CHAT_PLAYER",
	0x8D: "ID_CLUSTER",
	0x8E: "ID_PROTOCOL_MISMATCH",
	0x8F: "ID_PREFERRED_SPAWN_NAME",
	0x90: "ID_PROTOCOL_SYNC",
	0x91: "ID_SCHEMA_SYNC",
	0x92: "ID_PLACEID_VERIFICATION",
	0x93: "ID_DICTIONARY_FORMAT",
	0x94: "ID_HASH_MISMATCH",
	0x95: "ID_SECURITYKEY_MISMATCH",
	0x96: "ID_REQUEST_STATS",
	0x97: "ID_NEW_SCHEMA",
}

func (layers *PacketLayers) String() string {
	if layers.Subpacket != nil {
		return layers.Subpacket.String()
	}
	if layers.Main != nil {
		return layers.Main.String()
	} else {
		packetName, ok := PacketNames[layers.PacketType]
		if ok {
			return packetName
		}
	}
	return fmt.Sprintf("Packet 0x%02X", layers.PacketType)
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
	Flags RakNetFlags
	ACKs  []ACKRange
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

func (stream *extendedReader) DecodeRakNetLayer(reader PacketReader, packetType byte, layers *PacketLayers) (*RakNetLayer, error) {
	layer := NewRakNetLayer()

	var err error
	if packetType == 0x5 {
		_, err = stream.ReadByte()
		if err != nil {
			return layer, err
		}
		thisOfflineMessage := make([]byte, 0x10)
		err = stream.bytes(thisOfflineMessage, 0x10)
		if err != nil {
			return layer, err
		}

		if bytes.Compare(thisOfflineMessage, OfflineMessageID) != 0 {
			return layer, errors.New("offline message didn't match in packet 5")
		}

		layer.SimpleLayerID = packetType
		layer.payload = stream
		layer.IsSimple = true
		return layer, nil
	} else if packetType >= 0x6 && packetType <= 0x8 {
		_, err = stream.ReadByte()
		if err != nil {
			return layer, err
		}
		layer.IsSimple = true
		layer.payload = stream
		layer.SimpleLayerID = packetType
		return layer, nil
	}

	layer.Flags, err = stream.readRakNetFlags()
	if err != nil {
		return layer, err
	}
	if !layer.Flags.IsValid {
		return layer, errors.New("layer not a valid RakNet packet")
	}

	if layer.Flags.IsACK || layer.Flags.IsNAK {
		ackCount, err := stream.readUint16BE()
		if err != nil {
			return layer, err
		}
		var i uint16
		for i = 0; i < ackCount; i++ {
			var min, max uint32

			minEqualToMax, err := stream.readBoolByte()
			if err != nil {
				return layer, err
			}
			min, err = stream.readUint24LE()
			if err != nil {
				return layer, err
			}
			if minEqualToMax {
				max = min
			} else {
				max, err = stream.readUint24LE()
			}

			layer.ACKs = append(layer.ACKs, ACKRange{min, max})
		}
		return layer, nil
	} else {
		layer.DatagramNumber, err = stream.readUint24LE()
		if err != nil {
			return layer, err
		}

		layer.payload = stream
		return layer, nil
	}
}

func (layer *RakNetLayer) Serialize(writer PacketWriter, outStream *extendedWriter) error {
	err := outStream.writeRakNetFlags(layer.Flags)
	if err != nil {
		return err
	}

	if layer.Flags.IsACK || layer.Flags.IsNAK {
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
		err = outStream.writeUint24LE(layer.DatagramNumber)
		if err != nil {
			return err
		}

		content, err := ioutil.ReadAll(layer.payload)
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

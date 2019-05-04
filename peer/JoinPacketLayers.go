package peer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
)

// PasswordType describes a RakNet password type
type PasswordType int

const (
	// DefaultPassword refers to the default type used for most connections
	DefaultPassword PasswordType = iota
	// StudioPassword refers to the default type used for Roblox Studio connections
	StudioPassword
	// InvalidPassword means that the password type couldn't be identified
	InvalidPassword
)

// StudioPasswordBytes is the RakNet password used for Studio connections by default
var StudioPasswordBytes = []byte{0x5E, 0x11}

// DefaultPasswordBytes is the RakNet password used for Roblox connections by default
var DefaultPasswordBytes = []byte{0x37, 0x4F, 0x5E, 0x11, 0x6C, 0x45}

// IdentifyPassword identifies what RakNet password is being used
func IdentifyPassword(password []byte) PasswordType {
	switch {
	case bytes.Equal(password, StudioPasswordBytes):
		return StudioPassword
	case bytes.Equal(password, DefaultPasswordBytes):
		return DefaultPassword
	default:
		return InvalidPassword
	}
}

// Packet05Layer represents ID_OPEN_CONNECTION_REQUEST_1 - client -> server
type Packet05Layer struct {
	// RakNet protocol version, always 5
	ProtocolVersion  uint8
	MTUPaddingLength int
}

// Packet06Layer represents ID_OPEN_CONNECTION_REPLY_1 - server -> client
type Packet06Layer struct {
	// Server GUID
	GUID uint64
	// Use libcat encryption? Always false
	UseSecurity bool
	// MTU in bytes
	MTU uint16
}

// Packet07Layer represents ID_OPEN_CONNECTION_REQUEST_2 - client -> server
type Packet07Layer struct {
	// Server external IP address
	IPAddress *net.UDPAddr
	// MTU in bytes
	MTU uint16
	// Client GUID
	GUID uint64
}

// Packet08Layer represents ID_OPEN_CONNECTION_REPLY_2 - server -> client
type Packet08Layer struct {
	// Server GUID
	GUID uint64
	// Client external IP address
	IPAddress *net.UDPAddr
	// MTU in bytes
	MTU uint16
	// Use libcat encryption? Always false
	UseSecurity bool
}

// Packet09Layer represents ID_CONNECTION_REQUEST - client -> server
type Packet09Layer struct {
	// Client GUID
	GUID uint64
	// Timestamp of sending the request (seconds)
	Timestamp uint64
	// Use libcat encryption? Always false
	UseSecurity bool
	// Password: 2 or 6 bytes, always {0x5E, 0x11} in Studio, varies in real clients
	Password []byte
}

// Packet10Layer represents ID_CONNECTION_REQUEST_ACCEPTED - server -> client
type Packet10Layer struct {
	// Client IP address
	IPAddress   *net.UDPAddr
	SystemIndex uint16
	Addresses   [10]*net.UDPAddr
	// Timestamp from ID_CONNECTION_REQUEST
	SendPingTime uint64
	// Timestamp of sending reply (seconds)
	SendPongTime uint64
}

// Packet13Layer represents ID_NEW_INCOMING_CONNECTION - client -> server
type Packet13Layer struct {
	// Server IP address
	IPAddress *net.UDPAddr
	Addresses [10]*net.UDPAddr
	// SendPongTime from ID_CONNECTION_REQUEST_ACCEPTED
	SendPingTime uint64
	// Timestamp of sending reply (seconds)
	SendPongTime uint64
}

func (thisStream *extendedReader) DecodePacket05Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	var err error
	layer := &Packet05Layer{}
	layer.ProtocolVersion, err = thisStream.readUint8() // !! RakNetLayer will have read the offline message !!
	mtupad, err := ioutil.ReadAll(thisStream)
	if err != nil {
		return layer, err
	}
	layer.MTUPaddingLength = len(mtupad)

	return layer, nil
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet05Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x05)
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.WriteByte(byte(layer.ProtocolVersion))
	if err != nil {
		return err
	}
	empty := make([]byte, layer.MTUPaddingLength)
	err = stream.allBytes(empty)
	return err
}

func (layer *Packet05Layer) String() string {
	return fmt.Sprintf("ID_OPEN_CONNECTION_REQUEST_1: Version %d", layer.ProtocolVersion)
}

// TypeString implements RakNetPacket.TypeString()
func (Packet05Layer) TypeString() string {
	return "ID_OPEN_CONNECTION_REQUEST_1"
}

// Type implements RakNetPacket.Type()
func (Packet05Layer) Type() byte {
	return 5
}

func (thisStream *extendedReader) DecodePacket06Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet06Layer{}

	var err error
	layer.GUID, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisStream.readUint16BE()
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet06Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x06)
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.GUID)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(layer.UseSecurity)
	if err != nil {
		return err
	}
	err = stream.writeUint16BE(layer.MTU)
	return err
}
func (layer *Packet06Layer) String() string {
	return "ID_OPEN_CONNECTION_REPLY_1"
}

// TypeString impelements RakNetPacket.TypeString()
func (Packet06Layer) TypeString() string {
	return "ID_OPEN_CONNECTION_REPLY_1"
}

// Type implements RakNetPacket.Type()
func (Packet06Layer) Type() byte {
	return 6
}

func (thisStream *extendedReader) DecodePacket07Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet07Layer{}

	var err error
	layer.IPAddress, err = thisStream.readAddress()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisStream.readUint16BE()
	if err != nil {
		return layer, err
	}
	layer.GUID, err = thisStream.readUint64BE()
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet07Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x07)
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.writeAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	err = stream.writeUint16BE(layer.MTU)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.GUID)
	return err
}
func (layer *Packet07Layer) String() string {
	return "ID_OPEN_CONNECTION_REQUEST_2"
}

// TypeString impelements RakNetPacket.TypeString()
func (Packet07Layer) TypeString() string {
	return "ID_OPEN_CONNECTION_REQUEST_2"
}

// Type implements RakNetPacket.Type()
func (Packet07Layer) Type() byte {
	return 7
}

func (thisStream *extendedReader) DecodePacket08Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet08Layer{}

	var err error
	layer.GUID, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.IPAddress, err = thisStream.readAddress()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisStream.readUint16BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisStream.readBoolByte()
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet08Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x08)
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.GUID)
	if err != nil {
		return err
	}
	err = stream.writeAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	err = stream.writeUint16BE(layer.MTU)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(layer.UseSecurity)
	return err
}
func (layer *Packet08Layer) String() string {
	return "ID_OPEN_CONNECTION_REPLY_2"
}

// TypeString impelements RakNetPacket.TypeString()
func (Packet08Layer) TypeString() string {
	return "ID_OPEN_CONNECTION_REPLY_2"
}

// Type implements RakNetPacket.Type()
func (Packet08Layer) Type() byte {
	return 8
}

func (thisStream *extendedReader) DecodePacket09Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet09Layer{}

	var err error
	layer.GUID, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.Timestamp, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisStream.readBoolByte()
	if err != nil {
		return layer, err
	}
	// 2x 64 for timestamps, 8 for UseSecurity and 8 for PacketType

	layer.Password, err = ioutil.ReadAll(thisStream)
	if err != nil {
		return layer, err
	}
	if IdentifyPassword(layer.Password) == StudioPassword {
		layers.Root.Logger.Println("Detected Studio!")
		reader.Context().IsStudio = true
	}
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet09Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x09)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.GUID)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.Timestamp)
	if err != nil {
		return err
	}
	err = stream.writeBoolByte(layer.UseSecurity)
	if err != nil {
		return err
	}
	return stream.allBytes(layer.Password)
}
func (layer *Packet09Layer) String() string {
	return fmt.Sprintf("ID_CONNECTION_REQUEST: Password %X", layer.Password)
}

// TypeString impelements RakNetPacket.TypeString()
func (Packet09Layer) TypeString() string {
	return "ID_CONNECTION_REQUEST"
}

// Type implements RakNetPacket.Type()
func (Packet09Layer) Type() byte {
	return 9
}

func (thisStream *extendedReader) DecodePacket10Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet10Layer{}

	var err error
	layer.IPAddress, err = thisStream.readAddress()
	if err != nil {
		return layer, err
	}
	layer.SystemIndex, err = thisStream.readUint16BE()
	if err != nil {
		return layer, err
	}
	for i := 0; i < 10; i++ {
		layer.Addresses[i], err = thisStream.readAddress()
		if err != nil {
			return layer, err
		}
	}
	layer.SendPingTime, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisStream.readUint64BE()
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet10Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x10)
	if err != nil {
		return err
	}

	err = stream.writeAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	err = stream.writeUint16BE(layer.SystemIndex)
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		err = stream.writeAddress(layer.Addresses[i])
		if err != nil {
			return err
		}
	}
	err = stream.writeUint64BE(layer.SendPingTime)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.SendPongTime)
	if err != nil {
		return err
	}
	return err
}
func (layer *Packet10Layer) String() string {
	return "ID_CONNECTION_ACCEPTED"
}

// TypeString impelements RakNetPacket.TypeString()
func (Packet10Layer) TypeString() string {
	return "ID_CONNECTION_ACCEPTED"
}

// Type implements RakNetPacket.Type()
func (Packet10Layer) Type() byte {
	return 0x10
}

func (thisStream *extendedReader) DecodePacket13Layer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet13Layer{}

	var err error
	layer.IPAddress, err = thisStream.readAddress()
	if err != nil {
		return layer, err
	}
	for i := 0; i < 10; i++ {
		layer.Addresses[i], err = thisStream.readAddress()
		if err != nil {
			return layer, err
		}
	}
	layer.SendPingTime, err = thisStream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisStream.readUint64BE()
	return layer, err
}

// Serialize implements RakNetPacket.Serialize()
func (layer *Packet13Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(0x13)
	if err != nil {
		return err
	}

	err = stream.writeAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		err = stream.writeAddress(layer.Addresses[i])
		if err != nil {
			return err
		}
	}
	err = stream.writeUint64BE(layer.SendPingTime)
	if err != nil {
		return err
	}
	err = stream.writeUint64BE(layer.SendPongTime)
	if err != nil {
		return err
	}
	return err
}
func (layer *Packet13Layer) String() string {
	return "ID_NEW_INCOMING_CONNECTION"
}

// TypeString impelements RakNetPacket.TypeString()
func (Packet13Layer) TypeString() string {
	return "ID_NEW_INCOMING_CONNECTION"
}

// Type implements RakNetPacket.Type()
func (Packet13Layer) Type() byte {
	return 0x13
}

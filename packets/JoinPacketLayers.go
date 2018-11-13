package packets

import (
	"net"
    "github.com/gskartwii/roblox-dissector/util"
)

// ID_OPEN_CONNECTION_REQUEST_1 - client -> server
type ConnectionRequest1 struct {
	// RakNet protocol version, always 5
	ProtocolVersion uint8
	// internal
	maxLength uint
}

// ID_OPEN_CONNECTION_REPLY_1 - server -> client
type ConnectionReply1 struct {
	// Server GUID
	GUID uint64
	// Use libcat encryption? Always false
	UseSecurity bool
	// MTU in bytes
	MTU uint16
}

// ID_OPEN_CONNECTION_REQUEST_2 - client -> server
type ConnectionRequest2 struct {
	// Server external IP address
	IPAddress *net.UDPAddr
	// MTU in bytes
	MTU uint16
	// Client GUID
	GUID uint64
}

// ID_OPEN_CONNECTION_REPLY_2 - server -> client
type ConnectionReply2 struct {
	// Server GUID
	GUID uint64
	// Client external IP address
	IPAddress *net.UDPAddr
	// MTU in bytes
	MTU uint16
	// Use libcat encryption? Always false
	UseSecurity bool
}

// ID_CONNECTION_REQUEST - client -> server
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

// ID_CONNECTION_REQUEST_ACCEPTED - server -> client
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

// ID_NEW_INCOMING_CONNECTION - client -> server
type Packet13Layer struct {
	// Server IP address
	IPAddress *net.UDPAddr
	Addresses [10]*net.UDPAddr
	// SendPongTime from ID_CONNECTION_REQUEST_ACCEPTED
	SendPingTime uint64
	// Timestamp of sending reply (seconds)
	SendPongTime uint64
}

func NewConnectionRequest1() *ConnectionRequest1 {
	return &ConnectionRequest1{}
}
func NewConnectionReply1() *ConnectionReply1 {
	return &ConnectionReply1{}
}
func NewConnectionRequest2() *ConnectionRequest2 {
	return &ConnectionRequest2{}
}
func NewConnectionReply2() *ConnectionReply2 {
	return &ConnectionReply2{}
}
func NewPacket09Layer() *Packet09Layer {
	return &Packet09Layer{}
}
func NewPacket10Layer() *Packet10Layer {
	return &Packet10Layer{}
}
func NewPacket13Layer() *Packet13Layer {
	return &Packet13Layer{}
}

var voidOfflineMessage []byte = make([]byte, 0x10)

func (thisBitstream *PacketReaderBitstream) DecodeConnectionRequest1(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	var err error
	layer := NewConnectionRequest1()
	layer.ProtocolVersion, err = thisBitstream.ReadUint8() // !! RakNetLayer will have read the offline message !!
	return layer, err
}

func (layer *ConnectionRequest1) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
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
	empty := make([]byte, 1492-0x10-2-0x2A)
	err = stream.allBytes(empty)
	return err
}

func (thisBitstream *PacketReaderBitstream) DecodeConnectionReply1(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewConnectionReply1()

	var err error
	err = thisBitstream.bytes(voidOfflineMessage, 0x10)
	if err != nil {
		return layer, err
	}
	layer.GUID, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisBitstream.ReadUint16BE()
	return layer, err
}

func (layer *ConnectionReply1) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x06)
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.GUID)
	if err != nil {
		return err
	}
	err = stream.WriteBoolByte(layer.UseSecurity)
	if err != nil {
		return err
	}
	err = stream.WriteUint16BE(layer.MTU)
	return err
}

func (thisBitstream *PacketReaderBitstream) DecodeConnectionRequest2(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewConnectionRequest2()

	var err error
	err = thisBitstream.bytes(voidOfflineMessage, 0x10)
	if err != nil {
		return layer, err
	}
	layer.IPAddress, err = thisBitstream.ReadAddress()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisBitstream.ReadUint16BE()
	if err != nil {
		return layer, err
	}
	layer.GUID, err = thisBitstream.ReadUint64BE()
	return layer, err
}

func (layer *ConnectionRequest2) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x07)
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.WriteAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	err = stream.WriteUint16BE(layer.MTU)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.GUID)
	return err
}

func (thisBitstream *PacketReaderBitstream) DecodeConnectionReply2(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewConnectionReply2()
	

	var err error
	err = thisBitstream.bytes(voidOfflineMessage, 0x10)
	if err != nil {
		return layer, err
	}
	layer.GUID, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.IPAddress, err = thisBitstream.ReadAddress()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisBitstream.ReadUint16BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisBitstream.ReadBoolByte()
	return layer, err
}

func (layer *ConnectionReply2) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x08)
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.GUID)
	if err != nil {
		return err
	}
	err = stream.WriteAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	err = stream.WriteUint16BE(layer.MTU)
	if err != nil {
		return err
	}
	err = stream.WriteBoolByte(layer.UseSecurity)
	return err
}

func (thisBitstream *PacketReaderBitstream) DecodePacket09Layer(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket09Layer()
	

	var err error
	layer.GUID, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.Timestamp, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return layer, err
	}
	layer.Password, err = thisBitstream.ReadString(2)
	if layer.Password[0] == 0x5E && layer.Password[1] == 0x11 {
		reader.Context().IsStudio = true
	}
	return layer, err
}
func (layer *Packet09Layer) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x09)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.GUID)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.Timestamp)
	if err != nil {
		return err
	}
	err = stream.WriteBoolByte(layer.UseSecurity)
	if err != nil {
		return err
	}
	return stream.allBytes(layer.Password)
}

func (thisBitstream *PacketReaderBitstream) DecodePacket10Layer(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket10Layer()
	

	var err error
	layer.IPAddress, err = thisBitstream.ReadAddress()
	if err != nil {
		return layer, err
	}
	layer.SystemIndex, err = thisBitstream.ReadUint16BE()
	if err != nil {
		return layer, err
	}
	for i := 0; i < 10; i++ {
		layer.Addresses[i], err = thisBitstream.ReadAddress()
		if err != nil {
			return layer, err
		}
	}
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.ReadUint64BE()
	return layer, err
}
func (layer *Packet10Layer) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x10)
	if err != nil {
		return err
	}

	err = stream.WriteAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	err = stream.WriteUint16BE(layer.SystemIndex)
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		err = stream.WriteAddress(layer.Addresses[i])
		if err != nil {
			return err
		}
	}
	err = stream.WriteUint64BE(layer.SendPingTime)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.SendPongTime)
	if err != nil {
		return err
	}
	return err
}

func (thisBitstream *PacketReaderBitstream) DecodePacket13Layer(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket13Layer()
	

	var err error
	layer.IPAddress, err = thisBitstream.ReadAddress()
	if err != nil {
		return layer, err
	}
	for i := 0; i < 10; i++ {
		layer.Addresses[i], err = thisBitstream.ReadAddress()
		if err != nil {
			return layer, err
		}
	}
	layer.SendPingTime, err = thisBitstream.ReadUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.ReadUint64BE()
	return layer, err
}
func (layer *Packet13Layer) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteByte(0x13)
	if err != nil {
		return err
	}

	err = stream.WriteAddress(layer.IPAddress)
	if err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		err = stream.WriteAddress(layer.Addresses[i])
		if err != nil {
			return err
		}
	}
	err = stream.WriteUint64BE(layer.SendPingTime)
	if err != nil {
		return err
	}
	err = stream.WriteUint64BE(layer.SendPongTime)
	if err != nil {
		return err
	}
	return err
}

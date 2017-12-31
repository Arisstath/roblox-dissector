package peer
import "net"

// ID_OPEN_CONNECTION_REQUEST_1 - client -> server
type Packet05Layer struct {
	// RakNet protocol version, always 5
	ProtocolVersion uint8
}
// ID_OPEN_CONNECTION_REPLY_1 - server -> client
type Packet06Layer struct {
	// Server GUID
	GUID uint64
	// Use libcat encryption? Always false
	UseSecurity bool
	// MTU in bytes
	MTU uint16
}
// ID_OPEN_CONNECTION_REQUEST_2 - client -> server
type Packet07Layer struct {
	// Server external IP address
	IPAddress *net.UDPAddr
	// MTU in bytes
	MTU uint16
	// Client GUID
	GUID uint64
}
// ID_OPEN_CONNECTION_REPLY_2 - server -> client
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
	IPAddress *net.UDPAddr
	SystemIndex uint16
	Addresses [10]*net.UDPAddr
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

func NewPacket05Layer() *Packet05Layer {
	return &Packet05Layer{}
}
func NewPacket06Layer() *Packet06Layer {
	return &Packet06Layer{}
}
func NewPacket07Layer() *Packet07Layer {
	return &Packet07Layer{}
}
func NewPacket08Layer() *Packet08Layer {
	return &Packet08Layer{}
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

func decodePacket05Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	layer := NewPacket05Layer()
	thisBitstream := packet.stream
	layer.ProtocolVersion, err = thisBitstream.readUint8() // !! RakNetLayer will have read the offline message !!
	return layer, err
}

func (layer *Packet05Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
	err := stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.WriteByte(byte(layer.ProtocolVersion))
	if err != nil {
		return err
	}
	empty := make([]byte, 1492 - 0x10 - 1 - 0x1C)
	err = stream.allBytes(empty)
	return err
}

func decodePacket06Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket06Layer()
	thisBitstream := packet.stream

	var err error
	err = thisBitstream.bytes(voidOfflineMessage, 0x10)
	if err != nil {
		return layer, err
	}
	layer.GUID, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisBitstream.readUint16BE()
	return layer, err
}

func (layer *Packet06Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
	var err error
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

func decodePacket07Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket07Layer()
	thisBitstream := packet.stream

	var err error
	err = thisBitstream.bytes(voidOfflineMessage, 0x10)
	if err != nil {
		return layer, err
	}
	layer.IPAddress, err = thisBitstream.readAddress()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}
	layer.GUID, err = thisBitstream.readUint64BE()
	return layer, err
}

func (layer *Packet07Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
	var err error
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

func decodePacket08Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket08Layer()
	thisBitstream := packet.stream

	var err error
	err = thisBitstream.bytes(voidOfflineMessage, 0x10)
	if err != nil {
		return layer, err
	}
	layer.GUID, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.IPAddress, err = thisBitstream.readAddress()
	if err != nil {
		return layer, err
	}
	layer.MTU, err = thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisBitstream.readBoolByte()
	return layer, err
}

func (layer *Packet08Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
	var err error
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

func decodePacket09Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket09Layer()
	thisBitstream := packet.stream

	var err error
	layer.GUID, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.Timestamp, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.UseSecurity, err = thisBitstream.readBoolByte()
	if err != nil {
		return layer, err
	}
	layer.Password, err = thisBitstream.readString(2)
	if layer.Password[0] == 0x5E && layer.Password[1] == 0x11 {
		context.IsStudio = true
	}
	return layer, err
}
func (layer *Packet09Layer) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
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

func decodePacket10Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket10Layer()
	thisBitstream := packet.stream

	var err error
	layer.IPAddress, err = thisBitstream.readAddress()
	if err != nil {
		return layer, err
	}
	layer.SystemIndex, err = thisBitstream.readUint16BE()
	if err != nil {
		return layer, err
	}
	for i := 0; i < 10; i++ {
		layer.Addresses[i], err = thisBitstream.readAddress()
		if err != nil {
			return layer, err
		}
	}
	layer.SendPingTime, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.readUint64BE()
	return layer, err
}
func (layer *Packet10Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
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

func decodePacket13Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket13Layer()
	thisBitstream := packet.stream

	var err error
	layer.IPAddress, err = thisBitstream.readAddress()
	if err != nil {
		return layer, err
	}
	for i := 0; i < 10; i++ {
		layer.Addresses[i], err = thisBitstream.readAddress()
		if err != nil {
			return layer, err
		}
	}
	layer.SendPingTime, err = thisBitstream.readUint64BE()
	if err != nil {
		return layer, err
	}
	layer.SendPongTime, err = thisBitstream.readUint64BE()
	return layer, err
}
func (layer *Packet13Layer) serialize(isClient bool,context *CommunicationContext, stream *extendedWriter) error {
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

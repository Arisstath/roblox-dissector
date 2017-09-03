package peer
import "net"

type Packet05Layer struct {
	ProtocolVersion uint8
}
type Packet06Layer struct {
	GUID uint64
	UseSecurity bool
	MTU uint16
}
type Packet07Layer struct {
	IPAddress *net.UDPAddr
	MTU uint16
	GUID uint64
}
type Packet08Layer struct {
	GUID uint64
	IPAddress *net.UDPAddr
	MTU uint16
	UseSecurity bool
}
type Packet09Layer struct {
	GUID uint64
	Timestamp uint64
	UseSecurity bool
	Password []byte
}
type Packet10Layer struct {
	IPAddress *net.UDPAddr
	SystemIndex uint16
	Addresses [10]*net.UDPAddr
	SendPingTime uint64
	SendPongTime uint64
}
type Packet13Layer struct {
	IPAddress *net.UDPAddr
	Addresses [10]*net.UDPAddr
	SendPingTime uint64
	SendPongTime uint64
}

func NewPacket05Layer() Packet05Layer {
	return Packet05Layer{}
}
func NewPacket06Layer() Packet06Layer {
	return Packet06Layer{}
}
func NewPacket07Layer() Packet07Layer {
	return Packet07Layer{}
}
func NewPacket08Layer() Packet08Layer {
	return Packet08Layer{}
}
func NewPacket09Layer() Packet09Layer {
	return Packet09Layer{}
}
func NewPacket10Layer() Packet10Layer {
	return Packet10Layer{}
}
func NewPacket13Layer() Packet13Layer {
	return Packet13Layer{}
}
var VoidOfflineMessage []byte = make([]byte, 0x10)

func DecodePacket05Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket05Layer()
	thisBitstream := packet.Stream

	var err error
	err = thisBitstream.Bytes(VoidOfflineMessage, 0x10)
	if err != nil {
		return layer, err
	}
	layer.ProtocolVersion, err = thisBitstream.ReadUint8()
	return layer, err
}

func (layer *Packet05Layer) Serialize(stream *ExtendedWriter) error {
	err := stream.AllBytes(OfflineMessageID)
	if err != nil {
		return err
	}
	err = stream.WriteByte(byte(layer.ProtocolVersion))
	if err != nil {
		return err
	}
	empty := make([]byte, 1492 - 0x10 - 1)
	err = stream.AllBytes(empty)
	return err
}

func DecodePacket06Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket06Layer()
	thisBitstream := packet.Stream

	var err error
	err = thisBitstream.Bytes(VoidOfflineMessage, 0x10)
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

func (layer *Packet06Layer) Serialize(stream *ExtendedWriter) error {
	var err error
	err = stream.AllBytes(OfflineMessageID)
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

func DecodePacket07Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket07Layer()
	thisBitstream := packet.Stream

	var err error
	err = thisBitstream.Bytes(VoidOfflineMessage, 0x10)
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

func (layer *Packet07Layer) Serialize(stream *ExtendedWriter) error {
	var err error
	err = stream.AllBytes(OfflineMessageID)
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

func DecodePacket08Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket08Layer()
	thisBitstream := packet.Stream

	var err error
	err = thisBitstream.Bytes(VoidOfflineMessage, 0x10)
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

func (layer *Packet08Layer) Serialize(stream *ExtendedWriter) error {
	var err error
	err = stream.AllBytes(OfflineMessageID)
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

func DecodePacket09Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket09Layer()
	thisBitstream := packet.Stream

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
		context.IsStudio = true
	}
	return layer, err
}

func DecodePacket10Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket10Layer()
	thisBitstream := packet.Stream

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

func DecodePacket13Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket13Layer()
	thisBitstream := packet.Stream

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

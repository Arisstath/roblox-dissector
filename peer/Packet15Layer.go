package peer

// Disconnection reasons
const (
	// Sent when a data:HashItem packet or 0x8A packet contains a bad hash
	DISCONN_HASH = iota
	// Sent when a 0x8A packet contains the wrong security key
	DISCONN_SECURITYKEY = iota
	// Sent when a 0x8A or 0x90 packet contains an incompatible protocol number
	DISCONN_PROTOCOL = iota
	// Sent when a malformed packet is received by the remote peer
	DISCONN_RECEIVE = iota
	// Sent when a malformed stream packet is received by the remote peer
	DISCONN_STREAM = iota
	// Sent when the remote peer runs into an error while trying to send a packet
	DISCONN_SEND = iota
	// Send when a teleport place ID is invalid
	DISCONN_PLACEID = iota
	// Sent when a user with this id is in another game
	DISCONN_OTHER_GAME = iota
	// Sent when an invalid auth ticket is received
	DISCONN_AUTH_TICKET = iota
	// Unknown id. Possibly due to PMC timeout, or ack timeout?
	DISCONN_TIMEOUT = iota
	// Sent when a script kicks the player
	DISCONN_LUAKICK = iota
	// dataping -> onRemoteSysStats (cheater detection)
	DISCONN_REMOTESYSSTATS = iota
	// Unknown id. Possible due to PMC timeout, or ack timeout?
	DISCONN_TIMEOUT2 = iota
	// Sent on Team Create errors
	DISCONN_TEAMCREATE = iota
	// Sent when the Players service contains 0 players for too long.
	// Note that connections do not count toward active players.
	DISCONN_EMPTY_SERVER = iota
	// "Disconnected due to Security Key Mismatch"
	DISCONN_SECURITYKEY2 = iota
	// Sent when the same player tries to join the same game
	DISCONN_OTHER_DEVICE = iota

	// Catch-all disconnection for other reasons
	DISCONN_UNKNOWN = 0xFFFFFFFF
)

// ID_DISCONNECTION_NOTIFICATION - client <-> server
type Packet15Layer struct {
	Reason uint32
}

func NewPacket15Layer() *Packet15Layer {
	return &Packet15Layer{}
}

func DecodePacket15Layer(reader PacketReader, packet *UDPPacket) (RakNetPacket, error) {
	layer := NewPacket15Layer()
	thisBitstream := packet.stream

	var err error
	layer.Reason, err = thisBitstream.readUint32BE()
	return layer, err
}

func (layer *Packet15Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.WriteByte(0x15)
	if err != nil {
		return err
	}
	return stream.writeUint32BE(layer.Reason)
}

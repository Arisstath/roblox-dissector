package packets
import "github.com/gskartwii/roblox-dissector/util"

// ID_PING
type DataPing struct {
	// Always false
	IsPingBack bool
	Timestamp  uint64
	SendStats  uint32
	// Hack flags
	ExtraStats uint32
}

func (thisBitstream *PacketReaderBitstream) DecodeDataPing(reader util.PacketReader, layers *PacketLayers) (ReplicationSubpacket, error) {
	var err error
	inner := &DataPing{}

	inner.IsPingBack, err = thisBitstream.ReadBoolByte()
	if err != nil {
		return inner, err
	}

	inner.Timestamp, err = thisBitstream.bits(64)
	if err != nil {
		return inner, err
	}
	inner.SendStats, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}
	inner.ExtraStats, err = thisBitstream.ReadUint32BE()
	if err != nil {
		return inner, err
	}
	if inner.Timestamp&0x20 != 0 {
		inner.ExtraStats ^= 0xFFFFFFFF
	}

	return inner, err
}

func (layer *DataPing) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	var err error
	err = stream.WriteBoolByte(layer.IsPingBack)
	if err != nil {
		return err
	}
	err = stream.bits(64, layer.Timestamp)
	if err != nil {
		return err
	}
	err = stream.WriteUint32BE(layer.SendStats)
	if err != nil {
		return err
	}
	if layer.Timestamp&0x20 != 0 {
		layer.ExtraStats ^= 0xFFFFFFFF
	}

	err = stream.WriteUint32BE(layer.ExtraStats)
	return err
}

func (DataPing) Type() uint8 {
	return 5
}
func (DataPing) TypeString() string {
	return "ID_REPLIC_PING"
}

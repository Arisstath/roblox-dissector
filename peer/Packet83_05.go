package peer

import "errors"

// ID_PING
type Packet83_05 struct {
	PacketVersion uint8
	// Always false
	Timestamp uint64
	Fps1      float32
	Fps2      float32
	Fps3      float32
	Int1      uint32
	SendStats uint32
	// Hack flags
	ExtraStats uint32
}

func (thisStream *extendedReader) DecodePacket83_05(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_05{}

	inner.PacketVersion, err = thisStream.readUint8()
	if err != nil {
		return inner, err
	}

	if inner.PacketVersion <= 1 {
		inner.Timestamp, err = thisStream.readUint64BE()
		if err != nil {
			return inner, err
		}
	} else if inner.PacketVersion == 2 {
		inner.Int1, err = thisStream.readUint32BE()
		if err != nil {
			return inner, err
		}
		var timestamp uint32
		timestamp, err = thisStream.readUint32BE()
		inner.Timestamp = uint64(timestamp)
		if err != nil {
			return inner, err
		}
		inner.Fps1, err = thisStream.readFloat32BE()
		if err != nil {
			return inner, err
		}
		inner.Fps2, err = thisStream.readFloat32BE()
		if err != nil {
			return inner, err
		}
		inner.Fps3, err = thisStream.readFloat32BE()
		if err != nil {
			return inner, err
		}
	} else {
		return inner, errors.New("invalid packetversion")
	}
	inner.SendStats, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	inner.ExtraStats, err = thisStream.readUint32BE()
	if err != nil {
		return inner, err
	}
	if inner.Timestamp&0x20 != 0 {
		inner.ExtraStats ^= 0xFFFFFFFF
	}

	return inner, err
}

func (layer *Packet83_05) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	err = stream.WriteByte(layer.PacketVersion)
	if err != nil {
		return err
	}
	if layer.PacketVersion <= 1 {
		err = stream.writeUint64BE(layer.Timestamp)
		if err != nil {
			return err
		}
	} else if layer.PacketVersion == 2 {
		err = stream.writeUint32BE(layer.Int1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(uint32(layer.Timestamp))
		if err != nil {
			return err
		}
		err = stream.writeFloat32BE(layer.Fps1)
		if err != nil {
			return err
		}
		err = stream.writeFloat32BE(layer.Fps2)
		if err != nil {
			return err
		}
		err = stream.writeFloat32BE(layer.Fps3)
		if err != nil {
			return err
		}
	} else {
		return errors.New("invalid packetversion")
	}
	err = stream.writeUint32BE(layer.SendStats)
	if err != nil {
		return err
	}
	if layer.Timestamp&0x20 != 0 {
		layer.ExtraStats ^= 0xFFFFFFFF
	}

	err = stream.writeUint32BE(layer.ExtraStats)
	return err
}

func (Packet83_05) Type() uint8 {
	return 5
}
func (Packet83_05) TypeString() string {
	return "ID_REPLIC_PING"
}

func (layer *Packet83_05) String() string {
	// yes, these packets are boring
	return "ID_REPLIC_PING"
}

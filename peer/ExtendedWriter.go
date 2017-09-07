package peer
import "github.com/gskartwii/go-bitstream"
import "encoding/binary"
import "net"
import "github.com/gskartwii/rbxfile"

type ExtendedWriter struct {
	*bitstream.BitWriter
}

func (b *ExtendedWriter) Bits(len int, value uint64) error {
	return b.WriteBits(value, len)
}

func (b *ExtendedWriter) Bytes(len int, value []byte) error {
	for i := 0; i < len; i++ {
		err := b.WriteByte(value[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *ExtendedWriter) AllBytes(value []byte) error {
	return b.Bytes(len(value), value)
}

func (b *ExtendedWriter) WriteBool(value bool) error {
	return b.WriteBit(bitstream.Bit(value))
}

func (b *ExtendedWriter) WriteUint16BE(value uint16) error {
	dest := make([]byte, 2)
	binary.BigEndian.PutUint16(dest, value)
	return b.Bytes(2, dest)
}

func (b *ExtendedWriter) WriteUint32BE(value uint32) error {
	dest := make([]byte, 4)
	binary.BigEndian.PutUint32(dest, value)
	return b.Bytes(4, dest)
}
func (b *ExtendedWriter) WriteUint32LE(value uint32) error {
	dest := make([]byte, 4)
	binary.LittleEndian.PutUint32(dest, value)
	return b.Bytes(4, dest)
}

func (b *ExtendedWriter) WriteUint64BE(value uint64) error {
	dest := make([]byte, 8)
	binary.BigEndian.PutUint64(dest, value)
	return b.Bytes(8, dest)
}

func (b *ExtendedWriter) WriteUint24LE(value uint32) error {
	dest := make([]byte, 4)
	binary.LittleEndian.PutUint32(dest, value)
	return b.Bytes(3, dest)
}

func (b *ExtendedWriter) WriteBoolByte(value bool) error {
	if value {
		return b.WriteByte(1)
	} else {
		return b.WriteByte(0)
	}
}

func (b *ExtendedWriter) WriteAddress(value *net.UDPAddr) error {
	err := b.WriteByte(4)
	if err != nil {
		return err
	}
	err = b.Bytes(4, value.IP)
	if err != nil {
		return err
	}
	err = b.WriteUint16BE(uint16(value.Port))
	return err
}

func (b *ExtendedWriter) WriteASCII(value string) error {
	return b.AllBytes([]byte(value))
}

func (b *ExtendedWriter) WriteUintUTF8(value int) error {
    for value != 0 {
        nextValue := value >> 7
        if nextValue != 0 {
            err := b.WriteByte(byte(value&0x7F|0x80))
            if err != nil {
                return err
            }
        } else {
            err := b.WriteByte(byte(value&0x7F))
            if err != nil {
                return err
            }
        }
        value = nextValue
    }
    return nil
}

func (b *ExtendedWriter) WriteObject(object *rbxfile.Instance, isJoinData bool, context *CommunicationContext) error {
    var err error
    referent := object.Reference
    referentString := context.RefStringsByReferent[referent]

    if isJoinData {
        if referentString == "NULL2" {
            err = b.WriteByte(0)
            return err
        } else if referentString != "NULL" {
            err = b.WriteByte(uint8(len(referentString)))
            if err != nil {
                return err
            }
            err = b.WriteASCII(referentString)
        } else {
            err = b.WriteByte(0xFF)
        }
        if err != nil {
            return err
        }

        return b.WriteUint32LE(uint32(mustAtoi(referent)))
    }
    return nil
}

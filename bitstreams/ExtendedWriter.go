package bitstreams

import "github.com/gskartwii/go-bitstream"
import "github.com/gskartwii/roblox-dissector/util"
import "net"
import "errors"

type BitstreamWriter struct {
	*bitstream.BitWriter
}

func (b *BitstreamWriter) WriteN(data []byte, n int) error {
    m, err := b.Write(data)
    if err != nil {
        return err
    } else if m < n {
        return errors.New("couldn't write enough bytes")
    }
    return nil
}

func (b *BitstreamWriter) WriteAll(data []byte) error {
    return b.WriteN(data, len(data))
}

func (b *BitstreamWriter) WriteAddress(value *net.UDPAddr) error {
	err := b.WriteByte(4)
	if err != nil {
		return err
	}
	for i := 0; i < len(value.IP); i++ {
		value.IP[i] = value.IP[i] ^ 0xFF // bitwise NOT
	}
	err = b.WriteN(value.IP[len(value.IP)-4:], 4)
	if err != nil {
		return err
	}

	// in case the value will be used again
	for i := 0; i < len(value.IP); i++ {
		value.IP[i] = value.IP[i] ^ 0xFF
	}

	err = b.WriteUint16BE(uint16(value.Port))
	return err
}

func (b *BitstreamWriter) WriteJoinObject(reference util.Reference, context *util.CommunicationContext) error {
	var err error
	if reference.IsNull {
		err = b.WriteByte(0)
		return err
	}
	if reference.Scope == context.InstanceTopScope {
		err = b.WriteByte(0xFF)
	} else {
		err = b.WriteByte(uint8(len(reference.Scope)))
		if err != nil {
			return err
		}
		err = b.WriteASCII(reference.Scope)
	}
	if err != nil {
		return err
	}

	return b.WriteUint32LE(reference.Id)
}

func (b *BitstreamWriter) WriteReference(reference util.Reference, caches *util.Caches) error {
	var err error
	if reference.IsNull {
		return b.WriteByte(0x00)
	}
	err = b.WriteCachedObject(reference.Scope, caches)
	if err != nil {
		return err
	}
	return b.WriteUint32LE(reference.Id)
}
func (b *BitstreamWriter) WriteAnyObject(reference util.Reference, context *util.CommunicationContext, caches *util.Caches, isJoinData bool) error {
	if isJoinData {
		return b.WriteJoinObject(reference, context)
	}
	return b.WriteReference(reference, caches)
}

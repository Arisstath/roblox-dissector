package bitstreams

import "github.com/gskartwii/go-bitstream"
import "encoding/binary"
import "net"
import "github.com/gskartwii/rbxfile"
import "math"
import "bytes"

type BitstreamWriter struct {
	*bitstream.BitWriter
}

func (b *BitstreamWriter) WriteAddress(value *net.UDPAddr) error {
	err := b.WriteByte(4)
	if err != nil {
		return err
	}
	for i := 0; i < len(value.IP); i++ {
		value.IP[i] = value.IP[i] ^ 0xFF // bitwise NOT
	}
	err = b.bytes(4, value.IP[len(value.IP)-4:])
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

func (b *BitstreamWriter) WriteJoinObject(reference *Reference, context *CommunicationContext) error {
	var err error
	if reference == nil || reference.IsNull {
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

// TODO: Implement a similar system for readers, where it simply returns an instance
func (b *BitstreamWriter) WriteObject(reference *Reference, caches *Caches) error {
	var err error
	if reference == nil || reference.IsNull {
		return b.WriteByte(0x00)
	}
	err = b.WriteCachedObject(reference.Scope, caches)
	if err != nil {
		return err
	}
	return b.WriteUint32LE(reference.Id)
}
func (b *BitstreamWriter) WriteAnyObject(reference *Reference, writer PacketWriter, isJoinData bool) error {
	if isJoinData {
		return b.WriteJoinObject(reference, writer.Context())
	}
	return b.WriteObject(reference, writer.Caches())
}

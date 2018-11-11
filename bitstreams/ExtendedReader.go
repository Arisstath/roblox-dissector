package bitstreams

import "github.com/gskartwii/go-bitstream"
import "errors"
import "net"
import "math"
import "bytes"
import "io"
import "fmt"

type BitstreamReader struct {
	*bitstream.BitReader
}

func (b *BitstreamReader) ReadAddress() (*net.UDPAddr, error) {
	version, err := b.ReadUint8()
	if err != nil {
		return nil, err
	}
	if version != 4 {
		return nil, errors.New("Unsupported version")
	}
	var address net.IP = make([]byte, 4)
	err = b.bytes(address, 4)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 4; i++ {
		address[i] = address[i] ^ 0xFF // bitwise NOT
	}
	port, err := b.ReadUint16BE()
	if err != nil {
		return nil, err
	}

	return &net.UDPAddr{address, int(port), ""}, nil
}

func (b *BitstreamReader) ReadJoinReferent(context *CommunicationContext) (string, uint32, error) {
	stringLen, err := b.ReadUint8()
	if err != nil {
		return "", 0, err
	}
	if stringLen == 0x00 {
		return "null", 0, err
	}
	var ref string
	if stringLen != 0xFF {
		ref, err = b.ReadASCII(int(stringLen))
		if len(ref) != 0x23 {
			println("WARN: wrong ref len!! this should never happen, unless you are communicating with a non-standard peer")
		}
		if err != nil {
			return "", 0, err
		}
	} else {
		ref = context.InstanceTopScope
	}

	intVal, err := b.ReadUint32LE()
	if err != nil && err != io.EOF {
		return "", 0, err
	}

	return ref, intVal, nil
}

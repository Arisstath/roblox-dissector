package peer

import (
	"errors"
	"fmt"
)

// Packet83_0B represents ID_JOINDATA
type Packet83_0B struct {
	// Instances replicated by the server
	Instances []*ReplicationInstance
}

func (thisStream *extendedReader) DecodePacket83_0B(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	layer := &Packet83_0B{}

	arrayLen, err := thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}

	if arrayLen > 0x100000 {
		return layer, errors.New("sanity check: array len too long")
	}

	layer.Instances = make([]*ReplicationInstance, arrayLen)
	if arrayLen == 0 {
		return layer, nil
	}

	zstdStream, err := thisStream.RegionToZStdStream()
	if err != nil {
		return layer, err
	}

	deferred := newDeferredStrings(reader)
	var i uint32
	for i = 0; i < arrayLen; i++ {
		layer.Instances[i], err = decodeReplicationInstance(reader, &joinSerializeReader{zstdStream}, layers, deferred)
		if err != nil {
			return layer, err
		}
	}

	return layer, zstdStream.resolveDeferredStrings(deferred)
}

// Serialize implements Packet83Subpacket.Serialize()
func (layer *Packet83_0B) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	if layer.Instances == nil || len(layer.Instances) == 0 { // bail
		err = stream.writeUint32BE(0)
		return err
	}

	err = stream.writeUint32BE(uint32(len(layer.Instances)))
	if err != nil {
		return err
	}
	zstdStream := stream.wrapZstd()
	deferred := newWriteDeferredStrings(writer)

	for i := 0; i < len(layer.Instances); i++ {
		// It is safe to take .extendedWriter here
		// because the .extendedWriter will write to the compression/counter writemux
		err = layer.Instances[i].Serialize(writer, &joinSerializeWriter{zstdStream.extendedWriter}, deferred)
		if err != nil {
			zstdStream.Close()
			return err
		}
	}

	err = zstdStream.resolveDeferredStrings(deferred)
	if err != nil {
		zstdStream.Close()
		return err
	}

	return zstdStream.Close()
}

// Type implements Packet83Subpacket.Type()
func (Packet83_0B) Type() uint8 {
	return 0xB
}

// TypeString implements Packet83Subpacket.TypeString()
func (Packet83_0B) TypeString() string {
	return "ID_REPLIC_JOIN_DATA"
}

func (layer *Packet83_0B) String() string {
	return fmt.Sprintf("ID_REPLIC_JOIN_DATA: %d items", len(layer.Instances))
}

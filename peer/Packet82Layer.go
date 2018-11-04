package peer

// Outdated!
/*
import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"

	"github.com/gskartwii/go-bitstream"
)

// Descriptor item containing information about a class/property/event/type
type DescriptorItem struct {
	IDx     uint32
	OtherID uint32
	Name    string
}

// ID_TEACH_DESCRIPTOR_DICTIONARIES - server <-> client
// Contains descriptors to be negotiated by the peers, so that their uint32
// identifiers can be passed over the network
type Packet82Layer struct {
	ClassDescriptor    []*DescriptorItem
	PropertyDescriptor []*DescriptorItem
	EventDescriptor    []*DescriptorItem
	TypeDescriptor     []*DescriptorItem
}

func NewPacket82Layer() *Packet82Layer {
	return &Packet82Layer{}
}

func learnDictionary(decompressedStream *extendedReader, ContextDescriptor map[string]uint32) ([]*DescriptorItem, error) {
	var dictionary []*DescriptorItem
	dictionaryLength, _ := decompressedStream.readUint32BE()
	if dictionaryLength > 0x1000 {
		return dictionary, errors.New("sanity check: dictionary length exceeded maximum")
	}
	dictionary = make([]*DescriptorItem, dictionaryLength)
	var i uint32
	for i = 0; i < dictionaryLength; i++ {
		IDx, _ := decompressedStream.readUint32BE()
		nameLength, _ := decompressedStream.readUint16BE()
		name, _ := decompressedStream.readString(int(nameLength))
		otherID, _ := decompressedStream.readUint32BE()

		dictionary[i] = &DescriptorItem{IDx, otherID, string(name)}
		ContextDescriptor[string(name)] = IDx
	}
	return dictionary, nil
}

func teachDictionary(stream *extendedWriter, descriptor []*DescriptorItem) error {
	err := stream.writeUint32BE(uint32(len(descriptor)))
	if err != nil {
		return err
	}

	for _, item := range descriptor {
		err = stream.writeUint32BE(item.IDx)
		if err != nil {
			return err
		}
		err = stream.writeUint16BE(uint16(len(item.Name)))
		if err != nil {
			return err
		}
		err = stream.writeASCII(item.Name)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(item.OtherID)
		if err != nil {
			return err
		}
	}
	return nil
}

func learnDictionaryHuffman(decompressedStream *extendedReader, ContextDescriptor map[string]uint32) ([]*DescriptorItem, error) {
	var dictionary []*DescriptorItem
	dictionaryLength, _ := decompressedStream.readUint32BE()
	if dictionaryLength > 0x1000 {
		return dictionary, errors.New("sanity check: dictionary length exceeded maximum")
	}
	dictionary = make([]*DescriptorItem, dictionaryLength)
	var i uint32
	for i = 0; i < dictionaryLength; i++ {
		IDx, _ := decompressedStream.readUint32BE()
		name, _ := decompressedStream.readHuffman()

		dictionary[i] = &DescriptorItem{IDx, 0, string(name)}
		ContextDescriptor[string(name)] = IDx
	}
	return dictionary, nil
}

func DecodePacket82Layer(reader PacketReader, packet *UDPPacket) (RakNetPacket, error) {
	layer := NewPacket82Layer()
	thisBitstream := packet.stream

	var err error
	var decompressedStream *extendedReader
	if context.IsClient(packet.Source) {
		decompressedStream = thisBitstream

		layer.ClassDescriptor, err = learnDictionaryHuffman(decompressedStream, context.ClassDescriptor)
		if err != nil {
			return layer, err
		}

		layer.PropertyDescriptor, err = learnDictionaryHuffman(decompressedStream, context.PropertyDescriptor)
		if err != nil {
			return layer, err
		}

		layer.EventDescriptor, err = learnDictionaryHuffman(decompressedStream, context.EventDescriptor)
		if err != nil {
			return layer, err
		}

		layer.TypeDescriptor, err = learnDictionaryHuffman(decompressedStream, context.TypeDescriptor)
		if err != nil {
			return layer, err
		}
		return layer, nil
	} else {
		_, _ = thisBitstream.readUint32BE() // Skip compressed len
		gzipStream, err := gzip.NewReader(thisBitstream.GetReader())
		if err != nil {
			return layer, err
		}

		decompressedStream = &extendedReader{bitstream.NewReader(gzipStream)}

		layer.ClassDescriptor, err = learnDictionary(decompressedStream, context.ClassDescriptor)
		if err != nil {
			return layer, err
		}

		layer.PropertyDescriptor, err = learnDictionary(decompressedStream, context.PropertyDescriptor)
		if err != nil {
			return layer, err
		}

		layer.EventDescriptor, err = learnDictionary(decompressedStream, context.EventDescriptor)
		if err != nil {
			return layer, err
		}

		layer.TypeDescriptor, err = learnDictionary(decompressedStream, context.TypeDescriptor)
		if err != nil {
			return layer, err
		}

		return layer, nil
	}
}

func (layer *Packet82Layer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error
	// FIXME: Assume this peer is always a server

	err = stream.WriteByte(0x82)
	if err != nil {
		return err
	}
	gzipBuf := bytes.NewBuffer([]byte{})
	middleStream := gzip.NewWriter(gzipBuf)
	defer middleStream.Close()
	gzipStream := &extendedWriter{bitstream.NewWriter(middleStream)}
	err = teachDictionary(gzipStream, layer.ClassDescriptor)
	if err != nil {
		return err
	}
	err = teachDictionary(gzipStream, layer.PropertyDescriptor)
	if err != nil {
		return err
	}
	err = teachDictionary(gzipStream, layer.EventDescriptor)
	if err != nil {
		return err
	}
	err = teachDictionary(gzipStream, layer.TypeDescriptor)
	if err != nil {
		return err
	}
	err = gzipStream.Flush(bitstream.Zero)
	if err != nil {
		return err
	}
	err = middleStream.Flush()
	if err != nil {
		return err
	}
	err = middleStream.Close()
	if err != nil {
		return err
	}
	err = stream.writeUint32BE(uint32(gzipBuf.Len()))
	if err != nil {
		return err
	}
	err = stream.allBytes(gzipBuf.Bytes())
	return err
}
*/

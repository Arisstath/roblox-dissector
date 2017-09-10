package peer
import "compress/gzip"
import "github.com/gskartwii/go-bitstream"
import "errors"
import "bytes"

type DescriptorItem struct {
	IDx uint32
	OtherID uint32
	Name string
}

type Packet82Layer struct {
	ClassDescriptor []*DescriptorItem
	PropertyDescriptor []*DescriptorItem
	EventDescriptor []*DescriptorItem
	TypeDescriptor []*DescriptorItem
}

func NewPacket82Layer() Packet82Layer {
	return Packet82Layer{}
}

func LearnDictionary(decompressedStream *ExtendedReader, ContextDescriptor map[string]uint32) ([]*DescriptorItem, error) {
	var dictionary []*DescriptorItem
	dictionaryLength, _ := decompressedStream.ReadUint32BE()
	if dictionaryLength > 0x1000 {
		return dictionary, errors.New("sanity check: dictionary length exceeded maximum")
	}
	dictionary = make([]*DescriptorItem, dictionaryLength)
	var i uint32
	for i = 0; i < dictionaryLength; i++ {
		IDx, _ := decompressedStream.ReadUint32BE()
		nameLength, _ := decompressedStream.ReadUint16BE()
		name, _ := decompressedStream.ReadString(int(nameLength))
		otherID, _ := decompressedStream.ReadUint32BE()

		dictionary[i] = &DescriptorItem{IDx, otherID, string(name)}
		ContextDescriptor[string(name)] = IDx
	}
	return dictionary, nil
}

func TeachDictionary(stream *ExtendedWriter, descriptor []*DescriptorItem) error {
    err := stream.WriteUint32BE(uint32(len(descriptor)))
    if err != nil {
        return err
    }

    for _, item := range descriptor {
        err = stream.WriteUint32BE(item.IDx)
        if err != nil {
            return err
        }
        err = stream.WriteUint16BE(uint16(len(item.Name)))
        if err != nil {
            return err
        }
        err = stream.WriteASCII(item.Name)
        if err != nil {
            return err
        }
        err = stream.WriteUint32BE(item.OtherID)
        if err != nil {
            return err
        }
    }
    return nil
}

func LearnDictionaryHuffman(decompressedStream *ExtendedReader, ContextDescriptor map[string]uint32) ([]*DescriptorItem, error) {
	var dictionary []*DescriptorItem
	dictionaryLength, _ := decompressedStream.ReadUint32BE()
	if dictionaryLength > 0x1000 {
		return dictionary, errors.New("sanity check: dictionary length exceeded maximum")
	}
	dictionary = make([]*DescriptorItem, dictionaryLength)
	var i uint32
	for i = 0; i < dictionaryLength; i++ {
		IDx, _ := decompressedStream.ReadUint32BE()
		name, _ := decompressedStream.ReadHuffman()

		dictionary[i] = &DescriptorItem{IDx, 0, string(name)}
		ContextDescriptor[string(name)] = IDx
	}
	return dictionary, nil
}

func DecodePacket82Layer(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	layer := NewPacket82Layer()
	thisBitstream := packet.Stream

	var err error
	var decompressedStream *ExtendedReader
	context.MDescriptor.Lock()
	defer context.MDescriptor.Unlock() // Do not broadcast if parsing fails
	if context.IsClient(packet.Source) {
		decompressedStream = thisBitstream

		layer.ClassDescriptor, err = LearnDictionaryHuffman(decompressedStream, context.ClassDescriptor)
		if err != nil {
			return layer, err
		}

		layer.PropertyDescriptor, err = LearnDictionaryHuffman(decompressedStream, context.PropertyDescriptor)
		if err != nil {
			return layer, err
		}

		layer.EventDescriptor, err = LearnDictionaryHuffman(decompressedStream, context.EventDescriptor)
		if err != nil {
			return layer, err
		}

		layer.TypeDescriptor, err = LearnDictionaryHuffman(decompressedStream, context.TypeDescriptor)
		if err != nil {
			return layer, err
		}
		context.EDescriptorsParsed.Broadcast()
		return layer, nil
	} else {
		_, _ = thisBitstream.ReadUint32BE() // Skip compressed len
		gzipStream, err := gzip.NewReader(thisBitstream.GetReader())
		if err != nil {
			return layer, err
		}

		decompressedStream = &ExtendedReader{bitstream.NewReader(gzipStream)}

		layer.ClassDescriptor, err = LearnDictionary(decompressedStream, context.ClassDescriptor)
		if err != nil {
			return layer, err
		}

		layer.PropertyDescriptor, err = LearnDictionary(decompressedStream, context.PropertyDescriptor)
		if err != nil {
			return layer, err
		}

		layer.EventDescriptor, err = LearnDictionary(decompressedStream, context.EventDescriptor)
		if err != nil {
			return layer, err
		}

		layer.TypeDescriptor, err = LearnDictionary(decompressedStream, context.TypeDescriptor)
		if err != nil {
			return layer, err
		}

		context.EDescriptorsParsed.Broadcast()
		return layer, nil
	}
}

func (layer *Packet82Layer) Serialize(context *CommunicationContext, stream *ExtendedWriter) error {
    var err error
    // FIXME: Assume this peer is always a server

    err = stream.WriteByte(0x82)
    if err != nil {
        return err
    }
    gzipBuf := bytes.NewBuffer([]byte{})
    middleStream := gzip.NewWriter(gzipBuf)
    defer middleStream.Close()
    gzipStream := &ExtendedWriter{bitstream.NewWriter(middleStream)}
    err = TeachDictionary(gzipStream, layer.ClassDescriptor)
    if err != nil {
        return err
    }
    err = TeachDictionary(gzipStream, layer.PropertyDescriptor)
    if err != nil {
        return err
    }
    err = TeachDictionary(gzipStream, layer.EventDescriptor)
    if err != nil {
        return err
    }
    err = TeachDictionary(gzipStream, layer.TypeDescriptor)
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
    err = stream.WriteUint32BE(uint32(gzipBuf.Len()))
    if err != nil {
        return err
    }
    err = stream.AllBytes(gzipBuf.Bytes())
    return err
}

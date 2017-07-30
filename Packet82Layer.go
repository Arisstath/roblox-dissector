package main
import "compress/gzip"
import "github.com/google/gopacket"
import "github.com/gskartwii/go-bitstream"
import "errors"

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

func LearnDictionary(decompressedStream *ExtendedReader, ContextDescriptor map[uint32]string) ([]*DescriptorItem, error) {
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
		ContextDescriptor[IDx] = string(name)
	}
	return dictionary, nil
}

func LearnDictionaryHuffman(decompressedStream *ExtendedReader, ContextDescriptor map[uint32]string) ([]*DescriptorItem, error) {
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
		ContextDescriptor[IDx] = string(name)
	}
	return dictionary, nil
}

func DecodePacket82Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	println("Parsing 0x82")
	layer := NewPacket82Layer()

	var err error
	var decompressedStream *ExtendedReader
	context.MDescriptor.Lock()
	defer context.MDescriptor.Unlock() // Do not broadcast if parsing fails
	if context.PacketFromClient(packet) {
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

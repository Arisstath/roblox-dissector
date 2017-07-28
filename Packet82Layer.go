package main
import "compress/gzip"
import "github.com/google/gopacket"
import "github.com/dgryski/go-bitstream"

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

func LearnDictionary(decompressedStream *ExtendedReader, ContextDescriptor map[uint32]string) []*DescriptorItem {
	dictionaryLength, _ := decompressedStream.ReadUint32BE()
	dictionary := make([]*DescriptorItem, 0, dictionaryLength)
	var i uint32
	for i = 0; i < dictionaryLength; i++ {
		IDx, _ := decompressedStream.ReadUint32BE()
		nameLength, _ := decompressedStream.ReadUint16BE()
		name, _ := decompressedStream.ReadString(int(nameLength))
		otherID, _ := decompressedStream.ReadUint32BE()

		dictionary = append(dictionary, &DescriptorItem{IDx, otherID, string(name)})
		ContextDescriptor[IDx] = string(name)
	}
	return dictionary
}

func LearnDictionaryHuffman(decompressedStream *ExtendedReader, ContextDescriptor map[uint32]string) []*DescriptorItem {
	dictionaryLength, _ := decompressedStream.ReadUint32BE()
	dictionary := make([]*DescriptorItem, 0, dictionaryLength)
	var i uint32
	for i = 0; i < dictionaryLength; i++ {
		IDx, _ := decompressedStream.ReadUint32BE()
		name, _ := decompressedStream.ReadHuffman()

		dictionary = append(dictionary, &DescriptorItem{IDx, 0, string(name)})
		ContextDescriptor[IDx] = string(name)
	}
	return dictionary
}

func DecodePacket82Layer(thisBitstream *ExtendedReader, context *CommunicationContext, packet gopacket.Packet) (interface{}, error) {
	layer := NewPacket82Layer()

	var decompressedStream *ExtendedReader
	if PacketFromClient(packet, context) {
		decompressedStream = thisBitstream

		layer.ClassDescriptor = LearnDictionaryHuffman(decompressedStream, context.ClassDescriptor)
		layer.PropertyDescriptor = LearnDictionaryHuffman(decompressedStream, context.PropertyDescriptor)
		layer.EventDescriptor = LearnDictionaryHuffman(decompressedStream, context.EventDescriptor)
		layer.TypeDescriptor = LearnDictionaryHuffman(decompressedStream, context.TypeDescriptor)
		return layer, nil
	} else {
		_, _ = thisBitstream.ReadUint32BE() // Skip compressed len
		gzipStream, err := gzip.NewReader(thisBitstream.GetReader())
		if err != nil {
			return layer, err
		}

		decompressedStream = &ExtendedReader{bitstream.NewReader(gzipStream)}
		layer.ClassDescriptor = LearnDictionary(decompressedStream, context.ClassDescriptor)
		layer.PropertyDescriptor = LearnDictionary(decompressedStream, context.PropertyDescriptor)
		layer.EventDescriptor = LearnDictionary(decompressedStream, context.EventDescriptor)
		layer.TypeDescriptor = LearnDictionary(decompressedStream, context.TypeDescriptor)

		return layer, nil
	}
}

package main
import "compress/gzip"
import "github.com/google/gopacket"
import "github.com/gskartwii/go-bitstream"

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

		context.MClassDescriptor.Lock()
		layer.ClassDescriptor = LearnDictionaryHuffman(decompressedStream, context.ClassDescriptor)
		context.MClassDescriptor.Unlock()

		context.MPropertyDescriptor.Lock()
		layer.PropertyDescriptor = LearnDictionaryHuffman(decompressedStream, context.PropertyDescriptor)
		context.MPropertyDescriptor.Unlock()

		context.MEventDescriptor.Lock()
		layer.EventDescriptor = LearnDictionaryHuffman(decompressedStream, context.EventDescriptor)
		context.MEventDescriptor.Unlock()

		context.MTypeDescriptor.Lock()
		layer.TypeDescriptor = LearnDictionaryHuffman(decompressedStream, context.TypeDescriptor)
		context.MTypeDescriptor.Unlock()
		return layer, nil
	} else {
		_, _ = thisBitstream.ReadUint32BE() // Skip compressed len
		gzipStream, err := gzip.NewReader(thisBitstream.GetReader())
		if err != nil {
			return layer, err
		}

		decompressedStream = &ExtendedReader{bitstream.NewReader(gzipStream)}

		context.MClassDescriptor.Lock()
		layer.ClassDescriptor = LearnDictionary(decompressedStream, context.ClassDescriptor)
		context.MClassDescriptor.Unlock()

		context.MPropertyDescriptor.Lock()
		layer.PropertyDescriptor = LearnDictionary(decompressedStream, context.PropertyDescriptor)
		context.MPropertyDescriptor.Unlock()

		context.MEventDescriptor.Lock()
		layer.EventDescriptor = LearnDictionary(decompressedStream, context.EventDescriptor)
		context.MEventDescriptor.Unlock()

		context.MTypeDescriptor.Lock()
		layer.TypeDescriptor = LearnDictionary(decompressedStream, context.TypeDescriptor)
		context.MTypeDescriptor.Unlock()

		return layer, nil
	}
}

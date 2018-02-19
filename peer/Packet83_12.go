package peer
import "fmt"

// ID_HASH
type Packet83_12 struct {
	HashList []uint32
	SecurityTokens [3]uint32
}

func getRbxNonce(base uint32, query uint32) uint32 {
	baseState := base
	queryState := query
	var first13Bits uint32 = 0xFFF80000
	var second13Bits uint32 = 0xFFF80000 >> 13
	var first15Bits uint32 = 0xFFFE0000
	var second15Bits uint32 = 0xFFFE0000 >> 15

	baseState ^= baseState >> 16
	queryState ^= queryState >> 16

	baseState *= 0xA89ED915 // modinv 3
	queryState *= 0xA89ED915
	baseState ^= (baseState ^ first13Bits) >> 13
	queryState ^= (queryState ^ first13Bits) >> 13
	baseState ^= (baseState ^ second13Bits) >> 13
	queryState ^= (queryState ^ second13Bits) >> 13

	baseState *= 0xB6C92F47 // modinv 2
	queryState *= 0xB6C92F47
	baseState ^= (baseState ^ first15Bits) >> 15
	queryState ^= (queryState ^ first15Bits) >> 15
	baseState ^= (baseState ^ second15Bits) >> 15
	queryState ^= (queryState ^ second15Bits) >> 15
	
	baseState += 4

	queryState *= 0xA0FE3BCF // modinv 4
	queryState = queryState << (32 - 17) | queryState >> 17

	return (queryState - baseState) * 0xA89ED915
}

func decodePacket83_12(packet *UDPPacket, context *CommunicationContext) (interface{}, error) {
	var err error
	inner := &Packet83_12{}
	stream := packet.stream
	numItems, err := stream.readUint8()
	if err != nil {
		return inner, err
	}
	if numItems == 0xFF {
		println("noextranumitem")
		numItems = 0
	} else {
		numItems, err = stream.ReadUint8()
		if err != nil {
			return inner, err
		}
	}

	nonce, err := stream.readUint32BE()
	if err != nil {
		return inner, err
	}
	hashList := make([]uint32, numItems)
	for i := 0; i < int(numItems); i++ {
		hashList[i], err = stream.readUint32BE()
		if err != nil {
			return inner, err
		}
	}
	var tokens [3]uint32
	for i := 0; i < 3; i++ {
		tokens[i], err = stream.readUint32BE()
		if err != nil {
			return inner, err
		}
	}

	for i := numItems - 2; i > 0; i-- {
		hashList[i] ^= hashList[i - 1]
	}
	hashList[0] ^= nonce
	nonce ^= hashList[numItems - 1]
	nonceDiff := nonce - getRbxNonce(hashList[1], hashList[2])

	fmt.Println("hashlist", hashList, nonce, nonceDiff)

	return inner, nil
}

func (layer *Packet83_12) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
	return nil
}

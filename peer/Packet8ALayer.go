package peer
import "github.com/gskartwii/go-bitstream"
import "bytes"
import "crypto/aes"
import "crypto/cipher"
import "io"
import "errors"

func shuffleSlice(src []byte) ([]byte) {
	ShuffledSrc := make([]byte, 0, len(src))
	ShuffledSrc = append(ShuffledSrc, src[:0x10]...)
	for j := len(src) - 0x10; j >= 0x10; j -= 0x10 {
		ShuffledSrc = append(ShuffledSrc, src[j:j+0x10]...)
	}
	return ShuffledSrc
}

func calculateChecksum(data []byte) uint32 {
	var sum uint32 = 0
	var r uint16 = 55665
	var c1 uint16 = 52845
	var c2 uint16 = 22719
	for i := 0; i < len(data); i++ {
		char := data[i]
		cipher := (char ^ byte(r >> 8)) & 0xFF
		r = (uint16(cipher) + r)*c1 + c2
		sum += uint32(cipher)
	}
	return sum
}

// ID_SUBMIT_TICKET - client -> server
type Packet8ALayer struct {
	PlayerId int32
	ClientTicket []byte
	DataModelHash []byte
	// Always 36?
	ProtocolVersion uint32
	SecurityKey []byte
	Platform []byte
	RobloxProductName []byte
	SessionId []byte
	GoldenHash uint32
}

func NewPacket8ALayer() Packet8ALayer {
	return Packet8ALayer{}
}

func decodePacket8ALayer(packet *UDPPacket, context *CommunicationContext, data []byte) (interface{}, error) {
	layer := NewPacket8ALayer()
	block, e := aes.NewCipher([]byte{0xFE, 0xF9, 0xF0, 0xEB, 0xE2, 0xDD, 0xD4, 0xCF, 0xC6, 0xC1, 0xB8, 0xB3, 0xAA, 0xA5, 0x9C, 0x97})

	if e != nil {
		panic(e)
	}

	dest := make([]byte, len(data))
	for i, _ := range dest {
		dest[i] = 0xBA
	}
	c := cipher.NewCBCDecrypter(block, []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})

	ShuffledSrc := shuffleSlice(data)

	c.CryptBlocks(dest, ShuffledSrc)
	dest = shuffleSlice(dest)

	checkSum := calculateChecksum(dest[4:])
	thisBitstream := extendedReader{bitstream.NewReader(bytes.NewReader(dest))}
	storedChecksum, err := thisBitstream.readUint32LE()
	if err != nil {
		return layer, err
	}
	if storedChecksum != checkSum {
		println("checksum check failed!", storedChecksum, checkSum)
		return layer, errors.New("checksum check fail")
	}

	_, err = thisBitstream.ReadByte()
	if err != nil {
		return layer, err
	}
	paddingSizeByte, err := thisBitstream.ReadByte()
	if err != nil {
		return layer, err
	}
	PaddingSize := paddingSizeByte & 0xF

	void := make([]byte, PaddingSize)
	err = thisBitstream.bytes(void, int(PaddingSize))
	if err != nil {
		return layer, err
	}

	playerId, err := thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.PlayerId = int32(playerId)
	layer.ClientTicket, err = thisBitstream.readHuffman()
	if err != nil {
		return layer, err
	}
	layer.DataModelHash, err = thisBitstream.readHuffman()
	if err != nil {
		return layer, err
	}
	layer.ProtocolVersion, err = thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.SecurityKey, err = thisBitstream.readHuffman()
	if err != nil {
		return layer, err
	}
	layer.Platform, err = thisBitstream.readHuffman()
	if err != nil {
		return layer, err
	}
	layer.RobloxProductName, err = thisBitstream.readHuffman()
	if err == io.EOF {
		return layer, nil
	} else if err != nil {
		return layer, err
	}
	layer.SessionId, err = thisBitstream.readHuffman()
	if err != nil {
		return layer, err
	}
	layer.GoldenHash, err = thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}
	
	return layer, nil
}
func (layer *Packet8ALayer) serialize(isClient bool, context *CommunicationContext, stream *extendedWriter) error {
	rawBuffer := new(bytes.Buffer)
	rawStream := &extendedWriter{bitstream.NewWriter(rawBuffer)}
	var err error

	err = stream.WriteByte(0x8A)
	if err != nil {
		return err
	}
	err = rawStream.writeUint32BE(uint32(layer.PlayerId))
	if err != nil {
		return err
	}
	err = rawStream.writeHuffman(layer.ClientTicket)
	if err != nil {
		return err
	}
	err = rawStream.writeHuffman(layer.DataModelHash)
	if err != nil {
		return err
	}
	err = rawStream.writeUint32BE(layer.ProtocolVersion)
	if err != nil {
		return err
	}
	err = rawStream.writeHuffman(layer.SecurityKey)
	if err != nil {
		return err
	}
	err = rawStream.writeHuffman(layer.Platform)
	if err != nil {
		return err
	}
	err = rawStream.writeHuffman(layer.RobloxProductName)
	if err != nil {
		return err
	}
	err = rawStream.writeHuffman(layer.SessionId)
	if err != nil {
		return err
	}
	err = rawStream.writeUint32BE(layer.GoldenHash)
	if err != nil {
		return err
	}
	rawStream.Flush(bitstream.Bit(false))

	// create slice with 6-byte header and padding to align to 0x10-byte blocks
	length := rawBuffer.Len()
	paddingSize := 0xF - (length + 5)%0x10
	rawCopy := make([]byte, length + 6 + paddingSize)
	rawCopy[5] = byte(paddingSize & 0xF)
	copy(rawCopy[6+paddingSize:], rawBuffer.Bytes())

	checkSum := calculateChecksum(rawCopy[4:])
	rawCopy[3] = byte(checkSum >> 24 & 0xFF)
	rawCopy[2] = byte(checkSum >> 16 & 0xFF)
	rawCopy[1] = byte(checkSum >> 8 & 0xFF)
	rawCopy[0] = byte(checkSum & 0xFF)

	// CBC blocks are encrypted in a weird order
	dest := make([]byte, len(rawCopy))
	shuffledEncryptable := shuffleSlice(rawCopy)
	block, err := aes.NewCipher([]byte{0xFE, 0xF9, 0xF0, 0xEB, 0xE2, 0xDD, 0xD4, 0xCF, 0xC6, 0xC1, 0xB8, 0xB3, 0xAA, 0xA5, 0x9C, 0x97})
	if err != nil {
		return err
	}
	c := cipher.NewCBCEncrypter(block, []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})
	c.CryptBlocks(dest, shuffledEncryptable)
	dest = shuffleSlice(dest) // shuffle back to correct order

	return stream.allBytes(dest)
}

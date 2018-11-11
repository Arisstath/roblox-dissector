package packets

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
	"math/bits"

	"github.com/gskartwii/go-bitstream"
	"github.com/pierrec/xxHash/xxHash32"
)

func shuffleSlice(src []byte) []byte {
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
		cipher := (char ^ byte(r>>8)) & 0xFF
		r = (uint16(cipher)+r)*c1 + c2
		sum += uint32(cipher)
	}
	return sum
}

func hashClientTicket(ticket string) uint32 {
	var ecxHash uint32
	initHash := xxHash32.Checksum([]byte(ticket), 1)
	initHash += 0x557BB5D7
	initHash = bits.RotateLeft32(initHash, -7)
	initHash -= 0x557BB5D7
	initHash *= 0x443921D5
	initHash = bits.RotateLeft32(initHash, 0xD)
	ecxHash = 0x443921D5 - initHash
	ecxHash ^= 0x557BB5D7
	ecxHash = bits.RotateLeft32(ecxHash, -0x11)
	ecxHash += 0x11429402
	ecxHash = bits.RotateLeft32(ecxHash, 0x17)
	initHash = 0x99B4D7AC - ecxHash
	initHash = bits.RotateLeft32(initHash, -0x1D)
	initHash ^= 0x557BB5D7

	return initHash
}

// ID_SUBMIT_TICKET - client -> server
type AuthPacket struct {
	PlayerId      int64
	ClientTicket  string
	DataModelHash string
	// Always 36?
	ProtocolVersion   uint32
	SecurityKey       string
	Platform          string
	RobloxProductName string
	SessionId         string
	GoldenHash        uint32
}

func NewAuthPacket() *AuthPacket {
	return &AuthPacket{}
}

func (stream *PacketReaderBitstream) DecodeAuthPacket(reader util.PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewAuthPacket()

	lenBytes := bitsToBytes(uint(layers.Reliability.LengthInBits)) - 1 // -1 for packet id
	data, err := stream.readString(int(lenBytes))
	if err != nil {
		return layer, err
	}
	block, e := aes.NewCipher([]byte{0xFE, 0xF9, 0xF0, 0xEB, 0xE2, 0xDD, 0xD4, 0xCF, 0xC6, 0xC1, 0xB8, 0xB3, 0xAA, 0xA5, 0x9C, 0x97})

	if e != nil {
		panic(e)
	}

	dest := make([]byte, len(data))
	for i := range dest {
		dest[i] = 0xBA
	}
	c := cipher.NewCBCDecrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

	ShuffledSrc := shuffleSlice(data)

	c.CryptBlocks(dest, ShuffledSrc)
	dest = shuffleSlice(dest)

	checkSum := calculateChecksum(dest[4:])
	thisBitstream := PacketReaderBitstream{bitstream.NewReader(bytes.NewReader(dest))}
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

	playerId, err := thisBitstream.readVarsint64()
	if err != nil {
		return layer, err
	}
	layer.PlayerId = playerId
	layer.ClientTicket, err = thisBitstream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layer.DataModelHash, err = thisBitstream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layer.ProtocolVersion, err = thisBitstream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layer.SecurityKey, err = thisBitstream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layer.Platform, err = thisBitstream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layer.RobloxProductName, err = thisBitstream.readVarLengthString()
	if err == io.EOF {
		return layer, nil
	} else if err != nil {
		return layer, err
	}
	hash, err := thisBitstream.readUintUTF8()
	if err != nil {
		return layer, err
	}
	clientTicketHash := hashClientTicket(layer.ClientTicket)
	if hash != clientTicketHash {
		layers.Root.Logger.Printf("hash mismatch: read %8X != generated %8X\n", hash, clientTicketHash)
	} else {
		layers.Root.Logger.Printf("hash ok: %8X\n", hash)
	}

	layer.SessionId, err = thisBitstream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layer.GoldenHash, err = thisBitstream.readUint32BE() // 0xc001cafe on android - cool cafe!
	if err != nil {
		return layer, err
	}

	return layer, nil
}

func (layer *AuthPacket) Serialize(writer util.PacketWriter, stream *PacketWriterBitstream) error {
	rawBuffer := new(bytes.Buffer)
	rawStream := &PacketWriterBitstream{bitstream.NewWriter(rawBuffer)}
	var err error

	err = stream.WriteByte(0x8A)
	if err != nil {
		return err
	}
	err = rawStream.writeVarsint64(layer.PlayerId)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.ClientTicket)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.DataModelHash)
	if err != nil {
		return err
	}
	err = rawStream.writeUint32BE(layer.ProtocolVersion)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.SecurityKey)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.Platform)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.RobloxProductName)
	if err != nil {
		return err
	}
	err = rawStream.writeVarint64(uint64(hashClientTicket(layer.ClientTicket)))
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.SessionId)
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
	paddingSize := 0xF - (length+5)%0x10
	rawCopy := make([]byte, length+6+paddingSize)
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
	c := cipher.NewCBCEncrypter(block, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	c.CryptBlocks(dest, shuffledEncryptable)
	dest = shuffleSlice(dest) // shuffle back to correct order

	return stream.allBytes(dest)
}

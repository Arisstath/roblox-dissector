package bitstreams
import "github.com/DataDog/zstd"
import "compress/gzip"
import "encoding/binary"
import "errors"
import "fmt"
import "bytes"
import "github.com/gskartwii/go-bitstream"

var englishTree *huffmanEncodingTree

func init() {
	englishTree = generateHuffmanFromFrequencyTable([]uint32{ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 722, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11084, 58, 63, 1, 0, 31, 0, 317, 64, 64, 44, 0, 695, 62, 980, 266, 69, 67, 56, 7, 73, 3, 14, 2, 69, 1, 167, 9, 1, 2, 25, 94, 0, 195, 139, 34, 96, 48, 103, 56, 125, 653, 21, 5, 23, 64, 85, 44, 34, 7, 92, 76, 147, 12, 14, 57, 15, 39, 15, 1, 1, 1, 2, 3, 0, 3611, 845, 1077, 1884, 5870, 841, 1057, 2501, 3212, 164, 531, 2019, 1330, 3056, 4037, 848, 47, 2586, 2919, 4771, 1707, 535, 1106, 152, 1243, 100, 0, 2, 0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func (b *BitstreamReader) ReadCompressed(dest []byte, size uint32, unsignedData bool) error {
	var currentByte uint32 = (size >> 3) - 1
	var byteMatch, halfByteMatch byte

	if unsignedData {
		byteMatch = 0
		halfByteMatch = 0
	} else {
		byteMatch = 0xFF
		halfByteMatch = 0xF0
	}

	for currentByte > 0 {
		res, err := b.ReadBool()
		if err != nil {
			return err
		}
		if res {
			dest[currentByte] = byteMatch
			currentByte--
		} else {
            err = b.Read(dest[:currentByte+1])
			return err
		}
	}

	res, err := b.ReadBool()
	if err != nil {
		return err
	}

	if res {
		res, err := b.ReadBits(4)
		if err != nil {
			return err
		}
		dest[currentByte] = byte(res) | halfByteMatch
	} else {
		err := b.Read(dest[currentByte:currentByte+1])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *BitstreamReader) ReadUint32BECompressed(unsignedData bool) (uint32, error) {
	dest := make([]byte, 4)
	err := b.ReadCompressed(dest, 32, unsignedData)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(dest), nil
}

func (b *BitstreamReader) ReadHuffman() ([]byte, error) {
	var name []byte
	maxCharLen, err := b.ReadUint32BE()
	if err != nil {
		return name, err
	}
	sizeInBits, err := b.ReadUint32BECompressed(true)
	if err != nil {
		return name, err
	}

	if maxCharLen > 0x5000 || sizeInBits > 0x50000 {
		return name, errors.New("sanity check: exceeded maximum sizeinbits/maxcharlen of 0x5000")
	}
	name = make([]byte, maxCharLen)
	err = englishTree.decodeArray(b, uint(sizeInBits), uint(maxCharLen), name)

	return name, err
}

func (b *BitstreamReader) RegionToGZipStream() (*BitstreamReader, error) {
	compressedLen, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}
	println("compressedLen:", compressedLen)

	compressed := make([]byte, compressedLen)
	err = b.Read(compressed)
	if err != nil {
		return nil, err
	}
	fmt.Printf("compressed start: %v\n", compressed[:0x20])

	gzipStream, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	return &BitstreamReader{bitstream.NewReader(gzipStream)}, err
}

func (b *BitstreamReader) RegionToZStdStream() (*BitstreamReader, error) {
	compressedLen, err := b.ReadUint32BE()
	if err != nil {
		return nil, err
	}

	_, err = b.ReadUint32BE()
	if err != nil {
		return nil, err
	}

	compressed := make([]byte, compressedLen)
	err = b.Read(compressed)
	if err != nil {
		return nil, err
	}
	/*decompressed, err := zstd.Decompress(nil, compressed)
	println("decomp len", len(decompressed))
	if len(decompressed) > 0x20 {
		fmt.Printf("first bytes %#X\n", decompressed[:0x20])
	} else {
		fmt.Printf("first bytes %#X\n", decompressed)
	}*/

	zstdStream := zstd.NewReader(bytes.NewReader(compressed))
	return &BitstreamReader{bitstream.NewReader(zstdStream)}, nil
}

func (b *BitstreamReader) ReadFloat16BE(floatMin float32, floatMax float32) (float32, error) {
	scale, err := b.ReadUint16BE()
	if err != nil {
		return 0.0, err
	}

	outFloat := float32(scale)/65535.0*(floatMax-floatMin) + floatMin
	if outFloat < floatMin {
		return floatMin, nil
	} else if outFloat > floatMax {
		return floatMax, nil
	}
	return outFloat, nil
}

func (b *BitstreamWriter) WriteHuffman(value []byte) error {
	encodedBuffer := new(bytes.Buffer)
	encodedStream := &BitstreamWriter{bitstream.NewWriter(encodedBuffer)}

	bitLen, err := englishTree.encodeArray(encodedStream, value)
	if err != nil {
		return err
	}
	err = encodedStream.Flush(bitstream.Bit(false))
	if err != nil {
		return err
	}

	err = b.WriteUint32BE(uint32(len(value)))
	if err != nil {
		return err
	}
	err = b.WriteUint32BECompressed(uint32(bitLen))
	if err != nil {
		return err
	}
    n, err := b.Write(encodedBuffer.Bytes())
    if err != nil {
        return err
    } else if n < encodedBuffer.Len() {
        return errors.New("couldn't write enough bytes")
    }
    return nil
}

func (b *BitstreamWriter) WriteCompressed(value []byte, length uint32, isUnsigned bool) error {
	var byteMatch, halfByteMatch byte
	var err error
	if !isUnsigned {
		byteMatch = 0xFF
		halfByteMatch = 0xF0
	}
	var currentByte uint32
	for currentByte = length>>3 - 1; currentByte > 0; currentByte-- {
		isMatch := value[currentByte] == byteMatch
		err = b.WriteBool(isMatch)
		if err != nil {
			return err
		}
		if !isMatch {
            n, err := b.Write(value[:currentByte+1])
            if err != nil {
                return err
            } else if n < int(currentByte+1) {
                return errors.New("couldn't write enough bytes")
            }
            return nil
		}
	}
	lastByte := value[0]
	if lastByte&0xF0 == halfByteMatch {
		err = b.WriteBool(true)
		if err != nil {
			return err
		}
		return b.WriteBits(uint64(lastByte), 4)
	}
	err = b.WriteBool(false)
	if err != nil {
		return err
	}
	return b.WriteByte(lastByte)
}
func (b *BitstreamWriter) WriteUint32BECompressed(value uint32) error {
	val := make([]byte, 4)
	binary.BigEndian.PutUint32(val, value)
	return b.WriteCompressed(val, 32, true)
}

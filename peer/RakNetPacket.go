package peer

import "bytes"

import "github.com/gskartwii/go-bitstream"

func bufferToStream(buffer []byte) *extendedReader {
	return &extendedReader{bitstream.NewReader(bytes.NewReader(buffer))}
}

func bitsToBytes(bits uint) uint {
	return (bits + 7) >> 3
}

func bytesToBits(bytes uint) uint {
	return bytes << 3
}

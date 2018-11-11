package peer

import "bytes"

import "github.com/gskartwii/go-bitstream"
import "github.com/gskartwii/roblox-dissector/bitstreams"
import "github.com/gskartwii/roblox-dissector/packets"

func bufferToStream(buffer []byte) *PacketReaderBitstream {
	return &packets.PacketReaderBitstream{&bitstreams.BitstreamReader{bitstream.NewReader(bytes.NewReader(buffer))}}
}

func bitsToBytes(bits uint) uint {
	return (bits + 7) >> 3
}

func bytesToBits(bytes uint) uint {
	return bytes << 3
}

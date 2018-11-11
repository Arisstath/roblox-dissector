package packets

import "github.com/gskartwii/roblox-dissector/bitstreams"

type PacketReaderBitstream struct {
    *bitstreams.BitstreamReader
}

type PacketWriterBitstream struct {
    *bitstreams.BitstreamWriter
}

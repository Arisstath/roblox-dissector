package peer

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/DataDog/zstd"

	"github.com/olebedev/emitter"
)

type RawJoinDataBuffer struct {
	*Packet83_0B
	buf []byte
}

func (buf *RawJoinDataBuffer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	_, err := stream.Write(buf.buf)
	return err
}

// MaxJoinDataBytes tells how many bytes a JoinData layer can use at most
// Generally speaking, Roblox can usually handle about 100 kB
// but we set the limit a little lower here to be safe
// TODO: Move this to JoinDataStreamer?
const MaxJoinDataBytes = 80000

type countWriter struct {
	numBytes int
}

func (w *countWriter) Write(b []byte) (int, error) {
	thisLen := len(b)
	w.numBytes += thisLen

	return thisLen, nil
}

func newCountWriter() *countWriter {
	return &countWriter{}
}

// JoinDataStreamer is a helper struct that allows serialized
// JoinData objects to be created one at a time, while still constructing
// JoinData layers of appropriate length
type JoinDataStreamer struct {
	// BufferEmitter emits Packet83Subpackets on channel "join-data"
	// These buffers should be passed to PacketWriter.WritePacket()
	BufferEmitter    *emitter.Emitter
	compressedBuffer *bytes.Buffer
	compressor       *zstd.Writer
	writer           *joinSerializeWriter
	counter          *countWriter
	rawLayer         *Packet83_0B
	packetWriter     PacketWriter
}

func NewJoinDataStreamer(writer PacketWriter) *JoinDataStreamer {
	streamer := &JoinDataStreamer{
		BufferEmitter: emitter.New(0),
		packetWriter:  writer,
	}
	streamer.makeNewStream()

	return streamer
}

func (state *JoinDataStreamer) makeNewStream() *joinSerializeWriter {
	state.compressedBuffer = bytes.NewBuffer(nil)
	state.counter = newCountWriter()
	state.rawLayer = NewPacket83_0BLayer()
	state.compressor = zstd.NewWriter(state.compressedBuffer)

	writeMux := io.MultiWriter(state.compressor, state.counter)
	state.writer = &joinSerializeWriter{&extendedWriter{writeMux}}

	return state.writer
}

func (state *JoinDataStreamer) Flush() error {
	// If there's nothing to write, skip
	if len(state.rawLayer.Instances) == 0 {
		return nil
	}

	err := state.compressor.Close()
	if err != nil {
		return err
	}

	cachedBuffer := make([]byte, state.compressedBuffer.Len()+4+4+4)
	binary.BigEndian.PutUint32(cachedBuffer[:4], uint32(len(state.rawLayer.Instances)))
	binary.BigEndian.PutUint32(cachedBuffer[4:8], uint32(state.compressedBuffer.Len()))
	binary.BigEndian.PutUint32(cachedBuffer[8:12], uint32(state.counter.numBytes))

	copy(cachedBuffer[12:], state.compressedBuffer.Bytes())
	thisBuf := &RawJoinDataBuffer{
		Packet83_0B: state.rawLayer,
		buf:         cachedBuffer,
	}

	<-state.BufferEmitter.Emit("join-data", thisBuf)
	return nil
}

func (state *JoinDataStreamer) Close() error {
	err := state.Flush()
	if err != nil {
		return err
	}
	state.BufferEmitter.Off("*")
	return nil
}

func (state *JoinDataStreamer) AddInstance(instance *ReplicationInstance) error {
	if state.compressedBuffer.Len() > MaxJoinDataBytes {
		err := state.Flush()
		if err != nil {
			return err
		}
		state.makeNewStream()
	}

	state.rawLayer.Instances = append(state.rawLayer.Instances, instance)

	// TODO: Caching?
	return instance.Serialize(state.packetWriter, state.writer)
}

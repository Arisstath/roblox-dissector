package peer

import (
	"bytes"
	"sort"

	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
)

func min(x, y uint) uint {
	if x < y {
		return x
	}
	return y
}

// PacketWriter is an interface that can be passed to packet serializers
type PacketWriter interface {
	ContextualHandler
	SetToClient(bool)
	ToClient() bool
}

// DefaultPacketWriter is a struct used to write packets to a peer
// Pass packets in using WriteOffline/WriteGeneric/etc.
// and bind to the given emitters
type DefaultPacketWriter struct {
	contextualHandler
	// LayerEmitter provides a low-level interface for hooking into the
	// packet serialization process
	// Topics: full-reliable, offline, reliable, reliability, ack
	LayerEmitter *emitter.Emitter

	// ErrorEmitter never emits anything. It exists for compatibility
	ErrorEmitter *emitter.Emitter

	// Output sends the byte slice to be sent via UDP
	// It uses the "output" topic
	Output          *emitter.Emitter
	orderingIndex   uint32
	sequencingIndex uint32
	splitPacketID   uint16
	reliableNumber  uint32
	datagramNumber  uint32
	// Set this to true if the packets produced by this writer are sent to a client.
	toClient bool
}

// NewPacketWriter initializes a new DefaultPacketWriter
func NewPacketWriter() *DefaultPacketWriter {
	return &DefaultPacketWriter{
		// Ordering on output doesn't matter, hence we can set the cap high
		Output:       emitter.New(8),
		LayerEmitter: emitter.New(0),
		ErrorEmitter: emitter.New(0),

		contextualHandler: contextualHandler{
			caches:        new(Caches),
			sharedStrings: make(map[string]rbxfile.ValueSharedString),
		},
	}
}

// ToClient implements PacketWriter.ToClient
func (writer *DefaultPacketWriter) ToClient() bool {
	return writer.toClient
}

// SetToClient implements PacketWriter.SetToClient
func (writer *DefaultPacketWriter) SetToClient(val bool) {
	writer.toClient = val
}

func (writer *DefaultPacketWriter) output(bytes []byte) {
	<-writer.Output.Emit("udp", bytes)
}

// WriteOffline is used to write pre-connection packets (IDs 5-8). It doesn't use a
// ReliabilityLayer.
func (writer *DefaultPacketWriter) WriteOffline(packet RakNetPacket) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{buffer}

	err := stream.WriteByte(packet.Type())
	if err != nil {
		return err
	}
	err = stream.allBytes(OfflineMessageID)
	if err != nil {
		return err
	}

	err = packet.Serialize(writer, stream)
	if err != nil {
		return err
	}
	layers := &PacketLayers{
		PacketType:     packet.Type(),
		Main:           packet,
		OfflinePayload: buffer.Bytes(),
	}

	writer.output(layers.OfflinePayload)
	<-writer.LayerEmitter.Emit("offline", layers)
	return nil
}

// WriteRakNet writes the RakNetLayer contained in the PacketLayers
func (writer *DefaultPacketWriter) WriteRakNet(layers *PacketLayers) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{buffer}
	packet := layers.RakNet
	err := packet.Serialize(writer, stream)
	if err != nil {
		return err
	}

	writer.output(buffer.Bytes())
	return nil
}

func (writer *DefaultPacketWriter) createRakNet(packet *ReliabilityLayer, layers *PacketLayers) (*RakNetLayer, error) {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output)
	stream := &extendedWriter{buffer}
	err := packet.Serialize(writer, stream)
	if err != nil {
		return nil, err
	}

	payload := buffer.Bytes()
	raknet := &RakNetLayer{
		payload: bufferToStream(payload),
		Flags: RakNetFlags{
			IsValid: true,
		},
		DatagramNumber: writer.datagramNumber,
	}
	writer.datagramNumber++

	return raknet, nil
}

func (writer *DefaultPacketWriter) writeAsSplits(estHeaderLength int, data []byte, layers *PacketLayers) error {
	packet := layers.Reliability
	reliability := packet.Reliability
	realLen := len(data)
	splitBandwidth := 1472 - estHeaderLength
	requiredSplits := (realLen + splitBandwidth - 1) / splitBandwidth
	packet.HasSplitPacket = true
	packet.SplitPacketID = writer.splitPacketID
	writer.splitPacketID++
	packet.SplitPacketCount = uint32(requiredSplits)

	packet.SplitBuffer = &SplitPacketBuffer{
		NumReceivedSplits:  0,
		NextExpectedPacket: 0,
		RealLength:         uint32(realLen),
		ReliablePackets:    make([]*ReliablePacket, requiredSplits),
		RakNetPackets:      make([]*RakNetLayer, requiredSplits),
		HasPacketType:      true,
		PacketType:         layers.PacketType,
		UniqueID:           writer.context.uniqueID,
		Data:               data,
	}
	writer.context.uniqueID++
	layers.SplitPacket = packet.SplitBuffer

	var lastLayers *PacketLayers
	for i := 0; i < requiredSplits; i++ {
		thisPacket := packet.Copy()
		newLayers := &PacketLayers{
			Reliability: thisPacket,
			Timestamp:   layers.Timestamp,
			Main:        layers.Main,
			PacketType:  layers.PacketType,
			SplitPacket: layers.SplitPacket,
		}

		thisPacket.SelfData = data[splitBandwidth*i : min(uint(realLen), uint(splitBandwidth*(i+1)))]
		thisPacket.SplitPacketIndex = uint32(i)

		if reliability == Reliable || reliability == ReliableOrdered || reliability == ReliableSequenced {
			thisPacket.ReliableMessageNumber = writer.reliableNumber
			writer.reliableNumber++
		}

		thisReliabilityLayer := &ReliabilityLayer{[]*ReliablePacket{thisPacket}}
		thisRakNet, err := writer.createRakNet(thisReliabilityLayer, newLayers)
		newLayers.RakNet = thisRakNet
		newLayers.SplitPacket.ReliablePackets[i] = thisPacket
		newLayers.SplitPacket.RakNetPackets[i] = thisRakNet
		newLayers.SplitPacket.NumReceivedSplits = uint32(i + 1)
		newLayers.SplitPacket.NextExpectedPacket = uint32(i)
		newLayers.SplitPacket.IsFinal = i == requiredSplits-1

		<-writer.LayerEmitter.Emit("reliability", newLayers)
		<-writer.LayerEmitter.Emit("reliable", newLayers)

		err = writer.WriteRakNet(newLayers)
		if err != nil {
			return err
		}
		lastLayers = newLayers
	}
	<-writer.LayerEmitter.Emit("full-reliable", lastLayers)

	return nil
}

func (writer *DefaultPacketWriter) writeReliablePacket(data []byte, layers *PacketLayers, reliability uint8) error {
	realLen := len(data)
	estHeaderLength := 0x1C // UDP
	estHeaderLength += 4    // RakNet
	estHeaderLength++       // Reliability, has split
	estHeaderLength += 2    // len

	packet := &ReliablePacket{Reliability: reliability}
	layers.Reliability = packet

	if reliability >= 2 && reliability <= 4 {
		packet.ReliableMessageNumber = writer.reliableNumber
		writer.reliableNumber++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 4 {
		packet.SequencingIndex = writer.sequencingIndex
		writer.sequencingIndex++
		estHeaderLength += 3
	}
	if reliability == 1 || reliability == 3 || reliability == 4 || reliability == 7 {
		packet.OrderingChannel = 0
		packet.OrderingIndex = writer.orderingIndex
		writer.orderingIndex++
		estHeaderLength += 7
	}

	if realLen <= 1492-estHeaderLength { // Don't need to split
		packet.SelfData = data
		packet.LengthInBits = uint16(realLen * 8)
		packet.SplitPacketCount = 1

		thisPacketContainer := []*ReliablePacket{packet}
		rakNet, err := writer.createRakNet(&ReliabilityLayer{thisPacketContainer}, layers)
		if err != nil {
			return err
		}
		packet.SplitBuffer = &SplitPacketBuffer{
			IsFinal:            true,
			NumReceivedSplits:  1,
			NextExpectedPacket: 1,
			RealLength:         uint32(realLen),
			ReliablePackets:    thisPacketContainer,
			RakNetPackets:      []*RakNetLayer{rakNet},
			HasPacketType:      true,
			PacketType:         layers.PacketType,
			UniqueID:           writer.context.uniqueID,
			Data:               data,
		}
		writer.context.uniqueID++
		layers.RakNet = rakNet
		layers.SplitPacket = packet.SplitBuffer

		<-writer.LayerEmitter.Emit("reliability", layers)
		<-writer.LayerEmitter.Emit("reliable", layers)
		<-writer.LayerEmitter.Emit("full-reliable", layers)

		return writer.WriteRakNet(layers)
	}

	return writer.writeAsSplits(estHeaderLength, data, layers)
}

func (writer *DefaultPacketWriter) writeTimestamped(layers *PacketLayers, reliability uint8) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &extendedWriter{buffer}
	timestamp := layers.Timestamp
	generic := layers.Main
	err := stream.WriteByte(timestamp.Type())
	if err != nil {
		return err
	}
	err = timestamp.Serialize(writer, stream)
	if err != nil {
		return err
	}
	err = stream.WriteByte(generic.Type())
	if err != nil {
		return err
	}
	err = generic.Serialize(writer, stream)
	if err != nil {
		return err
	}

	err = writer.writeReliablePacket(buffer.Bytes(), layers, reliability)

	return err
}

func (writer *DefaultPacketWriter) writeGeneric(layers *PacketLayers, reliability uint8) error {
	output := make([]byte, 0, 1492)
	buffer := bytes.NewBuffer(output) // Will allocate more if needed
	stream := &extendedWriter{buffer}
	generic := layers.Main
	err := stream.WriteByte(layers.Main.Type())
	if err != nil {
		return err
	}
	err = generic.Serialize(writer, stream)
	if err != nil {
		return err
	}

	err = writer.writeReliablePacket(buffer.Bytes(), layers, reliability)

	return err
}

// WritePacket serializes the given RakNetPacket and outputs it.
// It uses the ReliableOrdered reliability setting.
func (writer *DefaultPacketWriter) WritePacket(generic RakNetPacket) error {
	layers := &PacketLayers{
		Main:       generic,
		PacketType: generic.Type(),
	}
	return writer.writeGeneric(layers, ReliableOrdered)
}

// WriteTimestamped serializes the given RakNetPacket using the given timestamp
// It uses the Unreliable reliability setting.
func (writer *DefaultPacketWriter) WriteTimestamped(timestamp *Packet1BLayer, generic RakNetPacket) error {
	layers := &PacketLayers{
		Timestamp:  timestamp,
		Main:       generic,
		PacketType: generic.Type(),
	}
	return writer.writeTimestamped(layers, Unreliable)
}

// WriteACKs writes an ACK/NAK packet for the given datagram numbers
func (writer *DefaultPacketWriter) WriteACKs(datagrams []int, isNAK bool) error {
	var ackStructure []ACKRange
	sort.Ints(datagrams)

	for _, ack := range datagrams {
		if len(ackStructure) == 0 {
			ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
			continue
		}

		inserted := false
		for i, ackRange := range ackStructure {
			if int(ackRange.Max) == ack {
				inserted = true
				break
			}
			if int(ackRange.Max+1) == ack {
				ackStructure[i].Max++
				inserted = true
				break
			}
		}
		if inserted {
			continue
		}

		ackStructure = append(ackStructure, ACKRange{uint32(ack), uint32(ack)})
	}

	result := &RakNetLayer{
		Flags: RakNetFlags{
			IsValid: true,
			IsACK:   !isNAK,
			IsNAK:   isNAK,
		},
		ACKs: ackStructure,
	}
	layers := &PacketLayers{
		RakNet: result,
	}
	<-writer.LayerEmitter.Emit("ack", layers)

	return writer.WriteRakNet(layers)
}

// Layers returns the emitter that emits packet layers while they are
// being generated
func (writer *DefaultPacketWriter) Layers() *emitter.Emitter {
	return writer.LayerEmitter
}

// Errors returns a no-op emitter
func (writer *DefaultPacketWriter) Errors() *emitter.Emitter {
	return writer.ErrorEmitter
}

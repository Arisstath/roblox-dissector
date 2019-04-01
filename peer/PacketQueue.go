package peer

type reliableMessageState struct {
	hasHandled map[uint32]bool // TODO: Better implementation
}

type sequenceState struct {
	highestIndex uint32
}
type orderingQueue struct {
	index [32]uint32
	queue [32]map[uint32]*PacketLayers // TODO: Use linked lists!
}

func (q *orderingQueue) add(layers *PacketLayers) {
	packet := layers.Reliability
	q.queue[packet.OrderingChannel][packet.OrderingIndex] = layers
}

func (q *orderingQueue) remove(layers *PacketLayers) {
	packet := layers.Reliability
	delete(q.queue[packet.OrderingChannel], packet.OrderingIndex)
}

func (q *orderingQueue) next(channel uint8) *PacketLayers {
	ordReadIndex := q.index[channel]
	packet, ok := q.queue[channel][ordReadIndex]
	if !ok {
		return nil
	}
	if packet.Reliability.SplitBuffer.IsFinal {
		q.index[channel]++
		q.remove(packet)
		return packet
	}

	return nil
}

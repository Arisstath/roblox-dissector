package peer

import "sync"

type rmNumberQueue struct {
	index uint32
	queue map[uint32]*PacketLayers // TODO: Use a linked list!
	lock  *sync.Mutex
}
type sequenceQueue rmNumberQueue
type orderingQueue struct {
	index [32]uint32
	queue [32]map[uint32]*PacketLayers // TODO: Use linked lists!
	lock  *sync.Mutex
}

type queues struct {
	rmNumberQueue *rmNumberQueue
	sequenceQueue *sequenceQueue
	orderingQueue *orderingQueue
}

func (q *queues) add(layers *PacketLayers) {
	packet := layers.Reliability
	reliability := packet.Reliability
	hasRMN := reliability >= 2 && reliability <= 4
	hasSeq := reliability == 1 || reliability == 4
	hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
	if hasRMN {
		q.rmNumberQueue.lock.Lock()
		q.rmNumberQueue.queue[packet.ReliableMessageNumber] = layers
		q.rmNumberQueue.lock.Unlock()
	}
	if hasSeq {
		q.sequenceQueue.lock.Lock()
		q.sequenceQueue.queue[packet.SequencingIndex] = layers
		q.sequenceQueue.lock.Unlock()
	}
	if hasOrd {
		q.orderingQueue.lock.Lock()
		q.orderingQueue.queue[packet.OrderingChannel][packet.OrderingIndex] = layers
		q.orderingQueue.lock.Unlock()
	}
}

func (q *queues) remove(layers *PacketLayers) {
	packet := layers.Reliability
	reliability := packet.Reliability
	hasRMN := reliability >= 2 && reliability <= 4
	hasSeq := reliability == 1 || reliability == 4
	hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
	if hasRMN {
		q.rmNumberQueue.lock.Lock()
		delete(q.rmNumberQueue.queue, packet.ReliableMessageNumber)
		q.rmNumberQueue.lock.Unlock()
	}
	if hasSeq {
		q.sequenceQueue.lock.Lock()
		delete(q.sequenceQueue.queue, packet.SequencingIndex)
		q.sequenceQueue.lock.Unlock()
	}
	if hasOrd {
		q.orderingQueue.lock.Lock()
		delete(q.orderingQueue.queue[packet.OrderingChannel], packet.OrderingIndex)
		q.orderingQueue.lock.Unlock()
	}
}

func (q *queues) next(channel uint8) *PacketLayers {
	rmnReadIndex := q.rmNumberQueue.index
	q.rmNumberQueue.lock.Lock()
	packet, ok := q.rmNumberQueue.queue[rmnReadIndex]
	if ok {
		q.rmNumberQueue.index++
		reliability := packet.Reliability.Reliability
		hasSeq := reliability == 1 || reliability == 4
		hasOrd := reliability == 1 || reliability == 4 || reliability == 3 || reliability == 7
		if !hasSeq && !hasOrd {
			q.rmNumberQueue.lock.Unlock()
			return packet
		}
	}
	q.rmNumberQueue.lock.Unlock()

	q.orderingQueue.lock.Lock()
	ordReadIndex := q.orderingQueue.index[channel]
	packet, ok = q.orderingQueue.queue[channel][ordReadIndex]
	if !ok {
		q.orderingQueue.lock.Unlock()
		return nil
	}
	if packet.Reliability.SplitBuffer.IsFinal {
		q.orderingQueue.index[channel]++
	}
	q.orderingQueue.lock.Unlock()

	reliability := packet.Reliability.Reliability
	hasSeq := reliability == 1 || reliability == 4
	if !hasSeq {
		return packet
	}

	q.sequenceQueue.lock.Lock()
	if packet.Reliability.SequencingIndex == q.sequenceQueue.index {
		if packet.Reliability.SplitBuffer.IsFinal {
			q.sequenceQueue.index++
		}
		q.sequenceQueue.lock.Unlock()
		return packet
	}
	q.sequenceQueue.lock.Unlock()
	return nil
}

func newPeerQueues() *queues {
	thisOrderingQueue := [32]map[uint32]*PacketLayers{}
	for i := 0; i < 32; i++ {
		thisOrderingQueue[i] = make(map[uint32]*PacketLayers)
	}
	return &queues{
		rmNumberQueue: &rmNumberQueue{
			index: 0,
			queue: make(map[uint32]*PacketLayers),
			lock:  new(sync.Mutex),
		},
		sequenceQueue: &sequenceQueue{
			index: 0,
			queue: make(map[uint32]*PacketLayers),
			lock:  new(sync.Mutex),
		},
		orderingQueue: &orderingQueue{
			queue: thisOrderingQueue,
			lock:  new(sync.Mutex),
		},
	}
}

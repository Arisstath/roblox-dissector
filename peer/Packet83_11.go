package peer

type MemoryStatsItem struct {
	Name   string
	Memory float64
}
type ServerMemoryStats struct {
	TotalServerMemory  float64
	DeveloperTags      []MemoryStatsItem
	InternalCategories []MemoryStatsItem
}
type DataStoreStats struct {
	Enabled                 bool
	GetAsync                uint32
	SetAndIncrementAsync    uint32
	UpdateAsync             uint32
	GetSortedAsync          uint32
	SetIncrementSortedAsync uint32
	OnUpdate                uint32
}

type JobStatsItem struct {
	Name  string
	Stat1 float32
	Stat2 float32
	Stat3 float32
}
type ScriptStatsItem struct {
	Name  string
	Stat1 float32
	Stat2 uint32
}

type Packet83_11 struct {
	Version uint32

	MemoryStats    ServerMemoryStats
	DataStoreStats DataStoreStats
	JobStats       []JobStatsItem
	ScriptStats    []ScriptStatsItem

	AvgPingMs             float32
	AvgPhysicsSenderPktPS float32
	TotalDataKBPS         float32
	TotalPhysicsKBPS      float32
	DataThroughputRatio   float32
}

func (thisBitstream *extendedReader) readMemoryStats() ([]MemoryStatsItem, error) {
	numItems, err := thisBitstream.readUint32BE()
	if err != nil {
		return nil, err
	}
	memoryStats := make([]MemoryStatsItem, numItems)
	for i := range memoryStats {
		name, err := thisBitstream.readUint32AndString()
		if err != nil {
			return memoryStats, err
		}
		memoryStats[i].Name = name.(string)

		memoryStats[i].Memory, err = thisBitstream.readFloat64BE()
		if err != nil {
			return memoryStats, err
		}
	}

	return memoryStats, nil
}

func (thisBitstream *extendedReader) DecodePacket83_11(reader PacketReader, layers *PacketLayers) (Packet83Subpacket, error) {
	var err error
	inner := &Packet83_11{}
	inner.Version, err = thisBitstream.readUint32BE()
	if err != nil {
		return inner, err
	}

	if inner.Version >= 5 {
		inner.MemoryStats.TotalServerMemory, err = thisBitstream.readFloat64BE()
		if err != nil {
			return inner, err
		}

		inner.MemoryStats.DeveloperTags, err = thisBitstream.readMemoryStats()
		if err != nil {
			return inner, err
		}
		inner.MemoryStats.InternalCategories, err = thisBitstream.readMemoryStats()
		if err != nil {
			return inner, err
		}
	}

	if inner.Version >= 3 {
		inner.DataStoreStats.Enabled, err = thisBitstream.readBoolByte()
		if err != nil {
			return inner, err
		}
		if inner.DataStoreStats.Enabled {
			inner.DataStoreStats.GetAsync, err = thisBitstream.readUint32BE()
			if err != nil {
				return inner, err
			}
			inner.DataStoreStats.SetAndIncrementAsync, err = thisBitstream.readUint32BE()
			if err != nil {
				return inner, err
			}
			inner.DataStoreStats.UpdateAsync, err = thisBitstream.readUint32BE()
			if err != nil {
				return inner, err
			}
			inner.DataStoreStats.GetSortedAsync, err = thisBitstream.readUint32BE()
			if err != nil {
				return inner, err
			}
			inner.DataStoreStats.SetIncrementSortedAsync, err = thisBitstream.readUint32BE()
			if err != nil {
				return inner, err
			}
			inner.DataStoreStats.OnUpdate, err = thisBitstream.readUint32BE()
			if err != nil {
				return inner, err
			}
		}
	}

	for isEnd, err := thisBitstream.readBoolByte(); !isEnd && err == nil; isEnd, err = thisBitstream.readBoolByte() {
		newJobItem := JobStatsItem{}
		println("reading a job")
		name, err := thisBitstream.readUint32AndString()
		if err != nil {
			return inner, err
		}
		newJobItem.Name = name.(string)
		println("job:", newJobItem.Name)

		newJobItem.Stat1, err = thisBitstream.readFloat32BE()
		if err != nil {
			return inner, err
		}
		newJobItem.Stat2, err = thisBitstream.readFloat32BE()
		if err != nil {
			return inner, err
		}
		newJobItem.Stat3, err = thisBitstream.readFloat32BE()
		if err != nil {
			return inner, err
		}

		inner.JobStats = append(inner.JobStats, newJobItem)
	}
	if err != nil {
		return inner, err
	}

	for isEnd, err := thisBitstream.readBoolByte(); !isEnd && err == nil; isEnd, err = thisBitstream.readBoolByte() {
		newScriptItem := ScriptStatsItem{}
		println("reading a script")
		name, err := thisBitstream.readUint32AndString()
		if err != nil {
			return inner, err
		}
		newScriptItem.Name = name.(string)
		println("script name:", newScriptItem.Name)

		newScriptItem.Stat1, err = thisBitstream.readFloat32BE()
		if err != nil {
			return inner, err
		}
		newScriptItem.Stat2, err = thisBitstream.readUint32BE()
		if err != nil {
			return inner, err
		}
		inner.ScriptStats = append(inner.ScriptStats, newScriptItem)
	}
	if err != nil {
		return inner, err
	}

	inner.AvgPingMs, err = thisBitstream.readFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.AvgPhysicsSenderPktPS, err = thisBitstream.readFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.TotalDataKBPS, err = thisBitstream.readFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.TotalPhysicsKBPS, err = thisBitstream.readFloat32BE()
	if err != nil {
		return inner, err
	}
	inner.DataThroughputRatio, err = thisBitstream.readFloat32BE()
	if err != nil {
		return inner, err
	}

	return inner, nil
}

func (stream *extendedWriter) writeMemoryStats(stats []MemoryStatsItem) error {
	err := stream.writeUint32BE(uint32(len(stats)))
	if err != nil {
		return err
	}

	for _, stat := range stats {
		err = stream.writeUint32AndString(stat.Name)
		if err != nil {
			return err
		}
		err = stream.writeFloat64BE(stat.Memory)
		if err != nil {
			return err
		}
	}

	return nil
}

func (layer *Packet83_11) Serialize(writer PacketWriter, stream *extendedWriter) error {
	err := stream.writeUint32BE(layer.Version)
	if err != nil {
		return err
	}

	if layer.Version >= 5 {
		err = stream.writeFloat64BE(layer.MemoryStats.TotalServerMemory)
		if err != nil {
			return err
		}
		err = stream.writeMemoryStats(layer.MemoryStats.DeveloperTags)
		if err != nil {
			return err
		}
		err = stream.writeMemoryStats(layer.MemoryStats.InternalCategories)
		if err != nil {
			return err
		}
	}

	if layer.Version >= 3 {
		err = stream.writeBoolByte(layer.DataStoreStats.Enabled)
		if err != nil {
			return err
		}
		if layer.DataStoreStats.Enabled {
			err = stream.writeUint32BE(layer.DataStoreStats.GetAsync)
			if err != nil {
				return err
			}
			err = stream.writeUint32BE(layer.DataStoreStats.SetAndIncrementAsync)
			if err != nil {
				return err
			}
			err = stream.writeUint32BE(layer.DataStoreStats.UpdateAsync)
			if err != nil {
				return err
			}
			err = stream.writeUint32BE(layer.DataStoreStats.GetSortedAsync)
			if err != nil {
				return err
			}
			err = stream.writeUint32BE(layer.DataStoreStats.SetIncrementSortedAsync)
			if err != nil {
				return err
			}
			err = stream.writeUint32BE(layer.DataStoreStats.OnUpdate)
			if err != nil {
				return err
			}
		}
	}

	for _, job := range layer.JobStats {
		// write isEnd
		err = stream.writeBoolByte(false)
		if err != nil {
			return err
		}
		err = stream.writeUint32AndString(job.Name)
		if err != nil {
			return err
		}
		err = stream.writeFloat32BE(job.Stat1)
		if err != nil {
			return err
		}
		err = stream.writeFloat32BE(job.Stat2)
		if err != nil {
			return err
		}
		err = stream.writeFloat32BE(job.Stat3)
		if err != nil {
			return err
		}
	}
	// write isEnd
	err = stream.writeBoolByte(true)

	for _, script := range layer.ScriptStats {
		// write isEnd
		err = stream.writeBoolByte(false)
		if err != nil {
			return err
		}
		err = stream.writeUint32AndString(script.Name)
		if err != nil {
			return err
		}
		err = stream.writeFloat32BE(script.Stat1)
		if err != nil {
			return err
		}
		err = stream.writeUint32BE(script.Stat2)
		if err != nil {
			return err
		}
	}
	// write isEnd
	err = stream.writeBoolByte(true)
	if err != nil {
		return err
	}

	err = stream.writeFloat32BE(layer.AvgPingMs)
	if err != nil {
		return err
	}
	err = stream.writeFloat32BE(layer.AvgPhysicsSenderPktPS)
	if err != nil {
		return err
	}
	err = stream.writeFloat32BE(layer.TotalDataKBPS)
	if err != nil {
		return err
	}
	err = stream.writeFloat32BE(layer.TotalPhysicsKBPS)
	if err != nil {
		return err
	}
	return stream.writeFloat32BE(layer.DataThroughputRatio)
}

func (Packet83_11) Type() uint8 {
	return 0x11
}
func (Packet83_11) TypeString() string {
	return "ID_REPLIC_STATS"
}

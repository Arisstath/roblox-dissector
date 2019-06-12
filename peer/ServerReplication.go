package peer

import (
	"github.com/olebedev/emitter"
)

func (client *ServerClient) sendReplicatedFirst() error {
	replicatedFirstStreamer := NewJoinDataStreamer(client.DefaultPacketWriter)
	replicatedFirstStreamer.BufferEmitter.On("join-data", func(e *emitter.Event) {
		err := client.WriteDataPackets(e.Args[0].(Packet83Subpacket))
		if err != nil {
			println("replicatedfirst error: ", err.Error())
		}
	}, emitter.Void)

	service := client.DataModel.FindService("ReplicatedFirst")
	config := joinDataConfigFor("ReplicatedFirst")
	err := client.Replicator.AddTopInstance(service, config.ReplicateChildren, config.ReplicateProperties, replicatedFirstStreamer)
	if err != nil {
		return err
	}
	err = replicatedFirstStreamer.Close()
	if err != nil {
		return err
	}
	// Tag: ReplicatedFirst finished!
	return client.WriteDataPackets(&Packet83_10{
		TagID: 12,
	})
}

func (client *ServerClient) sendContainer(streamer *JoinDataStreamer, config joinDataConfig) error {
	service := client.DataModel.FindService(config.ClassName)
	if service != nil {
		return client.Replicator.AddTopInstance(service, config.ReplicateChildren, config.ReplicateProperties, streamer)
	}
	return nil
}

func (client *ServerClient) sendContainers() error {
	var err error

	joinDataStreamer := NewJoinDataStreamer(client.DefaultPacketWriter)
	joinDataStreamer.BufferEmitter.On("join-data", func(e *emitter.Event) {
		err := client.WriteDataPackets(e.Args[0].(Packet83Subpacket))
		if err != nil {
			println("joindata error: ", err.Error())
		}
	}, emitter.Void)
	for _, dataConfig := range joinDataConfiguration {
		// Previously replicated for priority, don't duplicate
		if dataConfig.ClassName != "ReplicatedFirst" {
			err = client.sendContainer(joinDataStreamer, dataConfig)
			if err != nil {
				return err
			}
		}
	}
	err = joinDataStreamer.Close()
	if err != nil {
		return err
	}

	return client.WriteDataPackets(&Packet83_10{
		TagID: 13,
	})
}

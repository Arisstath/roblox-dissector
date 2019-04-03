package peer

import (
	"fmt"
)

// ID_SUBMIT_TICKET - client -> server
type Packet8ALayer struct {
	PlayerId      int64
	ClientTicket  string
	TicketHash    uint32
	DataModelHash string
	// Always 36?
	ProtocolVersion   uint32
	SecurityKey       string
	Platform          string
	RobloxProductName string
	SessionId         string
	GoldenHash        uint32
}

func NewPacket8ALayer() *Packet8ALayer {
	return &Packet8ALayer{}
}

func (stream *extendedReader) DecodePacket8ALayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := NewPacket8ALayer()

	lenBytes := bitsToBytes(uint(layers.Reliability.LengthInBits)) - 1 // -1 for packet id
	thisStream, err := stream.aesDecrypt(int(lenBytes))
	if err != nil {
		return layer, err
	}

	playerId, err := thisStream.readVarsint64()
	if err != nil {
		return layer, err
	}
	layer.PlayerId = playerId
	layers.Root.Logger.Println("playerid", playerId)
	layer.ClientTicket, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("ticket", layer.ClientTicket)
	layer.DataModelHash, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("dmhash", layer.DataModelHash)
	layer.ProtocolVersion, err = thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("protvr", layer.ProtocolVersion)
	layer.SecurityKey, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("key", layer.SecurityKey)
	layer.Platform, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("platform", layer.Platform)
	layer.RobloxProductName, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("prodname", layer.RobloxProductName)
	if !reader.Context().IsStudio {
		hash, err := thisStream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		layer.TicketHash = hash
		layers.Root.Logger.Println("hash", layer.TicketHash)
		hash2, err := thisStream.readUintUTF8()
		if err != nil {
			return layer, err
		}
		layers.Root.Logger.Println("hash2", hash2, "badfood check success: ", hash2 == layer.TicketHash-0xbadf00d)
	}

	layer.SessionId, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("sessid", layer.SessionId)
	layer.GoldenHash, err = thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("goldhash", layer.GoldenHash)

	return layer, nil
}

func (layer *Packet8ALayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error

	err = stream.WriteByte(0x8A)
	if err != nil {
		return err
	}
	rawStream := stream.aesEncrypt()
	err = rawStream.writeVarsint64(layer.PlayerId)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.ClientTicket)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.DataModelHash)
	if err != nil {
		return err
	}
	err = rawStream.writeUint32BE(layer.ProtocolVersion)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.SecurityKey)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.Platform)
	if err != nil {
		return err
	}
	err = rawStream.writeVarLengthString(layer.RobloxProductName)
	if err != nil {
		return err
	}
	if !writer.Context().IsStudio {
		err = rawStream.writeVarint64(uint64(layer.TicketHash))
		if err != nil {
			return err
		}
		err = rawStream.writeVarint64(uint64(layer.TicketHash - 0xbadf00d))
		if err != nil {
			return err
		}
	}
	err = rawStream.writeVarLengthString(layer.SessionId)
	if err != nil {
		return err
	}
	err = rawStream.writeUint32BE(layer.GoldenHash)
	if err != nil {
		return err
	}
	err = rawStream.Close()
	if err != nil {
		return err
	}
	return err
}

func (layer *Packet8ALayer) String() string {
	return fmt.Sprintf("ID_SUBMIT_TICKET: %s", layer.Platform)
}

func (Packet8ALayer) TypeString() string {
	return "ID_SUBMIT_TICKET"
}

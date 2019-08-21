package peer

import (
	"fmt"
)

// Packet8ALayer represents ID_SUBMIT_TICKET - client -> server
type Packet8ALayer struct {
	PlayerID      int64
	ClientTicket  string
	TicketHash    uint32
	LuauResponse  uint32
	DataModelHash string
	// Always 36?
	ProtocolVersion   uint32
	SecurityKey       string
	Platform          string
	RobloxProductName string
	SessionID         string
	GoldenHash        uint32
}

func (stream *extendedReader) DecodePacket8ALayer(reader PacketReader, layers *PacketLayers) (RakNetPacket, error) {
	layer := &Packet8ALayer{}

	lenBytes := bitsToBytes(uint(layers.Reliability.LengthInBits)) - 1 // -1 for packet id
	key := reader.Context().GenerateSubmitTicketKey()
	layers.Root.Logger.Println("using ticket key", string(key[:]))
	thisStream, err := stream.aesDecrypt(int(lenBytes), reader.Context().GenerateSubmitTicketKey())
	if err != nil {
		return layer, err
	}

	playerID, err := thisStream.readVarsint64()
	if err != nil {
		return layer, err
	}
	layer.PlayerID = playerID
	layers.Root.Logger.Println("playerid", playerID)
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
		hash3, err := thisStream.readSintUTF8()
		if err != nil {
			return layer, err
		}
		layer.LuauResponse = uint32(hash3)

		layers.Root.Logger.Println("hash2", hash2, "badfood check success: ", hash2 == layer.TicketHash-0xbadf00d)
		layers.Root.Logger.Println("win10 check: ", hash == Win10Settings().GenerateTicketHash(layer.ClientTicket))
		layers.Root.Logger.Println("win10 check luau: ", hash3 == int32(Win10Settings().GenerateLuauResponse(layer.ClientTicket)))
	}

	layer.SessionID, err = thisStream.readVarLengthString()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("sessid", layer.SessionID)
	layer.GoldenHash, err = thisStream.readUint32BE()
	if err != nil {
		return layer, err
	}
	layers.Root.Logger.Println("goldhash", layer.GoldenHash)

	return layer, nil
}

// Serialize implements RakNetPacket.Serialize
func (layer *Packet8ALayer) Serialize(writer PacketWriter, stream *extendedWriter) error {
	var err error

	rawStream := stream.aesEncrypt(writer.Context().GenerateSubmitTicketKey())
	err = rawStream.writeVarsint64(layer.PlayerID)
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
		err = rawStream.writeVarint64(uint64(layer.LuauResponse))
	}
	err = rawStream.writeVarLengthString(layer.SessionID)
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

// TypeString implements RakNetPacket.TypeString()
func (Packet8ALayer) TypeString() string {
	return "ID_SUBMIT_TICKET"
}

// Type implements RakNetPacket.Type()
func (Packet8ALayer) Type() byte {
	return 0x8A
}

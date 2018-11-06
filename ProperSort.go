package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
)

func NewProperSortModel(parent core.QObject_ITF) *gui.QStandardItemModel {
	standardModel := gui.NewQStandardItemModel(parent)
	return standardModel
}

func NewFilteringModel(parent core.QObject_ITF) (*gui.QStandardItemModel, *core.QSortFilterProxyModel) {
	standardModel := NewProperSortModel(parent)
	proxy := core.NewQSortFilterProxyModel(parent)
	proxy.SetSourceModel(standardModel)

	return standardModel, proxy
}

/*type FilterRule interface {
	Matches(packetType uint8, packet *peer.UDPPacket, layers *peer.PacketLayers, context *peer.CommunicationContext)
	String() string
}
type FilterSetting struct {
	Include bool
	Rule    FilterRule
}

func (setting *FilterSettings) String() string {
	if Include {
		return "Include " + setting.Rule.String()
	}
	return "Exclude " + setting.Rule.String()
}

type FilterSettings []FilterSetting

type BasicPacketTypeFilter struct {
	PacketType uint8
}

func (filter *BasicPacketTypeFilter) Matches(packetType uint8, packet *peer.UDPPacket, layers *peer.PacketLayers, context *peer.CommunicationContext) bool {
	return packetType == filter.PacketType
}
func (filter *BasicPacketTypeFilter) String() string {
	return fmt.Sprintf("PacketType %s", PacketNames[filter.PacketType])
}

type Packet83TypeFilter struct {
	PacketType uint8
}

func (filter *Packet83TypeFilter) Matches(packetType uint8, packet *peer.UDPPacket, layers *peer.PacketLayers, context *peer.CommunicationContext) bool {
	if packetType != 0x83 {
		return false
	}
	mainLayer := layers.Main.(*Packet83Layer)
	for _, packet := range mainLayer.SubPackets {
		if peer.Packet83ToType(packet) == filter.PacketType {
			return true
		}
	}
	return false
}
func (filter *Packet83TypeFilter) String() string {
	return fmt.Sprintf("Data packet %s", peer.Packet83ToTypeString(filter.PacketType))
}

type DirectionFilter struct {
	FromClient bool
}

func (filter *DirectionFilter) Matches(packetType uint8, packet *peer.UDPPacket, layers *peer.PacketLayers, context *peer.CommunicationContext) bool {
	return filter.FromClient == context.IsClient(packet.Source)
}
func (filter *DirectionFilter) String() string {
	if filter.FromClient {
		return "Client -> Server"
	} else {
		return "Server -> Client"
	}
}
*/

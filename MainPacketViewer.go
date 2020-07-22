package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
	"github.com/robloxapi/rbxapi/rbxapijson"
	"github.com/robloxapi/rbxapiref/fetch"
)

// Can't use https:// because the site is broken
const CDNURL = "http://setup.roblox.com/"

var latestRobloxAPI *rbxapijson.Root
var latestRobloxAPIChan chan struct{} // Closed when retrievement is done

func init() {
	latestRobloxAPIChan = make(chan struct{})
	go func() {
		defer func() {
			close(latestRobloxAPIChan)
		}()
		robloxApiClient := &fetch.Client{
			Client: &http.Client{},
			Config: fetch.Config{
				Builds:             []fetch.Location{fetch.NewLocation(CDNURL + "DeployHistory.txt")},
				Latest:             []fetch.Location{fetch.NewLocation(CDNURL + "versionQTStudio")},
				APIDump:            []fetch.Location{fetch.NewLocation(CDNURL + "$HASH-API-Dump.json")},
				ReflectionMetadata: []fetch.Location{fetch.NewLocation(CDNURL + "$HASH-RobloxStudio.zip#ReflectionMetadata.xml")},
				ExplorerIcons:      []fetch.Location{fetch.NewLocation(CDNURL + "$HASH-RobloxStudio.zip#RobloxStudioBeta.exe")},
			},
		}
		latestBuild, err := robloxApiClient.Latest()
		if err != nil {
			fmt.Println("Error retrieving API:", err.Error())
			return
		}
		apiDump, err := robloxApiClient.APIDump(latestBuild.Hash)
		if err != nil {
			fmt.Println("Error retrieving API:", err.Error())
			return
		}
		latestRobloxAPI = apiDump
	}()
}

func capabilitiesToString(cap uint64) string {
	var builder strings.Builder
	if cap&peer.CapabilityBasic == peer.CapabilityBasic {
		cap ^= peer.CapabilityBasic
		builder.WriteString("Basic, ")
	}
	if cap&peer.CapabilityServerCopiesPlayerGui3 != 0 {
		cap ^= peer.CapabilityServerCopiesPlayerGui3
		builder.WriteString("ServerCopiesPlayerGui, ")
	}
	if cap&peer.CapabilityDebugForceStreamingEnabled != 0 {
		cap ^= peer.CapabilityDebugForceStreamingEnabled
		builder.WriteString("DebugForceStreamingEnabled, ")
	}
	if cap&peer.CapabilityIHasMinDistToUnstreamed != 0 {
		cap ^= peer.CapabilityIHasMinDistToUnstreamed
		builder.WriteString("IHasMinDistToUnstreamed, ")
	}
	if cap&peer.CapabilityReplicateLuau != 0 {
		cap ^= peer.CapabilityReplicateLuau
		builder.WriteString("ReplicateLuau, ")
	}
	if cap&peer.CapabilityPositionBasedStreaming != 0 {
		cap ^= peer.CapabilityPositionBasedStreaming
		builder.WriteString("PositionBasedStreaming, ")
	}
	if cap&peer.CapabilityVersionedIDSync != 0 {
		cap ^= peer.CapabilityVersionedIDSync
		builder.WriteString("VersionedIDSync, ")
	}
	if cap&peer.CapabilityPubKeyExchange != 0 {
		cap ^= peer.CapabilityPubKeyExchange
		builder.WriteString("PubKeyExchange, ")
	}
	if cap&peer.CapabilitySystemAddressIsPeerId != 0 {
		cap ^= peer.CapabilitySystemAddressIsPeerId
		builder.WriteString("SystemAddressIsPeerId, ")
	}
	if cap&peer.CapabilityStreamingPrefetch != 0 {
		cap ^= peer.CapabilityStreamingPrefetch
		builder.WriteString("StreamingPrefetch, ")
	}
	if cap&peer.CapabilityTerrainReplicationUseLargerChunks != 0 {
		cap ^= peer.CapabilityTerrainReplicationUseLargerChunks
		builder.WriteString("TerrainReplicationUseLargerChunks , ")
	}
	if cap&peer.CapabilityUseBlake2BHashInSharedString != 0 {
		cap ^= peer.CapabilityUseBlake2BHashInSharedString
		builder.WriteString("UseBlake2BHashInSharedString, ")
	}
	if cap&peer.CapabilityUseSharedStringForScriptReplication != 0 {
		cap ^= peer.CapabilityUseSharedStringForScriptReplication
		builder.WriteString("UseSharedStringForScriptReplication, ")
	}

	if cap != 0 {
		fmt.Fprintf(&builder, "Unknown capabilities: %8X, ", cap)
	}

	if builder.Len() == 0 {
		return ""
	}
	return builder.String()[:builder.Len()-2]
}

func boxWithMargin() (*gtk.Box, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 4)
	if err != nil {
		return nil, err
	}
	box.SetMarginTop(8)
	box.SetMarginBottom(8)
	box.SetMarginStart(8)
	box.SetMarginEnd(8)
	return box, nil
}

func newLabelF(fmtS string, rest ...interface{}) (*gtk.Label, error) {
	label, err := gtk.LabelNew(fmt.Sprintf(fmtS, rest...))
	if err != nil {
		return nil, err
	}
	label.SetHAlign(gtk.ALIGN_START)
	return label, nil
}
func newWrappingLabelF(fmtS string, rest ...interface{}) (*gtk.Label, error) {
	label, err := gtk.LabelNew(fmt.Sprintf(fmtS, rest...))
	if err != nil {
		return nil, err
	}
	label.SetHAlign(gtk.ALIGN_START)

	label.SetLineWrap(true)
	label.SetLineWrapMode(pango.WRAP_WORD_CHAR)
	attrs := pango.AttrListNew()
	attrs.Insert(pango.AttrInsertHyphensNew(false))
	label.SetAttributes(attrs)

	return label, nil
}

func newIpAddressScrolledList(addrs []*net.UDPAddr) (*gtk.ScrolledWindow, error) {
	ipAddrStore, err := gtk.ListStoreNew(glib.TYPE_STRING)
	if err != nil {
		return nil, err
	}
	ipAddrView, err := gtk.TreeViewNewWithModel(ipAddrStore)
	if err != nil {
		return nil, err
	}
	colRender, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	column, err := gtk.TreeViewColumnNewWithAttribute("Address", colRender, "text", 0)
	if err != nil {
		return nil, err
	}
	ipAddrView.AppendColumn(column)
	for _, addr := range addrs {
		row := ipAddrStore.Append()
		err = ipAddrStore.SetValue(row, 0, addr.String())
		if err != nil {
			return nil, err
		}
	}
	ipAddrScrolledView, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	ipAddrScrolledView.Add(ipAddrView)
	ipAddrScrolledView.SetVExpand(true)
	return ipAddrScrolledView, nil
}

func openConnectionReq1Viewer(packet *peer.Packet05Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	versionLabel, err := newLabelF("Version: %d", packet.ProtocolVersion)
	if err != nil {
		return nil, err
	}
	box.Add(versionLabel)
	paddingLenLabel, err := newLabelF("Padding length: %d", packet.MTUPaddingLength)
	if err != nil {
		return nil, err
	}
	box.Add(paddingLenLabel)
	box.ShowAll()
	return box, nil
}
func openConnectionResp1Viewer(packet *peer.Packet06Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	guidLabel, err := newLabelF("GUID: %08X", packet.GUID)
	if err != nil {
		return nil, err
	}
	box.Add(guidLabel)
	securityLabel, err := newLabelF("Use security: %v", packet.UseSecurity)
	if err != nil {
		return nil, err
	}
	box.Add(securityLabel)
	mtuSizeLabel, err := newLabelF("MTU: %d", packet.MTU)
	if err != nil {
		return nil, err
	}
	box.Add(mtuSizeLabel)
	box.ShowAll()
	return box, nil
}
func openConnectionReq2Viewer(packet *peer.Packet07Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	ipAddrLabel, err := newLabelF("IP address: %s", packet.IPAddress)
	if err != nil {
		return nil, err
	}
	box.Add(ipAddrLabel)
	mtuSizeLabel, err := newLabelF("MTU: %d", packet.MTU)
	if err != nil {
		return nil, err
	}
	box.Add(mtuSizeLabel)
	guidLabel, err := newLabelF("GUID: %08X", packet.GUID)
	if err != nil {
		return nil, err
	}
	box.Add(guidLabel)
	supportedVersionLabel, err := newLabelF("Supported version: %d", packet.SupportedVersion)
	if err != nil {
		return nil, err
	}
	box.Add(supportedVersionLabel)
	capabilitiesLabel, err := newLabelF("Capabilities: %s", capabilitiesToString(packet.Capabilities))
	if err != nil {
		return nil, err
	}
	box.Add(capabilitiesLabel)
	box.ShowAll()
	return box, nil
}
func openConnectionResp2Viewer(packet *peer.Packet08Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	guidLabel, err := newLabelF("GUID: %08X", packet.GUID)
	if err != nil {
		return nil, err
	}
	box.Add(guidLabel)
	ipAddrLabel, err := newLabelF("IP address: %s", packet.IPAddress)
	if err != nil {
		return nil, err
	}
	box.Add(ipAddrLabel)
	mtuSizeLabel, err := newLabelF("MTU: %d", packet.MTU)
	if err != nil {
		return nil, err
	}
	box.Add(mtuSizeLabel)
	supportedVersionLabel, err := newLabelF("Supported version: %d", packet.SupportedVersion)
	if err != nil {
		return nil, err
	}
	box.Add(supportedVersionLabel)
	capabilitiesLabel, err := newLabelF("Capabilities: %s", capabilitiesToString(packet.Capabilities))
	if err != nil {
		return nil, err
	}
	box.Add(capabilitiesLabel)
	box.ShowAll()
	return box, nil
}
func connectionRequestViewer(packet *peer.Packet09Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	guidLabel, err := newLabelF("GUID: %08X", packet.GUID)
	if err != nil {
		return nil, err
	}
	box.Add(guidLabel)
	timeLabel, err := newLabelF("Timestamp: %d", packet.Timestamp)
	if err != nil {
		return nil, err
	}
	box.Add(timeLabel)
	securityLabel, err := newLabelF("Use security: %v", packet.UseSecurity)
	if err != nil {
		return nil, err
	}
	box.Add(securityLabel)
	passwordLabel, err := newLabelF("Password: %X", packet.Password)
	if err != nil {
		return nil, err
	}
	box.Add(passwordLabel)

	box.ShowAll()
	return box, nil
}
func connectionAcceptedViewer(packet *peer.Packet10Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	ipAddrLabel, err := newLabelF("IP address: %s", packet.IPAddress)
	if err != nil {
		return nil, err
	}
	box.Add(ipAddrLabel)
	indexLabel, err := newLabelF("System index: %d", packet.SystemIndex)
	if err != nil {
		return nil, err
	}
	box.Add(indexLabel)
	remotesLabel, err := newLabelF("Remote IP addresses:")
	if err != nil {
		return nil, err
	}
	box.Add(remotesLabel)

	ipAddrScrolledView, err := newIpAddressScrolledList(packet.Addresses[:])
	if err != nil {
		return nil, err
	}
	box.Add(ipAddrScrolledView)

	sendPingTime, err := newLabelF("Send ping time: %d", packet.SendPingTime)
	if err != nil {
		return nil, err
	}
	box.Add(sendPingTime)
	sendPongTime, err := newLabelF("Send pong time: %d", packet.SendPongTime)
	if err != nil {
		return nil, err
	}
	box.Add(sendPongTime)

	box.ShowAll()
	return box, nil
}
func newIncomingConnectionViewer(packet *peer.Packet13Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	ipAddrLabel, err := newLabelF("IP address: %s", packet.IPAddress)
	if err != nil {
		return nil, err
	}
	box.Add(ipAddrLabel)
	remotesLabel, err := newLabelF("Remote IP addresses:")
	if err != nil {
		return nil, err
	}
	box.Add(remotesLabel)

	ipAddrScrolledView, err := newIpAddressScrolledList(packet.Addresses[:])
	if err != nil {
		return nil, err
	}
	box.Add(ipAddrScrolledView)

	sendPingTime, err := newLabelF("Send ping time: %d", packet.SendPingTime)
	if err != nil {
		return nil, err
	}
	box.Add(sendPingTime)
	sendPongTime, err := newLabelF("Send pong time: %d", packet.SendPongTime)
	if err != nil {
		return nil, err
	}
	box.Add(sendPongTime)

	box.ShowAll()
	return box, nil
}
func disconnectionNotificationViewer(packet *peer.Packet15Layer) (gtk.IWidget, error) {
	reasonLabel, err := gtk.LabelNew(fmt.Sprintf("Resolving disconnection reason (%d)...", packet.Reason))
	if err != nil {
		return nil, err
	}
	reasonLabel.SetHAlign(gtk.ALIGN_START)
	reasonLabel.SetVAlign(gtk.ALIGN_START)
	reasonLabel.SetMarginTop(8)
	reasonLabel.SetMarginBottom(8)
	reasonLabel.SetMarginStart(8)
	reasonLabel.SetMarginEnd(8)
	if packet.Reason == -1 {
		reasonLabel.SetText("Generic disconnection -1")
	} else {
		handleAPIAvailable := func() {
			if !reasonLabel.IsVisible() {
				return
			}
			disconnectionEnum := latestRobloxAPI.GetEnum("ConnectionError")
			if disconnectionEnum == nil {
				reasonLabel.SetText(fmt.Sprintf("ConnectionError Enum not available; disconnection reason %d", packet.Reason))
				return
			}
			for _, item := range disconnectionEnum.GetEnumItems() {
				if item.GetValue() == int(packet.Reason) {
					reasonLabel.SetText("Reason: " + item.GetName())
					return
				}
			}
			reasonLabel.SetText(fmt.Sprintf("Unknown disconnection reason (%d)", packet.Reason))
		}
		if latestRobloxAPI != nil {
			handleAPIAvailable()
		} else {
			go func() {
				if latestRobloxAPI != nil {
					glib.IdleAdd(func() bool {
						handleAPIAvailable()
						return false
					})
				}
				wait := true
				for wait {
					_, wait = <-latestRobloxAPIChan
				}
				glib.IdleAdd(func() bool {
					handleAPIAvailable()
					return false
				})
			}()
		}
	}
	reasonLabel.ShowAll()

	return reasonLabel, nil
}
func topReplicViewer(packet *peer.Packet81Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	streamJob, err := newLabelF("StreamingEnabled: %v", packet.StreamJob)
	if err != nil {
		return nil, err
	}
	box.Add(streamJob)
	filteringEnabled, err := newLabelF("FilteringEnabled: %v", packet.FilteringEnabled)
	if err != nil {
		return nil, err
	}
	box.Add(filteringEnabled)
	bool1, err := newLabelF("Bool1: %v", packet.Bool1)
	if err != nil {
		return nil, err
	}
	box.Add(bool1)
	bool2, err := newLabelF("Bool2: %v", packet.Bool2)
	if err != nil {
		return nil, err
	}
	box.Add(bool2)
	bool3, err := newLabelF("Bool3: %v", packet.Bool3)
	if err != nil {
		return nil, err
	}
	box.Add(bool3)
	characterAutoSpawn, err := newLabelF("CharacterAutoSpawn (?): %v", packet.CharacterAutoSpawn)
	if err != nil {
		return nil, err
	}
	box.Add(characterAutoSpawn)
	peerId, err := newLabelF("Client peerid: %d", packet.PeerID)
	if err != nil {
		return nil, err
	}
	box.Add(peerId)

	model, err := gtk.ListStoreNew(
		glib.TYPE_INT,     // id
		glib.TYPE_STRING,  // class name
		glib.TYPE_STRING,  // reference
		glib.TYPE_BOOLEAN, // watch changes
		glib.TYPE_BOOLEAN, // watch children
	)
	if err != nil {
		return nil, err
	}
	view, err := gtk.TreeViewNewWithModel(model)
	if err != nil {
		return nil, err
	}
	colRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	for i, name := range []string{"ID", "Class", "Reference", "Watch changes", "Watch children"} {
		col, err := gtk.TreeViewColumnNewWithAttribute(
			name,
			colRenderer,
			"text",
			i,
		)
		if err != nil {
			return nil, err
		}
		view.AppendColumn(col)
	}
	for i, item := range packet.Items {
		row := model.Append()
		model.SetValue(row, 0, i)
		model.SetValue(row, 1, item.Schema.Name)
		model.SetValue(row, 2, item.Instance.Ref.String())
		model.SetValue(row, 3, item.WatchChanges)
		model.SetValue(row, 4, item.WatchChildren)
	}
	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolled.SetVExpand(true)
	scrolled.Add(view)
	separator, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	box.Add(separator)
	label, err := gtk.LabelNew("Top replication containers:")
	if err != nil {
		return nil, err
	}
	label.SetHAlign(gtk.ALIGN_START)
	box.Add(label)
	box.Add(scrolled)

	box.ShowAll()

	return box, nil
}
func submitTicketViewer(packet *peer.Packet8ALayer) (gtk.IWidget, error) {
	scrollWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrollWindow.SetMarginTop(8)
	scrollWindow.SetMarginBottom(8)
	scrollWindow.SetMarginStart(8)
	scrollWindow.SetMarginEnd(8)
	scrollWindow.SetVExpand(true)
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 4)
	if err != nil {
		return nil, err
	}

	playerID, err := newLabelF("Player ID: %d", packet.PlayerID)
	if err != nil {
		return nil, err
	}
	box.Add(playerID)
	clientTicket, err := newWrappingLabelF("Client ticket: %s", packet.ClientTicket)
	if err != nil {
		return nil, err
	}
	box.Add(clientTicket)
	ticketHash, err := newLabelF("Ticket hash: %08X", packet.TicketHash)
	if err != nil {
		return nil, err
	}
	box.Add(ticketHash)
	luauResponse, err := newLabelF("Luau response: %08X", packet.LuauResponse)
	if err != nil {
		return nil, err
	}
	box.Add(luauResponse)
	dataModelHash, err := newLabelF("DataModel hash: %08x", packet.DataModelHash)
	if err != nil {
		return nil, err
	}
	box.Add(dataModelHash)
	protocolVersion, err := newLabelF("Protocol version: %d", packet.ProtocolVersion)
	if err != nil {
		return nil, err
	}
	box.Add(protocolVersion)
	securityKey, err := newWrappingLabelF("Security key: %s", packet.SecurityKey)
	if err != nil {
		return nil, err
	}
	box.Add(securityKey)
	platform, err := newWrappingLabelF("Platform: %s", packet.Platform)
	if err != nil {
		return nil, err
	}
	box.Add(platform)
	robloxProductName, err := newWrappingLabelF("Roblox product name: %s", packet.RobloxProductName)
	if err != nil {
		return nil, err
	}
	box.Add(robloxProductName)
	sessionID, err := newWrappingLabelF("Session ID: %s", packet.SessionID)
	if err != nil {
		return nil, err
	}
	box.Add(sessionID)
	goldenHash, err := newLabelF("Golden hash: %08X", packet.GoldenHash)
	if err != nil {
		return nil, err
	}
	box.Add(goldenHash)

	scrollWindow.Add(box)
	scrollWindow.ShowAll()
	return scrollWindow, nil
}
func clusterViewer(packet *peer.Packet8DLayer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func protocolSyncViewer(packet *peer.Packet90Layer) (gtk.IWidget, error) {
	scrollWindow, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrollWindow.SetMarginTop(8)
	scrollWindow.SetMarginBottom(8)
	scrollWindow.SetMarginStart(8)
	scrollWindow.SetMarginEnd(8)
	scrollWindow.SetVExpand(true)
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 4)
	if err != nil {
		return nil, err
	}

	schemaVersion, err := newLabelF("Schema version: %d", packet.SchemaVersion)
	if err != nil {
		return nil, err
	}
	box.Add(schemaVersion)
	int1, err := newLabelF("Int 1: %d", packet.Int1)
	if err != nil {
		return nil, err
	}
	box.Add(int1)
	int2, err := newLabelF("Int 2: %d", packet.Int2)
	if err != nil {
		return nil, err
	}
	box.Add(int2)

	flagsLabel, err := gtk.LabelNew("Requested FFlags:")
	if err != nil {
		return nil, err
	}
	flagsLabel.SetHAlign(gtk.ALIGN_START)
	box.Add(flagsLabel)

	model, err := gtk.TreeStoreNew(
		glib.TYPE_INT,
		glib.TYPE_STRING,
	)
	if err != nil {
		return nil, err
	}
	view, err := gtk.TreeViewNewWithModel(model)
	if err != nil {
		return nil, err
	}
	renderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	for i, title := range []string{"Index", "Flag name"} {
		col, err := gtk.TreeViewColumnNewWithAttribute(
			title,
			renderer,
			"text",
			i,
		)
		if err != nil {
			return nil, err
		}
		view.AppendColumn(col)
	}
	for i, flag := range packet.RequestedFlags {
		model.InsertWithValues(nil, nil, -1, []int{0, 1}, []interface{}{i, flag})
	}
	view.SetVExpand(true)
	box.Add(view)

	joinData, err := newWrappingLabelF("Join data: %s", packet.JoinData)
	if err != nil {
		return nil, err
	}
	box.Add(joinData)
	pubKeyData, err := newLabelF("Public key data: %X", packet.PubKeyData)
	if err != nil {
		return nil, err
	}
	box.Add(pubKeyData)

	for i, ver := range packet.VersionID {
		label, err := newLabelF("Version id %d: %08X", i, uint32(ver))
		if err != nil {
			return nil, err
		}
		box.Add(label)
	}

	scrollWindow.Add(box)
	scrollWindow.ShowAll()
	return scrollWindow, nil
}
func dictionaryFormatViewer(packet *peer.Packet93Layer) (gtk.IWidget, error) {
	box, err := boxWithMargin()
	if err != nil {
		return nil, err
	}
	protocolSchemaSync, err := newLabelF("Protocol schema sync: %v", packet.ProtocolSchemaSync)
	if err != nil {
		return nil, err
	}
	box.Add(protocolSchemaSync)
	apiDictionaryCompression, err := newLabelF("API dictionary compression: %v", packet.APIDictionaryCompression)
	if err != nil {
		return nil, err
	}
	box.Add(apiDictionaryCompression)

	flagsLabel, err := newLabelF("FFlags:")
	if err != nil {
		return nil, err
	}
	box.Add(flagsLabel)
	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	model, err := gtk.TreeStoreNew(
		glib.TYPE_STRING,
		glib.TYPE_BOOLEAN,
	)
	if err != nil {
		return nil, err
	}
	view, err := gtk.TreeViewNewWithModel(model)
	if err != nil {
		return nil, err
	}
	renderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	for i, title := range []string{"Flag name", "Value"} {
		col, err := gtk.TreeViewColumnNewWithAttribute(
			title,
			renderer,
			"text",
			i,
		)
		if err != nil {
			return nil, err
		}
		view.AppendColumn(col)
	}
	for name, value := range packet.Params {
		model.InsertWithValues(nil, nil, -1, []int{0, 1}, []interface{}{name, value})
	}
	scrolled.Add(view)
	scrolled.SetVExpand(true)
	box.Add(scrolled)

	box.ShowAll()
	return box, nil
}
func schemaViewer(packet *peer.Packet97Layer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}
func luauChallengeViewer(packet *peer.Packet9BLayer) (gtk.IWidget, error) {
	return nil, errors.New("unimplemented")
}

func viewerForMainPacket(packet peer.RakNetPacket) (gtk.IWidget, error) {
	switch packet.Type() {
	case 0x00, 0x03, 0x83, 0x84, 0x85, 0x86, 0x87, 0x8F, 0x92, 0x96, 0x98:
		return blanketViewer(packet.String())
	case 0x05:
		return openConnectionReq1Viewer(packet.(*peer.Packet05Layer))
	case 0x06:
		return openConnectionResp1Viewer(packet.(*peer.Packet06Layer))
	case 0x07:
		return openConnectionReq2Viewer(packet.(*peer.Packet07Layer))
	case 0x08:
		return openConnectionResp2Viewer(packet.(*peer.Packet08Layer))
	case 0x09:
		return connectionRequestViewer(packet.(*peer.Packet09Layer))
	case 0x10:
		return connectionAcceptedViewer(packet.(*peer.Packet10Layer))
	case 0x13:
		return newIncomingConnectionViewer(packet.(*peer.Packet13Layer))
	case 0x15:
		return disconnectionNotificationViewer(packet.(*peer.Packet15Layer))
	case 0x81:
		return topReplicViewer(packet.(*peer.Packet81Layer))
	case 0x8A:
		return submitTicketViewer(packet.(*peer.Packet8ALayer))
	case 0x8D:
		return clusterViewer(packet.(*peer.Packet8DLayer))
	case 0x90:
		return protocolSyncViewer(packet.(*peer.Packet90Layer))
	case 0x93:
		return dictionaryFormatViewer(packet.(*peer.Packet93Layer))
	case 0x97:
		return schemaViewer(packet.(*peer.Packet97Layer))
	case 0x9B:
		return luauChallengeViewer(packet.(*peer.Packet9BLayer))
	default:
		return nil, fmt.Errorf("unimplemented packet type %02X", packet.Type())
	}
}

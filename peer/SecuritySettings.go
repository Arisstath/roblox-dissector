package peer

import (
	"bytes"
	"encoding/json"
	"math/bits"

	"github.com/pierrec/xxHash/xxHash32"
)

type joinData struct {
	CharacterAppearance   string
	GameChatType          string
	FollowUserID          int64  `json:"FollowUserId"`
	OSPlatform            string `json:"OsPlatform"`
	AccountAge            int32
	SuperSafeChat         bool
	VRDevice              string `json:"VrDevice"`
	MembershipType        string
	Locale2IDRef          string `json:"Locale2IdRef"`
	RawJoinData           string
	Locale2ID             string `json:"Locale2Id"`
	UserName              string
	IsTeleportedIn        bool
	LocaleID              string `json:"LocaleId"`
	CharacterAppearanceID int64  `json:"CharacterAppearanceId"`
	UserID                int64  `json:"UserId"`
}

func (d joinData) JSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(&d)
	// drop trailing newline
	return buffer.Bytes()[:buffer.Len()-1], err
}

// SecurityHandler describes an interface that provides emulation
// of a certain Roblox client
type SecurityHandler interface {
	// GenerateIdResponse should provide a response to a challenge
	// given in Packet83_09_05
	GenerateIDResponse(challenge uint32) uint32
	// PatchTicketPacket should change the parameters in a Packet8ALayer
	// appropriately
	PatchTicketPacket(*Packet8ALayer)
	// GenerateTicketHash should implement the hashing algorithm
	// used for auth ticket hashes in Packet8ALayer
	GenerateTicketHash(ticket string) uint32
	// GenerateLuauResponse should implement the Luau calback based
	// hashing algorithm used in Packet8ALayer
	GenerateLuauResponse(ticket string) uint32
	// OSPlatform should return a string recognized by Roblox
	// that names the Roblox client platform (Win32, Windows_Universal, Android, etc.)
	OSPlatform() string
	// UserAgent should return a user agent string to be used in
	// HTTP requests
	UserAgent() string
	// VersionID should return the version id to be passed to
	// ID_PROTOCOL_SYNC serialization as well as used to generate the
	// ID_SUBMIT_TICKET encryption key
	VersionID() [5]int32
}
type securitySettings struct {
	RakPassword   []byte
	GoldenHash    uint32
	SecurityKey   string
	DataModelHash string
	osPlatform    string
	userAgent     string
}
type windows10SecuritySettings struct {
	securitySettings
}

// Win10Settings returns a SecurityHandler that imitates
// a Win10Universal client (Windows Store version)
// You can optionally pass a custom UserAgent as an argument
func Win10Settings(args ...string) SecurityHandler {
	settings := &windows10SecuritySettings{}
	if len(args) != 0 {
		settings.userAgent = args[0]
	} else {
		settings.userAgent = "Roblox/WinINet"
	}
	settings.osPlatform = "Windows_Universal"

	return settings
}
func (settings *windows10SecuritySettings) GenerateIDResponse(challenge uint32) uint32 {
	return 0x39CB866E - challenge
}

func (settings *windows10SecuritySettings) GenerateLuauResponse(ticket string) uint32 {
	simpleHash := xxHash32.Checksum([]byte(ticket), 1)
	shuffledHash := settings.GenerateTicketHash(ticket)

	return uint32(ResolveLuaChallenge(int32(shuffledHash), int32(simpleHash)))
}
func (settings *windows10SecuritySettings) GenerateTicketHash(ticket string) uint32 {
	initHash := xxHash32.Checksum([]byte(ticket), 1)
	result := -(bits.RotateLeft32(0x11429402-bits.RotateLeft32(bits.RotateLeft32((0x557BB5D7-bits.RotateLeft32(0x557BB5D7*(bits.RotateLeft32(initHash+0x557BB5D7, -7)-0x443921D5), -0xD))^0x557BB5D7, -17)-0x664B2854, 0x17), -0x1D) ^ 0x443921D5)
	return result
}
func (settings *windows10SecuritySettings) PatchTicketPacket(packet *Packet8ALayer) {
	packet.SecurityKey = "2e427f51c4dab762fe9e3471c6cfa1650841723b!d1caf34d619f3bc73e0841daeb75c925\x1E"
	packet.GoldenHash = 0xC001CAFE
	packet.DataModelHash = "ios,ios"
	packet.Platform = settings.osPlatform
	packet.TicketHash = settings.GenerateTicketHash(packet.ClientTicket)
	packet.LuauResponse = settings.GenerateLuauResponse(packet.ClientTicket)
}
func (settings *windows10SecuritySettings) VersionID() [5]int32 {
	return [5]int32{
		-0x14C19E2F,
		+0x75089FCA,
		-0x36BEC40D,
		+0x2501A2F1,
		+0x7EEA9B30,
	}
}

func (settings *securitySettings) OSPlatform() string {
	return settings.osPlatform
}
func (settings *securitySettings) UserAgent() string {
	return settings.userAgent
}

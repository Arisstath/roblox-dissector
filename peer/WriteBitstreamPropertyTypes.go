package peer

import (
	"errors"
	"fmt"
	"math"
	"net"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/robloxapi/rbxfile"
)

func (b *extendedWriter) writeUDim(val rbxfile.ValueUDim) error {
	err := b.writeFloat32BE(val.Scale)
	if err != nil {
		return err
	}
	return b.writeUint32BE(uint32(val.Offset))
}
func (b *extendedWriter) writeUDim2(val rbxfile.ValueUDim2) error {
	err := b.writeUDim(val.X)
	if err != nil {
		return err
	}
	return b.writeUDim(val.Y)
}

func (b *extendedWriter) writeRay(val rbxfile.ValueRay) error {
	err := b.writeVector3Simple(val.Origin)
	if err != nil {
		return err
	}
	return b.writeVector3Simple(val.Direction)
}

func (b *extendedWriter) writeRegion3(val datamodel.ValueRegion3) error {
	err := b.writeVector3Simple(val.Start)
	if err != nil {
		return err
	}
	return b.writeVector3Simple(val.End)
}

func (b *extendedWriter) writeRegion3int16(val datamodel.ValueRegion3int16) error {
	err := b.writeVector3int16(val.Start)
	if err != nil {
		return err
	}
	return b.writeVector3int16(val.End)
}

func (b *extendedWriter) writeColor3(val rbxfile.ValueColor3) error {
	err := b.writeFloat32BE(val.R)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.G)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.B)
}
func (b *extendedWriter) writeColor3uint8(val rbxfile.ValueColor3uint8) error {
	err := b.WriteByte(val.R)
	if err != nil {
		return err
	}
	err = b.WriteByte(val.G)
	if err != nil {
		return err
	}
	return b.WriteByte(val.B)
}
func (b *extendedWriter) writeVector2(val rbxfile.ValueVector2) error {
	err := b.writeFloat32BE(val.X)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Y)
}
func (b *extendedWriter) writeVector3Simple(val rbxfile.ValueVector3) error {
	err := b.writeFloat32BE(val.X)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Y)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Z)
}
func (b *extendedWriter) writeVector3(val rbxfile.ValueVector3) error {
	/*if math.Mod(float64(val.X), 0.5) != 0 ||
		math.Mod(float64(val.Y), 0.1) != 0 ||
		math.Mod(float64(val.Z), 0.5) != 0 ||
		val.X > 511.5 ||
		val.X < -511.5 ||
		val.Y > 204.7 ||
		val.Y < 0 ||
		val.Z > 511.5 ||
		val.Z < -511.5 {
		err = b.writeBoolByte(false)
		if err != nil {
			return err
		}
		err = b.writeVector3Simple(val)
		return err
	}
	err = b.writeBoolByte(true)
	if err != nil {
		return err
	}
	xScaled := uint16(val.X * 2)
	yScaled := uint16(val.Y * 10)
	zScaled := uint16(val.Z * 2)
	xScaled = (xScaled >> 8 & 7) | ((xScaled & 0xFF) << 3)
	yScaled = (yScaled >> 8 & 7) | ((yScaled & 0xFF) << 3)
	zScaled = (zScaled >> 8 & 7) | ((zScaled & 0xFF) << 3)
	err = b.bits(11, uint64(xScaled))
	if err != nil {
		return err
	}
	err = b.bits(11, uint64(yScaled))
	if err != nil {
		return err
	}
	err = b.bits(11, uint64(zScaled))
	return err*/
	return errors.New("v3comp not implemented")
}

func (b *extendedWriter) writeVector2int16(val rbxfile.ValueVector2int16) error {
	err := b.writeUint16BE(uint16(val.X))
	if err != nil {
		return err
	}
	return b.writeUint16BE(uint16(val.Y))
}
func (b *extendedWriter) writeVector3int16(val rbxfile.ValueVector3int16) error {
	err := b.writeUint16BE(uint16(val.X))
	if err != nil {
		return err
	}
	err = b.writeUint16BE(uint16(val.Y))
	if err != nil {
		return err
	}
	return b.writeUint16BE(uint16(val.Z))
}

func (b *extendedWriter) writePBool(val rbxfile.ValueBool) error {
	return b.writeBoolByte(bool(val))
}
func (b *extendedWriter) writePSint(val rbxfile.ValueInt) error {
	return b.writeUint32BE(uint32(val))
}
func (b *extendedWriter) writePFloat(val rbxfile.ValueFloat) error {
	return b.writeFloat32BE(float32(val))
}
func (b *extendedWriter) writePDouble(val rbxfile.ValueDouble) error {
	return b.writeFloat64BE(float64(val))
}

func (b *extendedWriter) writeAxes(val rbxfile.ValueAxes) error {
	write := 0
	if val.X {
		write |= 4
	}
	if val.Y {
		write |= 2
	}
	if val.Z {
		write |= 1
	}
	return b.writeUint32BE(uint32(write))
}
func (b *extendedWriter) writeFaces(val rbxfile.ValueFaces) error {
	write := 0
	if val.Right {
		write |= 32
	}
	if val.Top {
		write |= 16
	}
	if val.Back {
		write |= 8
	}
	if val.Left {
		write |= 4
	}
	if val.Bottom {
		write |= 2
	}
	if val.Front {
		write |= 1
	}
	return b.writeUint32BE(uint32(write))
}

func (b *extendedWriter) writeBrickColor(val rbxfile.ValueBrickColor) error {
	return b.writeUint16BE(uint16(val))
}

func (b *extendedWriter) writeVarLengthString(val string) error {
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	return b.writeASCII(val)
}

func (b *extendedWriter) writeLuauProtectedStringRaw(val rbxfile.ValueProtectedString) error {
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	err = b.allBytes([]byte(val))
	if err != nil {
		return err
	}
	err = b.writeUintUTF8(0x1C)
	if err != nil {
		return err
	}
	// TODO: Writes a zero signature for now
	return b.allBytes(make([]byte, 0x1C))
}

func (b *joinSerializeWriter) writeLuauProtectedString(val rbxfile.ValueProtectedString) error {
	return b.writeLuauProtectedStringRaw(val)
}
func (b *extendedWriter) writeLuauProtectedString(val rbxfile.ValueProtectedString, caches *Caches) error {
	return b.writeLuauCachedProtectedString([]byte(val), caches)
}

func (b *extendedWriter) writeNewPString(val rbxfile.ValueString, caches *Caches) error {
	return b.writeCached(string(val), caches)
}
func (b *extendedWriter) writePStringNoCache(val rbxfile.ValueString) error {
	return b.writeVarLengthString(string(val))
}

func (b *extendedWriter) writeNewProtectedString(val rbxfile.ValueProtectedString, caches *Caches) error {
	return b.writeNewCachedProtectedString([]byte(val), caches)
}
func (b *extendedWriter) writeNewBinaryString(val rbxfile.ValueBinaryString) error {
	return b.writeVarLengthString(string(val))
}
func (b *extendedWriter) writeNewContent(val rbxfile.ValueContent, caches *Caches) error {
	return b.writeCachedContent(string(val), caches)
}

func (b *joinSerializeWriter) writeNewPString(val rbxfile.ValueString) error {
	return b.extendedWriter.writePStringNoCache(val)
}
func (b *joinSerializeWriter) writeNewProtectedString(val rbxfile.ValueProtectedString) error {
	return b.extendedWriter.writePStringNoCache(rbxfile.ValueString(val))
}
func (b *joinSerializeWriter) writeNewContent(val rbxfile.ValueContent) error {
	return b.writeVarLengthString(string(val))
}

func (b *extendedWriter) writeCFrameSimple(val rbxfile.ValueCFrame) error {
	return errors.New("simple CFrame not implemented")
}

func rotMatrixToQuaternion(r [9]float32) [4]float32 {
	q := float32(math.Sqrt(float64(1+r[0*3+0]+r[1*3+1]+r[2*3+2])) / 2)
	return [4]float32{
		(r[2*3+1] - r[1*3+2]) / (4 * q),
		(r[0*3+2] - r[2*3+0]) / (4 * q),
		(r[1*3+0] - r[0*3+1]) / (4 * q),
		q,
	}
} // So nice to not have to worry about normalization on this side!
func (b *extendedWriter) writeCFrame(val rbxfile.ValueCFrame) error {
	err := b.writeVector3Simple(val.Position)
	if err != nil {
		return err
	}
	err = b.writeBoolByte(false) // Not going to bother with lookup stuff
	if err != nil {
		return err
	}

	return b.writePhysicsMatrix(val.Rotation)
}

func (b *extendedWriter) writeSintUTF8(val int32) error {
	return b.writeUintUTF8(uint32(val)<<1 ^ -(uint32(val) >> 31))
}
func (b *extendedWriter) writeNewPSint(val rbxfile.ValueInt) error {
	return b.writeSintUTF8(int32(val))
}
func (b *extendedWriter) writeVarint64(value uint64) error {
	if value == 0 {
		return b.WriteByte(0)
	}
	for value != 0 {
		nextValue := value >> 7
		if nextValue != 0 {
			err := b.WriteByte(byte(value&0x7F | 0x80))
			if err != nil {
				return err
			}
		} else {
			err := b.WriteByte(byte(value & 0x7F))
			if err != nil {
				return err
			}
		}
		value = nextValue
	}
	return nil
}
func (b *extendedWriter) writeVarsint64(val int64) error {
	return b.writeVarint64(uint64(val)<<1 ^ -(uint64(val) >> 63))
}
func (b *extendedWriter) writeVarsint32(val int32) error {
	return b.writeVarint64(uint64(uint32(val)<<1^-(uint32(val)>>31)) & 0xFFFFFFFF)
}

var typeToNetworkConvTable = map[rbxfile.Type]uint8{
	rbxfile.TypeString:                   PropertyTypeString,
	rbxfile.TypeBinaryString:             PropertyTypeBinaryString,
	rbxfile.TypeProtectedString:          PropertyTypeProtectedString0,
	rbxfile.TypeContent:                  PropertyTypeContent,
	rbxfile.TypeBool:                     PropertyTypeBool,
	rbxfile.TypeInt:                      PropertyTypeInt,
	rbxfile.TypeFloat:                    PropertyTypeFloat,
	rbxfile.TypeDouble:                   PropertyTypeDouble,
	rbxfile.TypeUDim:                     PropertyTypeUDim,
	rbxfile.TypeUDim2:                    PropertyTypeUDim2,
	rbxfile.TypeRay:                      PropertyTypeRay,
	rbxfile.TypeFaces:                    PropertyTypeFaces,
	rbxfile.TypeAxes:                     PropertyTypeAxes,
	rbxfile.TypeBrickColor:               PropertyTypeBrickColor,
	rbxfile.TypeColor3:                   PropertyTypeColor3,
	rbxfile.TypeVector2:                  PropertyTypeVector2,
	rbxfile.TypeVector3:                  PropertyTypeComplicatedVector3,
	rbxfile.TypeCFrame:                   PropertyTypeComplicatedCFrame,
	datamodel.TypeToken:                  PropertyTypeEnum,
	datamodel.TypeReference:              PropertyTypeInstance,
	rbxfile.TypeVector3int16:             PropertyTypeVector3int16,
	rbxfile.TypeVector2int16:             PropertyTypeVector2int16,
	datamodel.TypeNumberSequence:         PropertyTypeNumberSequence,
	datamodel.TypeColorSequence:          PropertyTypeColorSequence,
	rbxfile.TypeNumberRange:              PropertyTypeNumberRange,
	rbxfile.TypeRect2D:                   PropertyTypeRect2D,
	rbxfile.TypePhysicalProperties:       PropertyTypePhysicalProperties,
	rbxfile.TypeColor3uint8:              PropertyTypeColor3uint8,
	datamodel.TypeNumberSequenceKeypoint: PropertyTypeNumberSequenceKeypoint,
	datamodel.TypeColorSequenceKeypoint:  PropertyTypeColorSequenceKeypoint,
	datamodel.TypeSystemAddress:          PropertyTypeSystemAddress,
	datamodel.TypeMap:                    PropertyTypeMap,
	datamodel.TypeDictionary:             PropertyTypeDictionary,
	datamodel.TypeArray:                  PropertyTypeArray,
	datamodel.TypeTuple:                  PropertyTypeTuple,
	rbxfile.TypeInt64:                    PropertyTypeInt64,
	datamodel.TypePathWaypoint:           PropertyTypePathWaypoint,
	datamodel.TypeDeferredString:         PropertyTypeSharedString,
}

func typeToNetwork(val rbxfile.Value) (uint8, bool) {
	if val == nil {
		return 0, true
	}
	typ, ok := typeToNetworkConvTable[val.Type()]
	return typ, ok
}
func isCorrectType(val rbxfile.Value, expected uint8) bool {
	typ, ok := typeToNetworkConvTable[val.Type()]
	if !ok {
		return false
	}
	switch expected {
	case PropertyTypeString, PropertyTypeStringNoCache:
		return typ == PropertyTypeString
	case PropertyTypeProtectedString0, PropertyTypeProtectedString1, PropertyTypeProtectedString2, PropertyTypeProtectedString3:
		return typ == PropertyTypeProtectedString0
	case PropertyTypeSimpleVector3, PropertyTypeComplicatedVector3:
		return typ == PropertyTypeComplicatedVector3
	case PropertyTypeSimpleCFrame, PropertyTypeComplicatedCFrame:
		return typ == PropertyTypeComplicatedCFrame
	default:
		return typ == expected
	}
}

func (b *extendedWriter) writeSerializedValueGeneric(val rbxfile.Value, valueType uint8, deferred writeDeferredStrings) error {
	if val == nil {
		return errors.New("can't write nil value")
	}

	if !isCorrectType(val, valueType) {
		return fmt.Errorf("bad value type: expected %d, got %T", valueType, val)
	}
	var err error
	switch valueType {
	case PropertyTypeEnum:
		err = b.writeNewEnumValue(val.(datamodel.ValueToken))
	case PropertyTypeBinaryString:
		err = b.writeNewBinaryString(val.(rbxfile.ValueBinaryString))
	case PropertyTypeBool:
		err = b.writePBool(val.(rbxfile.ValueBool))
	case PropertyTypeInt:
		err = b.writeNewPSint(val.(rbxfile.ValueInt))
	case PropertyTypeFloat:
		err = b.writePFloat(val.(rbxfile.ValueFloat))
	case PropertyTypeDouble:
		err = b.writePDouble(val.(rbxfile.ValueDouble))
	case PropertyTypeUDim:
		err = b.writeUDim(val.(rbxfile.ValueUDim))
	case PropertyTypeUDim2:
		err = b.writeUDim2(val.(rbxfile.ValueUDim2))
	case PropertyTypeRay:
		err = b.writeRay(val.(rbxfile.ValueRay))
	case PropertyTypeFaces:
		err = b.writeFaces(val.(rbxfile.ValueFaces))
	case PropertyTypeAxes:
		err = b.writeAxes(val.(rbxfile.ValueAxes))
	case PropertyTypeBrickColor:
		err = b.writeBrickColor(val.(rbxfile.ValueBrickColor))
	case PropertyTypeColor3:
		err = b.writeColor3(val.(rbxfile.ValueColor3))
	case PropertyTypeColor3uint8:
		err = b.writeColor3uint8(val.(rbxfile.ValueColor3uint8))
	case PropertyTypeVector2:
		err = b.writeVector2(val.(rbxfile.ValueVector2))
	case PropertyTypeSimpleVector3:
		err = b.writeVector3Simple(val.(rbxfile.ValueVector3))
	case PropertyTypeComplicatedVector3:
		err = b.writeVector3(val.(rbxfile.ValueVector3))
	case PropertyTypeVector2int16:
		err = b.writeVector2int16(val.(rbxfile.ValueVector2int16))
	case PropertyTypeVector3int16:
		err = b.writeVector3int16(val.(rbxfile.ValueVector3int16))
	case PropertyTypeSimpleCFrame:
		err = b.writeCFrameSimple(val.(rbxfile.ValueCFrame))
	case PropertyTypeComplicatedCFrame:
		err = b.writeCFrame(val.(rbxfile.ValueCFrame))
	case PropertyTypeNumberSequence:
		err = b.writeNumberSequence(val.(datamodel.ValueNumberSequence))
	case PropertyTypeNumberSequenceKeypoint:
		err = b.writeNumberSequenceKeypoint(val.(datamodel.ValueNumberSequenceKeypoint))
	case PropertyTypeNumberRange:
		err = b.writeNumberRange(val.(rbxfile.ValueNumberRange))
	case PropertyTypeColorSequence:
		err = b.writeColorSequence(val.(datamodel.ValueColorSequence))
	case PropertyTypeColorSequenceKeypoint:
		err = b.writeColorSequenceKeypoint(val.(datamodel.ValueColorSequenceKeypoint))
	case PropertyTypeRect2D:
		err = b.writeRect2D(val.(rbxfile.ValueRect2D))
	case PropertyTypePhysicalProperties:
		err = b.writePhysicalProperties(val.(rbxfile.ValuePhysicalProperties))
	case PropertyTypeInt64:
		err = b.writeVarsint64(int64(val.(rbxfile.ValueInt64)))
	case PropertyTypeStringNoCache:
		err = b.writePStringNoCache(val.(rbxfile.ValueString))
	case PropertyTypePathWaypoint:
		err = b.writePathWaypoint(val.(datamodel.ValuePathWaypoint))
	case PropertyTypeSharedString:
		err = b.writeSharedString(val.(*datamodel.ValueDeferredString), deferred)
	default:
		return errors.New("Unsupported property type: " + strconv.Itoa(int(valueType)))
	}
	return err
}

func (b *extendedWriter) writeNewTypeAndValue(val rbxfile.Value, writer PacketWriter, deferred writeDeferredStrings) error {
	var err error
	valueType, ok := typeToNetwork(val)
	if !ok {
		fmt.Printf("Invalid network type: %T\n", val)
		return errors.New("invalid network type")
	}
	err = b.WriteByte(uint8(valueType))
	// if it's nil:
	if valueType == 0 {
		return nil
	}
	if valueType == 7 {
		err = b.writeUint16BE(val.(datamodel.ValueToken).ID)
		if err != nil {
			return err
		}
	}
	return b.WriteSerializedValue(val, writer, valueType, deferred)
}

func (b *extendedWriter) writeNewTuple(val datamodel.ValueTuple, writer PacketWriter, deferred writeDeferredStrings) error {
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeNewTypeAndValue(val[i], writer, deferred)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeNewArray(val datamodel.ValueArray, writer PacketWriter, deferred writeDeferredStrings) error {
	return b.writeNewTuple(datamodel.ValueTuple(val), writer, deferred)
}

func (b *extendedWriter) writeNewDictionary(val datamodel.ValueDictionary, writer PacketWriter, deferred writeDeferredStrings) error {
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for key, value := range val {
		err = b.writeUintUTF8(uint32(len(key)))
		if err != nil {
			return err
		}
		err = b.writeASCII(key)
		if err != nil {
			return err
		}
		err = b.writeNewTypeAndValue(value, writer, deferred)
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeNewMap(val datamodel.ValueMap, writer PacketWriter, deferred writeDeferredStrings) error {
	return b.writeNewDictionary(datamodel.ValueDictionary(val), writer, deferred)
}

func (b *extendedWriter) writeNumberSequenceKeypoint(val datamodel.ValueNumberSequenceKeypoint) error {
	err := b.writeFloat32BE(val.Time)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Value)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Envelope)
	return err
}
func (b *extendedWriter) writeNumberSequence(val datamodel.ValueNumberSequence) error {
	err := b.writeUint32BE(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeNumberSequenceKeypoint(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeNumberRange(val rbxfile.ValueNumberRange) error {
	err := b.writeFloat32BE(val.Min)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Max)
}

func (b *extendedWriter) writeColorSequenceKeypoint(val datamodel.ValueColorSequenceKeypoint) error {
	err := b.writeFloat32BE(val.Time)
	if err != nil {
		return err
	}
	err = b.writeColor3(val.Value)
	if err != nil {
		return err
	}
	return b.writeFloat32BE(val.Envelope)
}
func (b *extendedWriter) writeColorSequence(val datamodel.ValueColorSequence) error {
	err := b.writeUint32BE(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeColorSequenceKeypoint(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *extendedWriter) writeNewEnumValue(val datamodel.ValueToken) error {
	return b.writeUintUTF8(val.Value)
}

func (b *extendedWriter) writeSystemAddressRaw(val datamodel.ValueSystemAddress) error {
	var err error
	addr := net.UDPAddr(val)
	if err != nil {
		return err
	}

	tmpIPAddr := [...]byte{addr.IP[3], addr.IP[2], addr.IP[1], addr.IP[0]}

	err = b.bytes(4, tmpIPAddr[:])
	if err != nil {
		return err
	}

	return b.writeUint16BE(uint16(addr.Port))
}

func (b *extendedWriter) writeSystemAddress(val datamodel.ValueSystemAddress, caches *Caches) error {
	return b.writeCachedSystemAddress(val, caches)
}
func (b *joinSerializeWriter) writeSystemAddress(val datamodel.ValueSystemAddress) error {
	return b.writeSystemAddressRaw(val)
}

func (b *extendedWriter) writeRect2D(val rbxfile.ValueRect2D) error {
	err := b.writeFloat32BE(val.Min.X)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Min.Y)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Max.X)
	if err != nil {
		return err
	}
	err = b.writeFloat32BE(val.Max.Y)
	return err
}

func (b *extendedWriter) writePhysicalProperties(val rbxfile.ValuePhysicalProperties) error {
	err := b.writeBoolByte(val.CustomPhysics)
	if err != nil {
		return err
	}
	if val.CustomPhysics {
		err := b.writeFloat32BE(val.Density)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.Friction)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.Elasticity)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.FrictionWeight)
		if err != nil {
			return err
		}
		err = b.writeFloat32BE(val.ElasticityWeight)
	}
	return err
}

func (b *extendedWriter) writePathWaypoint(val datamodel.ValuePathWaypoint) error {
	err := b.writeVector3Simple(val.Position)
	if err != nil {
		return err
	}
	return b.writeUint32BE(val.Action)
}

func (b *extendedWriter) writeSharedString(val *datamodel.ValueDeferredString, deferred writeDeferredStrings) error {
	if len(val.Hash) != 0x10 {
		return errors.New("invalid deferred hash")
	}
	err := b.writeASCII(val.Hash)
	if err != nil {
		return err
	}

	deferred.Defer(val)

	return nil
}

func (b *extendedWriter) writeCoordsMode0(val rbxfile.ValueVector3) error {
	return b.writeVector3Simple(val)
}
func (b *extendedWriter) writeCoordsMode1(val rbxfile.ValueVector3) error {
	valRange := float32(math.Max(math.Max(math.Abs(float64(val.X)), math.Abs(float64(val.Y))), math.Abs(float64(val.Z))))
	err := b.writeFloat32BE(valRange)
	if err != nil {
		return err
	}
	if valRange <= 0.0000099999997 {
		return nil
	}
	err = b.writeUint16BE(uint16(val.X/valRange*32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.writeUint16BE(uint16(val.Y/valRange*32767.0 + 32767.0))
	if err != nil {
		return err
	}
	err = b.writeUint16BE(uint16(val.Z/valRange*32767.0 + 32767.0))
	return err
}
func (b *extendedWriter) writeCoordsMode2(val rbxfile.ValueVector3) error {
	xShort := uint16((val.X + 1024.0) * 16.0)
	yShort := uint16((val.Y + 1024.0) * 16.0)
	zShort := uint16((val.Z + 1024.0) * 16.0)

	err := b.writeUint16BE((xShort&0x7F)<<7 | (xShort >> 8))
	if err != nil {
		return err
	}
	err = b.writeUint16BE((yShort&0x7F)<<7 | (yShort >> 8))
	if err != nil {
		return err
	}
	err = b.writeUint16BE((zShort&0x7F)<<7 | (zShort >> 8))
	return err
}

func (b *extendedWriter) writePhysicsCoords(val rbxfile.ValueVector3) error {
	var xModifier, yModifier, zModifier float32
	var err error
	xAbs := math.Abs(float64(val.X))
	yAbs := math.Abs(float64(val.Y))
	zAbs := math.Abs(float64(val.Z))
	largest := xAbs
	if yAbs > xAbs {
		largest = yAbs
	}
	if zAbs > largest {
		largest = zAbs
	}

	_, exp := math.Frexp(largest)
	if exp < 0 {
		exp = 0
	} else if exp > 31 {
		exp = 31
	}

	scale := float32(math.Exp2(float64(exp)))

	xScale := float32(-1.0)
	yScale := float32(-1.0)
	zScale := float32(-1.0)

	if val.X/scale > -1.0 {
		xScale = val.X / scale
	}
	if val.Y/scale > -1.0 {
		yScale = val.Y / scale
	}
	if val.Z/scale > -1.0 {
		zScale = val.Z / scale
	}

	if xScale > 1.0 {
		xScale = 1.0
	}
	if yScale > 1.0 {
		yScale = 1.0
	}
	if zScale > 1.0 {
		zScale = 1.0
	}

	xModifier = -0.5
	yModifier = -0.5
	zModifier = -0.5
	if val.X > 0.0 {
		xModifier = 0.5
	}
	if val.Y > 0.0 {
		yModifier = 0.5
	}
	if val.Z > 0.0 {
		zModifier = 0.5
	}

	if exp <= 4.0 {
		xScale *= 1023.0
		yScale *= 1023.0
		zScale *= 1023.0

		xScale += xModifier
		yScale += yModifier
		zScale += zModifier

		xScaleInt := int32(xScale)
		yScaleInt := int32(yScale)
		zScaleInt := int32(zScale)

		xSign := xScaleInt >> 10 & 1
		ySign := yScaleInt >> 10 & 1
		zSign := zScaleInt >> 10 & 1

		var header = uint8(exp << 3)
		header |= uint8(xSign << 2)
		header |= uint8(ySign << 1)
		header |= uint8(zSign << 0)
		err = b.WriteByte(header)
		if err != nil {
			return err
		}

		var val1 uint32
		val1 |= uint32(xScaleInt<<20) & 0x3FF00000
		val1 |= uint32(yScaleInt&0x3FF) << 10
		val1 |= uint32(zScaleInt) & 0x3FF

		err = b.writeUint32BE(val1)
		if err != nil {
			return err
		}
	} else if exp > 10.0 {
		xScale *= 2097200.0
		yScale *= 2097200.0
		zScale *= 2097200.0

		xScale += xModifier
		yScale += yModifier
		zScale += zModifier

		xScaleInt := int32(xScale)
		yScaleInt := int32(yScale)
		zScaleInt := int32(zScale)

		xSign := xScaleInt >> 21 & 1
		ySign := yScaleInt >> 21 & 1
		zSign := zScaleInt >> 21 & 1

		var header = uint8(exp << 3)
		header |= uint8(xSign << 2)
		header |= uint8(ySign << 1)
		header |= uint8(zSign << 0)
		err = b.WriteByte(header)
		if err != nil {
			return err
		}

		var val1, val2 uint32
		val1 |= uint32(xScaleInt << 11)
		val1 |= uint32((yScaleInt >> 10) & 0x7FF)

		val2 |= uint32((yScaleInt << 21) & 0x7FE00000)
		val2 |= uint32(zScaleInt & 0x1FFFFF)

		err = b.writeUint32BE(val1)
		if err != nil {
			return err
		}
		err = b.writeUint32BE(val2)
		if err != nil {
			return err
		}
	} else {
		xScale *= 65535.0
		yScale *= 65535.0
		zScale *= 65535.0

		xScale += xModifier
		yScale += yModifier
		zScale += zModifier

		xScaleInt := int32(xScale)
		yScaleInt := int32(yScale)
		zScaleInt := int32(zScale)

		xSign := xScaleInt >> 16 & 1
		ySign := yScaleInt >> 16 & 1
		zSign := zScaleInt >> 16 & 1

		var header = uint8(exp << 3)
		header |= uint8(xSign << 2)
		header |= uint8(ySign << 1)
		header |= uint8(zSign << 0)
		err = b.WriteByte(header)
		if err != nil {
			return err
		}

		err = b.writeUint16BE(uint16(xScaleInt))
		if err != nil {
			return err
		}
		err = b.writeUint16BE(uint16(yScaleInt))
		if err != nil {
			return err
		}
		err = b.writeUint16BE(uint16(zScaleInt))
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *extendedWriter) writeMatrixMode0(val [9]float32) error {
	var err error
	q := rotMatrixToQuaternion(val)
	b.writeFloat32BE(q[3])
	for i := 0; i < 3; i++ {
		err = b.writeFloat32BE(q[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeMatrixMode1(val [9]float32) error {
	q := rotMatrixToQuaternion(val)
	err := b.writeBoolByte(q[3] < 0) // sqrt doesn't return negative numbers
	if err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		err = b.writeBoolByte(q[i] < 0)
		if err != nil {
			return err
		}
	}
	for i := 0; i < 3; i++ {
		err = b.writeUint16LE(uint16(math.Abs(float64(q[i]))))
		if err != nil {
			return err
		}
	}
	return nil
}
func (b *extendedWriter) writeMatrixMode2(val [9]float32) error {
	return b.writeMatrixMode1(val)
}
func (b *extendedWriter) writePhysicsMatrix(val [9]float32) error {
	var err error
	quat := rotMatrixToQuaternion(val)
	largestIndex := 0
	largest := math.Abs(float64(quat[0]))
	for i := 1; i < 4; i++ {
		if math.Abs(float64(quat[i])) > largest {
			largest = math.Abs(float64(quat[i]))
			largestIndex = i
		}
	}
	indexSet := quaternionIndices[largestIndex]
	norm := float32(math.Sqrt(float64(quat[0]*quat[0] + quat[1]*quat[1] + quat[2]*quat[2] + quat[3]*quat[3])))
	for i := 0; i < 4; i++ {
		quat[i] /= norm
	}
	if quat[largestIndex] < 0.0 {
		for i := 0; i < 4; i++ {
			quat[i] = -quat[i]
		}
	}

	val1 := quat[indexSet[0]] * math.Sqrt2 * 16383.0
	val2 := quat[indexSet[1]] * math.Sqrt2 * 16383.0
	val3 := quat[indexSet[2]] * math.Sqrt2 * 16383.0

	if quat[indexSet[0]] < 0.0 {
		val1 -= 0.5
	} else {
		val1 += 0.5
	}
	if quat[indexSet[1]] < 0.0 {
		val2 -= 0.5
	} else {
		val2 += 0.5
	}
	if quat[indexSet[2]] < 0.0 {
		val3 -= 0.5
	} else {
		val3 += 0.5
	}

	val1Int := int32(val1) & 0x7FFF
	val2Int := int32(val2) & 0x7FFF
	val3Int := int32(val3) & 0x7FFF

	err = b.writeUint16BE(uint16(val1Int))
	if err != nil {
		return err
	}

	var val2Encoded uint32
	val2Encoded |= uint32(largestIndex << 30)
	val2Encoded |= uint32(val2Int << 15)
	val2Encoded |= uint32(val3Int << 0)

	err = b.writeUint32BE(uint32(val2Encoded))
	return err
}
func (b *extendedWriter) writePhysicsCFrame(val rbxfile.ValueCFrame) error {
	err := b.writePhysicsCoords(val.Position)
	if err != nil {
		return err
	}
	return b.writePhysicsMatrix(val.Rotation)
}

func (b *extendedWriter) writePhysicsVelocity(val rbxfile.ValueVector3) error {
	var err error
	var xModifier, yModifier, zModifier, xScale, yScale, zScale float32
	xAbs := math.Abs(float64(val.X))
	yAbs := math.Abs(float64(val.Y))
	zAbs := math.Abs(float64(val.Z))
	largest := xAbs
	if yAbs > xAbs {
		largest = yAbs
	}
	if zAbs > largest {
		largest = zAbs
	}

	if largest < 0.001 {
		return b.WriteByte(0) // no velocity!
	}

	_, exp := math.Frexp(largest)
	if exp < 0 {
		exp = 0
	} else if exp > 14 {
		exp = 14
	}

	scale := float32(math.Exp2(float64(exp)))

	xScale = -1.0
	yScale = -1.0
	zScale = -1.0

	if val.X/scale > -1.0 {
		xScale = val.X / scale
	}
	if val.Y/scale > -1.0 {
		yScale = val.Y / scale
	}
	if val.Z/scale > -1.0 {
		zScale = val.Z / scale
	}

	if xScale > 1.0 {
		xScale = 1.0
	}
	if yScale > 1.0 {
		yScale = 1.0
	}
	if zScale > 1.0 {
		zScale = 1.0
	}

	xModifier = -0.5
	yModifier = -0.5
	zModifier = -0.5
	if val.X > 0.0 {
		xModifier = 0.5
	}
	if val.Y > 0.0 {
		yModifier = 0.5
	}
	if val.Z > 0.0 {
		zModifier = 0.5
	}

	xScale *= 2047.0
	yScale *= 2047.0
	zScale *= 2047.0

	xScale += xModifier
	yScale += yModifier
	zScale += zModifier

	xScaleInt := int32(xScale)
	yScaleInt := int32(yScale)
	zScaleInt := int32(zScale)

	var header uint8
	header |= uint8((exp + 1) << 4)
	header |= uint8(zScaleInt & 0xF)
	err = b.WriteByte(header)
	if err != nil {
		return err
	}

	var val1 uint32
	val1 |= uint32((zScaleInt >> 4) & 0xFF)
	val1 |= uint32(yScaleInt << 8)
	val1 |= uint32(xScaleInt << 20)

	err = b.writeUint32BE(val1)
	return err
}

func (b *extendedWriter) writeMotor(motor PhysicsMotor) error {
	hasCoords := false
	hasRotation := false
	norm := motor.Position.X*motor.Position.X + motor.Position.Y*motor.Position.Y + motor.Position.Z*motor.Position.Z
	// I don't understand the point of the following code, other than
	// norm != 0.0. Why do we need to check if v4 is less than normAbs?
	if norm != 0.0 {
		normAbs := math.Abs(float64(norm))
		normPlus1 := normAbs + 1.0
		v4 := 1.0 / 100000.0
		if !math.IsInf(normPlus1, 0) {
			v4 = normPlus1 / 100000.0
		}
		if v4 < normAbs {
			hasCoords = true
		}
	}

	motorRot := motor.Rotation
	trace := motorRot[0] + motorRot[4] + motorRot[8]
	traceCos := 0.5 * (trace - 1.0)
	angle := math.Acos(float64(traceCos))
	if angle != 0.0 {
		angleAbs := math.Abs(float64(angle))
		anglePlus1 := angleAbs + 1.0
		v7 := 1.0 / 100000.0
		if !math.IsInf(anglePlus1, 0) {
			v7 = anglePlus1 / 100000.0
		}
		if v7 < angleAbs {
			hasRotation = true
		}
	}

	var header uint8
	if hasCoords {
		header |= 1 << 0
	}
	if hasRotation {
		header |= 1 << 1
	}
	err := b.WriteByte(header)
	if err != nil {
		return err
	}

	if hasCoords {
		err = b.writePhysicsCoords(motor.Position)
		if err != nil {
			return err
		}
	}

	if hasRotation {
		quat := rotMatrixToQuaternion(motor.Rotation)
		largestIndex := 0
		largest := math.Abs(float64(quat[0]))
		for i := 1; i < 4; i++ {
			if math.Abs(float64(quat[i])) > largest {
				largest = math.Abs(float64(quat[i]))
				largestIndex = i
			}
		}
		indexSet := quaternionIndices[largestIndex]
		rotationNorm := float32(math.Sqrt(float64(quat[0]*quat[0] + quat[1]*quat[1] + quat[2]*quat[2] + quat[3]*quat[3])))
		for i := 0; i < 4; i++ {
			quat[i] /= rotationNorm
		}
		if quat[largestIndex] < 0.0 {
			for i := 0; i < 4; i++ {
				quat[i] = -quat[i]
			}
		}

		val1 := quat[indexSet[0]] * math.Sqrt2 * 511.0
		val2 := quat[indexSet[1]] * math.Sqrt2 * 511.0
		val3 := quat[indexSet[2]] * math.Sqrt2 * 511.0

		if quat[indexSet[0]] < 0.0 {
			val1 -= 0.5
		} else {
			val1 += 0.5
		}
		if quat[indexSet[1]] < 0.0 {
			val2 -= 0.5
		} else {
			val2 += 0.5
		}
		if quat[indexSet[2]] < 0.0 {
			val3 -= 0.5
		} else {
			val3 += 0.5
		}

		val1Int := int32(val1) & 0x3FF
		val2Int := int32(val2) & 0x3FF
		val3Int := int32(val3) & 0x3FF

		var val1Encoded uint32
		val1Encoded |= uint32(largestIndex << 30)
		val1Encoded |= uint32(val1Int << 20)
		val1Encoded |= uint32(val2Int << 10)
		val1Encoded |= uint32(val3Int << 0)

		err = b.writeUint32BE(uint32(val1Encoded))
		return err
	}
	return nil
}

func (b *extendedWriter) writeMotors(val []PhysicsMotor) error {
	err := b.writeUintUTF8(uint32(len(val)))
	if err != nil {
		return err
	}
	for i := 0; i < len(val); i++ {
		err = b.writeMotor(val[i])
		if err != nil {
			return err
		}
	}
	return nil
}

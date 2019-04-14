package peer

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/robloxapi/rbxfile"
)

func rotToString(rot [9]float32) string {
	m := rot[:]
	var builder strings.Builder
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&builder, "\n%f %f %f", m[0], m[1], m[2])
		m = m[3:]
	}

	return builder.String()
}

func rotationEqual(ex [9]float32, try [9]float32) bool {
	for i := 0; i < 9; i++ {
		if ex[i]-try[i] > 0.001 || try[i]-ex[i] > 0.001 {
			return false
		}
	}
	return true
}

func expectCFrame(t *testing.T, expect rbxfile.ValueCFrame, try rbxfile.ValueCFrame) {
	if expect.Position != try.Position {
		t.Errorf("Position was incorrect, got %s, expected %s", try.Position, expect.Position)
	}

	if !rotationEqual(expect.Rotation, try.Rotation) {
		t.Errorf("Rotation was incorrect, got %s, expected %s", rotToString(try.Rotation), rotToString(expect.Rotation))
	}
}

func tryOneCFrame(t *testing.T, testValue rbxfile.ValueCFrame) {
	var cframeTestBuffer bytes.Buffer
	writer := &extendedWriter{&cframeTestBuffer}
	err := writer.writeCFrame(testValue)
	if err != nil {
		t.Fatal(err.Error())
	}

	testReader := &extendedReader{&cframeTestBuffer}
	gottenCFrame, err := testReader.readCFrame()
	if err != nil {
		t.Fatal(err.Error())
	}

	expectCFrame(t, testValue, gottenCFrame)
}

func TestCFrame(t *testing.T) {
	tryOneCFrame(t, rbxfile.ValueCFrame{
		Position: rbxfile.ValueVector3{X: 1, Y: 2, Z: 3},
		Rotation: [9]float32{
			1, 0, 0,
			0, 1, 0,
			0, 0, 1,
		},
	})

	tryOneCFrame(t, rbxfile.ValueCFrame{
		Position: rbxfile.ValueVector3{X: 117, Y: 224, Z: 5.3},
		Rotation: [9]float32{
			-0, -1, -0,
			0, 0, -1,
			1, 0, 0,
		},
	})

	tryOneCFrame(t, rbxfile.ValueCFrame{
		Position: rbxfile.ValueVector3{X: 0, Y: 0, Z: 0},
		Rotation: [9]float32{
			0.391828269, 0.114024453, 0.912945271,
			0.599104345, 0.721456409, -0.347238451,
			-0.698243856, 0.68300736, 0.214374468,
		},
	})
}

func TestRotation(t *testing.T) {
	float1 := []byte{0x3f, 0x80, 0, 0}
	var testBuf []byte
	// Vector3
	testBuf = append(testBuf, float1...)
	testBuf = append(testBuf, float1...)
	testBuf = append(testBuf, float1...)

	// Special #0
	testBuf = append(testBuf, 2)

	reader := &extendedReader{bytes.NewReader(testBuf)}
	cf, err := reader.readCFrame()
	if err != nil {
		t.Fatal(err.Error())
	}

	expectCFrame(t, rbxfile.ValueCFrame{
		Position: rbxfile.ValueVector3{X: 1, Y: 1, Z: 1},
		Rotation: [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1},
	}, cf)
}

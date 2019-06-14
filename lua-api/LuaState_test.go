package api

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/robloxapi/rbxfile"
	"github.com/robloxapi/rbxfile/xml"
	lua "github.com/yuin/gopher-lua"
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

func TestCFrame(t *testing.T) {
	schemaFile, err := os.Open("testdata/schema_studio.txt")
	if err != nil {
		t.Fatal(err.Error())
	}
	schema, err := peer.ParseSchema(schemaFile)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = schemaFile.Close()
	if err != nil {
		t.Error(err.Error())
	}

	dmFile, err := os.Open("testdata/baseplate.rbxlx")
	if err != nil {
		t.Fatal(err.Error())
	}
	dmRoot, err := xml.Deserialize(dmFile, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	instanceDictionary := datamodel.NewInstanceDictionary()
	thisRoot := datamodel.FromRbxfile(instanceDictionary, dmRoot)
	peer.NormalizeDataModel(thisRoot, schema)

	state := NewState(lua.Options{
		IncludeGoStackTrace: true,
	})
	state.RegisterDataModel(thisRoot)
	state.RegisterSchema(schema)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	state.SetContext(ctx)

	err = state.DoString(`
	local workspace = game.Workspace;
	local baseplate = workspace.Baseplate;
	local newPosition = as({X = 5, Y = 117, Z = 5.2}, "Vector3");
	local newRotation = {1,0,0, 0,1,0, 0,0,1};
	baseplate.CFrame = as({Position = newPosition, Rotation = newRotation}, "CFrame");
	`)
	if err != nil {
		t.Error(err.Error())
	}

	expectCFrame(t, rbxfile.ValueCFrame{
		Position: rbxfile.ValueVector3{X: 5, Y: 117, Z: 5.2},
		Rotation: [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1},
	}, thisRoot.FindService("Workspace").FindFirstChild("Baseplate").Get("CFrame").(rbxfile.ValueCFrame))
}

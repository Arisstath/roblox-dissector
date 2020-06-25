package main

import (
	"fmt"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/robloxapi/rbxfile"
	"github.com/therecipe/qt/widgets"
)

var noLocalDefaults = map[string](map[string]rbxfile.Value){
	"StarterGui": map[string]rbxfile.Value{
		"Archivable":            rbxfile.ValueBool(true),
		"Name":                  rbxfile.ValueString("StarterGui"),
		"ResetPlayerGuiOnSpawn": rbxfile.ValueBool(false),
		"RobloxLocked":          rbxfile.ValueBool(false),
		// TODO: Set token ID correctly here_
		"ScreenOrientation":  datamodel.ValueToken{Value: 0},
		"ShowDevelopmentGui": rbxfile.ValueBool(true),
		"Tags":               rbxfile.ValueBinaryString(""),
	},
	"Workspace": map[string]rbxfile.Value{
		"Archivable": rbxfile.ValueBool(true),
		// TODO: Set token ID correctly here_
		"AutoJointsMode":             datamodel.ValueToken{Value: 0},
		"CollisionGroups":            rbxfile.ValueString(""),
		"DataModelPlaceVersion":      rbxfile.ValueInt(0),
		"ExpSolverEnabled_Replicate": rbxfile.ValueBool(true),
		"ExplicitAutoJoints":         rbxfile.ValueBool(true),
		"FallenPartsDestroyHeight":   rbxfile.ValueFloat(-500.0),
		"FilteringEnabled":           rbxfile.ValueBool(true),
		"Gravity":                    rbxfile.ValueFloat(196.2),
		"ModelInPrimary":             rbxfile.ValueCFrame{},
		"Name":                       rbxfile.ValueString("Workspace"),
		"PrimaryPart":                datamodel.ValueReference{Instance: nil, Reference: datamodel.NullReference},
		"RobloxLocked":               rbxfile.ValueBool(false),
		"StreamingEnabled":           rbxfile.ValueBool(false),
		"StreamingMinRadius":         rbxfile.ValueInt(0),
		"StreamingTargetRadius":      rbxfile.ValueInt(0),
		"Tags":                       rbxfile.ValueBinaryString(""),
		"TerrainWeldsFixed":          rbxfile.ValueBool(true),
	},
	"StarterPack": map[string]rbxfile.Value{
		"Archivable":   rbxfile.ValueBool(true),
		"Name":         rbxfile.ValueString("StarterPack"),
		"RobloxLocked": rbxfile.ValueBool(false),
		"Tags":         rbxfile.ValueBinaryString(""),
	},
	"TeleportService": map[string]rbxfile.Value{
		"Archivable":   rbxfile.ValueBool(true),
		"Name":         rbxfile.ValueString("Teleport Service"), // intentional
		"RobloxLocked": rbxfile.ValueBool(false),
		"Tags":         rbxfile.ValueBinaryString(""),
	},
	"LocalizationService": map[string]rbxfile.Value{
		"Archivable":           rbxfile.ValueBool(true),
		"IsTextScraperRunning": rbxfile.ValueBool(false),
		"LocaleManifest":       rbxfile.ValueString("en-us"),
		"Name":                 rbxfile.ValueString("LocalizationService"),
		"RobloxLocked":         rbxfile.ValueBool(false),
		"ShouldUseCloudTable":  rbxfile.ValueBool(false),
		"Tags":                 rbxfile.ValueBinaryString(""),
		"WebTableContents":     rbxfile.ValueString(""),
	},
	"Players": map[string]rbxfile.Value{
		"Archivable":               rbxfile.ValueBool(true),
		"MaxPlayersInternal":       rbxfile.ValueInt(6),
		"Name":                     rbxfile.ValueString("Players"),
		"PreferredPlayersInternal": rbxfile.ValueInt(6),
		"RespawnTime":              rbxfile.ValueFloat(5.0),
		"RobloxLocked":             rbxfile.ValueBool(false),
		"Tags":                     rbxfile.ValueBinaryString(""),
	},
}

// normalizeTypes changes the types of instances from binary format types to network types
func normalizeTypes(children []*datamodel.Instance, schema *peer.NetworkSchema) {
	for _, instance := range children {
		defaultValues, ok := noLocalDefaults[instance.ClassName]
		if ok {
			for _, prop := range schema.SchemaForClass(instance.ClassName).Properties {
				if _, ok = instance.Properties[prop.Name]; !ok {
					println("Adding missing default value", instance.ClassName, prop.Name)
					instance.Properties[prop.Name] = defaultValues[prop.Name]
				}
			}
		}
		if val, ok := instance.Properties["AttributesReplicate"]; ok && val == nil {
			println("Adding missing AttributesReplicate")
			instance.Properties["AttributesReplicate"] = rbxfile.ValueString("")
		}

		// hack: color is saved in the wrong format
		if instance.ClassName == "Part" {
			color := instance.Get("Color")
			if color != nil {
				instance.Set("Color3uint8", color)
				delete(instance.Properties, "Color")
			}
		}

		for name, prop := range instance.Properties {
			propSchema := schema.SchemaForClass(instance.ClassName).SchemaForProp(name)
			if propSchema == nil {
				fmt.Printf("Warning: %s.%s doesn't exist in schema! Stripping this property.\n", instance.ClassName, name)
				delete(instance.Properties, name)
				continue
			}
			switch propSchema.Type {
			case peer.PropertyTypeProtectedString0,
				peer.PropertyTypeProtectedString1,
				peer.PropertyTypeProtectedString2,
				peer.PropertyTypeProtectedString3:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueProtectedString(prop.(rbxfile.ValueString))
				}
			case peer.PropertyTypeContent:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueContent(prop.(rbxfile.ValueString))
				}
			case peer.PropertyTypeEnum:
				instance.Properties[name] = datamodel.ValueToken{ID: propSchema.EnumID, Value: prop.(datamodel.ValueToken).Value}
			case peer.PropertyTypeBinaryString:
				// This type may be encoded correctly depending on the format
				if _, ok = prop.(rbxfile.ValueString); ok {
					instance.Properties[name] = rbxfile.ValueBinaryString(prop.(rbxfile.ValueString))
				}
			case peer.PropertyTypeColor3uint8:
				if _, ok = prop.(rbxfile.ValueColor3); ok {
					propc3 := prop.(rbxfile.ValueColor3)
					instance.Properties[name] = rbxfile.ValueColor3uint8{R: uint8(propc3.R * 255), G: uint8(propc3.G * 255), B: uint8(propc3.B * 255)}
				}
			case peer.PropertyTypeBrickColor:
				if _, ok = prop.(rbxfile.ValueInt); ok {
					instance.Properties[name] = rbxfile.ValueBrickColor(prop.(rbxfile.ValueInt))
				}
			}
		}
		normalizeTypes(instance.Children, schema)
	}
}

func normalizeChildren(instances []*datamodel.Instance, schema *peer.NetworkSchema) {
	for _, inst := range instances {
		newChildren := make([]*datamodel.Instance, 0, len(inst.Children))
		for _, child := range inst.Children {
			class := schema.SchemaForClass(child.ClassName)
			if class == nil {
				fmt.Printf("Warning: %s doesn't exist in schema! Stripping this instance.\n", child.ClassName)
				continue
			}

			newChildren = append(newChildren, child)
		}

		inst.Children = newChildren
		normalizeChildren(inst.Children, schema)
	}
}

func normalizeServices(root *datamodel.DataModel, schema *peer.NetworkSchema) {
	newInstances := make([]*datamodel.Instance, 0, len(root.Instances))
	for _, serv := range root.Instances {
		class := schema.SchemaForClass(serv.ClassName)
		if class == nil {
			fmt.Printf("Warning: %s doesn't exist in schema! Stripping this instance.\n", serv.ClassName)
			continue
		}

		newInstances = append(newInstances, serv)
	}

	root.Instances = newInstances
}

func normalizeRoot(root *datamodel.DataModel, schema *peer.NetworkSchema) {
	normalizeServices(root, schema)
	// Clear children of some services if they exist
	players := root.FindService("Players")
	if players != nil {
		players.Children = nil
	}
	joints := root.FindService("JointsService")
	if joints != nil {
		joints.Children = nil
	}
	normalizeServices(root, schema)
	normalizeChildren(root.Instances, schema)
	normalizeTypes(root.Instances, schema)
}

func NewServerStartWidget(parent widgets.QWidget_ITF, settings *ServerSettings, callback func(*ServerSettings)) {
	window := widgets.NewQWidget(parent, 1)
	window.SetWindowTitle("Start server...")
	layout := widgets.NewQFormLayout(window)

	rbxlLayout := NewFileBrowseLayout(window, false, settings.RBXLLocation, "Find place file...", "RBXLX files (*.rbxlx)")
	layout.AddRow4("RBXLX location:", rbxlLayout)

	schemaLayout := NewFileBrowseLayout(window, false, settings.SchemaLocation, "Find schema...", "Text files (*.txt)")
	layout.AddRow4("Schema location:", schemaLayout)

	// HACK: convenience
	if settings.Port == "" {
		settings.Port = "53640"
	}
	port := widgets.NewQLineEdit2(settings.Port, nil)
	layout.AddRow3("Port number:", port)

	startButton := widgets.NewQPushButton2("Start", nil)
	startButton.ConnectReleased(func() {
		window.Close()
		settings.Port = port.Text()
		settings.SchemaLocation = schemaLayout.FileName()
		settings.RBXLLocation = rbxlLayout.FileName()
		callback(settings)
	})
	layout.AddRow5(startButton)

	window.SetLayout(layout)
	window.Show()
}

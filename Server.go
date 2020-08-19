package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Gskartwii/roblox-dissector/datamodel"
	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/gotk3/gotk3/gtk"
	"github.com/olebedev/emitter"
	"github.com/robloxapi/rbxfile"
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
				peer.PropertyTypeProtectedString3,
				peer.PropertyTypeLuauString:
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

func NewServerStartWidget(callback func(string, string, uint16)) error {
	builder, err := gtk.BuilderNewFromFile("res/serverstartwidget.ui")
	if err != nil {
		return err
	}
	rbxlxChooser_, err := builder.GetObject("rbxlxchooser")
	if err != nil {
		return err
	}
	rbxlxChooser, ok := rbxlxChooser_.(*gtk.FileChooserButton)
	if !ok {
		return invalidUi("rbxlxchooser")
	}
	schemaChooser_, err := builder.GetObject("schemachooser")
	if err != nil {
		return err
	}
	schemaChooser, ok := schemaChooser_.(*gtk.FileChooserButton)
	if !ok {
		return invalidUi("schemachooser")
	}
	portEntry_, err := builder.GetObject("portentry")
	if err != nil {
		return err
	}
	portEntry, ok := portEntry_.(*gtk.Entry)
	if !ok {
		return invalidUi("portentry")
	}
	cancelButton_, err := builder.GetObject("cancelbutton")
	if err != nil {
		return err
	}
	cancelButton, ok := cancelButton_.(*gtk.Button)
	if !ok {
		return invalidUi("cancelbutton")
	}
	okButton_, err := builder.GetObject("okbutton")
	if err != nil {
		return err
	}
	okButton, ok := okButton_.(*gtk.Button)
	if !ok {
		return invalidUi("okbutton")
	}

	win_, err := builder.GetObject("serverstartwindow")
	if err != nil {
		return err
	}
	win, ok := win_.(*gtk.Window)
	if !ok {
		return invalidUi("serverstartwindow")
	}

	cancelButton.Connect("clicked", func() {
		win.Destroy()
	})

	okButton.Connect("clicked", func() {
		schemaName := schemaChooser.GetFilename()
		if schemaName == "" {
			ShowError(win, errors.New("schema is missing"), "Please choose a schema")
			return
		}
		rbxlxName := rbxlxChooser.GetFilename()
		if rbxlxName == "" {
			ShowError(win, errors.New("rbxlx is missing"), "Please choose a rbxlx file")
			return
		}
		port, err := portEntry.GetText()
		if err != nil {
			ShowError(win, err, "Failed to get port")
			return
		}
		var portNum int
		if port == "" {
			portNum = 53640
		} else {
			portNum, err = strconv.Atoi(port)
			if err != nil {
				ShowError(win, err, "Failed to get port")
				return
			}
			if uint(portNum) > 0xFFFF {
				ShowError(win, errors.New("port is out of range"), "Failed to get port")
				return
			}
		}
		callback(schemaName, rbxlxName, uint16(portNum))
		win.Destroy()
	})

	win.Show()
	return nil
}

func CaptureFromServer(ctx context.Context, session *CaptureSession, server *peer.CustomServer) {
	server.ClientEmitter.On("client", func(e *emitter.Event) {
		client := e.Args[0].(*peer.ServerClient)
		session.AddConversation(&Conversation{
			Client:       client.Address,
			Server:       client.Server.Address,
			ClientReader: client.DefaultPacketReader,
			ServerReader: client.DefaultPacketWriter,
			Context:      client.Context,
		})
	}, emitter.Void)
}

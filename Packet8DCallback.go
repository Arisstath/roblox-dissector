package main

import (
	"fmt"
	"math"
	"strconv"
	"unsafe"

	"github.com/Gskartwii/roblox-dissector/peer"
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

var TerrainMaterials = [...]string{
	"Air",
}
var TerrainColors = [...][3]float32{
	[3]float32{0, 0, 0},
}

var vertexData = [...]float32{
	-0.5, -0.5, -0.5, 0.0, 0.0, -1.0,
	0.5, -0.5, -0.5, 0.0, 0.0, -1.0,
	0.5, 0.5, -0.5, 0.0, 0.0, -1.0,
	0.5, 0.5, -0.5, 0.0, 0.0, -1.0,
	-0.5, 0.5, -0.5, 0.0, 0.0, -1.0,
	-0.5, -0.5, -0.5, 0.0, 0.0, -1.0,

	-0.5, -0.5, 0.5, 0.0, 0.0, 1.0,
	0.5, -0.5, 0.5, 0.0, 0.0, 1.0,
	0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
	0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
	-0.5, 0.5, 0.5, 0.0, 0.0, 1.0,
	-0.5, -0.5, 0.5, 0.0, 0.0, 1.0,

	-0.5, 0.5, 0.5, -1.0, 0.0, 0.0,
	-0.5, 0.5, -0.5, -1.0, 0.0, 0.0,
	-0.5, -0.5, -0.5, -1.0, 0.0, 0.0,
	-0.5, -0.5, -0.5, -1.0, 0.0, 0.0,
	-0.5, -0.5, 0.5, -1.0, 0.0, 0.0,
	-0.5, 0.5, 0.5, -1.0, 0.0, 0.0,

	0.5, 0.5, 0.5, 1.0, 0.0, 0.0,
	0.5, 0.5, -0.5, 1.0, 0.0, 0.0,
	0.5, -0.5, -0.5, 1.0, 0.0, 0.0,
	0.5, -0.5, -0.5, 1.0, 0.0, 0.0,
	0.5, -0.5, 0.5, 1.0, 0.0, 0.0,
	0.5, 0.5, 0.5, 1.0, 0.0, 0.0,

	-0.5, -0.5, -0.5, 0.0, -1.0, 0.0,
	0.5, -0.5, -0.5, 0.0, -1.0, 0.0,
	0.5, -0.5, 0.5, 0.0, -1.0, 0.0,
	0.5, -0.5, 0.5, 0.0, -1.0, 0.0,
	-0.5, -0.5, 0.5, 0.0, -1.0, 0.0,
	-0.5, -0.5, -0.5, 0.0, -1.0, 0.0,

	-0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
	0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
	0.5, 0.5, 0.5, 0.0, 1.0, 0.0,
	0.5, 0.5, 0.5, 0.0, 1.0, 0.0,
	-0.5, 0.5, 0.5, 0.0, 1.0, 0.0,
	-0.5, 0.5, -0.5, 0.0, 1.0, 0.0,
}

const vertexShaderSrc = `#version 330 core
layout (location = 0) in vec3 posAttr;
layout (location = 1) in vec3 normalAttr;
uniform vec4 colAttr;
out vec4 col;
uniform mat4 projection;
uniform mat4 view;
uniform mat4 model;
out vec3 FragPos;
out vec3 Normal;
void main() {
	col = colAttr;
	gl_Position = projection * view * model * vec4(posAttr, 1.0);
	FragPos = vec3(model * vec4(posAttr, 1.0));
	Normal = mat3(transpose(inverse(model))) * normalAttr;
}
` + "\x00"
const fragmentShaderSrc = `#version 330 core
in vec4 col;
in vec3 FragPos;
in vec3 Normal;
out vec4 fragColor;
void main() {
	vec3 lightColor = vec3(1.0, 1.0, 1.0);
	float ambientStrength = 0.4;
	vec3 ambient = ambientStrength * lightColor;

	vec3 lightPos = vec3(32.0, 32.0, 32.0);
	vec3 norm = normalize(Normal);
	vec3 lightDir = normalize(lightPos - FragPos);
	float diff = max(dot(norm, lightDir), 0.0);
	vec3 diffuse = diff * lightColor;
	vec3 result = (ambient + diffuse) * col.rgb;

	fragColor = vec4(result, 1.0);
}` + "\x00"

type TerrainChunkViewer struct {
	OpenGLWindow

	_       func() `constructor:"init"`
	Program *gui.QOpenGLShaderProgram

	PositionLocation   uint
	NormalLocation     uint
	ColorLocation      uint
	ProjectionLocation uint
	ViewLocation       uint
	ModelLocation      uint
	Frame              float32

	CameraPosition  *gui.QVector3D
	CameraYaw       float32
	CameraPitch     float32
	CameraFront     *gui.QVector3D
	CameraUp        *gui.QVector3D
	transformations []*gui.QMatrix4x4
	colors          []*gui.QColor

	KeysDown          map[core.Qt__Key]struct{}
	MouseDown         bool
	IsFirstMouseDown  bool
	LastMousePosition *core.QPoint
}

func (w *TerrainChunkViewer) event(event *core.QEvent) bool {
	switch event.Type() {
	case core.QEvent__UpdateRequest:
		w.updateKeys()
		w.OpenGLWindow.renderNow()
		return true
	default:
		return w.EventDefault(event)
	}
}

func (w *TerrainChunkViewer) init() {
	w.CameraPosition = gui.NewQVector3D3(32.0, 32.0, 32.0)
	w.CameraUp = gui.NewQVector3D3(0, 1, 0)
	w.CameraYaw = -90
	w.CameraPitch = -45
	w.CameraFront = gui.NewQVector3D3(0, -1, -1).Normalized()
	w.OpenGLWindow.initializeLazy = w.initialize
	w.OpenGLWindow.renderLazy = w.render

	w.KeysDown = make(map[core.Qt__Key]struct{})
	w.ConnectExposeEvent(w.OpenGLWindow.exposeEvent)
	w.ConnectEvent(w.event)
	w.ConnectKeyPressEvent(func(evt *gui.QKeyEvent) {
		if evt.IsAutoRepeat() {
			evt.Ignore()
		}
		w.KeysDown[core.Qt__Key(evt.Key())] = struct{}{}
	})
	w.ConnectKeyReleaseEvent(func(evt *gui.QKeyEvent) {
		if evt.IsAutoRepeat() {
			evt.Ignore()
		}
		delete(w.KeysDown, core.Qt__Key(evt.Key()))
	})

	w.IsFirstMouseDown = true
	w.ConnectMousePressEvent(func(evt *gui.QMouseEvent) {
		w.MouseDown = true
	})

	w.ConnectMouseReleaseEvent(func(evt *gui.QMouseEvent) {
		w.MouseDown = false
		w.IsFirstMouseDown = true
	})
}

func (w *TerrainChunkViewer) IsKeyDown(key core.Qt__Key) bool {
	_, ok := w.KeysDown[key]
	return ok
}

func (w *TerrainChunkViewer) initialize() {
	w.Program = gui.NewQOpenGLShaderProgram(w)
	w.Program.AddShaderFromSourceCode(gui.QOpenGLShader__Vertex, vertexShaderSrc)
	w.Program.AddShaderFromSourceCode(gui.QOpenGLShader__Fragment, fragmentShaderSrc)
	w.Program.Link()

	w.ColorLocation = uint(w.Program.UniformLocation("colAttr"))
	w.ProjectionLocation = uint(w.Program.UniformLocation("projection"))
	w.ViewLocation = uint(w.Program.UniformLocation("view"))
	w.ModelLocation = uint(w.Program.UniformLocation("model"))

	w.GlVertexAttribPointer(0, 3, GLFloat, false, 6*4, unsafe.Pointer(&vertexData[0]))
	w.GlVertexAttribPointer(1, 3, GLFloat, false, 6*4, unsafe.Pointer(&vertexData[3]))

	w.GlEnableVertexAttribArray(0)
	w.GlEnableVertexAttribArray(1)
}

func (w *TerrainChunkViewer) updateKeys() {
	speed := float32(0.25)
	sensitivity := float32(0.15)
	if w.IsKeyDown(core.Qt__Key_W) {
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(speed*w.CameraFront.X(), speed*w.CameraFront.Y(), speed*w.CameraFront.Z()))
	}
	if w.IsKeyDown(core.Qt__Key_S) {
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(-speed*w.CameraFront.X(), -speed*w.CameraFront.Y(), -speed*w.CameraFront.Z()))
	}
	if w.IsKeyDown(core.Qt__Key_A) {
		norm := gui.QVector3D_CrossProduct(w.CameraFront, w.CameraUp).Normalized()
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(-speed*norm.X(), -speed*norm.Y(), -speed*norm.Z()))
	}
	if w.IsKeyDown(core.Qt__Key_D) {
		norm := gui.QVector3D_CrossProduct(w.CameraFront, w.CameraUp).Normalized()
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(speed*norm.X(), speed*norm.Y(), speed*norm.Z()))
	}

	if w.MouseDown {
		prevPosition := w.LastMousePosition
		currPosition := gui.QCursor_Pos()
		w.LastMousePosition = currPosition
		if !w.IsFirstMouseDown {
			mouseDelta := core.NewQPoint2(currPosition.X()-prevPosition.X(), currPosition.Y()-prevPosition.Y())
			w.CameraYaw += sensitivity * float32(mouseDelta.X())
			w.CameraPitch += sensitivity * float32(mouseDelta.Y())
			if w.CameraPitch > 89 {
				w.CameraPitch = 89
			} else if w.CameraPitch < -89 {
				w.CameraPitch = -89
			}

			yawRad := float64(w.CameraYaw) / 180.0 * math.Pi
			pitchRad := float64(w.CameraPitch) / 180.0 * math.Pi
			w.CameraFront.SetX(float32(math.Cos(yawRad) * math.Cos(pitchRad)))
			w.CameraFront.SetY(float32(math.Sin(pitchRad)))
			w.CameraFront.SetZ(float32(math.Sin(yawRad) * math.Cos(pitchRad)))
			w.CameraFront.Normalize()
		}
		w.IsFirstMouseDown = false
	}
}

func (w *TerrainChunkViewer) Clear() {
	w.transformations = make([]*gui.QMatrix4x4, 0)
	w.colors = make([]*gui.QColor, 0)
}

func (w *TerrainChunkViewer) AddChunk(chunk *peer.Chunk) {
	sideLength := chunk.SideLength
	baseX := float32(chunk.ChunkIndex.X) * float32(sideLength)
	baseY := float32(chunk.ChunkIndex.Y) * float32(sideLength)
	baseZ := float32(chunk.ChunkIndex.Z) * float32(sideLength)
	// too lazy to calculate i using x,y,z
	i := 0
	for x := uint32(0); x < sideLength; x++ {
		for y := uint32(0); y < sideLength; y++ {
			for z := uint32(0); z < sideLength; z++ {
				thisCell := chunk.CellCube[x][y][z]
				material := thisCell.Material
				occupancy := thisCell.Occupancy
				// don't draw these cells
				if occupancy == 0 {
					continue
				}

				model := gui.NewQMatrix4x4()
				// position the model at the appropriate coordinates
				model.Translate3(baseX+float32(x), baseY+float32(y), baseZ+float32(z))
				// scale the cube by making its height indicate the occupancy
				model.Scale3(1.0, float32(occupancy)/255.0, 1.0)
				w.transformations = append(w.transformations, model)

				if material > 30 {
					println("warning: material higher than 15:", material)
				}
				color := gui.QColor_FromHsl(int(float32(material)/35.0*360.0), 255, 127, 255)
				w.colors = append(w.colors, color)

				i++
			}
		}
	}
}

func addQVector3(vec1, vec2 *gui.QVector3D) *gui.QVector3D {
	return gui.NewQVector3D3(vec1.X()+vec2.X(), vec1.Y()+vec2.Y(), vec1.Z()+vec2.Z())
}

func (w *TerrainChunkViewer) render(painter *gui.QPainter) {
	retinaScale := int(w.DevicePixelRatio())
	w.GlViewport(0, 0, w.Width()*retinaScale, w.Height()*retinaScale)

	w.GlClear(GLColorBufferBit | GLDepthBufferBit)
	w.GlClearColor(0, 0, 0, 1)

	w.Program.Bind()
	w.GlEnable(GLDepthTest)

	projection := gui.NewQMatrix4x4()
	// TODO: dynamic update
	projection.Perspective(60, float32(w.Width())/float32(w.Height()), 0.1, 100.0)
	view := gui.NewQMatrix4x4()
	view.LookAt(w.CameraPosition, addQVector3(w.CameraPosition, w.CameraFront), w.CameraUp)
	w.Program.SetUniformValue23(int(w.ProjectionLocation), projection)
	w.Program.SetUniformValue23(int(w.ViewLocation), view)

	for i := 0; i < len(w.transformations); i++ {
		matrix := w.transformations[i]
		color := w.colors[i]
		w.Program.SetUniformValue23(int(w.ModelLocation), matrix)
		w.Program.SetUniformValue10(int(w.ColorLocation), color)
		w.GlDrawArrays(GLTriangles, 0, 36)
	}

	w.Program.Release()

	w.Frame++
}

func MaterialToString(material uint8) string {
	if len(TerrainMaterials) > int(material) {
		return TerrainMaterials[material]
	}

	return fmt.Sprintf("0x%02X", material)
}

func ShowPacket8D(layerLayout *widgets.QVBoxLayout, context *peer.CommunicationContext, layers *peer.PacketLayers) {
	MainLayer := layers.Main.(*peer.Packet8DLayer)

	labelForSubpackets := NewQLabelF("Terrain cluster (%d chunks):", len(MainLayer.Chunks))
	layerLayout.AddWidget(labelForSubpackets, 0, 0)

	chunkList := widgets.NewQTreeView(nil)
	standardModel := NewProperSortModel(nil)
	standardModel.SetHorizontalHeaderLabels([]string{"Index", "Location", "Dimensions"})

	rootNode := standardModel.InvisibleRootItem()
	for i, chunk := range MainLayer.Chunks {
		indexItem := NewIntItem(i)
		locationItem := NewStringItem(chunk.ChunkIndex.String())
		countItem := NewQStandardItemF("%d x %d x %d", chunk.SideLength, chunk.SideLength, chunk.SideLength)

		rootNode.AppendRow([]*gui.QStandardItem{indexItem, locationItem, countItem})
	}

	chunkList.SetModel(standardModel)
	chunkList.SetSelectionMode(1)
	chunkList.SetSortingEnabled(true)

	chunkList.ConnectClicked(func(index *core.QModelIndex) {
		thisIndex, _ := strconv.Atoi(standardModel.Item(index.Row(), 0).Data(0).ToString())
		chunk := &MainLayer.Chunks[thisIndex]

		viewer := NewTerrainChunkViewer(nil)
		format := gui.NewQSurfaceFormat()
		format.SetSamples(16)
		format.SetDepthBufferSize(24)
		viewer.SetFormat(format)
		viewer.Resize2(640, 480)
		viewer.AddChunk(chunk)
		viewer.SetTitle(fmt.Sprintf("Sala Terrain Viewer: Chunk at %s", chunk.ChunkIndex))
		viewer.Show()
		viewer.setAnimating(true)
	})
	layerLayout.AddWidget(chunkList, 0, 0)

	viewAll := widgets.NewQPushButton2("View all chunks...", nil)
	viewAll.ConnectReleased(func() {
		viewer := NewTerrainChunkViewer(nil)
		format := gui.NewQSurfaceFormat()
		format.SetSamples(16)
		format.SetDepthBufferSize(24)
		viewer.SetFormat(format)
		viewer.Resize2(640, 480)
		for _, chunk := range MainLayer.Chunks {
			viewer.AddChunk(&chunk)
		}
		viewer.SetTitle("Sala Terrain Viewer: Chunk cluster")
		viewer.Show()
		viewer.setAnimating(true)
	})
	layerLayout.AddWidget(viewAll, 0, 0)
}

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
	"Water",
	"Grass",
	"Slate",
	"Concrete",
	"Brick",
	"Sand",
	"Wood Planks",
	"Rock",
	"Snow",
}
var TerrainColors = [...][3]float32{
	[3]float32{0, 0, 0},
	[3]float32{12, 84, 92},    // water
	[3]float32{106, 127, 63},  // grass
	[3]float32{63, 127, 107},  // slate
	[3]float32{127, 102, 63},  // concrete
	[3]float32{138, 86, 62},   // brick
	[3]float32{143, 126, 95},  // sand
	[3]float32{139, 109, 79},  // wood planks
	[3]float32{102, 108, 111}, // rock
	[3]float32{195, 199, 218}, // snow
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
layout (location = 2) in vec4 colAttr;
layout (location = 3) in mat4 model;
layout (location = 7) in mat3 modelNormal;
out vec4 col;
uniform mat4 projectionView;
out vec3 FragPos;
out vec3 Normal;
void main() {
	col = colAttr;
	gl_Position = projectionView * model * vec4(posAttr, 1.0);
	FragPos = vec3(model * vec4(posAttr, 1.0));
	Normal = modelNormal * normalAttr;
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

	ProjectionViewLocation uint
	Frame                  float32

	CameraPosition *gui.QVector3D
	CameraYaw      float32
	CameraPitch    float32
	CameraFront    *gui.QVector3D
	CameraUp       *gui.QVector3D
	ProjectionView *gui.QMatrix4x4

	updateBuffers bool
	modelBuffer   []float32
	normalBuffer  []float32
	colorBuffer   []float32

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

		w.renderLater()
	})
	w.ConnectKeyReleaseEvent(func(evt *gui.QKeyEvent) {
		if evt.IsAutoRepeat() {
			evt.Ignore()
		}
		delete(w.KeysDown, core.Qt__Key(evt.Key()))

		w.renderLater()
	})

	w.IsFirstMouseDown = true
	w.ConnectMousePressEvent(func(evt *gui.QMouseEvent) {
		w.MouseDown = true

		w.renderLater()
	})

	w.ConnectMouseReleaseEvent(func(evt *gui.QMouseEvent) {
		w.MouseDown = false
		w.IsFirstMouseDown = true

		w.renderLater()
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

	w.GlVertexAttribPointer(0, 3, GLFloat, false, 6*4, unsafe.Pointer(&vertexData[0]))
	w.GlVertexAttribPointer(1, 3, GLFloat, false, 6*4, unsafe.Pointer(&vertexData[3]))

	w.GlEnableVertexAttribArray(0)
	w.GlEnableVertexAttribArray(1)
}

func (w *TerrainChunkViewer) updateKeys() {
	renderLater := false
	speed := float32(0.25)
	sensitivity := float32(0.15)
	if w.IsKeyDown(core.Qt__Key_W) {
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(speed*w.CameraFront.X(), speed*w.CameraFront.Y(), speed*w.CameraFront.Z()))
		renderLater = true
	}
	if w.IsKeyDown(core.Qt__Key_S) {
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(-speed*w.CameraFront.X(), -speed*w.CameraFront.Y(), -speed*w.CameraFront.Z()))
		renderLater = true
	}
	if w.IsKeyDown(core.Qt__Key_A) {
		norm := gui.QVector3D_CrossProduct(w.CameraFront, w.CameraUp).Normalized()
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(-speed*norm.X(), -speed*norm.Y(), -speed*norm.Z()))
		renderLater = true
	}
	if w.IsKeyDown(core.Qt__Key_D) {
		norm := gui.QVector3D_CrossProduct(w.CameraFront, w.CameraUp).Normalized()
		w.CameraPosition = addQVector3(w.CameraPosition, gui.NewQVector3D3(speed*norm.X(), speed*norm.Y(), speed*norm.Z()))
		renderLater = true
	}

	if w.MouseDown {
		renderLater = true
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

	if renderLater {
		w.renderLater()
	}
}

func (w *TerrainChunkViewer) Clear() {
	w.updateBuffers = true
	w.modelBuffer = make([]float32, 0)
	w.normalBuffer = make([]float32, 0)
	w.colorBuffer = make([]float32, 0)

	w.renderLater()
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
				col0 := model.Column(0)
				col1 := model.Column(1)
				col2 := model.Column(2)
				col3 := model.Column(3)
				w.modelBuffer = append(w.modelBuffer,
					col0.X(), col0.Y(), col0.Z(), col0.W(),
					col1.X(), col1.Y(), col1.Z(), col1.W(),
					col2.X(), col2.Y(), col2.Z(), col2.W(),
					col3.X(), col3.Y(), col3.Z(), col3.W())

				inverted := model.Inverted(nil)
				iCol0 := inverted.Column(0)
				iCol1 := inverted.Column(1)
				iCol2 := inverted.Column(2)
				w.normalBuffer = append(w.normalBuffer,
					iCol0.X(), iCol1.X(), iCol2.X(),
					iCol0.Y(), iCol1.Y(), iCol2.Y(),
					iCol0.Z(), iCol1.Z(), iCol2.Z())

				if int(material) < len(TerrainColors) {
					cData := TerrainColors[material]
					w.colorBuffer = append(w.colorBuffer, cData[0]/255.0, cData[1]/255.0, cData[2]/255.0)
				} else {
					if material > 35 {
						println("warning: material higher than 35:", material)
					}
					color := gui.QColor_FromHsl(int(float32(int(material)-len(TerrainColors))/(35.0-float32(len(TerrainColors)))*360.0), 255, 127, 255)
					w.colorBuffer = append(w.colorBuffer, float32(color.RedF()), float32(color.GreenF()), float32(color.BlueF()))
				}
				i++
			}
		}
	}

	w.updateBuffers = true
	w.renderLater()
}

func (w *TerrainChunkViewer) RegisterBuffers() {
	w.updateBuffers = false
	if len(w.colorBuffer) == 0 || len(w.modelBuffer) == 0 {
		w.GlDisableVertexAttribArray(2)
		w.GlDisableVertexAttribArray(3)
		w.GlDisableVertexAttribArray(4)
		w.GlDisableVertexAttribArray(5)
		w.GlDisableVertexAttribArray(6)
		w.GlDisableVertexAttribArray(7)
		w.GlDisableVertexAttribArray(8)
		w.GlDisableVertexAttribArray(9)
		return
	}
	w.GlVertexAttribPointer(2, 3, GLFloat, false, 3*4, unsafe.Pointer(&w.colorBuffer[0]))
	w.GlVertexAttribPointer(3, 4, GLFloat, false, 4*4*4, unsafe.Pointer(&w.modelBuffer[0]))
	w.GlVertexAttribPointer(4, 4, GLFloat, false, 4*4*4, unsafe.Pointer(&w.modelBuffer[4]))
	w.GlVertexAttribPointer(5, 4, GLFloat, false, 4*4*4, unsafe.Pointer(&w.modelBuffer[8]))
	w.GlVertexAttribPointer(6, 4, GLFloat, false, 4*4*4, unsafe.Pointer(&w.modelBuffer[12]))
	w.GlVertexAttribPointer(7, 3, GLFloat, false, 4*3*3, unsafe.Pointer(&w.normalBuffer[0]))
	w.GlVertexAttribPointer(8, 3, GLFloat, false, 4*3*3, unsafe.Pointer(&w.normalBuffer[3]))
	w.GlVertexAttribPointer(9, 3, GLFloat, false, 4*3*3, unsafe.Pointer(&w.normalBuffer[6]))

	w.GlVertexAttribDivisor(2, 1)
	w.GlVertexAttribDivisor(3, 1)
	w.GlVertexAttribDivisor(4, 1)
	w.GlVertexAttribDivisor(5, 1)
	w.GlVertexAttribDivisor(6, 1)
	w.GlVertexAttribDivisor(7, 1)
	w.GlVertexAttribDivisor(8, 1)
	w.GlVertexAttribDivisor(9, 1)

	w.GlEnableVertexAttribArray(2)
	w.GlEnableVertexAttribArray(3)
	w.GlEnableVertexAttribArray(4)
	w.GlEnableVertexAttribArray(5)
	w.GlEnableVertexAttribArray(6)
	w.GlEnableVertexAttribArray(7)
	w.GlEnableVertexAttribArray(8)
	w.GlEnableVertexAttribArray(9)

	println("will draw", len(w.colorBuffer)/3, "cubes instanced")
}

func addQVector3(vec1, vec2 *gui.QVector3D) *gui.QVector3D {
	return gui.NewQVector3D3(vec1.X()+vec2.X(), vec1.Y()+vec2.Y(), vec1.Z()+vec2.Z())
}

func (w *TerrainChunkViewer) render(painter *gui.QPainter) {
	retinaScale := int(w.DevicePixelRatio())
	w.GlViewport(0, 0, w.Width()*retinaScale, w.Height()*retinaScale)

	projectionView := gui.NewQMatrix4x4()
	projectionView.Perspective(60, float32(w.Width())/float32(w.Height()), 0.1, 100.0)
	projectionView.LookAt(w.CameraPosition, addQVector3(w.CameraPosition, w.CameraFront), w.CameraUp)
	w.ProjectionView = projectionView

	if w.updateBuffers {
		w.RegisterBuffers()
	}

	w.Program.Bind()
	w.GlEnable(GLDepthTest)

	w.Program.SetUniformValue23(int(w.ProjectionViewLocation), w.ProjectionView)
	w.GlDrawArraysInstanced(GLTriangles, 0, 36, len(w.colorBuffer)/3)

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
		viewer.Clear()
		viewer.AddChunk(chunk)
		viewer.SetTitle(fmt.Sprintf("Sala Terrain Viewer: Chunk at %s", chunk.ChunkIndex))
		viewer.Show()
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
		viewer.Clear()
		progressDialog := widgets.NewQProgressDialog2("Generating vertex attribute buffers...", "Cancel", 0, len(MainLayer.Chunks), viewAll, 0)
		progressDialog.SetWindowTitle("Preparing terrain viewer")
		progressDialog.SetWindowModality(core.Qt__WindowModal)
		for i, chunk := range MainLayer.Chunks {
			viewer.AddChunk(&chunk)
			progressDialog.SetValue(i)
			if progressDialog.WasCanceled() {
				viewer.DestroyQWindow()
				return
			}
		}
		progressDialog.SetValue(len(MainLayer.Chunks))
		viewer.SetTitle("Sala Terrain Viewer: Chunk cluster")
		viewer.Show()
	})
	layerLayout.AddWidget(viewAll, 0, 0)
}

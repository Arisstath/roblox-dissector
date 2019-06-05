package main

import (
	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
)

const (
	GLColorBufferBit uint = 0x4000
	GLDepthBufferBit uint = 0x100
	GLFloat          uint = 0x1406
	GLTriangles      uint = 4
	GLDepthTest      uint = 0xB71
)

type OpenGLWindow struct {
	gui.QWindow
	gui.QOpenGLFunctions

	_ func() `constructor:"init"`

	Context     *gui.QOpenGLContext
	PaintDevice *gui.QOpenGLPaintDevice

	initializeLazy func()
	renderLazy     func(*gui.QPainter)

	Animating bool
}

func (w *OpenGLWindow) init() {
	w.QOpenGLFunctions = *gui.NewQOpenGLFunctions()
	w.SetSurfaceType(gui.QSurface__OpenGLSurface)
}

func (w *OpenGLWindow) render() {
	if w.PaintDevice == nil {
		w.PaintDevice = gui.NewQOpenGLPaintDevice()
	}

	w.GlClear(GLColorBufferBit | GLDepthBufferBit)

	w.PaintDevice.SetSize(w.Size())

	painter := gui.NewQPainter2(w.PaintDevice)
	w.renderLazy(painter)
	painter.DestroyQPainter()
}

func (w *OpenGLWindow) renderLater() {
	w.RequestUpdate()
}

func (w *OpenGLWindow) event(event *core.QEvent) bool {
	switch event.Type() {
	case core.QEvent__UpdateRequest:
		w.renderNow()
		return true
	default:
		return w.EventDefault(event)
	}
}

func (w *OpenGLWindow) exposeEvent(*gui.QExposeEvent) {
	if w.IsExposed() {
		w.renderNow()
	}
}

func (w *OpenGLWindow) renderNow() {
	if !w.IsExposed() {
		return
	}

	var needsInitialize bool

	if w.Context == nil {
		w.Context = gui.NewQOpenGLContext(w)
		w.Context.SetFormat(w.RequestedFormat())
		w.Context.Create()

		needsInitialize = true
	}

	w.Context.MakeCurrent(w)

	if needsInitialize {
		w.InitializeOpenGLFunctions()
		w.initializeLazy()
	}

	w.render()

	w.Context.SwapBuffers(w)

	if w.Animating {
		w.renderLater()
	}
}

func (w *OpenGLWindow) setAnimating(animating bool) {
	w.Animating = animating

	if animating {
		w.renderLater()
	}
}

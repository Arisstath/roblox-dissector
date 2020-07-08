package main

import (
	"github.com/gotk3/gotk3/gtk"
)

type PacketDetailsViewer struct {
	mainWidget *gtk.Notebook
}

func NewPacketDetailsViewer() (*PacketDetailsViewer, error) {
	viewer := &PacketDetailsViewer{}

	builder, err := gtk.BuilderNewFromFile("packetdetailsviewer.ui")
	if err != nil {
		return nil, err
	}
	notebook_, err := builder.GetObject("packetdetailsnotebook")
	if err != nil {
		return nil, err
	}

	notebook, ok := notebook_.(*gtk.Notebook)
	if !ok {
		return nil, invalidUi("notebook")
	}

	notebookContainer_, err := builder.GetObject("packetdetailsviewerwindow")
	if err != nil {
		return nil, err
	}
	notebookContainer, ok := notebookContainer_.(*gtk.Window)
	if !ok {
		return nil, invalidUi("notebook container")
	}
	notebookContainer.Remove(notebook)

	viewer.mainWidget = notebook

	return viewer, nil
}

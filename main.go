package main

import (
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type State struct {
	startBtn widget.Button
}

func main() {
	eventLoop(app.NewWindow())
	app.Main()
}

func eventLoop(w *app.Window) {
	gofont.Register()
	var (
		state State
		gtx   = layout.NewContext(w.Queue())
		th    = material.NewTheme()
	)

	for e := range w.Events() {
		if evt, ok := e.(system.FrameEvent); ok {
			gtx.Reset(evt.Config, evt.Size)
			th.Button("Start").Layout(gtx, &state.startBtn)
			evt.Frame(gtx.Ops)
		}
	}
}

package main

import (
	"image"
	"image/color"
	"math/rand"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/font/gofont"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

const FieldDimensions = 25
const TileSize = 30

type board [FieldDimensions][FieldDimensions]bool

type State struct {
	board      board // liveness of all cells
	running    bool
	startBtn   widget.Button
	randomBtn  widget.Button
	clearBtn   widget.Button
	lastUpdate time.Time   // time when the game was stopped last
	pressedOn  image.Point // the cell last clicked on
}

func main() {
	go eventLoop(app.NewWindow(
		app.Title("Connways Game of Life"),
		app.Size(unit.Px(FieldDimensions*TileSize), unit.Px(FieldDimensions*TileSize+36))),
	)
	app.Main()
}

func eventLoop(w *app.Window) {
	gofont.Register()
	th := material.NewTheme()
	gtx := layout.NewContext(w.Queue())

	state := new(State)
	state.pressedOn = image.Point{-1, -1}
	var boardHandler event.Key = new(int) // a unique key

	for e := range w.Events() {
		if evt, ok := e.(system.FrameEvent); ok {
			gtx.Reset(evt.Config, evt.Size)

			processInputs(gtx, state, boardHandler)

			layout.NW.Layout(gtx, func() {
				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func() {
						drawPlayground(gtx, state, boardHandler)
					}),
					layout.Rigid(func() {
						drawControls(gtx, th, state)
					}),
				)
			})

			evt.Frame(gtx.Ops)
		}
	}
}

func processInputs(gtx *layout.Context, state *State, boardHandler event.Key) bool {
	repaint := false

	if state.startBtn.Clicked(gtx) {
		state.running = !state.running
		repaint = true
	}
	if state.randomBtn.Clicked(gtx) && !state.running {
		for y := 0; y < FieldDimensions; y++ {
			for x := 0; x < FieldDimensions; x++ {
				tint := false
				if rand.Float32() < 0.2 {
					tint = true
				}
				state.board[x][y] = tint
			}
		}
		repaint = true
	}
	if state.clearBtn.Clicked(gtx) && !state.running {
		state.board = board{}
		repaint = true
	}

	// check for clicks on the board
	for _, bEvt := range gtx.Events(boardHandler) {
		evt := bEvt.(pointer.Event)

		if !((evt.Buttons == pointer.ButtonLeft && evt.Type == pointer.Press) ||
			evt.Type == pointer.Release) {

			if evt.Type == pointer.Press {
				state.pressedOn = image.Point{-1, -1}
			}
			break
		}

		x := int(evt.Position.X) / TileSize
		y := int(evt.Position.Y) / TileSize
		if evt.Type == pointer.Press {
			state.pressedOn = image.Point{x, y}
		} else if evt.Type == pointer.Release && (image.Point{x, y}) == state.pressedOn {
			state.board[x][y] = !state.board[x][y]
			repaint = true
		}
	}
	return repaint
}

func drawPlayground(gtx *layout.Context, state *State, boardHandler event.Key) {

	if state.running && gtx.Now().After(state.lastUpdate.Add(200*time.Millisecond)) {
		gameLoop(state)
		state.lastUpdate = gtx.Now()
	}
	op.InvalidateOp{state.lastUpdate.Add(200 * time.Millisecond)}.Add(gtx.Ops)

	// draw gray background
	paint.ColorOp{color.RGBA{0xF0, 0xF0, 0xF0, 0xFF}}.Add(gtx.Ops)
	paint.PaintOp{
		Rect: f32.Rectangle{
			Max: f32.Point{FieldDimensions * TileSize, FieldDimensions * TileSize},
		},
	}.Add(gtx.Ops)

	// draw tiles
	paint.ColorOp{color.RGBA{0, 0, 0, 0xFF}}.Add(gtx.Ops)
	for y := 0; y < FieldDimensions; y++ {
		for x := 0; x < FieldDimensions; x++ {
			if state.board[x][y] {
				paint.PaintOp{
					Rect: f32.Rectangle{
						Min: f32.Point{float32(x * TileSize), float32(y * TileSize)},
						Max: f32.Point{float32((x + 1) * TileSize), float32((y + 1) * TileSize)},
					},
				}.Add(gtx.Ops)
			}
		}
	}

	// set size
	gtx.Dimensions = layout.Dimensions{
		image.Point{FieldDimensions * TileSize, FieldDimensions * TileSize}, 0,
	}

	// register input handler
	pointer.Rect(image.Rectangle{
		Max: image.Point{FieldDimensions * TileSize, FieldDimensions * TileSize},
	}).Add(gtx.Ops)
	pointer.InputOp{Key: boardHandler}.Add(gtx.Ops)
}

func drawControls(gtx *layout.Context, th *material.Theme, state *State) {
	layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
		layout.Flexed(1./3, func() {
			txt := "Stopped"
			if state.running {
				txt = "Running"
			}
			th.Button(txt).Layout(gtx, &state.startBtn)
		}),
		layout.Flexed(1./3, func() {
			b := th.Button("Random")
			if state.running {
				b.Background = color.RGBA{0x88, 0x88, 0x88, 0xFF}
			}
			b.Layout(gtx, &state.randomBtn)
		}),
		layout.Flexed(1./3, func() {
			b := th.Button("Clear")
			if state.running {
				b.Background = color.RGBA{0x88, 0x88, 0x88, 0xFF}
			}
			b.Layout(gtx, &state.clearBtn)
		}),
	)
}

func countAliveNeigh(b *board, x, y int) int {
	ctr := 0
	env := []int{-1, 0, 1}
	for _, dy := range env {
		for _, dx := range env {
			if dy == 0 && dx == 0 {
				continue
			}
			if x+dx >= FieldDimensions || x+dx < 0 || y+dy >= FieldDimensions || y+dy < 0 {
				continue
			}
			if b[x+dx][y+dy] {
				ctr++
			}
		}
	}
	return ctr
}

func gameLoop(state *State) {
	newBoard := board{}

	for y := 0; y < FieldDimensions; y++ {
		for x := 0; x < FieldDimensions; x++ {
			count := countAliveNeigh(&state.board, x, y)
			if state.board[x][y] {
				if count == 2 || count == 3 {
					newBoard[x][y] = true
				}
			} else {
				if count == 3 {
					newBoard[x][y] = true
				}
			}
		}
	}
	state.board = newBoard
}

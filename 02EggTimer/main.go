package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var progress float32
var progressIncrementer chan float32
var boiling bool
var boilDurationInput widget.Editor
var boilDuration float32

func main() {

	progressIncrementer = make(chan float32)
	go func() {
		for {
			time.Sleep(time.Second / 25)
			progressIncrementer <- 0.004
		}
	}()
	go func() {
		window := new(app.Window)
		app.Title("My first page")
		app.Size(unit.Dp(400), unit.Dp(600))
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window) error {

	go func() {
		for range progressIncrementer {
			if boiling && progress < 1 {
				progress += 1.0 / 25.0 / boilDuration
				if progress >= 1 {
					progress = 1
				}
				// Force a redraw by invalidating the frame
				window.Invalidate()
			}
		}
	}()
	var ops op.Ops

	// startButton is a clickable widget
	var startButton widget.Clickable

	// th defines the material design style
	th := material.NewTheme()
	for {
		switch event := window.Event().(type) {
		case app.DestroyEvent:
			return event.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, event)
			if startButton.Clicked(gtx) {
				boiling = !boiling
				if progress >= 1 {
					progress = 0
				}

				// Read from the input box
				inputString := boilDurationInput.Text()
				inputString = strings.TrimSpace(inputString)
				inputFloat, _ := strconv.ParseFloat(inputString, 32)
				boilDuration = float32(inputFloat)
				boilDuration = boilDuration / (1 - progress)
			}
			if boiling && progress < 1 {
				boilRemain := (1 - progress) * boilDuration
				// Format to 1 decimal.
				inputStr := fmt.Sprintf("%.1f", math.Round(float64(boilRemain)*10)/10)
				// Update the text in the inputbox
				boilDurationInput.SetText(inputStr)
			}
			// Adding Layout to the button using layout
			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(gtx,
				//Egg Display
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						var eggPath clip.Path
						op.Offset(image.Pt(gtx.Dp(200), gtx.Dp(125))).Add(gtx.Ops)
						eggPath.Begin(gtx.Ops)
						// Rotate from 0 to 360 degrees
						for deg := 0.0; deg <= 360; deg++ {

							// Egg math (really) at this brilliant site. Thanks!
							// https://observablehq.com/@toja/egg-curve
							// Convert degrees to radians
							rad := deg / 360 * 2 * math.Pi
							// Trig gives the distance in X and Y direction
							cosT := math.Cos(rad)
							sinT := math.Sin(rad)
							// Constants to define the eggshape
							a := 110.0
							b := 150.0
							d := 20.0
							// The x/y coordinates
							x := a * cosT
							y := -(math.Sqrt(b*b-d*d*cosT*cosT) + d*sinT) * sinT
							// Finally the point on the outline
							p := f32.Pt(float32(x), float32(y))
							// Draw the line to this point
							eggPath.LineTo(p)
						}
						// Close the path
						eggPath.Close()

						// Get hold of the actual clip
						eggArea := clip.Outline{Path: eggPath.End()}.Op()

						// Fill the shape
						// color := color.NRGBA{R: 255, G: 239, B: 174, A: 255}
						color := color.NRGBA{R: 255, G: uint8(239 * (1 - progress)), B: uint8(174 * (1 - progress)), A: 255}
						paint.FillShape(gtx.Ops, color, eggArea)

						d := image.Point{Y: 375}
						return layout.Dimensions{Size: d}
					},
				),
				// Input Button

				// Progress Bar
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						ed := material.Editor(th, &boilDurationInput, "sec")
						boilDurationInput.SingleLine = true
						boilDurationInput.Alignment = text.Middle
						margins := layout.Inset{
							Top:    unit.Dp(0),
							Right:  unit.Dp(170),
							Bottom: unit.Dp(40),
							Left:   unit.Dp(170),
						}
						// ... and borders ...
						border := widget.Border{
							Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
							CornerRadius: unit.Dp(3),
							Width:        unit.Dp(2),
						}
						// ... before laying it out, one inside the other
						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								return border.Layout(gtx, ed.Layout)
							})
					},
				),
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						bar := material.ProgressBar(th, progress) // Here progress is used
						return bar.Layout(gtx)
					},
				),
				// Start Stop Button
				layout.Rigid(
					func(gtx layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(35),
							Left:   unit.Dp(35),
						}
						return margins.Layout(gtx,
							func(gtx layout.Context) layout.Dimensions {
								var text string
								text = "Start"
								if boiling && progress < 1 {
									text = "Stop"
								}
								if boiling && progress >= 1 {
									text = "Finished"
								}
								btn := material.Button(th, &startButton, text)
								return btn.Layout(gtx)
							},
						)
					}),
			)
			event.Frame(gtx.Ops)
		}
	}
}

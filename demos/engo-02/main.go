package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

type DefaultScene struct{}

var (
	zoomSpeed   float32 = -0.125
	scrollSpeed float32 = 700

	border float32 = 10000

	//worldWidth  int = 10000
	//worldHeight int = 10000

	borderLeft, borderRight, borderTop, borderBottom uint64
)

func (*DefaultScene) Preload() {}

// Setup is called before the main loop is started
func (*DefaultScene) Setup(w *ecs.World) {
	common.CameraBounds = engo.AABB{Min: engo.Point{X: -border / 2, Y: -border / 2}, Max: engo.Point{X: border / 2, Y: border / 2}}
	common.SetBackground(color.RGBA{55, 55, 55, 255})

	w.AddSystem(&common.RenderSystem{})
	w.AddSystem(&common.CollisionSystem{})
	w.AddSystem(&MovingSystem{entities: map[uint64]movingEntity{}})
	w.AddSystem(&ShapeSpawnSystem{})

	// Adding camera controllers so we can verify it doesn't break when we move
	w.AddSystem(common.NewKeyboardScroller(scrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis))
	w.AddSystem(&common.MouseZoomer{ZoomSpeed: zoomSpeed})
	w.AddSystem(&common.MouseRotator{RotationSpeed: 0.125})

	new_border(w, border)

	engo.Mailbox.Dispatch(
		common.CameraMessage{
			Axis:        common.YAxis,
			Value:       0,
			Incremental: false,
		})
	engo.Mailbox.Dispatch(
		common.CameraMessage{
			Axis:        common.XAxis,
			Value:       0,
			Incremental: false,
		})

}

func (*DefaultScene) Type() string { return "Game" }

func main() {
	opts := engo.RunOptions{
		Title:          "Shapes Demo",
		Width:          1000,
		Height:         1000,
		StandardInputs: true,
		MSAA:           4, // This one is not mandatory, but makes the shapes look so much better when rotating the camera
	}
	engo.Run(opts, &DefaultScene{})
}

func new_border(w *ecs.World, border float32) {

	c := color.RGBA{200, 200, 200, 255}
	borderTop = new_line(w, -border/2, -10-border/2, border, 10, c)
	borderBottom = new_line(w, -border/2, border/2, border, 10, c)
	borderLeft = new_line(w, -10-border/2, -border/2, 10, border, c)
	borderRight = new_line(w, border/2, -border/2, 10, border, c)

}

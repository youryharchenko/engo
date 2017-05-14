package main

import (
	"image/color"
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

type Shape struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	common.CollisionComponent
	MovingComponent
}

func NewShape(world *ecs.World, position engo.Point) {
	new_circle(world,
		position.X,
		position.Y,
		float32(rand.Intn(100)),
		color.RGBA{
			uint8(rand.Intn(255)),
			uint8(rand.Intn(255)),
			uint8(rand.Intn(255)),
			uint8(rand.Intn(255)),
		})
}

func DestroyShape(world *ecs.World, shape ecs.BasicEntity) {
	del_circle(world, shape)
}

func new_circle(w *ecs.World, x float32, y float32, r float32, c color.RGBA) {
	circle := Shape{BasicEntity: ecs.NewBasic()}
	circle.SpaceComponent = common.SpaceComponent{Position: engo.Point{X: x, Y: y}, Width: 2 * r, Height: 2 * r}
	circle.RenderComponent = common.RenderComponent{Drawable: common.Circle{}, Color: c}
	circle.MovingComponent = MovingComponent{r: r, v: engo.Point{X: float32(200 - rand.Intn(400)), Y: float32(200 - rand.Intn(400))}}
	circle.CollisionComponent = common.CollisionComponent{Main: true, Solid: true}

	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&circle.BasicEntity, &circle.RenderComponent, &circle.SpaceComponent)
		case *common.CollisionSystem:
			sys.Add(&circle.BasicEntity, &circle.CollisionComponent, &circle.SpaceComponent)
		case *MovingSystem:
			sys.Add(&circle.BasicEntity, &circle.MovingComponent, &circle.SpaceComponent)
		}
	}
}

func del_circle(w *ecs.World, e ecs.BasicEntity) {
	for _, system := range w.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Remove(e)
		case *common.CollisionSystem:
			sys.Remove(e)
		case *MovingSystem:
			sys.Remove(e)
		}
	}
}

func new_line(world *ecs.World, x float32, y float32, w float32, h float32, c color.RGBA) uint64 {
	line := Shape{BasicEntity: ecs.NewBasic()}
	line.SpaceComponent = common.SpaceComponent{Position: engo.Point{X: x, Y: y}, Width: w, Height: h}
	line.RenderComponent = common.RenderComponent{Drawable: common.Rectangle{}, Color: c}
	line.CollisionComponent = common.CollisionComponent{Main: false, Solid: true}

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&line.BasicEntity, &line.RenderComponent, &line.SpaceComponent)
		case *common.CollisionSystem:
			sys.Add(&line.BasicEntity, &line.CollisionComponent, &line.SpaceComponent)
		}
	}
	return line.BasicEntity.ID()
}

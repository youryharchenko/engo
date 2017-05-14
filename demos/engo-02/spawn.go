package main

import (
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
)

type ShapeSpawnSystem struct {
	world *ecs.World
}

func (spawn *ShapeSpawnSystem) New(w *ecs.World) {
	spawn.world = w
	engo.Mailbox.Listen("MovingMessage", func(message engo.Message) {
		e := message.(MovingMessage).Entity.BasicEntity
		DestroyShape(w, *e)
	})
}

func (*ShapeSpawnSystem) Remove(ecs.BasicEntity) {}

func (spawn *ShapeSpawnSystem) Update(dt float32) {
	// 2% change of spawning a shape each frame
	if countMovingEntity > 200 || rand.Float32() < .98 {
		return
	}

	position := engo.Point{
		X: rand.Float32()*1000 + engo.GameWidth()/2 - 1000,
		Y: rand.Float32()*1000 + engo.GameHeight()/2 - 1000,
	}
	NewShape(spawn.world, position)
}

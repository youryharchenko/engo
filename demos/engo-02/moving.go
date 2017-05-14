package main

import (
	"log"
	"math"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

var countMovingEntity uint64

type MovingComponent struct {
	v engo.Point
	r float32
}

type MovingMessage struct {
	Entity movingEntity
}

func (MovingMessage) Type() string { return "MovingMessage" }

type movingEntity struct {
	*ecs.BasicEntity
	*MovingComponent
	*common.SpaceComponent
}

type MovingSystem struct {
	//entities []movingEntity
	count    uint64
	entities map[uint64]movingEntity
}

func (m *MovingSystem) New(*ecs.World) {
	engo.Mailbox.Listen("CollisionMessage", func(message engo.Message) {
		//log.Println("collision")

		collision, isCollision := message.(common.CollisionMessage)
		if isCollision {
			// See if we also have that Entity, and if so, change the speed
			if e1, ok := m.entities[collision.Entity.BasicEntity.ID()]; ok {
				switch collision.To.BasicEntity.ID() {
				case borderLeft, borderRight:
					e1.MovingComponent.v.X *= -1
				case borderTop, borderBottom:
					e1.MovingComponent.v.Y *= -1
				default:
					if e2, ok := m.entities[collision.To.BasicEntity.ID()]; ok {
						c1 := e1.SpaceComponent.Center()
						c2 := e2.SpaceComponent.Center()
						dist := c1.PointDistance(c2)

						r1 := e1.MovingComponent.r
						r2 := e2.MovingComponent.r

						if dist <= r1*1.01+r2*1.01 && dist >= r1*0.09+r2*0.09 {
							mag := float32(math.Sqrt(float64((c1.X-c2.X)*(c1.X-c2.X) + (c1.Y-c2.Y)*(c1.Y-c2.Y))))
							ex := engo.Point{X: (c1.X - c2.X) / mag, Y: (c1.Y - c2.Y) / mag}
							ey := engo.Point{X: -ex.Y, Y: ex.X}

							log.Println(c1, c2, mag, ex, ey)

							d := ex.X*ey.Y - ey.X*ex.Y
							exs := engo.Point{X: ey.Y / d, Y: -ex.Y / d}
							eys := engo.Point{X: -ey.X / d, Y: ex.X / d}

							log.Println(ex.X*exs.X+ey.X*exs.Y,
								ex.X*eys.X+ey.X*eys.Y,
								ex.Y*exs.X+ey.Y*exs.Y,
								ex.Y*eys.X+ey.Y*eys.Y,
							)

							m1 := r1 * r1 * 3.14
							m2 := r2 * r2 * 3.14

							x1n := e1.MovingComponent.v.X
							x2n := e2.MovingComponent.v.X
							y1n := e1.MovingComponent.v.Y
							y2n := e2.MovingComponent.v.Y

							//x1 := ex.X * (x1n*ex.X + y1n*ex.Y)
							//y1 := ey.Y * (x1n*ey.X + y1n*ey.Y)
							//x2 := ex.X * (x2n*ex.X + y2n*ex.Y)
							//y2 := ey.Y * (x2n*ey.X + y2n*ey.Y)

							x1 := (x1n*ex.X + y1n*ex.Y)
							y1 := (x1n*ey.X + y1n*ey.Y)
							x2 := (x2n*ex.X + y2n*ex.Y)
							y2 := (x2n*ey.X + y2n*ey.Y)

							x1a := ((m1-m2)*x1 + 2*m2*x2) / (m1 + m2)
							y1a := ((m1-m2)*y1 + 2*m2*y2) / (m1 + m2)
							x2a := ((m2-m1)*x2 + 2*m1*x1) / (m1 + m2)
							y2a := ((m2-m1)*y2 + 2*m1*y1) / (m1 + m2)

							//e1.MovingComponent.v.X = ((m1-m2)*x1 + 2*m2*x2) / (m1 + m2)
							//e1.MovingComponent.v.Y = ((m1-m2)*y1 + 2*m2*y2) / (m1 + m2)
							//e2.MovingComponent.v.X = ((m2-m1)*x2 + 2*m1*x1) / (m1 + m2)
							//e2.MovingComponent.v.Y = ((m2-m1)*y2 + 2*m1*y1) / (m1 + m2)
							x1e := (x1a*exs.X + y1a*exs.Y) //x1a
							y1e := (x1a*eys.X + y1a*eys.Y) //y1a
							x2e := (x2a*exs.X + y2a*exs.Y) //x2a
							y2e := (x2a*eys.X + y2a*eys.Y) //y2a

							e1.MovingComponent.v.X = x1e
							e1.MovingComponent.v.Y = y1e
							e2.MovingComponent.v.X = x2e
							e2.MovingComponent.v.Y = y2e

							log.Println(x1n+x2n, x1e+x2e)
							log.Println(y1n+y2n, y1e+y2e)

							log.Println(m1*x1n+m2*x2n, m1*x1e+m2*x2e)
							log.Println(m1*y1n+m2*y2n, m1*y1e+m2*y2e)
						}
						//log.Println(collision.Entity.BasicEntity.ID(), collision.To.BasicEntity.ID())
						//c1 := e1.SpaceComponent.Center()
						//log.Println("Distance", c1.PointDistance(e2.SpaceComponent.Center()))
						//colliding(e1.MovingComponent, e2.MovingComponent, e1.SpaceComponent, e2.SpaceComponent)
					}
				}
			}
		}
	})
}

func (m *MovingSystem) Add(basic *ecs.BasicEntity, moving *MovingComponent, space *common.SpaceComponent) {
	//m.entities = append(m.entities, movingEntity{basic, moving, space})
	m.entities[basic.ID()] = movingEntity{basic, moving, space}
	countMovingEntity++
}

func (m *MovingSystem) Remove(basic ecs.BasicEntity) {

	delete(m.entities, basic.ID())
	//log.Printf("MovingSystem Remove>> id: %d, x: %f, y: %f", e.BasicEntity.ID(), e.SpaceComponent.Position.X, e.SpaceComponent.Position.Y)
	countMovingEntity--
}

func (m *MovingSystem) Update(dt float32) {
	//movingMultiplier := float32(100)
	//if e, ok := m.entities[1]; ok {

	if m.count > 100 {
		m.count = 0
		log.Printf("countMovingEntity: %d", countMovingEntity)
		//log.Printf("MovingSystem Update>> id: %d, x: %f, y: %f", e.BasicEntity.ID(), e.SpaceComponent.Position.X, e.SpaceComponent.Position.Y)
	} else {
		m.count++
	}
	//}

	for _, e := range m.entities {

		y := e.SpaceComponent.Position.Y + e.SpaceComponent.Height/2
		x := e.SpaceComponent.Position.X + e.SpaceComponent.Width/2
		if x < -5010 || y < -5010 || x > 5010 || y > 5010 {
			engo.Mailbox.Dispatch(MovingMessage{
				Entity: e,
			})
		}

		e.SpaceComponent.Position.X += e.MovingComponent.v.X * dt
		e.SpaceComponent.Position.Y += e.MovingComponent.v.Y * dt

		//var directionX, directionY float32
		//if e.MovingComponent.X > 0 {
		//if rand.Intn(2) == 1 {
		//	directionX = 1.0
		//} else {
		//	directionX = -1.0
		//}
		//if e.MovingComponent.Y > 0 {
		//if rand.Intn(2) == 1 {
		//	directionY = 1.0
		//} else {
		//	directionY = -1.0
		//}

		//e.MovingComponent.X += movingMultiplier * dt * directionX
		//e.MovingComponent.Y += movingMultiplier * dt * directionY
	}
}

func colliding(c1 *MovingComponent, c2 *MovingComponent, s1 *common.SpaceComponent, s2 *common.SpaceComponent) {
	log.Println("Start colliding", c1, c2, c1.r+c2.r)
	//Vec2 rv = B.velocity - A.velocity
	rv := c1.v
	prv := &rv
	prv.Subtract(c2.v)
	log.Println("rv", rv)
	// Calculate relative velocity in terms of the normal direction
	//float velAlongNormal = DotProduct( rv, normal )
	line := engo.Line{P1: c2.v, P2: c1.v}
	//normal, _ := rv.Normalize()
	normal := line.Normal()
	log.Println("normal", normal)
	velAlongNormal := engo.DotProduct(*prv, normal)
	log.Println("velAlongNormal", velAlongNormal)
	// Do not resolve if velocities are separating
	if velAlongNormal > 0 {
		//return;

	} else {
		// Calculate restitution
		//float e = min( A.restitution, B.restitution)
		e := float32(1)

		m1 := c1.r * c1.r * 3.14
		m2 := c2.r * c2.r * 3.14
		p1 := engo.Point{X: c1.v.X * m1, Y: c1.v.Y * m1}
		p2 := engo.Point{X: c2.v.X * m2, Y: c2.v.Y * m2}
		log.Println("m1, m2, p1, p2, suma", m1, m2, p1, p2, engo.Point{X: p1.X + p2.X, Y: p1.Y + p2.Y})
		// Calculate impulse scalar
		//float j = -(1 + e) * velAlongNormal
		//j /= 1 / A.mass + 1 / B.mass
		j := -(1 + e) * velAlongNormal
		j /= 1/m1 + 1/m2
		log.Println("j", j)

		// Apply impulse
		//Vec2 impulse = j * normal

		imp1 := normal
		imp2 := normal

		pImp1 := &imp1
		pImp1.MultiplyScalar(j)
		pImp2 := &imp2
		pImp2.MultiplyScalar(j)

		//A.velocity -= 1 / A.mass * impulse
		pImp1.MultiplyScalar(1 / m1)
		log.Println("pImp1", *pImp1)
		c1.v.Subtract(*pImp1)
		log.Println("c1.v", c1.v)

		//B.velocity += 1 / B.mass * impulse
		pImp2.MultiplyScalar(1 / m2)
		log.Println("pImp2", *pImp2)
		c2.v.Add(*pImp2)
		log.Println("c2.v", c2.v)

		c1X := c1.v.X - pImp1.X
		c1Y := c1.v.Y - pImp1.Y
		np1 := engo.Point{X: c1X * m1, Y: c1Y * m1}

		c2X := c2.v.X + pImp2.X
		c2Y := c2.v.Y + pImp2.Y
		np2 := engo.Point{X: c2X * m2, Y: c2Y * m2}
		log.Println("m1, m2, np1, np2, nsuma", m1, m2, np1, np2, engo.Point{X: np1.X + np2.X, Y: np1.Y + np2.Y})
	}
}

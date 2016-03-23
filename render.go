package engi

import (
	"image/color"
	"math"

	"github.com/paked/engi/ecs"
	"github.com/paked/webgl"
)

const (
	// HighestGround is the highest PriorityLevel that will be rendered
	HighestGround PriorityLevel = 50
	// HUDGround is a PriorityLevel from which everything isn't being affected by the Camera
	HUDGround    PriorityLevel = 40
	Foreground   PriorityLevel = 30
	MiddleGround PriorityLevel = 20
	ScenicGround PriorityLevel = 10
	// Background is the lowest PriorityLevel that will be rendered
	Background PriorityLevel = 0
	// Hidden indicates that it should not be rendered by the RenderSystem
	Hidden PriorityLevel = -1
)

const (
	RenderSystemPriority = -1000
)

type PriorityLevel int

type Drawable interface {
	Texture() *webgl.Texture
	Width() float32
	Height() float32
	View() (float32, float32, float32, float32)
}

type renderChangeMessage struct {
	entity      *ecs.Entity
	oldPriority PriorityLevel
	newPriority PriorityLevel
}

func (renderChangeMessage) Type() string {
	return "renderChangeMessage"
}

type RenderComponent struct {
	scale        Point
	Label        string
	priority     PriorityLevel
	Transparency float32
	Color        color.Color

	drawable      Drawable
	buffer        *webgl.Buffer
	bufferContent []float32
}

func NewRenderComponent(d Drawable, scale Point, label string) *RenderComponent {
	rc := &RenderComponent{
		Label:        label,
		Transparency: 1,
		Color:        color.White,

		scale:    scale,
		priority: MiddleGround,
	}
	rc.SetDrawable(d)

	return rc
}

func (r *RenderComponent) SetPriority(p PriorityLevel) {
	r.priority = p
	Mailbox.Dispatch(renderChangeMessage{})
}

func (r *RenderComponent) Priority() PriorityLevel {
	return r.priority
}

func (r *RenderComponent) SetDrawable(d Drawable) {
	r.drawable = d
	r.preloadTexture()
}

func (r *RenderComponent) Drawable() Drawable {
	return r.drawable
}

func (r *RenderComponent) SetScale(scale Point) {
	r.scale = scale
	r.preloadTexture()
}

func (r *RenderComponent) Scale() Point {
	return r.scale
}

func (*RenderComponent) Type() string {
	return "RenderComponent"
}

// Init is called to initialize the RenderElement
func (ren *RenderComponent) preloadTexture() {
	if ren.drawable == nil || headless {
		return
	}

	ren.bufferContent = ren.generateBufferContent()

	ren.buffer = Gl.CreateBuffer()
	Gl.BindBuffer(Gl.ARRAY_BUFFER, ren.buffer)
	Gl.BufferData(Gl.ARRAY_BUFFER, ren.bufferContent, Gl.STATIC_DRAW)

	// TODO: ask why this doesn't work
	// ren.bufferContent = make([]float32, 0)
}

// generateBufferContent computes information about the 4 vertices needed to draw the texture, which should
// be stored in the buffer
func (ren *RenderComponent) generateBufferContent() []float32 {
	scaleX := ren.scale.X
	scaleY := ren.scale.Y
	transparency := float32(1.0)
	c := ren.Color

	fx := float32(0)
	fy := float32(0)
	fx2 := ren.drawable.Width()
	fy2 := ren.drawable.Height()

	if scaleX != 1 || scaleY != 1 {
		//fx *= scaleX
		//fy *= scaleY
		fx2 *= scaleX
		fy2 *= scaleY
	}

	x1 := fx
	y1 := fy
	x2 := fx
	y2 := fy2
	x3 := fx2
	y3 := fy2
	x4 := fx2
	y4 := fy

	colorR, colorG, colorB, _ := c.RGBA()

	red := colorR
	green := colorG << 8
	blue := colorB << 16
	alpha := uint32(transparency*255.0) << 24

	tint := math.Float32frombits((alpha | blue | green | red) & 0xfeffffff)

	u, v, u2, v2 := ren.drawable.View()

	return []float32{x1, y1, u, v, tint, x4, y4, u2, v, tint, x3, y3, u2, v2, tint, x2, y2, u, v2, tint}
}

type RenderSystem struct {
	ecs.LinearSystem

	renders map[PriorityLevel][]*ecs.Entity
	changed bool
	world   *ecs.World
}

func (rs *RenderSystem) New(w *ecs.World) {
	rs.renders = make(map[PriorityLevel][]*ecs.Entity)
	rs.world = w

	if !headless {
		if !Shaders.setup {
			Shaders.def.Initialize(Width(), Height())

			hud := &HUDShader{}
			hud.Initialize(Width(), Height())
			for i := HUDGround; i <= HighestGround; i++ {
				Shaders.Register(i, hud)
			}

			Shaders.setup = true
		}
	}

	Mailbox.Listen("renderChangeMessage", func(m Message) {
		rs.changed = true
	})
}

func (rs *RenderSystem) AddEntity(e *ecs.Entity) {
	rs.changed = true
	rs.LinearSystem.AddEntity(e)
}

func (rs *RenderSystem) RemoveEntity(e *ecs.Entity) {
	rs.changed = true
	rs.LinearSystem.RemoveEntity(e)
}

func (rs *RenderSystem) Pre() {
	if !headless {
		Gl.Clear(Gl.COLOR_BUFFER_BIT)
	}

	if !rs.changed {
		return
	}

	rs.renders = make(map[PriorityLevel][]*ecs.Entity)
}

func (rs *RenderSystem) Post() {
	if headless {
		return
	}

	var currentShader Shader

	for i := Background; i <= HighestGround; i++ {
		if len(rs.renders[i]) == 0 {
			continue
		}

		// Retrieve a batch, may be the default one -- then use it if we arent already using it
		s := Shaders.Get(i)
		if s != currentShader {
			if currentShader != nil {
				currentShader.Post()
			}
			s.Pre()
			currentShader = s
		}

		// Then render everything for this level
		for _, entity := range rs.renders[i] {
			var (
				render *RenderComponent
				space  *SpaceComponent
				ok     bool
			)

			if render, ok = entity.ComponentFast(render).(*RenderComponent); !ok {
				continue // with other entities
			}

			if space, ok = entity.ComponentFast(space).(*SpaceComponent); !ok {
				continue // with other entities
			}

			s.Draw(render.drawable.Texture(), render.buffer, render.drawable.Width(), render.drawable.Height(), space.Position.X, space.Position.Y, space.Rotation)
		}
	}

	if currentShader != nil {
		currentShader.Post()
	}

	rs.changed = false
}

func (rs *RenderSystem) UpdateEntity(entity *ecs.Entity, dt float32) {
	if !rs.changed {
		return
	}

	var render *RenderComponent
	var ok bool

	if render, ok = entity.ComponentFast(render).(*RenderComponent); !ok {
		return
	}

	rs.renders[render.priority] = append(rs.renders[render.priority], entity)
}

func (*RenderSystem) Type() string {
	return "RenderSystem"
}

func (*RenderSystem) Priority() int {
	return RenderSystemPriority
}

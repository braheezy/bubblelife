package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Sphere represents a 3D sphere (bubble) in the Game of Life.
type Sphere struct {
	Position     mgl32.Vec3
	CurrentState bool    // Current alive/dead state
	NextState    bool    // Next alive/dead state after GoL logic
	Radius       float32 // The radius for rendering (0 = dead, 1 = fully alive)
	Animating    bool    // Is the sphere currently animating its radius
}

var sphereVAO, instanceVBO, instanceRadiusVBO uint32

// NewSphere creates a new sphere at a specific position.
func NewSphere(position mgl32.Vec3) *Sphere {
	sphere := &Sphere{
		Position: position,
	}

	return sphere
}

// Initialize the buffer for storing instance-specific data (sphere positions)
func InitInstanceBuffer(spheres []*Sphere) {
	// Extract the positions and radii of the spheres
	positions := make([]mgl32.Vec3, len(spheres))
	radii := make([]float32, len(spheres)) // Store radii for each sphere

	for i, sphere := range spheres {
		positions[i] = sphere.Position
		radii[i] = sphere.Radius // Use sphere.Radius (0 for dead spheres, >0 for alive spheres)
	}

	// Generate and bind VAO first
	gl.GenVertexArrays(1, &sphereVAO)
	gl.BindVertexArray(sphereVAO)

	// Generate and bind instance VBO for positions
	gl.GenBuffers(1, &instanceVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(positions)*3*4, gl.Ptr(positions), gl.STATIC_DRAW)

	// Enable instance attribute for position (Vec3)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.Ptr(nil))
	gl.VertexAttribDivisor(0, 1) // Each instance uses a different position

	// Generate and bind instance VBO for radii
	gl.GenBuffers(1, &instanceRadiusVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceRadiusVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(radii)*4, gl.Ptr(radii), gl.STATIC_DRAW)

	// Enable instance attribute for radius (float)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 1, gl.FLOAT, false, 1*4, gl.Ptr(nil))
	gl.VertexAttribDivisor(1, 1) // Each instance uses a different radius

	// Unbind VAO
	gl.BindVertexArray(0)
}

// Update buffer for radii dynamically as we animate
func UpdateRadiiBuffer(spheres []*Sphere) {
	// Extract the updated radii of the spheres
	radii := make([]float32, len(spheres))
	for i, sphere := range spheres {
		radii[i] = sphere.Radius
	}

	// Update instance VBO for radii (update the buffer on GPU)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceRadiusVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(radii)*4, gl.Ptr(radii))
}

// Render renders the sphere by binding its VAO and issuing a draw call.
func renderSpheres(shader *Shader, sphereCount int) {
	// Use the shader program
	shader.use()

	// Bind the VAO (which now contains only the sphere positions)
	gl.BindVertexArray(sphereVAO)

	// Draw all instances of the spheres with one draw call (rendering points)
	gl.DrawArraysInstanced(gl.POINTS, 0, 1, int32(sphereCount))

	// Unbind
	gl.BindVertexArray(0)
}

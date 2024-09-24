package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Sphere represents a 3D sphere (bubble) in the Game of Life.
type Sphere struct {
	Position mgl32.Vec3
	IsAlive  bool
	Radius   float32
}

var sphereVAO, instanceVBO uint32

// NewSphere creates a new sphere at a specific position.
func NewSphere(position mgl32.Vec3, radius float32) *Sphere {
	sphere := &Sphere{
		Position: position,
		IsAlive:  true,
		Radius:   radius,
	}

	return sphere
}

// Initialize the buffer for storing instance-specific data (sphere positions)
func InitInstanceBuffer(spheres []*Sphere) {
	// Extract the positions of the spheres
	positions := make([]mgl32.Vec3, len(spheres))
	for i, sphere := range spheres {
		positions[i] = sphere.Position
	}

	// Generate and bind instance VBO for sphere positions
	gl.GenBuffers(1, &instanceVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(positions)*3*4, gl.Ptr(positions), gl.STATIC_DRAW)

	// Generate and bind VAO
	gl.GenVertexArrays(1, &sphereVAO)
	gl.BindVertexArray(sphereVAO)

	// Enable instance attribute for position (Vec3)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.Ptr(nil))
	gl.VertexAttribDivisor(0, 1) // Each instance uses a different position

	// Unbind VAO
	gl.BindVertexArray(0)
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

package main

import (
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Sphere represents a 3D sphere (bubble) in the Game of Life.
type Sphere struct {
	Position     mgl32.Vec3
	CurrentState bool       // Current alive/dead state
	NextState    bool       // Next alive/dead state after GoL logic
	Radius       float32    // The radius for rendering (0 = dead, 1 = fully alive)
	Animating    bool       // Is the sphere currently animating its radius
	Color        mgl32.Vec3 // Color of the sphere
	GroupID      int        // Group ID to distinguish clusters of alive spheres
}

var sphereVAO, instanceVBO, instanceRadiusVBO, instanceColorVBO uint32

// NewSphere creates a new sphere at a specific position.
func NewSphere(position mgl32.Vec3) *Sphere {
	sphere := &Sphere{
		Position: position,
		Color:    white,
	}

	return sphere
}

// InitInstanceBuffer initializes the buffer for storing instance-specific data (sphere positions and colors).
func InitInstanceBuffer(spheres []*Sphere) {
	// Generate the VAO
	gl.GenVertexArrays(1, &sphereVAO)
	gl.BindVertexArray(sphereVAO) // Bind the VAO

	// Extract the positions, radii, and colors of the spheres
	positions := make([]mgl32.Vec3, len(spheres))
	radii := make([]float32, len(spheres))
	colors := make([]mgl32.Vec3, len(spheres)) // Add colors array

	for i, sphere := range spheres {
		positions[i] = sphere.Position
		radii[i] = sphere.Radius
		colors[i] = sphere.Color // Populate with each sphere's color
	}

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

	// Generate and bind instance VBO for colors
	gl.GenBuffers(1, &instanceColorVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceColorVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(colors)*3*4, gl.Ptr(colors), gl.STATIC_DRAW)

	// Enable instance attribute for color (Vec3)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 3*4, gl.Ptr(nil))
	gl.VertexAttribDivisor(2, 1) // Each instance uses a different color

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

// UpdateColorBuffer updates the instance color buffer on the GPU.
func UpdateColorBuffer(spheres []*Sphere) {
	// Extract the updated colors of the spheres
	colors := make([]mgl32.Vec3, len(spheres))
	for i, sphere := range spheres {
		colors[i] = sphere.Color
	}

	// Update instance VBO for colors (update the buffer on GPU)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceColorVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(colors)*3*4, gl.Ptr(colors))
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

// FindNeighbors finds neighboring spheres for the given position within the grid.
func FindNeighbors(spheres []*Sphere, index int, N, M int, spacing float32) []int {
	neighbors := []int{}
	currentSphere := spheres[index]

	// Define the possible neighbor offsets in 3D (6 possible directions)
	offsets := []mgl32.Vec3{
		{spacing, 0, 0}, {-spacing, 0, 0}, // X-axis neighbors
		{0, spacing, 0}, {0, -spacing, 0}, // Y-axis neighbors
		{0, 0, spacing}, {0, 0, -spacing}, // Z-axis neighbors
	}

	// Iterate through spheres to find those close enough to be neighbors
	for i, sphere := range spheres {
		if i != index && sphere.CurrentState {
			for _, offset := range offsets {
				if sphere.Position.ApproxEqualThreshold(currentSphere.Position.Add(offset), 0.01) {
					neighbors = append(neighbors, i)
					break
				}
			}
		}
	}

	return neighbors
}

// BFS to find groups of connected alive spheres.
func FindGroups(spheres []*Sphere, N, M int, spacing float32) int {
	visited := make([]bool, len(spheres))
	groupID := 0

	for i := 0; i < len(spheres); i++ {
		if !visited[i] && spheres[i].CurrentState {
			// Perform BFS to find all connected alive spheres
			queue := []int{i}
			visited[i] = true
			spheres[i].GroupID = groupID

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]

				neighbors := FindNeighbors(spheres, curr, N, M, spacing)
				for _, neighbor := range neighbors {
					if !visited[neighbor] {
						visited[neighbor] = true
						spheres[neighbor].GroupID = groupID
						queue = append(queue, neighbor)
					}
				}
			}
			groupID++
		}
	}

	return groupID // Return total number of groups found
}

// AssignColorsToGroups assigns a unique color to each group of connected spheres.
func AssignColorsToGroups(spheres []*Sphere, numGroups int) {
	groupBaseColors := make([]mgl32.Vec3, numGroups)

	// Generate base pastel colors for each group
	for g := 0; g < numGroups; g++ {
		groupBaseColors[g] = mgl32.Vec3{
			0.6 + rand.Float32()*0.3, // Random pastel-like color (R)
			0.6 + rand.Float32()*0.3, // Random pastel-like color (G)
			0.6 + rand.Float32()*0.3, // Random pastel-like color (B)
		}
	}

	// Assign each sphere in a group a slightly varied color based on the base color
	for _, sphere := range spheres {
		if sphere.CurrentState {
			baseColor := groupBaseColors[sphere.GroupID]
			sphere.Color = mgl32.Vec3{
				baseColor.X(), // Slight variation in R
				baseColor.Y(), // Slight variation in G
				baseColor.Z(), // Slight variation in B
			}
		}
	}
}

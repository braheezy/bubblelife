package main

import (
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Bubble represents a 3D sphere (bubble) in the Game of Life.
type Bubble struct {
	Position mgl32.Vec3
	// Current alive/dead state
	CurrentState bool
	// Next alive/dead state
	NextState bool
	// The radius for rendering (0 = dead, 1 = fully alive, other = transitioning)
	Radius float32
	// Is the bubble currently animating its radius
	Animating bool
	// Color of the bubble
	Color mgl32.Vec3
	// Group ID to distinguish clusters of alive bubbles
	GroupID int
}

var bubbleVAO, instanceVBO, instanceRadiusVBO, instanceColorVBO uint32

// NewBubble creates a new bubble at a specific position.
func NewBubble(position mgl32.Vec3) *Bubble {
	bubble := &Bubble{
		Position: position,
		Color:    textColor,
	}

	return bubble
}

// initInstanceBuffer initializes the buffer for storing instance-specific data (bubble positions and colors).
func initInstanceBuffer(bubbles []*Bubble) {
	// Generate the VAO
	gl.GenVertexArrays(1, &bubbleVAO)
	gl.BindVertexArray(bubbleVAO) // Bind the VAO

	// Extract the positions, radii, and colors of the bubbles
	positions := make([]mgl32.Vec3, len(bubbles))
	radii := make([]float32, len(bubbles))
	colors := make([]mgl32.Vec3, len(bubbles))

	for i, bubble := range bubbles {
		positions[i] = bubble.Position
		radii[i] = bubble.Radius
		colors[i] = bubble.Color
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
	gl.VertexAttribDivisor(1, 1)

	// Generate and bind instance VBO for colors
	gl.GenBuffers(1, &instanceColorVBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceColorVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(colors)*3*4, gl.Ptr(colors), gl.STATIC_DRAW)

	// Enable instance attribute for color (Vec3)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 3*4, gl.Ptr(nil))
	gl.VertexAttribDivisor(2, 1)

	// Unbind VAO
	gl.BindVertexArray(0)
}

// Update buffer for radii dynamically as we animate
func updateRadiiBuffer(bubbles []*Bubble) {
	// Extract the updated radii of the bubbles
	radii := make([]float32, len(bubbles))
	for i, bubble := range bubbles {
		radii[i] = bubble.Radius
	}

	// Update instance VBO for radii (update the buffer on GPU)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceRadiusVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(radii)*4, gl.Ptr(radii))
}

// updateColorBuffer updates the instance color buffer on the GPU.
func updateColorBuffer(bubbles []*Bubble) {
	// Extract the updated colors of the bubbles
	colors := make([]mgl32.Vec3, len(bubbles))
	for i, bubble := range bubbles {
		colors[i] = bubble.Color
	}

	// Update instance VBO for colors (update the buffer on GPU)
	gl.BindBuffer(gl.ARRAY_BUFFER, instanceColorVBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(colors)*3*4, gl.Ptr(colors))
}

// Render renders the bubble by binding its VAO and issuing a draw call.
func renderBubbles(shader *Shader, bubbleCount int) {
	// Use the shader program
	shader.use()

	// Bind the VAO (which now contains only the bubble positions)
	gl.BindVertexArray(bubbleVAO)

	// Draw all instances of the bubbles with one draw call (rendering points)
	gl.DrawArraysInstanced(gl.POINTS, 0, 1, int32(bubbleCount))

	// Unbind
	gl.BindVertexArray(0)
}

// FindNeighbors finds neighboring bubbles for the given position within the grid.
func FindNeighbors(bubbles []*Bubble, index int, N, M int, spacing float32) []int {
	neighbors := []int{}
	currentBubble := bubbles[index]

	// Define the possible neighbor offsets in 3D (6 possible directions)
	offsets := []mgl32.Vec3{
		{spacing, 0, 0}, {-spacing, 0, 0}, // X-axis neighbors
		{0, spacing, 0}, {0, -spacing, 0}, // Y-axis neighbors
		{0, 0, spacing}, {0, 0, -spacing}, // Z-axis neighbors
	}

	// Iterate through bubbles to find those close enough to be neighbors
	for i, bubble := range bubbles {
		if i != index && bubble.CurrentState {
			for _, offset := range offsets {
				if bubble.Position.ApproxEqualThreshold(currentBubble.Position.Add(offset), 0.01) {
					neighbors = append(neighbors, i)
					break
				}
			}
		}
	}

	return neighbors
}

// BFS to find groups of connected alive bubbles.
func findGroups(bubbles []*Bubble, N, M int, spacing float32) int {
	visited := make([]bool, len(bubbles))
	groupID := 0

	for i := 0; i < len(bubbles); i++ {
		if !visited[i] && bubbles[i].CurrentState {
			// Perform BFS to find all connected alive bubbles
			queue := []int{i}
			visited[i] = true
			bubbles[i].GroupID = groupID

			for len(queue) > 0 {
				curr := queue[0]
				queue = queue[1:]

				neighbors := FindNeighbors(bubbles, curr, N, M, spacing)
				for _, neighbor := range neighbors {
					if !visited[neighbor] {
						visited[neighbor] = true
						bubbles[neighbor].GroupID = groupID
						queue = append(queue, neighbor)
					}
				}
			}
			groupID++
		}
	}

	return groupID
}

// assignColorsToGroups assigns a unique color to each group of connected bubbles.
func assignColorsToGroups(bubbles []*Bubble, numGroups int) {
	groupBaseColors := make([]mgl32.Vec3, numGroups)

	// Generate base pastel colors for each group
	for g := 0; g < numGroups; g++ {
		groupBaseColors[g] = mgl32.Vec3{
			0.6 + rand.Float32()*0.3, // Random pastel-like color (R)
			0.6 + rand.Float32()*0.3, // Random pastel-like color (G)
			0.6 + rand.Float32()*0.3, // Random pastel-like color (B)
		}
	}

	// Assign each bubble in a group a slightly varied color based on the base color
	for _, bubble := range bubbles {
		if bubble.CurrentState {
			baseColor := groupBaseColors[bubble.GroupID]
			bubble.Color = mgl32.Vec3{
				baseColor.X(), // Slight variation in R
				baseColor.Y(), // Slight variation in G
				baseColor.Z(), // Slight variation in B
			}
		}
	}
}

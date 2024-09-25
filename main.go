package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Settings
const (
	windowWidth  = 800
	windowHeight = 600
	// bubbles
	N              = 10
	M              = 20
	spacing        = 1.5
	animationSpeed = 3.0
)

var (
	// Track time stats related to frame speed to account for different
	// computer performance
	deltaTime = 0.0 // time between current frame and last frame
	lastFrame = 0.0 // time of last frame
	// Last mouse positions, initially in the center of the window
	lastX = float64(windowWidth / 2)
	lastY = float64(windowHeight / 2)
	// Handle when mouse first enters window and has large offset to center
	firstMouse = true
	camera     *Camera
	// Track FPS timing
	frameCount        = 0
	lastFPSUpdateTime = 0.0
	fps               = 0.0
	white             = mgl32.Vec3{1.0, 1.0, 1.0}
	generation        int
)

func init() {

	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()

	camera = NewDefaultCameraAtPosition(mgl32.Vec3{0.0, 0.0, 5.0})
}

func initGL() *glfw.Window {
	//* GLFW init and configure
	err := glfw.Init()
	if err != nil {
		log.Fatal(err)
	}
	// Using hints, set various options for the window we're about to create.
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	// Compatibility profile allows more deprecated function calls over core profile.
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.Resizable, glfw.False)

	//* GLFW window creation
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "BubbleLife", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()
	//* Callbacks
	// Set the function that is run every time the viewport is resized by the user.
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	// Listen to mouse events
	window.SetKeyCallback(keyCallback)
	window.SetCursorPosCallback(mouseCallback)
	window.SetScrollCallback(scrollCallback)
	// Tell glfw to capture and hide the cursor
	// window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	//* Load OS-specific OpenGL function pointers
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	//* OpenGL configuration
	gl.Viewport(0, 0, windowWidth, windowHeight)

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.PROGRAM_POINT_SIZE)
	return window
}

func main() {
	window := initGL()

	// Load shaders
	shader, err := NewShader("shaders/shader.vs", "shaders/shader.fs", "")
	if err != nil {
		log.Fatalln("Failed to load shaders:", err)
	}

	// Create pillar of spheres (positions only)
	spheres := createPillarOfSpheres(N, M, spacing, 194741008) // 10x10 grid, 20 height, 1.5 units spacing

	// Init buffers for sphere positions
	InitInstanceBuffer(spheres)

	// Setup view/projection matrices
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), windowWidth/windowHeight, 0.1, 100.0)
	shader.use()
	shader.setMat4("projection", projection)

	text := NewTextRenderer(windowWidth, windowHeight)
	text.Load("fonts/ocraext.ttf", 24)

	var lastGoLUpdateTime float64 = 0.0
	var goLUpdateInterval float64 = 1.0
	for !window.ShouldClose() {
		// calculate time stats
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame

		// Calculate FPS every 1 second
		frameCount++
		if currentFrame-lastFPSUpdateTime >= 1.0 {
			fps = float64(frameCount) / (currentFrame - lastFPSUpdateTime)
			lastFPSUpdateTime = currentFrame
			frameCount = 0
		}

		// Handle user input.
		processInput(window)

		//* render
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		if currentFrame-lastGoLUpdateTime >= goLUpdateInterval {
			updateGameOfLife(spheres, N, M)
			lastGoLUpdateTime = currentFrame // Reset the last update time
			generation++
		}

		animateSphereRadius(spheres, deltaTime)
		UpdateRadiiBuffer(spheres)

		aliveCount := 0
		for _, sphere := range spheres {
			if sphere.CurrentState && sphere.Radius > 0.0 {
				aliveCount++
			}
		}

		text.RenderText(fmt.Sprintf("FPS:%0.2f", fps), 5.0, 5.0, 1.0, white)
		text.RenderText(fmt.Sprintf("Spheres: %d/%d", aliveCount, len(spheres)), 5.0, 30.0, 1.0, white)
		text.RenderText(fmt.Sprintf("Generation #: %d", generation), 5.0, 60.0, 1.0, white)

		// Use the shader program
		shader.use()
		view := camera.getViewMatrix()

		// Pass matrices to the shader
		shader.setMat4("view", view)

		renderSpheres(shader, len(spheres))

		window.SwapBuffers()
		glfw.PollEvents()
	}
	glfw.Terminate()
}

var neighborOffsets = [26][3]int{
	{-1, -1, -1}, {-1, -1, 0}, {-1, -1, 1}, {-1, 0, -1}, {-1, 0, 0}, {-1, 0, 1}, {-1, 1, -1}, {-1, 1, 0}, {-1, 1, 1},
	{0, -1, -1}, {0, -1, 0}, {0, -1, 1}, {0, 0, -1}, {0, 0, 1}, {0, 1, -1}, {0, 1, 0}, {0, 1, 1},
	{1, -1, -1}, {1, -1, 0}, {1, -1, 1}, {1, 0, -1}, {1, 0, 0}, {1, 0, 1}, {1, 1, -1}, {1, 1, 0}, {1, 1, 1},
}

func countAliveNeighbors(spheres []*Sphere, N, M int, x, y, z int) int {
	aliveNeighbors := 0

	for _, offset := range neighborOffsets {
		nx := x + offset[0]
		ny := y + offset[1]
		nz := z + offset[2]

		// Bounds check
		if nx >= 0 && nx < N && ny >= 0 && ny < M && nz >= 0 && nz < N {
			index := (nx * M * N) + (ny * N) + nz
			if spheres[index].CurrentState {
				aliveNeighbors++
			}
		}
	}

	return aliveNeighbors
}

func updateGameOfLife(spheres []*Sphere, N, M int) {
	// Apply Game of Life rules for 3D
	for x := 0; x < N; x++ {
		for y := 0; y < M; y++ {
			for z := 0; z < N; z++ {
				index := (x * M * N) + (y * N) + z
				sphere := spheres[index]

				aliveNeighbors := countAliveNeighbors(spheres, N, M, x, y, z)

				if sphere.CurrentState {
					// Apply 3D GoL rules for alive cells
					if aliveNeighbors < 4 || aliveNeighbors > 9 {
						sphere.NextState = false // Cell dies
					} else {
						sphere.NextState = true // Cell survives
					}
				} else {
					// Apply 3D GoL rules for dead cells
					if aliveNeighbors >= 5 && aliveNeighbors <= 7 {
						sphere.NextState = true // Cell is born
					}
				}
			}
		}
	}
}

// Function to animate radius changes
func animateSphereRadius(spheres []*Sphere, deltaTime float64) {
	for _, sphere := range spheres {
		// Check if there's a state change that needs to be animated
		if sphere.CurrentState != sphere.NextState {
			sphere.Animating = true
		}

		// Animate based on the state change
		if sphere.Animating {
			if sphere.NextState && sphere.Radius < 1.0 {
				// Growing animation
				sphere.Radius += float32(deltaTime * animationSpeed)
				if sphere.Radius >= 1.0 {
					sphere.Radius = 1.0
					sphere.CurrentState = true // Commit the new state
					sphere.Animating = false
				}
			} else if !sphere.NextState && sphere.Radius > 0.0 {
				// Shrinking animation
				sphere.Radius -= float32(deltaTime * animationSpeed)
				if sphere.Radius <= 0.0 {
					sphere.Radius = 0.0
					sphere.CurrentState = false // Commit the new state
					sphere.Animating = false
				}
			}
		}
	}
}

// createPillarOfSpheres generates an NxN grid of spheres stacked vertically into a pillar.
func createPillarOfSpheres(N, M int, spacing float32, seed int64) []*Sphere {
	spheres := make([]*Sphere, 0)

	rnd := rand.New(rand.NewSource(seed))

	// Iterate through the grid to create spheres at specific positions
	for x := 0; x < N; x++ {
		for y := 0; y < M; y++ {
			for z := 0; z < N; z++ {
				position := mgl32.Vec3{
					float32(x) * spacing,
					float32(y) * spacing,
					float32(z) * spacing,
				}

				// Create a new sphere with a default radius (not used in shaders, just kept for logical structure)
				sphere := NewSphere(position)

				if rnd.Float32() < 0.4 {
					sphere.CurrentState = true
					sphere.NextState = true
					sphere.Radius = 1.0
				}

				spheres = append(spheres, sphere)
			}
		}
	}
	return spheres
}

// framebufferSizeCallback is called when the gl viewport is resized.
func framebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}
func processInput(w *glfw.Window) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}

	if w.GetKey(glfw.KeyW) == glfw.Press {
		camera.processKeyboard(FORWARD, float32(deltaTime))
	}
	if w.GetKey(glfw.KeyS) == glfw.Press {
		camera.processKeyboard(BACKWARD, float32(deltaTime))
	}
	if w.GetKey(glfw.KeyA) == glfw.Press {
		camera.processKeyboard(LEFT, float32(deltaTime))
	}
	if w.GetKey(glfw.KeyD) == glfw.Press {
		camera.processKeyboard(RIGHT, float32(deltaTime))
	}

	if w.GetKey(glfw.KeyLeftShift) == glfw.Press {
		// Tell glfw to capture and hide the cursor
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	}
	if w.GetKey(glfw.KeyRightShift) == glfw.Press {
		// Tell glfw to show and stop capturing cursor
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
}

// keyCallback is called when the gl viewport is resized.
func keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}
}
func scrollCallback(w *glfw.Window, xOffset float64, yOffset float64) {
	camera.processMouseScroll(float32(yOffset))
}

// mouseCallback is called every time the mouse is moved. x, y are current positions of the mouse
func mouseCallback(w *glfw.Window, x float64, y float64) {
	if firstMouse {
		// prevent large visual jump
		lastX = x
		lastY = y
		firstMouse = false
	}
	// calculate mouse offset since last frame
	xOffset := x - lastX
	yOffset := lastY - y
	lastX = x
	lastY = y

	camera.processMouseMovement(float32(xOffset), float32(yOffset), true)
}

func checkGLError() {
	err := gl.GetError()
	if err != 0 {
		fmt.Printf("OpenGL error: %v\n", err)
	}
}

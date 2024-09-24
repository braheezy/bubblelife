package main

import (
	"fmt"
	"log"
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
	N       = 10
	M       = 20
	spacing = 1.5
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
	spheres := createPillarOfSpheres(10, 20, 1.5) // 10x10 grid, 20 height, 1.5 units spacing

	// Init buffers for sphere positions
	InitInstanceBuffer(spheres)

	// Setup view/projection matrices
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), windowWidth/windowHeight, 0.1, 100.0)
	shader.use()
	shader.setMat4("projection", projection)

	text := NewTextRenderer(windowWidth, windowHeight)
	text.Load("fonts/ocraext.ttf", 24)

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

		text.RenderText(fmt.Sprintf("FPS:%0.2f", fps), 5.0, 5.0, 1.0, white)
		text.RenderText(fmt.Sprintf("Spheres: %d", len(spheres)), 5.0, 30.0, 1.0, white)

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

// createPillarOfSpheres generates an NxN grid of spheres stacked vertically into a pillar.
func createPillarOfSpheres(N, M int, spacing float32) []*Sphere {
	spheres := make([]*Sphere, 0)

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
				sphere := NewSphere(position, 0.5)
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

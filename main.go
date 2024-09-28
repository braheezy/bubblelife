package main

import (
	"image"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"unsafe"

	_ "github.com/mdouchement/hdr/codec/rgbe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Settings
const (
	windowWidth   = 800
	windowHeight  = 600
	bubbleSpacing = 1.5
	// when a bubble pops, how fast that animation happens
	animationSpeed = 3.0
	// resolution to use for background
	resolution = int32(4096)
)

var (
	// Track time stats related to frame speed to account for different computer performance
	// time between current frame and last frame
	deltaTime = 0.0
	lastFrame = 0.0

	// Last mouse positions, initially in the center of the window
	lastX = float64(windowWidth / 2)
	lastY = float64(windowHeight / 2)

	// Handle when mouse first enters window and has large offset to center
	firstMouse = true

	// user-controlled fly camera
	camera *Camera

	// Track FPS timing
	frameCount        = 0
	lastFPSUpdateTime = 0.0
	fps               = 0.0

	// help track generation speed, it doesn't update every render frame
	lastGenerationTime float64

	// scene settings
	generation      int
	initialSeed     = int64(42)
	pillarN         = 10
	pillarM         = 20
	bubbles         []*Bubble
	generationSpeed = 5.0
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()
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
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	//* Load OS-specific OpenGL function pointers
	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	//* OpenGL configuration
	gl.Viewport(0, 0, windowWidth, windowHeight)

	gl.Enable(gl.DEPTH_TEST)

	// for text rendering
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// for points rendering
	gl.Enable(gl.PROGRAM_POINT_SIZE)

	// for cubemap
	gl.DepthFunc(gl.LEQUAL)
	gl.Enable(gl.TEXTURE_CUBE_MAP_SEAMLESS)

	return window
}

func main() {
	window := initGL()

	//* Load shaders
	shader, err := NewShader("shaders/shader.vs", "shaders/shader.fs", "")
	if err != nil {
		log.Fatalln("Failed to load shaders:", err)
	}
	backgroundShader, err := NewShader("shaders/background.vs", "shaders/background.fs", "")
	if err != nil {
		log.Fatalln("Failed to load shaders:", err)
	}
	equirectangularToCubemapShader, err := NewShader("shaders/cubemap.vs", "shaders/equirectangular.fs", "")
	if err != nil {
		log.Fatalln("Failed to load shaders:", err)
	}

	//* Load textures
	hdrTexture := loadHDRTexture("nebula.hdr")
	envCubemap := setupCubemap(hdrTexture, equirectangularToCubemapShader)
	// Create pillar of bubbles (positions only)
	bubbles = createPillarOfBubbles(pillarN, pillarM, bubbleSpacing, initialSeed)

	// Init buffers for bubble positions
	initInstanceBuffer(bubbles)

	// calculate sane starting camera position based on pillar size
	pillarWidth := (float32(pillarN) - 1) * bubbleSpacing
	pillarHeight := (float32(pillarM) - 1) * bubbleSpacing
	pillarDepth := (float32(pillarN) - 1) * bubbleSpacing
	maxDimension := pillarWidth
	if pillarHeight > maxDimension {
		maxDimension = pillarHeight
	}
	if pillarDepth > maxDimension {
		maxDimension = pillarDepth
	}
	// Field of view (in radians) and aspect ratio
	fov := mgl32.DegToRad(45.0) // Assuming the FOV is 45 degrees
	aspectRatio := float32(windowWidth) / float32(windowHeight)
	// Calculate the distance from the center of the pillar to the camera
	// Based on the formula: distance = (maxDimension / 2) / tan(fov / 2)
	distance := (maxDimension / 2) / float32(math.Tan(float64(fov)/2))
	if aspectRatio < 1.0 {
		// If the window is taller than wide, increase the distance to fit the height
		distance /= aspectRatio
	}
	cameraPos := mgl32.Vec3{pillarWidth / 2, pillarHeight / 2, distance / 2}
	camera = NewDefaultCameraAtPosition(cameraPos)

	// Setup view/projection matrices
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), windowWidth/windowHeight, 0.1, 100.0)
	shader.use()
	shader.setMat4("projection", projection)

	// lights
	shader.setVec3("lightDir", mgl32.Vec3{1.0, 1.0, 1.0})
	shader.setVec3("lightColor", mgl32.Vec3{0.8, 0.8, 0.8})
	shader.setVec3("ambientLight", iris)

	// Set up bubble effect uniforms
	shader.setFloat("bubbleThickness", 0.08)
	shader.setFloat("fresnelStrength", 0.2)
	shader.setFloat("transparency", 0.8)

	// set skybox from background texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)
	shader.setInt("skybox", 0)

	backgroundShader.use()
	backgroundShader.setInt("environmentMap", 0)
	backgroundShader.setMat4("projection", projection)

	textRenderer := NewTextRenderer(windowWidth, windowHeight)
	textRenderer.Load("fonts/ocraext.ttf", 24)

	//* render loop
	for !window.ShouldClose() {
		// calculate time stats
		currentFrame := glfw.GetTime()
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame
		glfw.PollEvents()

		// Calculate FPS
		frameCount++
		if currentFrame-lastFPSUpdateTime >= 1.0 {
			fps = float64(frameCount) / (currentFrame - lastFPSUpdateTime)
			lastFPSUpdateTime = currentFrame
			frameCount = 0
		}

		processInput(window)

		// update generation if enough time has passed
		if currentFrame-lastGenerationTime >= generationSpeed {
			updateGameOfLife(bubbles, pillarN, pillarM)
			lastGenerationTime = currentFrame
			generation++

			// The goal is to find populations of bubbles and give them the same color. It's not great but
			// results in a visually pleasing effect
			numGroups := findGroups(bubbles, pillarN, pillarM, bubbleSpacing)
			assignColorsToGroups(bubbles, numGroups)
			updateColorBuffer(bubbles)
		}

		// "alive" and "dead" is transitioned by growing/shrinking the radius of the bubble. that animation can happen
		// over multiple frames, so update that here.
		animateBubbleRadius(bubbles, deltaTime)
		updateRadiiBuffer(bubbles)
		aliveCount := 0
		for _, bubble := range bubbles {
			if bubble.CurrentState && bubble.Radius > 0.0 {
				aliveCount++
			}
		}

		//* render
		gl.ClearColor(0.0, 0.0, 0.0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		shader.use()
		view := camera.getViewMatrix()
		shader.setMat4("view", view)
		shader.setVec3("viewPos", camera.position)

		renderBubbles(shader, len(bubbles))

		backgroundShader.use()
		backgroundShader.setMat4("view", mgl32.Mat4(view).Mat3().Mat4())
		// Bind the cubemap texture
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)

		// Render the cubemap (as the background)
		renderCube()

		if showUI {
			// draw all UI elements
			renderUI(textRenderer, bubbles, fps, aliveCount, generation)
		}

		window.SwapBuffers()
	}
	glfw.Terminate()
}

func setupCubemap(textureID uint32, equirectangularToCubemapShader *Shader) uint32 {
	var envCubemap uint32
	gl.GenTextures(1, &envCubemap)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, envCubemap)

	for i := 0; i < 6; i++ {
		gl.TexImage2D(uint32(gl.TEXTURE_CUBE_MAP_POSITIVE_X+i), 0, gl.RGB16, resolution, resolution, 0, gl.RGB, gl.FLOAT, nil)
	}

	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	var captureFBO uint32
	gl.GenFramebuffers(1, &captureFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)
	// Bind and convert HDRI to cubemap using a shader (similar to previous code snippets)
	gl.BindFramebuffer(gl.FRAMEBUFFER, captureFBO)

	captureProjection := mgl32.Perspective(mgl32.DegToRad(90.0), 1.0, 0.1, 10.0)
	captureViews := []mgl32.Mat4{
		mgl32.LookAtV(mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{1.0, 0.0, 0.0}, mgl32.Vec3{0.0, -1.0, 0.0}),
		mgl32.LookAtV(mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{-1.0, 0.0, 0.0}, mgl32.Vec3{0.0, -1.0, 0.0}),
		mgl32.LookAtV(mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{0.0, 1.0, 0.0}, mgl32.Vec3{0.0, 0.0, 1.0}),
		mgl32.LookAtV(mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{0.0, -1.0, 0.0}, mgl32.Vec3{0.0, 0.0, -1.0}),
		mgl32.LookAtV(mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{0.0, 0.0, 1.0}, mgl32.Vec3{0.0, -1.0, 0.0}),
		mgl32.LookAtV(mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{0.0, 0.0, -1.0}, mgl32.Vec3{0.0, -1.0, 0.0}),
	}
	equirectangularToCubemapShader.use()
	equirectangularToCubemapShader.setInt("equirectangularMap", 0)
	equirectangularToCubemapShader.setMat4("projection", captureProjection)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	gl.Viewport(0, 0, resolution, resolution)
	for i := 0; i < 6; i++ {
		equirectangularToCubemapShader.setMat4("view", captureViews[i])
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, uint32(gl.TEXTURE_CUBE_MAP_POSITIVE_X+i), envCubemap, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		renderCube()
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// restore viewport
	gl.Viewport(0, 0, int32(windowWidth), int32(windowHeight))

	return envCubemap
}

func loadHDRTexture(path string) uint32 {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	hdrImg, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := hdrImg.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	pixelData := make([]float32, 0, width*height*3) // 3 for RGB channels
	for y := bounds.Max.Y; y > bounds.Min.Y; y-- {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := hdrImg.At(x, y).RGBA() // HDR images do not use alpha, ignore it
			// Convert from uint32 range (0-65535) to float32 range (0.0-1.0)
			pixelData = append(pixelData, float32(r)/65535.0, float32(g)/65535.0, float32(b)/65535.0)
		}
	}

	var hdrTexture uint32
	gl.GenTextures(1, &hdrTexture)
	gl.BindTexture(gl.TEXTURE_2D, hdrTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB16, int32(width), int32(height), 0, gl.RGB, gl.FLOAT, gl.Ptr(pixelData))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	return hdrTexture
}

// renderCube() renders a 1x1 3D cube in NDC.
var cubeVAO, cubeVBO uint32

func renderCube() {
	if cubeVAO == 0 {
		// initialize
		vertices := []float32{
			// back face
			-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
			1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
			1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 0.0, // bottom-right
			1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 1.0, 1.0, // top-right
			-1.0, -1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 0.0, // bottom-left
			-1.0, 1.0, -1.0, 0.0, 0.0, -1.0, 0.0, 1.0, // top-left
			// front face
			-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
			1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 0.0, // bottom-right
			1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
			1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 1.0, 1.0, // top-right
			-1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 1.0, // top-left
			-1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, // bottom-left
			// left face
			-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
			-1.0, 1.0, -1.0, -1.0, 0.0, 0.0, 1.0, 1.0, // top-left
			-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
			-1.0, -1.0, -1.0, -1.0, 0.0, 0.0, 0.0, 1.0, // bottom-left
			-1.0, -1.0, 1.0, -1.0, 0.0, 0.0, 0.0, 0.0, // bottom-right
			-1.0, 1.0, 1.0, -1.0, 0.0, 0.0, 1.0, 0.0, // top-right
			// right face
			1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
			1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
			1.0, 1.0, -1.0, 1.0, 0.0, 0.0, 1.0, 1.0, // top-right
			1.0, -1.0, -1.0, 1.0, 0.0, 0.0, 0.0, 1.0, // bottom-right
			1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 1.0, 0.0, // top-left
			1.0, -1.0, 1.0, 1.0, 0.0, 0.0, 0.0, 0.0, // bottom-left
			// bottom face
			-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
			1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 1.0, 1.0, // top-left
			1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
			1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 1.0, 0.0, // bottom-left
			-1.0, -1.0, 1.0, 0.0, -1.0, 0.0, 0.0, 0.0, // bottom-right
			-1.0, -1.0, -1.0, 0.0, -1.0, 0.0, 0.0, 1.0, // top-right
			// top face
			-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
			1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
			1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 1.0, 1.0, // top-right
			1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, // bottom-right
			-1.0, 1.0, -1.0, 0.0, 1.0, 0.0, 0.0, 1.0, // top-left
			-1.0, 1.0, 1.0, 0.0, 1.0, 0.0, 0.0, 0.0, // bottom-left
		}
		gl.GenVertexArrays(1, &cubeVAO)
		gl.GenBuffers(1, &cubeVBO)
		// fill buffer
		gl.BindBuffer(gl.ARRAY_BUFFER, cubeVBO)
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(vertices[0])), gl.Ptr(vertices), gl.STATIC_DRAW)
		// link vertex attributes
		gl.BindVertexArray(cubeVAO)
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 8*int32(unsafe.Sizeof(float32(0))), gl.Ptr(nil))
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(8*unsafe.Sizeof(float32(0))), gl.Ptr(3*unsafe.Sizeof(float32(0))))
		gl.EnableVertexAttribArray(2)
		gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(8*unsafe.Sizeof(float32(0))), gl.Ptr(6*unsafe.Sizeof(float32(0))))
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	}
	// render cube
	gl.BindVertexArray(cubeVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 36)
	gl.BindVertexArray(0)
}

func countAliveNeighbors(bubbles []*Bubble, N, M int, x, y, z int) int {
	aliveNeighbors := 0

	// Iterate through all possible neighbor coordinates (-1, 0, 1) for x, y, z
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			for dz := -1; dz <= 1; dz++ {
				// Skip the bubble itself (dx, dy, dz all zero)
				if dx == 0 && dy == 0 && dz == 0 {
					continue
				}

				// Calculate neighbor coordinates with wrapping
				nx := (x + dx + N) % N // Wrap around for x-axis
				ny := (y + dy + M) % M // Wrap around for y-axis
				nz := (z + dz + N) % N // Wrap around for z-axis

				// Calculate the index of the neighbor
				neighborIndex := (nx * M * N) + (ny * N) + nz

				// Count alive neighbors
				if bubbles[neighborIndex].CurrentState {
					aliveNeighbors++
				}
			}
		}
	}

	return aliveNeighbors
}

func updateGameOfLife(bubbles []*Bubble, N, M int) {
	// Apply Game of Life rules for 3D
	for x := 0; x < N; x++ {
		for y := 0; y < M; y++ {
			for z := 0; z < N; z++ {
				index := (x * M * N) + (y * N) + z
				bubble := bubbles[index]

				aliveNeighbors := countAliveNeighbors(bubbles, N, M, x, y, z)

				if bubble.CurrentState {
					// Apply 3D GoL rules for alive cells
					if aliveNeighbors < 4 || aliveNeighbors > 9 {
						bubble.NextState = false // Cell dies
					} else {
						bubble.NextState = true // Cell survives
					}
				} else {
					// Apply 3D GoL rules for dead cells
					if aliveNeighbors >= 5 && aliveNeighbors <= 7 {
						bubble.NextState = true // Cell is born
					}
				}
			}
		}
	}
}

// Function to animate radius changes
func animateBubbleRadius(bubbles []*Bubble, deltaTime float64) {
	for _, bubble := range bubbles {
		// Check if there's a state change that needs to be animated
		if bubble.CurrentState != bubble.NextState {
			bubble.Animating = true
		}

		// Animate based on the state change
		if bubble.Animating {
			if bubble.NextState && bubble.Radius < 1.0 {
				// Growing animation
				bubble.Radius += float32(deltaTime * animationSpeed)
				if bubble.Radius >= 1.0 {
					bubble.Radius = 1.0
					bubble.CurrentState = true // Commit the new state
					bubble.Animating = false
				}
			} else if !bubble.NextState && bubble.Radius > 0.0 {
				// Shrinking animation
				bubble.Radius -= float32(deltaTime * animationSpeed)
				if bubble.Radius <= 0.0 {
					bubble.Radius = 0.0
					bubble.CurrentState = false // Commit the new state
					bubble.Animating = false
				}
			}
		}
	}
}

// createPillarOfBubbles generates an NxN grid of bubbles stacked vertically into a pillar.
func createPillarOfBubbles(N, M int, spacing float32, seed int64) []*Bubble {
	bubbles := make([]*Bubble, 0)

	rnd := rand.New(rand.NewSource(seed))

	// Iterate through the grid to create bubbles at specific positions
	for x := 0; x < N; x++ {
		for y := 0; y < M; y++ {
			for z := 0; z < N; z++ {
				position := mgl32.Vec3{
					float32(x) * spacing,
					float32(y) * spacing,
					float32(z) * spacing,
				}

				// Create a new bubble with a default radius (not used in shaders, just kept for logical structure)
				bubble := NewBubble(position)

				if rnd.Float32() < 0.4 {
					bubble.CurrentState = true
					bubble.NextState = true
					bubble.Radius = 1.0
				}

				bubbles = append(bubbles, bubble)
			}
		}
	}

	numGroups := findGroups(bubbles, N, M, spacing)

	// 2. Assign colors to each group of bubbles
	assignColorsToGroups(bubbles, numGroups)
	return bubbles
}

// framebufferSizeCallback is called when the gl viewport is resized.
func framebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
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

func recreatePillar(N, M int) {
	bubbles = createPillarOfBubbles(N, M, bubbleSpacing, uiSeed)
	numGroups := findGroups(bubbles, N, M, bubbleSpacing)
	assignColorsToGroups(bubbles, numGroups)
	initInstanceBuffer(bubbles)
	pillarM = M
	pillarN = N
}

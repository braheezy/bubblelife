package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	menuY   = float32(100.0)
	spacing = float32(30.0)
)

// Variables to store UI state
var (
	// Toggles whether the UI is shown or not
	showUI = false
	// Set default UI values to scene values
	uiN, uiM          = pillarN, pillarM
	uiSeed            = initialSeed
	uiGenerationSpeed = generationSpeed

	// styling + colors
	white     = mgl32.Vec3{1.0, 1.0, 1.0}
	love      = mgl32.Vec3{235.0 / 255.0, 111.0 / 255.0, 146.0 / 255.0}
	iris      = mgl32.Vec3{196.0 / 255.0, 167.0 / 255.0, 231.0 / 255.0}
	textColor = white
	// Highlight the selected option
	highlightColor = love

	// debounce keys
	tabPressed       = false
	downPressed      = false
	upPressed        = false
	leftPressed      = false
	rightPressed     = false
	numberKeyPressed = make(map[glfw.Key]bool)
	backspacePressed bool
	enterPressed     bool
	shiftPressed     bool

	// Buffer to store typed input for the seed
	inputBuffer string
	// currently selected UI element
	selectedOption = 0
)

// renderUI renders the simple overlay menu when the user presses Tab.
func renderUI(text *TextRenderer, bubble []*Bubble, fps float64, aliveCount int, generation int) {
	// scene stats
	text.RenderText(fmt.Sprintf("FPS: %.2f", fps), 5.0, 5.0, 1.0, textColor)
	text.RenderText(fmt.Sprintf("bubbles: %d/%d", aliveCount, len(bubble)), 5.0, 30.0, 1.0, textColor)
	text.RenderText(fmt.Sprintf("generation #: %d", generation), 5.0, 60.0, 1.0, textColor)

	// Display the menu title
	text.RenderText("settings (tab to toggle, arrow keys to navigate)", 5.0, menuY, 1.0, textColor)

	// Pillar Size - N
	if selectedOption == 0 {
		text.RenderText(fmt.Sprintf("pillar width: %d", uiN), 5.0, menuY+spacing, 1.2, highlightColor)
	} else {
		text.RenderText(fmt.Sprintf("pillar width: %d", uiN), 5.0, menuY+spacing, 1.0, textColor)
	}

	// Pillar Size - M
	if selectedOption == 1 {
		text.RenderText(fmt.Sprintf("pillar height: %d", uiM), 5.0, menuY+2*spacing, 1.2, highlightColor)
	} else {
		text.RenderText(fmt.Sprintf("pillar heigh: %d", uiM), 5.0, menuY+2*spacing, 1.0, textColor)
	}

	// Seed
	if selectedOption == 2 {
		if len(inputBuffer) > 0 {
			text.RenderText(fmt.Sprintf("seed: %s", inputBuffer), 5.0, menuY+3*spacing, 1.2, highlightColor)
		} else {
			text.RenderText(fmt.Sprintf("seed: %d", uiSeed), 5.0, menuY+3*spacing, 1.2, highlightColor)
		}
	} else {
		text.RenderText(fmt.Sprintf("seed: %d", uiSeed), 5.0, menuY+3*spacing, 1.0, textColor)
	}

	// Generation Speed
	if selectedOption == 3 {
		text.RenderText(fmt.Sprintf("generation rate: %.2f sec", uiGenerationSpeed), 5.0, menuY+4*spacing, 1.2, highlightColor)
	} else {
		text.RenderText(fmt.Sprintf("generation rate: %.2f sec", uiGenerationSpeed), 5.0, menuY+4*spacing, 1.0, textColor)
	}

}

// Helper function to validate and clamp the seed value between 1 and int64
func validateAndClampSeed(input string) int64 {
	seed, err := strconv.Atoi(input)
	if err != nil || seed < 1 {
		return 1
	}
	return int64(math.Min(float64(seed), float64(math.MaxInt64)))
}

func processInput(w *glfw.Window) {
	//* quit on escape
	if w.GetKey(glfw.KeyEscape) == glfw.Press {
		w.SetShouldClose(true)
	}

	//* Toggle menu on tab
	if w.GetKey(glfw.KeyTab) == glfw.Press && !tabPressed {
		tabPressed = true
		showUI = !showUI
	}
	if w.GetKey(glfw.KeyTab) == glfw.Release {
		tabPressed = false
	}

	if showUI {
		// Track whether we need to recreate the pillar
		var pillarChanged bool = false

		//* Navigate UI items using Up/Down
		if w.GetKey(glfw.KeyDown) == glfw.Press && !downPressed {
			// Move down (wrap around)
			selectedOption = (selectedOption + 1) % 4
			downPressed = true
		}
		if w.GetKey(glfw.KeyDown) == glfw.Release {
			downPressed = false
		}

		if w.GetKey(glfw.KeyUp) == glfw.Press && !upPressed {
			// Move up (wrap around)
			selectedOption = (selectedOption - 1 + 4) % 4
			upPressed = true
		}
		if w.GetKey(glfw.KeyUp) == glfw.Release {
			upPressed = false
		}

		//* Handle value changes with left/right arrow keys for selected option
		if selectedOption == 0 { //* Pillar Size N
			if w.GetKey(glfw.KeyLeft) == glfw.Press && !leftPressed {
				// Decrease N, but not below 2
				uiN = max(2, uiN-1)
				leftPressed = true
				pillarChanged = true
			}
			if w.GetKey(glfw.KeyRight) == glfw.Press && !rightPressed {
				uiN++
				rightPressed = true
				pillarChanged = true
			}
		} else if selectedOption == 1 { //* Pillar Size M
			if w.GetKey(glfw.KeyLeft) == glfw.Press && !leftPressed {
				// Decrease M, but not below 2
				uiM = max(2, uiM-1)
				leftPressed = true
				pillarChanged = true
			}
			if w.GetKey(glfw.KeyRight) == glfw.Press && !rightPressed {
				uiM++
				rightPressed = true
				pillarChanged = true
			}
		} else if selectedOption == 2 { //* Seed input
			// Handle numerical input for the seed
			for key := glfw.Key0; key <= glfw.Key9; key++ {
				// Check if the key is pressed and hasn't been handled yet
				if w.GetKey(key) == glfw.Press && !numberKeyPressed[key] {
					inputBuffer += string(rune('0' + key - glfw.Key0))
					numberKeyPressed[key] = true
				}

				// Reset the key state when it is released
				if w.GetKey(key) == glfw.Release {
					numberKeyPressed[key] = false
				}
			}

			// Backspace key to delete last digit (with debounce)
			if w.GetKey(glfw.KeyBackspace) == glfw.Press && !backspacePressed && len(inputBuffer) > 0 {
				inputBuffer = inputBuffer[:len(inputBuffer)-1]
				backspacePressed = true
			}
			if w.GetKey(glfw.KeyBackspace) == glfw.Release {
				backspacePressed = false
			}

			// Enter key to confirm the seed input (with debounce)
			if w.GetKey(glfw.KeyEnter) == glfw.Press && !enterPressed && len(inputBuffer) > 0 {
				uiSeed = validateAndClampSeed(inputBuffer)
				inputBuffer = ""
				pillarChanged = true
				enterPressed = true
			}
			if w.GetKey(glfw.KeyEnter) == glfw.Release {
				enterPressed = false
			}
		} else if selectedOption == 3 { //* Generation speed
			if w.GetKey(glfw.KeyLeft) == glfw.Press && !leftPressed {
				uiGenerationSpeed = max(0, uiGenerationSpeed-1)
				leftPressed = true
			}
			if w.GetKey(glfw.KeyRight) == glfw.Press && !rightPressed {
				uiGenerationSpeed++
				rightPressed = true
			}
			generationSpeed = uiGenerationSpeed
		}

		// Release left/right key press flags
		if w.GetKey(glfw.KeyLeft) == glfw.Release {
			leftPressed = false
		}
		if w.GetKey(glfw.KeyRight) == glfw.Release {
			rightPressed = false
		}

		// If the pillar size changed or the seed was updated, recreate the pillar
		if pillarChanged {
			recreatePillar(uiN, uiM)
		}
	}

	// Handle camera movement when UI is not being shown
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

	// Allow escaping window
	if w.GetKey(glfw.KeyLeftShift) == glfw.Press && !shiftPressed {
		shiftPressed = true
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
	if w.GetKey(glfw.KeyLeftShift) == glfw.Release {
		shiftPressed = false
	}

}

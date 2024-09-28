package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Variables to store UI state
var (
	showUI = false // Toggles whether the UI is shown or not
	// mouseLocked       = true             // Toggles whether the mouse is locked or unlocked
	uiN, uiM          = pillarN, pillarM // Default N and M values (can be updated via UI)
	uiSeed            = initialSeed      // Default seed value
	uiGenerationSpeed = generationSpeed  // Default generation speed (seconds)
)

// // Toggle the UI visibility and mouse lock state
// func ToggleUI(w *glfw.Window) {
// 	showUI = !showUI
// 	if showUI {
// 		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal) // Unlock mouse
// 		mouseLocked = false
// 	} else {
// 		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled) // Lock mouse
// 		mouseLocked = true
// 	}
// }

var selectedOption = 0

// RenderUI renders the simple overlay menu when the user presses Tab.
func RenderUI(window *glfw.Window, text *TextRenderer, spheres []*Sphere, fps float64, aliveCount int, generation int) {
	// Always render the status text (FPS, sphere count, generation number)
	text.RenderText(fmt.Sprintf("FPS: %.2f", fps), 5.0, 5.0, 1.0, white)
	text.RenderText(fmt.Sprintf("Spheres: %d/%d", aliveCount, len(spheres)), 5.0, 30.0, 1.0, white)
	text.RenderText(fmt.Sprintf("Generation #: %d", generation), 5.0, 60.0, 1.0, white)

	// Render the UI if it's toggled on
	if showUI {
		menuY := float32(100.0)
		spacing := float32(30.0)

		// Display the menu title
		text.RenderText("UI Menu (Press Tab to Toggle)", 5.0, menuY, 1.0, white)

		// Highlight the selected option
		highlightColor := mgl32.Vec3{0.2, 1.0, 0.2} // Greenish highlight color

		// Render the options (highlight if selected)

		// Pillar Size - N
		if selectedOption == 0 {
			text.RenderText(fmt.Sprintf("Pillar Size - N: %d", uiN), 5.0, menuY+spacing, 1.2, highlightColor) // Highlighted
		} else {
			text.RenderText(fmt.Sprintf("Pillar Size - N: %d", uiN), 5.0, menuY+spacing, 1.0, white) // Regular
		}

		// Pillar Size - M
		if selectedOption == 1 {
			text.RenderText(fmt.Sprintf("Pillar Size - M: %d", uiM), 5.0, menuY+2*spacing, 1.2, highlightColor) // Highlighted
		} else {
			text.RenderText(fmt.Sprintf("Pillar Size - M: %d", uiM), 5.0, menuY+2*spacing, 1.0, white) // Regular
		}

		// Seed
		if selectedOption == 2 {
			if len(inputBuffer) > 0 {
				text.RenderText(fmt.Sprintf("Seed: %s", inputBuffer), 5.0, menuY+3*spacing, 1.2, highlightColor) // Show typed input
			} else {
				text.RenderText(fmt.Sprintf("Seed: %d", uiSeed), 5.0, menuY+3*spacing, 1.2, highlightColor) // Show current seed
			}
		} else {
			text.RenderText(fmt.Sprintf("Seed: %d", uiSeed), 5.0, menuY+3*spacing, 1.0, white) // Regular
		}

		// Generation Speed
		if selectedOption == 3 {
			text.RenderText(fmt.Sprintf("Generation Speed: %.2f sec", uiGenerationSpeed), 5.0, menuY+4*spacing, 1.2, highlightColor) // Highlighted
		} else {
			text.RenderText(fmt.Sprintf("Generation Speed: %.2f sec", uiGenerationSpeed), 5.0, menuY+4*spacing, 1.0, white) // Regular
		}
	}
}

// Helper function to validate and clamp the seed value between 1 and int32
func validateAndClampSeed(input string) int64 {
	// Convert the string to an int32 (safely)
	seed, err := strconv.Atoi(input)
	if err != nil || seed < 1 {
		return 1
	}
	return int64(math.Min(float64(seed), float64(math.MaxInt64)))
}

# bubblelife
A 3D interpretation of [Conway's Game of Life](https://www.wikiwand.com/en/articles/Conway%27s_Game_of_Life).

Made with bubbles! :speech_balloon:

https://github.com/user-attachments/assets/aafd42b5-59e9-416c-baa7-662810e6ce0f.mp4

## Usage
You can use Go or a container engine to run this project.

**Container**

Run it:

```bash

# Assuming X11 compatible host
docker run --rm -v /tmp/.X11-unix:/tmp/.X11-unix:rw -e DISPLAY=$DISPLAY ghcr.io/braheezy/bubblelife:v0.1.0
```
**Go**

You need the dependencies that [go-gl](https://github.com/go-gl/gl) has. That's a C compiler, and on Debian flavors, the `libgl1-mesa-dev` package.

Then:

```bash
go install github.com/braheezy/bubblelife@v0.1.0
bubblelife
```

## Keybindings

|Action|	Keybinding|	Description|
|--- |--- |---|
|Quit	|Esc|	Closes the window/application.
|Toggle UI menu	|Tab|	Toggles the visibility of the UI menu.
|Navigate UI (down)	|Down Arrow	|Moves down through UI options.
|Navigate UI (up)	|Up Arrow	|Moves up through UI options.
|Adjust Pillar width	|Left/Right Arrow (when option 0)|	Adjusts the pillar size N (min 2).
|Adjust Pillar height	|Left/Right Arrow (when option 1)|	Adjusts the pillar size M (min 2).
|Enter seed	|0-9 (when option 2)|	Enters numerical input for the seed.
|Delete last seed digit	|Backspace|	Removes the last digit entered for the seed.
|Confirm seed	|Enter|	Confirms the seed input.
|Adjust Generation Speed	|Left/Right Arrow (when option 3)|	Adjusts the generation speed.
|Camera movement (forward)	|W|	Moves the camera forward.
|Camera movement (backward)	|S|	Moves the camera backward.
|Camera movement (left)	|A|	Moves the camera to the left.
|Camera movement (right)	|D|	Moves the camera to the right.
|Unlock cursor	|Left Shift|	Unlocks the cursor and allows free cursor movement.

## Design
After completing the LearnOpenGL tutorial series, I wanted my own project to use some of the skills I learned. For graphics concepts, I use:

- a cubemap to create the background from. The cubemap is computed and created at runtime from an HDRi image.
- instanced rendering of points for good performance. Several other attempts before this tried to render spheres with vertices and there was poor FPS when viewing 2000 bubbles
- text rendering! which means opentype font parsing. always tricky
- hand-crafted UI system with input handling
- Blinn-Phong shading. And the entire sphere is "faked" in the fragment shader

The size of the pillar is chosen and the seed value is used to set the initial alive/dead population state. The Game is then set into motion. To extend to 3D, I check more neighbors than the original rules used for 2D. The Game algorithm does wrapped boundary checking, treating the pillar as a torus essentially.

For fun, related neighbors are given the same color. This is done by selecting cells, doing a BFS search to find neighbors, and labelling that group with an ID to enforce a different random color later. This works out quite nice.

## Credits
I got the HDRi file for the background [from here](https://www.artstation.com/marketplace/p/6Koj/nebula-hdri).

The font is freely available [online](https://www.fontsupply.com/fonts/O/Ocraext.html).

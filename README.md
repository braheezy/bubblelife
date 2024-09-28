# bubblelife
A 3D interpretation of Conway's Game of Life.

Made with bubbles! :speech_balloon:

![demo](./demo.mp4)

## Usage
You can use Go or a container engine to run this project.

**Container**

Clone the project, then build a container and run it:

```bash
    git clone https://github.com/braheezy/bubblelife.git
    cd bubblelife
    docker build . -t localhost/bubblelife:latest
    # Assuming X11 compatible host
    docker run --rm -v /tmp/.X11-unix:/tmp/.X11-unix:rw -e DISPLAY=$DISPLAY localhost/bubblelife:latest bubblelife
```
**Go**

You need the dependencies that [go-gl](https://github.com/go-gl/gl) has. That's a C compiler, and on Debian flavors, the `libgl1-mesa-dev` package.

Then:

```bash
    go install github.com/braheezy/bubblelife
    bubblelife
```

## Credits
I got the HDRi file for the background [from here](https://www.artstation.com/marketplace/p/6Koj/nebula-hdri).

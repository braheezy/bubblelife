FROM mcr.microsoft.com/devcontainers/go:1

RUN apt update
RUN apt install -y libgl1-mesa-dev xorg-dev glslang-tools

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o /usr/bin/bubblelife

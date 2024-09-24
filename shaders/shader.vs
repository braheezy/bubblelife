#version 410 core

layout(location = 0) in vec3 instancePosition;  // Position (center) of each sphere

uniform mat4 projection;
uniform mat4 view;

void main() {
    // Transform the sphere center into clip space
    gl_Position = projection * view * vec4(instancePosition, 1.0);

    // Set the point size based on the distance from the camera to the sphere
    float distance = length(vec3(view * vec4(instancePosition, 1.0)));
    gl_PointSize = 500.0 / distance;  // Adjust the factor to scale point size as needed
}

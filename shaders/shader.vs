#version 410 core

layout(location = 0) in vec3 instancePosition;  // Sphere position (center)
layout(location = 1) in float instanceRadius;   // Sphere radius (per-instance)
layout(location = 2) in vec3 instanceColor;     // Sphere color (per-instance)

// Output to the fragment shader
out vec3 fragPosition;  // Pass the fragment's world position
out float radius;       // Pass the radius to the fragment shader
out vec3 fragColor;     // Pass the color to the fragment shader

uniform mat4 projection;
uniform mat4 view;

void main() {
    // Transform the sphere center into clip space
    gl_Position = projection * view * vec4(instancePosition, 1.0);

    // Set the point size based on the distance from the camera to the sphere
    float distance = length(vec3(view * vec4(instancePosition, 1.0)));
    gl_PointSize = clamp(500.0 / distance, 10.0, 100.0);  // Adjust point size for distance

    // Pass the radius, position, and color to the fragment shader
    radius = instanceRadius;
    fragPosition = instancePosition;
    fragColor = instanceColor;  // Pass per-instance color to the fragment shader
}

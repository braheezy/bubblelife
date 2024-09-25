#version 410 core

layout(location = 0) in vec3 instancePosition;  // Sphere position (center)
layout(location = 1) in float instanceRadius;   // Sphere radius (per-instance)

// Output to the geometry or fragment shader
out float radius;  // Pass the radius to the fragment shader

uniform mat4 projection;
uniform mat4 view;

void main() {
    // Transform the sphere center into clip space
    gl_Position = projection * view * vec4(instancePosition, 1.0);

    // Set the point size based on the distance from the camera to the sphere
    float distance = length(vec3(view * vec4(instancePosition, 1.0)));

    // Adjust point size (you can tune this formula as needed)
    gl_PointSize = clamp(500.0 / distance, 10.0, 100.0);

    // Pass the radius to the fragment shader
    radius = instanceRadius;
}

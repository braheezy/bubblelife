#version 410 core

// bubble position (center)
layout(location = 0) in vec3 instancePosition;
 // bubble radius (per-instance)
layout(location = 1) in float instanceRadius;
  // bubble color (per-instance)
layout(location = 2) in vec3 instanceColor;

// Output to the fragment shader
// Pass the fragment's world position
out vec3 fragPosition;
// Pass the radius to the fragment shader
out float radius;
// Pass the color to the fragment shader
out vec3 fragColor;

uniform mat4 projection;
uniform mat4 view;

void main() {
    // Transform the bubble center into clip space
    gl_Position = projection * view * vec4(instancePosition, 1.0);

    // Set the point size based on the distance from the camera to the bubble
    float distance = length(vec3(view * vec4(instancePosition, 1.0)));
    // Adjust point size for distance
    gl_PointSize = clamp(500.0 / distance, 10.0, 100.0);

    // Pass the radius, position, and color to the fragment shader
    radius = instanceRadius;
    fragPosition = instancePosition;
    // Pass per-instance color to the fragment shader
    fragColor = instanceColor;
}

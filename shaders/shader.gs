#version 410 core

layout(points) in;               // Input: single point for the center of the sphere
layout(triangle_strip, max_vertices = 100) out;  // Output: a set of triangles for the sphere

uniform mat4 projection;
uniform mat4 view;

const float radius = 0.5;  // Sphere radius
const int stacks = 10;     // Number of stacks (latitude divisions)
const int slices = 10;     // Number of slices (longitude divisions)

void main() {
    vec3 center = gl_in[0].gl_Position.xyz;  // Sphere center from vertex shader

    // Loop through spherical coordinates and generate vertices
    for (int i = 0; i <= stacks; ++i) {
        float phi1 = float(i) / float(stacks) * 3.14159;     // Latitude: 0 to PI
        float phi2 = float(i + 1) / float(stacks) * 3.14159; // Next latitude

        for (int j = 0; j <= slices; ++j) {
            float theta = float(j) / float(slices) * 6.28318;  // Longitude: 0 to 2*PI

            // First vertex at current latitude
            vec3 vertex1 = center + radius * vec3(
                sin(phi1) * cos(theta),  // X
                cos(phi1),               // Y
                sin(phi1) * sin(theta)   // Z
            );

            // Second vertex at next latitude
            vec3 vertex2 = center + radius * vec3(
                sin(phi2) * cos(theta),  // X
                cos(phi2),               // Y
                sin(phi2) * sin(theta)   // Z
            );

            // Project and emit first vertex (current latitude)
            gl_Position = projection * view * vec4(vertex1, 1.0);
            EmitVertex();

            // Project and emit second vertex (next latitude)
            gl_Position = projection * view * vec4(vertex2, 1.0);
            EmitVertex();
        }

        EndPrimitive();  // Finish the strip for this stack
    }
}

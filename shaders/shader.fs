#version 410 core

in float radius;  // Radius passed from the vertex shader

out vec4 FragColor;

uniform vec3 sphereColor = vec3(0.784, 0.635, 0.784);  // Lilac color
uniform vec3 lightDir = vec3(0.0, 0.0, 1.0);  // Light coming from the camera direction

void main() {
    // Calculate the fragment's position within the point sprite (normalized coordinates from -1 to 1)
    vec2 coord = 2.0 * gl_PointCoord - vec2(1.0);  // gl_PointCoord is in range [0, 1], map it to [-1, 1]

    // Compute the distance from the center of the point sprite to this fragment
    float dist = length(coord);

    // If the distance is greater than the per-instance radius, discard this fragment
    if (dist > radius) {
        discard;
    }

    // Calculate the z-coordinate (depth) of the point on the sphere's surface using the Pythagorean theorem
    float z = sqrt(radius * radius - dist * dist);

    // Normal for basic lighting (assuming the sphere is facing the camera)
    vec3 normal = normalize(vec3(coord, z));

    // Calculate the lighting using Lambertian reflectance (dot product of normal and light direction)
    float lighting = max(dot(normal, lightDir), 0.0);

    // Apply the lighting to the sphere's color
    FragColor = vec4(sphereColor * lighting, 1.0);
}

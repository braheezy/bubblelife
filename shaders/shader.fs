#version 410 core

in vec3 fragPosition;  // Fragment's world position
in float radius;       // Sphere radius passed from vertex shader
out vec4 FragColor;    // Final fragment color

// Uniforms
uniform vec3 sphereColor;     // Color of the sphere (albedo)
uniform vec3 lightPos;        // Light position
uniform vec3 lightColor;      // Light color
uniform vec3 viewPos;         // Camera position
uniform vec3 ambientLight;    // Ambient light color

// Cubemap sampler for background reflections
uniform samplerCube skybox;

// Blinn-Phong lighting calculation
void main() {
    // Calculate the fragment's position within the point sprite (normalized coordinates from -1 to 1)
    vec2 coord = 2.0 * gl_PointCoord - vec2(1.0);  // gl_PointCoord is in range [0, 1], map to [-1, 1]

    // Compute the distance from the center of the point sprite to this fragment
    float dist = length(coord);

    // If the distance is greater than the per-instance radius, discard this fragment (for spherical effect)
    if (dist > radius) {
        discard;
    }

    // Calculate the z-coordinate (depth) of the point on the sphere's surface using the Pythagorean theorem
    float z = sqrt(radius * radius - dist * dist);

    // Compute the fragment's actual position in world space
    vec3 fragPos = vec3(fragPosition.xy + coord * radius, fragPosition.z + z * radius);

    // Compute normal at the fragment's point on the sphere's surface
    vec3 normal = normalize(vec3(coord, z));  // Sphere's surface normal

    // Lighting calculations:
    // 1. Ambient light
    vec3 ambient = ambientLight * sphereColor;

    // 2. Diffuse light (Lambertian reflectance)
    vec3 lightDir = normalize(lightPos - fragPos);
    float diff = max(dot(normal, lightDir), 0.0);
    vec3 diffuse = diff * lightColor * sphereColor;

    // 3. Specular light (Blinn-Phong)
    vec3 viewDir = normalize(viewPos - fragPos);  // Direction from fragment to camera
    vec3 halfwayDir = normalize(lightDir + viewDir);  // Halfway vector for Blinn-Phong
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 64.0);  // Specular intensity (shininess = 64)
    vec3 specular = spec * lightColor;

    // Combine lighting components
    vec3 result = ambient + diffuse + specular;

    // Apply cube map reflection (optional enhancement)
    vec3 reflection = reflect(-viewDir, normal);
    vec3 reflectedColor = texture(skybox, reflection).rgb;

    // Mix lighting with reflection for some environmental reflection (adjust mix factor as needed)
    vec3 finalColor = mix(result, reflectedColor, 0.15);  // 15% reflection, can be tuned

    FragColor = vec4(finalColor, 1.0);  // Final color with full alpha
}

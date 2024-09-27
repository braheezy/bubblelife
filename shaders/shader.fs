#version 410 core

in vec3 fragPosition;  // Fragment's world position
in vec3 fragColor;     // Color passed from the vertex shader (fix name to match vertex shader)
in float radius;       // Sphere radius passed from vertex shader
out vec4 FragColor;    // Final fragment color

// Uniforms
uniform vec3 lightDir;             // Direction of the directional light
uniform vec3 lightColor;           // Color of the light
uniform vec3 viewPos;              // Camera position
uniform vec3 ambientLight;         // Ambient light color

// Bubble effect parameters
uniform float bubbleThickness;     // Controls iridescence (thin-film effect)
uniform float fresnelStrength;     // Fresnel effect strength for reflections
uniform float transparency;        // Base transparency level for the sphere

// Cubemap sampler for background reflections
uniform samplerCube skybox;

// Fresnel Schlick approximation
float fresnelSchlick(float cosTheta, float strength) {
    return strength + (1.0 - strength) * pow(1.0 - cosTheta, 5.0);
}

// Iridescence color shift (thin-film interference)
vec3 iridescence(float thickness, vec3 normal, vec3 viewDir) {
    float angle = max(dot(normal, viewDir), 0.0);
    float shift = (1.0 - angle) * thickness;

    // Subtle color shift for iridescence effect
    vec3 colorShift = vec3(
        0.9 + 0.1 * sin(12.0 * shift),  // Red channel shift
        0.9 + 0.1 * sin(12.0 * shift + 2.0),  // Green channel shift
        0.9 + 0.1 * sin(12.0 * shift + 4.0)   // Blue channel shift
    );

    return colorShift;
}

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
    vec3 sphereColor = fragColor;  // Use the passed color from the vertex shader
    // 1. Ambient light (stronger for overall brightness)
    vec3 ambient = ambientLight * sphereColor * 1.2;  // Further boosted ambient

    // 2. Diffuse lighting (Directional Light)
    vec3 lightDirection = normalize(-lightDir);  // Light direction, invert for shading
    float diff = max(dot(normal, lightDirection), 0.0);  // Lambertian reflectance
    vec3 diffuse = diff * lightColor * sphereColor * 0.7;  // Softer diffuse light

    // 3. View-based lighting: This adds light contribution based on the camera's view position
    vec3 viewDir = normalize(viewPos - fragPos);  // Direction from fragment to camera
    float viewLight = max(dot(normal, viewDir), 0.0);  // View-based lighting factor
    vec3 viewLighting = viewLight * lightColor * sphereColor * 0.5;  // Half the strength of normal diffuse

    // 4. Specular lighting (Blinn-Phong model)
    vec3 halfwayDir = normalize(lightDirection + viewDir);  // Halfway vector for Blinn-Phong
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 64.0);  // Shininess factor (adjustable)
    vec3 specular = spec * lightColor * 0.5;  // Softer specular highlights

    // 5. Fresnel reflection (strong reflections at glancing angles)
    float fresnel = fresnelSchlick(dot(normal, viewDir), fresnelStrength);
    vec3 reflection = texture(skybox, reflect(-viewDir, normal)).rgb;

    // 6. Iridescence effect (subtle color shift)
    vec3 iridescentColor = iridescence(bubbleThickness, normal, viewDir);

    // Combine lighting, reflection, and iridescence with the base color
    vec3 resultColor = ambient + diffuse + viewLighting + specular;

    // Increase reflection blending to make the spheres brighter
    vec3 finalColor = mix(resultColor, reflection * iridescentColor, fresnel * 0.4);

    // Output the color with transparency
    FragColor = vec4(resultColor, transparency);  // Final color with transparency
}

#version 410 core

// Fragment's world position
in vec3 fragPosition;
// Color passed from the vertex shader (fix name to match vertex shader)
in vec3 fragColor;
// bubble radius passed from vertex shader
in float radius;
// Final fragment color
out vec4 FragColor;

// Uniforms
// Direction of the directional light
uniform vec3 lightDir;
// Color of the light
uniform vec3 lightColor;
// Camera position
uniform vec3 viewPos;
// Ambient light color
uniform vec3 ambientLight;

// Bubble effect parameters
// Controls iridescence (thin-film effect)
uniform float bubbleThickness;
// Fresnel effect strength for reflections
uniform float fresnelStrength;
// Base transparency level for the bubble
uniform float transparency;

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
    // gl_PointCoord is in range [0, 1], map to [-1, 1]
    vec2 coord = 2.0 * gl_PointCoord - vec2(1.0);

    // Compute the distance from the center of the point sprite to this fragment
    float dist = length(coord);

    // If the distance is greater than the per-instance radius, discard this fragment (for spherical effect)
    if (dist > radius) {
        discard;
    }

    // Calculate the z-coordinate (depth) of the point on the bubble's surface using the Pythagorean theorem
    float z = sqrt(radius * radius - dist * dist);

    // Compute the fragment's actual position in world space
    vec3 fragPos = vec3(fragPosition.xy + coord * radius, fragPosition.z + z * radius);

    // Compute normal at the fragment's point on the bubble's surface
    vec3 normal = normalize(vec3(coord, z));

    // Lighting calculations:
    vec3 bubbleColor = fragColor;
    // 1. Ambient light (stronger for overall brightness)
    vec3 ambient = ambientLight * bubbleColor * 1.2;

    // 2. Diffuse lighting (Directional Light)
    // Light direction, invert for shading
    vec3 lightDirection = normalize(-lightDir);
    // Lambertian reflectance
    float diff = max(dot(normal, lightDirection), 0.0);
    vec3 diffuse = diff * lightColor * bubbleColor * 0.7;

    // 3. View-based lighting: This adds light contribution based on the camera's view position
    // Direction from fragment to camera
    vec3 viewDir = normalize(viewPos - fragPos);
     // View-based lighting factor
    float viewLight = max(dot(normal, viewDir), 0.0);
    // Half the strength of normal diffuse
    vec3 viewLighting = viewLight * lightColor * bubbleColor * 0.5;

    // 4. Specular lighting (Blinn-Phong model)
     // Halfway vector for Blinn-Phong
    vec3 halfwayDir = normalize(lightDirection + viewDir);
    // Shininess factor (adjustable)
    float spec = pow(max(dot(normal, halfwayDir), 0.0), 64.0);
    // Softer specular highlights
    vec3 specular = spec * lightColor * 0.5;

    // 5. Fresnel reflection (strong reflections at glancing angles)
    float fresnel = fresnelSchlick(dot(normal, viewDir), fresnelStrength);
    vec3 reflection = texture(skybox, reflect(-viewDir, normal)).rgb;

    // 6. Iridescence effect (subtle color shift)
    vec3 iridescentColor = iridescence(bubbleThickness, normal, viewDir);

    // Combine lighting, reflection, and iridescence with the base color
    vec3 resultColor = ambient + diffuse + viewLighting + specular;

    // Increase reflection blending to make the bubble's brighter
    vec3 finalColor = mix(resultColor, reflection * iridescentColor, fresnel * 0.4);

    // Output the color with transparency
    FragColor = vec4(resultColor, transparency);
}

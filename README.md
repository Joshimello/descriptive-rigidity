# Descriptive Rigidity

A lightweight API server that generates 3D character deformation data from natural language descriptions using AI. Built for integration with As-Rigid-As-Possible (ARAP) deformation systems.

## Overview

This project provides a simple HTTP API that takes control point data and text prompts to generate realistic deformation amounts for 3D character animation. It uses OpenAI's GPT-4o-mini model to interpret animation descriptions and output appropriate delta transformations while respecting ARAP rigidity constraints.

**Key Features:**
- Natural language to 3D deformation translation
- ARAP-aware deformation generation
- REST API with JSON input/output
- Support for humanoid character rigs
- Realistic motion constraints
- Multi-frame animation sequences

## Quick Start

### Prerequisites
- Go 1.24+ 
- OpenAI API key

### Setup
1. Clone the repository
2. Create a `.env` file with your OpenAI API key:
   ```
   OPENAI_API_KEY=your_api_key_here
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Start the server:
   ```bash
   ./start_server.sh
   ```

The server will start on port 8080 (or the port specified in the `PORT` environment variable).

## API Reference

### POST /generate-deformations

Generate deformation amounts for control points based on a text description. Supports both single-frame and multi-frame animations.

**Request Body:**
```json
{
  "control_points": [
    {
      "id": 0,
      "role": "left arm",
      "position": [1.0, 2.0, 0.0]
    },
    {
      "id": 1,
      "role": "right arm", 
      "position": [-1.0, 2.0, 0.0]
    }
  ],
  "prompt": "make the character wave",
  "length": 1
}
```

**Parameters:**
- `control_points`: Array of control points with id, role, and position
- `prompt`: Natural language description of the desired animation
- `length`: Number of animation frames to generate (must be > 0)

**Response:**
Returns an array of deformation frames. Each frame contains deformations for each control point.

```json
[
  {
    "0": {"delta_x": 0.0, "delta_y": 0.0, "delta_z": 0.0},
    "1": {"delta_x": 0.5, "delta_y": 1.0, "delta_z": 0.2}
  }
]
```

**Multi-frame Example:**
```json
{
  "control_points": [...],
  "prompt": "make the character walk forward",
  "length": 4
}
```

Response:
```json
[
  {
    "0": {"delta_x": 0.1, "delta_y": 0.0, "delta_z": 0.0},
    "1": {"delta_x": 0.0, "delta_y": 0.0, "delta_z": 0.0},
    "2": {"delta_x": -0.1, "delta_y": 0.0, "delta_z": 0.0}
  },
  {
    "0": {"delta_x": 0.2, "delta_y": 0.0, "delta_z": 0.0},
    "1": {"delta_x": 0.0, "delta_y": 0.0, "delta_z": 0.0},
    "2": {"delta_x": -0.2, "delta_y": 0.0, "delta_z": 0.0}
  },
  {
    "0": {"delta_x": 0.1, "delta_y": 0.0, "delta_z": 0.0},
    "1": {"delta_x": 0.0, "delta_y": 0.0, "delta_z": 0.0},
    "2": {"delta_x": -0.1, "delta_y": 0.0, "delta_z": 0.0}
  },
  {
    "0": {"delta_x": 0.0, "delta_y": 0.0, "delta_z": 0.0},
    "1": {"delta_x": 0.0, "delta_y": 0.0, "delta_z": 0.0},
    "2": {"delta_x": 0.0, "delta_y": 0.0, "delta_z": 0.0}
  }
]
```

## Integration Examples

### JavaScript

```javascript
async function generateDeformations(controlPoints, prompt) {
  const response = await fetch('http://localhost:8080/generate-deformations', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      control_points: controlPoints,
      prompt: prompt,
      length: 1  // or specify number of frames for multi-frame animation
    })
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  return await response.json();
}

// Usage example
const controlPoints = [
  { id: 0, role: "left arm", position: [1.0, 2.0, 0.0] },
  { id: 1, role: "right arm", position: [-1.0, 2.0, 0.0] },
  { id: 2, role: "head", position: [0.0, 7.0, 0.0] }
];

generateDeformations(controlPoints, "make the character wave")
  .then(frames => {
    console.log('Generated animation frames:', frames);
    // Apply deformations to your 3D model
    frames.forEach((deformations, frameIndex) => {
      console.log(`Frame ${frameIndex}:`, deformations);
      Object.entries(deformations).forEach(([id, delta]) => {
        // Apply delta_x, delta_y, delta_z to control point with id
        applyDeformation(parseInt(id), delta.delta_x, delta.delta_y, delta.delta_z);
      });
    });
  })
  .catch(error => console.error('Error:', error));

// Multi-frame animation example
async function animateCharacter() {
  const response = await fetch('http://localhost:8080/generate-deformations', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      control_points: controlPoints,
      prompt: "make the character walk forward",
      length: 8  // Generate 8 frames
    })
  });
  
  const frames = await response.json();
  
  // Play animation by applying frames sequentially
  for (let i = 0; i < frames.length; i++) {
    applyFrame(frames[i]);
    await new Promise(resolve => setTimeout(resolve, 100)); // 100ms per frame
  }
}
```

### C++

```cpp
#include <iostream>
#include <string>
#include <curl/curl.h>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

// Callback function to write response data
static size_t WriteCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    userp->append((char*)contents, size * nmemb);
    return size * nmemb;
}

class DeformationGenerator {
private:
    std::string apiUrl;
    
public:
    DeformationGenerator(const std::string& url = "http://localhost:8080/generate-deformations") 
        : apiUrl(url) {}
    
    json generateDeformations(const json& controlPoints, const std::string& prompt) {
        CURL* curl;
        CURLcode res;
        std::string response;
        
        curl = curl_easy_init();
        if (!curl) {
            throw std::runtime_error("Failed to initialize CURL");
        }
        
        // Prepare request data
        json requestData = {
            {"control_points", controlPoints},
            {"prompt", prompt},
            {"length", 1}  // or specify frames for multi-frame animation
        };
        std::string jsonString = requestData.dump();
        
        // Set CURL options
        curl_easy_setopt(curl, CURLOPT_URL, apiUrl.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonString.c_str());
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
        
        // Set headers
        struct curl_slist* headers = nullptr;
        headers = curl_slist_append(headers, "Content-Type: application/json");
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        
        // Perform request
        res = curl_easy_perform(curl);
        
        // Cleanup
        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
        
        if (res != CURLE_OK) {
            throw std::runtime_error("CURL request failed: " + std::string(curl_easy_strerror(res)));
        }
        
        return json::parse(response);
    }
};

// Usage example
int main() {
    try {
        DeformationGenerator generator;
        
        // Define control points
        json controlPoints = json::array({
            {{"id", 0}, {"role", "left arm"}, {"position", {1.0, 2.0, 0.0}}},
            {{"id", 1}, {"role", "right arm"}, {"position", {-1.0, 2.0, 0.0}}},
            {{"id", 2}, {"role", "head"}, {"position", {0.0, 7.0, 0.0}}}
        });
        
        // Generate deformations
        json deformations = generator.generateDeformations(controlPoints, "make the character wave");
        
        // Process results (now returns an array of frames)
        std::cout << "Generated animation frames:\n" << deformations.dump(2) << std::endl;
        
        // Apply each frame of deformations to your 3D model
        for (size_t frameIndex = 0; frameIndex < deformations.size(); ++frameIndex) {
            const auto& frame = deformations[frameIndex];
            std::cout << "Frame " << frameIndex << ":\n";
            
            for (auto& [id, delta] : frame.items()) {
                int controlPointId = std::stoi(id);
                double deltaX = delta["delta_x"];
                double deltaY = delta["delta_y"];
                double deltaZ = delta["delta_z"];
                
                // Apply to your 3D system
                applyDeformation(controlPointId, deltaX, deltaY, deltaZ);
            }
        }
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
    
    return 0;
}

// Your deformation application function
void applyDeformation(int controlPointId, double deltaX, double deltaY, double deltaZ) {
    // Implement your specific deformation logic here
    std::cout << "Applying deformation to control point " << controlPointId 
              << ": (" << deltaX << ", " << deltaY << ", " << deltaZ << ")" << std::endl;
}
```

**C++ Dependencies:**
- libcurl for HTTP requests
- nlohmann/json for JSON parsing

Install on Ubuntu/Debian:
```bash
sudo apt-get install libcurl4-openssl-dev nlohmann-json3-dev
```

Compile:
```bash
g++ -std=c++17 your_file.cpp -lcurl -o deformation_client
```

## Testing

Run the included test script to verify the API is working:

```bash
./test_api.sh
```

This will send a sample request and display the response.

## Common Control Point Roles

- `"head"`, `"neck"`
- `"left arm"`, `"right arm"`, `"left hand"`, `"right hand"`
- `"left leg"`, `"right leg"`, `"left foot"`, `"right foot"`
- `"spine"`, `"chest"`, `"pelvis"`

## Animation Prompts Examples

Single-frame animations (static poses):
- `"make the character wave"` (length: 1)
- `"character jumps with arms up"` (length: 1)
- `"bend forward to pick something up"` (length: 1)
- `"raise left hand to the head"` (length: 1)

Multi-frame animations (motion sequences):
- `"make the character walk naturally forward"` (length: 8-12)
- `"character does jumping jacks"` (length: 4-6)
- `"character waves enthusiastically"` (length: 6-8)
- `"character turns around"` (length: 4-8)
- `"character sits down"` (length: 6-10)

## License

MIT License
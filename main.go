package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sashabaranov/go-openai"
)

// Input struct for the API request
type ControlPoint struct {
	ID       int       `json:"id"`
	Role     string    `json:"role"`
	Position []float64 `json:"position"`
}

type RequestPayload struct {
	ControlPoints []ControlPoint `json:"control_points"`
	Prompt        string         `json:"prompt"`
	Length        int            `json:"length"`
}

// Output struct for deformation amounts
type Deformation struct {
	DeltaX float64 `json:"delta_x"`
	DeltaY float64 `json:"delta_y"`
	DeltaZ float64 `json:"delta_z"`
}

type OpenAIResponse struct {
	Frames []map[string]Deformation `json:"frames"`
}

type ResponsePayload []map[int]Deformation

// System prompt for GPT-4o-mini
const systemPrompt = `
You are an animation generation assistant integrated with an As-Rigid-As-Possible (ARAP) deformation system. Your task is to generate a JSON array containing multiple frames of deformation amounts for each control point of a 3D character model based on a user-provided text prompt, control point data, and animation length. The deformation amounts represent changes in position (delta_x, delta_y, delta_z) for each control point to achieve the described animation while preserving ARAP rigidity constraints (minimize stretching, prioritize local rigidity).

**Input**:
- **Control Points**: A list of control points with id (integer), role (e.g., "left leg", "right arm", "head"), and position (x, y, z coordinates as floats).
- **Prompt**: A text description of the desired animation (e.g., "make the character wave", "make the character walk naturally forward").
- **Length**: The number of animation frames to generate (integer).
- **Context**: Assume a 3D humanoid character model with a standard rig (arms, legs, head).

**Output**:
- A JSON array where each element represents one frame of animation.
- Each frame is a JSON object where each key is a control point id (as a string), and the value is an object with delta_x, delta_y, delta_z (deformation amounts in the same units as the input positions).
- The frames should create a smooth animation sequence for the described motion (e.g., for "walk", alternate leg movements; for "wave", arm going up and down).
- Ensure deformations are plausible for a humanoid character and respect ARAP constraints (small, localized changes for non-moving parts; smooth transitions for moving parts).
- If the prompt affects only specific control points (e.g., "wave" primarily involves the arm), set deformation amounts for unaffected points (e.g., legs, head) to zero or minimal values.
- For cyclical animations (like walking), ensure the sequence can loop smoothly by making the last frame transition well back to the first frame.

**Example Input**:
{
  "control_points": [
    {"id": 0, "role": "left leg", "position": [1, 2, 0]},
    {"id": 1, "role": "right arm", "position": [-1, 2, 0]},
    {"id": 2, "role": "head", "position": [0, 7, 0]}
  ],
  "prompt": "make the character wave",
  "length": 3
}

**Example Output**:
[
  {
    "0": {"delta_x": 0, "delta_y": 0, "delta_z": 0},
    "1": {"delta_x": 0.2, "delta_y": 0.5, "delta_z": 0.1},
    "2": {"delta_x": 0, "delta_y": 0, "delta_z": 0}
  },
  {
    "0": {"delta_x": 0, "delta_y": 0, "delta_z": 0},
    "1": {"delta_x": 0.5, "delta_y": 1.0, "delta_z": 0.2},
    "2": {"delta_x": 0, "delta_y": 0, "delta_z": 0}
  },
  {
    "0": {"delta_x": 0, "delta_y": 0, "delta_z": 0},
    "1": {"delta_x": 0.2, "delta_y": 0.3, "delta_z": 0.1},
    "2": {"delta_x": 0, "delta_y": 0, "delta_z": 0}
  }
]

**Instructions**:
1. Interpret the prompt to identify which control points are involved in the animation and the type of motion.
2. Generate the specified number of frames that create a smooth animation sequence.
3. Keep deformations small and realistic (e.g., within ±1 unit unless specified) to maintain ARAP rigidity.
4. Set deformation amounts to zero for control points unaffected by the animation.
5. For cyclical motions, ensure smooth looping by making frame transitions natural.
6. Output only the JSON array with deformation frames, no additional text.
`

// Handler for the /generate-deformations endpoint
func generateDeformations(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var payload RequestPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate input
	if len(payload.ControlPoints) == 0 || payload.Prompt == "" || payload.Length <= 0 {
		http.Error(w, "Missing control_points, prompt, or invalid length", http.StatusBadRequest)
		return
	}

	// Fix duplicate IDs by reassigning unique IDs (assuming typo in input)
	idMap := make(map[int]int)
	uniqueID := 0
	for i, cp := range payload.ControlPoints {
		if _, exists := idMap[cp.ID]; !exists {
			idMap[cp.ID] = uniqueID
			payload.ControlPoints[i].ID = uniqueID
			uniqueID++
		} else {
			payload.ControlPoints[i].ID = idMap[cp.ID]
		}
	}

	// Initialize OpenAI client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		http.Error(w, "OpenAI API key not configured", http.StatusInternalServerError)
		return
	}
	client := openai.NewClient(apiKey)

	// Prepare input for GPT-4o-mini
	inputJSON, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to serialize input", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Sending payload to OpenAI: %s", string(inputJSON))

	// Call GPT-4o-mini
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: string(inputJSON),
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("OpenAI API error: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse OpenAI response
	responseContent := resp.Choices[0].Message.Content
	log.Printf("OpenAI Response Content: %s", responseContent)
	
	var openaiResp OpenAIResponse
	if err := json.Unmarshal([]byte(responseContent), &openaiResp); err != nil {
		log.Printf("Failed to parse OpenAI response: %v", err)
		log.Printf("Response content was: %s", responseContent)
		http.Error(w, fmt.Sprintf("Failed to parse OpenAI response: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Convert string keys to integers and create deformations array
	deformations := make(ResponsePayload, len(openaiResp.Frames))
	for frameIndex, frame := range openaiResp.Frames {
		frameMap := make(map[int]Deformation)
		for idStr, deformation := range frame {
			id := 0
			if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
				log.Printf("Invalid ID format: %s", idStr)
				continue
			}
			frameMap[id] = deformation
		}
		deformations[frameIndex] = frameMap
	}

	// Adjust IDs back to original (if they were remapped)
	adjustedDeformations := make(ResponsePayload, len(deformations))
	for frameIndex, frame := range deformations {
		adjustedFrame := make(map[int]Deformation)
		for originalID, newID := range idMap {
			if deformation, exists := frame[newID]; exists {
				adjustedFrame[originalID] = deformation
			}
		}
		adjustedDeformations[frameIndex] = adjustedFrame
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(adjustedDeformations); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Set up router
	http.HandleFunc("/generate-deformations", generateDeformations)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/omnikam04/release-notes-generator/internal/external/bugsby"
	"github.com/omnikam04/release-notes-generator/internal/models"
)

// AIReleaseNoteResponse represents the structured JSON response from AI
type AIReleaseNoteResponse struct {
	ReleaseNote         string   `json:"release_note"`
	Confidence          float64  `json:"confidence"`
	Reasoning           string   `json:"reasoning"`
	AlternativeVersions []string `json:"alternative_versions"`
}

// BuildReleaseNotePrompt constructs a prompt for AI to generate a release note
func BuildReleaseNotePrompt(bug *models.Bug, commits []*bugsby.ParsedCommitInfo) string {
	var builder strings.Builder

	// System instruction with AID1711 guidelines
	builder.WriteString("You are a technical writer creating release notes for network operating system bugs.\n\n")
	builder.WriteString("MANDATORY RELEASE NOTE GUIDELINES (AID1711):\n\n")

	builder.WriteString("AUDIENCE & FOCUS:\n")
	builder.WriteString("- Write for CUSTOMERS and field teams, NOT internal engineering\n")
	builder.WriteString("- Focus on customer-visible issue/symptom, NOT internal fix details\n")
	builder.WriteString("- Answer: What will customers notice? What conditions trigger this issue?\n\n")

	builder.WriteString("FORMAT & CONTENT:\n")
	builder.WriteString("- Keep it brief (1-2 sentences)\n")
	builder.WriteString("- MUST include: when the problem occurs (required configuration) and the impact\n")
	builder.WriteString("- Use past tense for fixes (e.g., 'Resolved', 'Fixed', 'Corrected')\n")
	builder.WriteString("- If workaround exists, add as second line (do NOT say 'no known workarounds')\n\n")

	builder.WriteString("AVOID INTERNAL JARGON:\n")
	builder.WriteString("- NO internal architectural names (e.g., 'HW LAG', 'SW LAG')\n")
	builder.WriteString("- NO codenames (e.g., Jericho, Sand, Broadcom chip numbers)\n")
	builder.WriteString("- NO bug IDs in the note text\n")
	builder.WriteString("- NO specific EOS version numbers in the note text\n")
	builder.WriteString("- AVOID: crash, segfault, assert, race condition\n\n")

	builder.WriteString("AGENT/SYSTEM LANGUAGE:\n")
	builder.WriteString("- If agent dies: 'the [Agent Name] agent can restart unexpectedly'\n")
	builder.WriteString("- If system goes down: 'the system can restart unexpectedly' or 'reset unexpectedly'\n\n")

	builder.WriteString("SPELLING & CAPITALIZATION:\n")
	builder.WriteString("- Use American English spelling\n")
	builder.WriteString("- Protocol names/acronyms in ALL CAPS (BGP, OSPF, MLAG, VXLAN)\n")
	builder.WriteString("- Specific spellings: 'running config', 'route map', 'next hop', 'port channel' (not hyphenated)\n")
	builder.WriteString("- Use 'workaround' as a noun\n\n")

	builder.WriteString("DO NOT:\n")
	builder.WriteString("- Comment on likelihood (avoid 'rare', 'infrequently', etc.)\n\n")

	// Bug information
	builder.WriteString("=== BUG INFORMATION ===\n\n")
	builder.WriteString(fmt.Sprintf("Bug ID: %s\n", bug.BugsbyID))
	builder.WriteString(fmt.Sprintf("Title: %s\n", bug.Title))
	builder.WriteString(fmt.Sprintf("Severity: %s\n", bug.Severity))
	builder.WriteString(fmt.Sprintf("Priority: %s\n", bug.Priority))

	if bug.Component != "" {
		builder.WriteString(fmt.Sprintf("Component: %s\n", bug.Component))
	}

	if bug.Release != "" {
		builder.WriteString(fmt.Sprintf("Release: %s\n", bug.Release))
	}

	if bug.Description != nil && *bug.Description != "" {
		builder.WriteString(fmt.Sprintf("\nDescription:\n%s\n", *bug.Description))
	}

	// Commit information
	if len(commits) > 0 {
		builder.WriteString("\n=== CODE CHANGES ===\n\n")
		builder.WriteString(fmt.Sprintf("Number of commits: %d\n\n", len(commits)))

		for i, commit := range commits {
			builder.WriteString(fmt.Sprintf("Commit %d:\n", i+1))

			if commit.Title != "" {
				builder.WriteString(fmt.Sprintf("  Title: %s\n", commit.Title))
			}

			if commit.ChangeID != "" {
				builder.WriteString(fmt.Sprintf("  Change ID: %s\n", commit.ChangeID))
			}

			if commit.Message != "" {
				// Truncate long messages
				message := commit.Message
				if len(message) > 500 {
					message = message[:500] + "..."
				}
				builder.WriteString(fmt.Sprintf("  Message: %s\n", message))
			}

			builder.WriteString("\n")
		}
	} else {
		builder.WriteString("\n=== CODE CHANGES ===\n\n")
		builder.WriteString("No commit information available.\n\n")
	}

	// Output format instruction
	builder.WriteString("=== OUTPUT FORMAT ===\n\n")
	builder.WriteString("Return a JSON object with the following structure:\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"release_note\": \"<your release note text>\",\n")
	builder.WriteString("  \"confidence\": <0.0-1.0>,\n")
	builder.WriteString("  \"reasoning\": \"<brief explanation of your confidence score>\",\n")
	builder.WriteString("  \"alternative_versions\": [\"<alternative 1>\", \"<alternative 2>\"]\n")
	builder.WriteString("}\n\n")

	builder.WriteString("EXAMPLE OUTPUT:\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"release_note\": \"Resolved issue preventing packet capture on C-360 APs in Dual 5G mode\",\n")
	builder.WriteString("  \"confidence\": 0.85,\n")
	builder.WriteString("  \"reasoning\": \"Bug affects C-360 AP packet capture in specific mode. Combined info from 3 commits.\",\n")
	builder.WriteString("  \"alternative_versions\": [\n")
	builder.WriteString("    \"Fixed packet capture for C-360 APs in Dual 5G mode\",\n")
	builder.WriteString("    \"Resolved packet capture failure on C-360 access points\"\n")
	builder.WriteString("  ]\n")
	builder.WriteString("}\n\n")

	builder.WriteString("Generate the release note following ALL AID1711 guidelines above.\n")
	builder.WriteString("Return ONLY valid JSON, no additional text.\n")

	return builder.String()
}

// BuildReleaseNotePromptSimple constructs a simpler prompt when no commits are available
func BuildReleaseNotePromptSimple(bug *models.Bug) string {
	var builder strings.Builder

	// Use same AID1711 guidelines as detailed prompt
	builder.WriteString("You are a technical writer creating release notes following AID1711 guidelines.\n\n")
	builder.WriteString("IMPORTANT: Write for CUSTOMERS, focus on customer-visible symptoms, avoid internal jargon.\n\n")

	builder.WriteString(fmt.Sprintf("Bug ID: %s\n", bug.BugsbyID))
	builder.WriteString(fmt.Sprintf("Title: %s\n", bug.Title))
	builder.WriteString(fmt.Sprintf("Severity: %s\n", bug.Severity))

	if bug.Component != "" {
		builder.WriteString(fmt.Sprintf("Component: %s\n", bug.Component))
	}

	if bug.Description != nil && *bug.Description != "" {
		builder.WriteString(fmt.Sprintf("\nDescription: %s\n", *bug.Description))
	}

	builder.WriteString("\n\nReturn JSON format:\n")
	builder.WriteString("{\n")
	builder.WriteString("  \"release_note\": \"<1-2 sentence customer-facing note>\",\n")
	builder.WriteString("  \"confidence\": <0.0-1.0>,\n")
	builder.WriteString("  \"reasoning\": \"<why this confidence>\",\n")
	builder.WriteString("  \"alternative_versions\": [\"<alt 1>\", \"<alt 2>\"]\n")
	builder.WriteString("}\n")

	return builder.String()
}

// ParseAIResponse parses the JSON response from AI and returns the structured data
func ParseAIResponse(response string) (*AIReleaseNoteResponse, error) {
	// Clean up the response
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	// Parse JSON
	var aiResponse AIReleaseNoteResponse
	if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
		// If JSON parsing fails, try to extract plain text as fallback
		return &AIReleaseNoteResponse{
			ReleaseNote:         response,
			Confidence:          0.5, // Low confidence for non-JSON response
			Reasoning:           "Failed to parse JSON response, using raw text",
			AlternativeVersions: []string{},
		}, nil
	}

	// Validate confidence is in valid range
	if aiResponse.Confidence < 0.0 {
		aiResponse.Confidence = 0.0
	} else if aiResponse.Confidence > 1.0 {
		aiResponse.Confidence = 1.0
	}

	return &aiResponse, nil
}

// ExtractReleaseNoteFromResponse extracts just the release note text (backward compatibility)
func ExtractReleaseNoteFromResponse(response string) string {
	parsed, err := ParseAIResponse(response)
	if err != nil || parsed == nil {
		return strings.TrimSpace(response)
	}
	return parsed.ReleaseNote
}

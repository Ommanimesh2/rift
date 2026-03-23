package output

import (
	"encoding/json"
	"fmt"

	"github.com/Ommanimesh2/rift/internal/security"
)

// SARIF v2.1.0 structures for GitHub Code Scanning integration.

// SarifReport is the top-level SARIF document.
type SarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SarifRun `json:"runs"`
}

// SarifRun represents a single analysis run.
type SarifRun struct {
	Tool    SarifTool     `json:"tool"`
	Results []SarifResult `json:"results"`
}

// SarifTool describes the analysis tool.
type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

// SarifDriver describes the tool driver.
type SarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Version        string      `json:"version"`
	Rules          []SarifRule `json:"rules"`
}

// SarifRule defines a rule for a security finding type.
type SarifRule struct {
	ID               string          `json:"id"`
	ShortDescription SarifMessage    `json:"shortDescription"`
	FullDescription  SarifMessage    `json:"fullDescription"`
	DefaultConfig    SarifRuleConfig `json:"defaultConfiguration"`
}

// SarifRuleConfig holds the default severity level.
type SarifRuleConfig struct {
	Level string `json:"level"`
}

// SarifMessage is a SARIF message with text.
type SarifMessage struct {
	Text string `json:"text"`
}

// SarifResult is a single finding.
type SarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SarifMessage    `json:"message"`
	Locations []SarifLocation `json:"locations"`
}

// SarifLocation represents a file location.
type SarifLocation struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
}

// SarifPhysicalLocation is a physical file path.
type SarifPhysicalLocation struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
}

// SarifArtifactLocation is a URI-based artifact location.
type SarifArtifactLocation struct {
	URI string `json:"uri"`
}

// sarifRules defines all supported security event rules.
var sarifRules = []SarifRule{
	{
		ID:               "rift/new_suid",
		ShortDescription: SarifMessage{Text: "New file with SUID bit"},
		FullDescription:  SarifMessage{Text: "A newly added file has the SUID bit set, allowing it to execute with the file owner's privileges."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
	{
		ID:               "rift/new_sgid",
		ShortDescription: SarifMessage{Text: "New file with SGID bit"},
		FullDescription:  SarifMessage{Text: "A newly added file has the SGID bit set, allowing it to execute with the file group's privileges."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
	{
		ID:               "rift/suid_added",
		ShortDescription: SarifMessage{Text: "SUID bit added to existing file"},
		FullDescription:  SarifMessage{Text: "An existing file gained the SUID bit, allowing it to execute with the file owner's privileges."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
	{
		ID:               "rift/sgid_added",
		ShortDescription: SarifMessage{Text: "SGID bit added to existing file"},
		FullDescription:  SarifMessage{Text: "An existing file gained the SGID bit, allowing it to execute with the file group's privileges."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
	{
		ID:               "rift/new_executable",
		ShortDescription: SarifMessage{Text: "New executable file"},
		FullDescription:  SarifMessage{Text: "A newly added file has execute permission bits set."},
		DefaultConfig:    SarifRuleConfig{Level: "warning"},
	},
	{
		ID:               "rift/world_writable",
		ShortDescription: SarifMessage{Text: "World-writable file"},
		FullDescription:  SarifMessage{Text: "A file is world-writable, meaning any user in the container can modify it."},
		DefaultConfig:    SarifRuleConfig{Level: "warning"},
	},
	{
		ID:               "rift/perm_escalation",
		ShortDescription: SarifMessage{Text: "Permission escalation"},
		FullDescription:  SarifMessage{Text: "A file's permission bits were broadened, granting more access than before."},
		DefaultConfig:    SarifRuleConfig{Level: "warning"},
	},
	{
		ID:               "rift/secret_private_key",
		ShortDescription: SarifMessage{Text: "Private key detected"},
		FullDescription:  SarifMessage{Text: "A file containing a private key was found in the image."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
	{
		ID:               "rift/secret_aws_key",
		ShortDescription: SarifMessage{Text: "AWS access key detected"},
		FullDescription:  SarifMessage{Text: "A file containing an AWS access key (AKIA...) was found in the image."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
	{
		ID:               "rift/secret_api_token",
		ShortDescription: SarifMessage{Text: "API token detected"},
		FullDescription:  SarifMessage{Text: "A file containing an API key or token pattern was found in the image."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
	{
		ID:               "rift/secret_file_path",
		ShortDescription: SarifMessage{Text: "Secret file detected"},
		FullDescription:  SarifMessage{Text: "A file matching a known secret file pattern (e.g., .env, private key, credentials) was found."},
		DefaultConfig:    SarifRuleConfig{Level: "error"},
	},
}

// sarifRuleID maps SecurityEventKind to SARIF rule IDs.
var sarifRuleID = map[security.SecurityEventKind]string{
	security.KindNewSUID:        "rift/new_suid",
	security.KindNewSGID:        "rift/new_sgid",
	security.KindSUIDAdded:      "rift/suid_added",
	security.KindSGIDAdded:      "rift/sgid_added",
	security.KindNewExecutable:  "rift/new_executable",
	security.KindWorldWritable:  "rift/world_writable",
	security.KindPermEscalation:  "rift/perm_escalation",
	security.KindSecretPrivateKey: "rift/secret_private_key",
	security.KindSecretAWSKey:     "rift/secret_aws_key",
	security.KindSecretAPIToken:   "rift/secret_api_token",
	security.KindSecretFilePath:   "rift/secret_file_path",
}

// sarifLevel maps SecurityEventKind to SARIF severity levels.
var sarifLevel = map[security.SecurityEventKind]string{
	security.KindNewSUID:        "error",
	security.KindNewSGID:        "error",
	security.KindSUIDAdded:      "error",
	security.KindSGIDAdded:      "error",
	security.KindNewExecutable:  "warning",
	security.KindWorldWritable:  "warning",
	security.KindPermEscalation:  "warning",
	security.KindSecretPrivateKey: "error",
	security.KindSecretAWSKey:     "error",
	security.KindSecretAPIToken:   "error",
	security.KindSecretFilePath:   "error",
}

// FormatSARIF produces a SARIF v2.1.0 JSON report from security events.
func FormatSARIF(events []security.SecurityEvent, image1, image2, version string) ([]byte, error) {
	results := make([]SarifResult, 0, len(events))

	for _, ev := range events {
		ruleID, ok := sarifRuleID[ev.Kind]
		if !ok {
			continue
		}
		level := sarifLevel[ev.Kind]

		msg := fmt.Sprintf("%s: %s (mode: %04o → %04o) comparing %s → %s",
			ev.Kind, ev.Path, ev.Before, ev.After, image1, image2)

		results = append(results, SarifResult{
			RuleID:  ruleID,
			Level:   level,
			Message: SarifMessage{Text: msg},
			Locations: []SarifLocation{
				{
					PhysicalLocation: SarifPhysicalLocation{
						ArtifactLocation: SarifArtifactLocation{
							URI: ev.Path,
						},
					},
				},
			},
		})
	}

	report := SarifReport{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []SarifRun{
			{
				Tool: SarifTool{
					Driver: SarifDriver{
						Name:           "rift",
						InformationURI: "https://github.com/Ommanimesh2/rift",
						Version:        version,
						Rules:          sarifRules,
					},
				},
				Results: results,
			},
		},
	}

	return json.MarshalIndent(report, "", "  ")
}

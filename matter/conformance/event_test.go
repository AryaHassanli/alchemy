package conformance

import (
	"strings"
	"testing"
)

func TestValidateEventConformance(t *testing.T) {
	isFeature := func(id string) bool {
		switch id {
		case "FeatureBit1", "FeatureBit2", "ON", "OFF", "LOW", "OTHER":
			return true
		}
		return false
	}

	tests := []struct {
		name        string
		conformance string
		valid       bool
		errContains string
		lookup      func(id string) bool
	}{
		{
			name:        "Event1: M",
			conformance: "M",
			valid:       true,
		},
		{
			name:        "Event2: FeatureBit1",
			conformance: "FeatureBit1",
			valid:       true,
		},
		{
			name:        "Event3: FeatureBit1 & FeatureBit2",
			conformance: "FeatureBit1 & FeatureBit2",
			valid:       true,
		},
		{
			name:        "Event4: Attribute1 & FeatureBit1",
			conformance: "Attribute1 & FeatureBit1",
			valid:       true,
		},
		{
			name:        "Event5: O",
			conformance: "O",
			valid:       false,
			errContains: "conformance is purely optional",
		},
		{
			name:        "Event6: Attribute1",
			conformance: "Attribute1",
			valid:       false,
			errContains: "conformance is not tied to a feature bit",
		},
		{
			name:        "Event7: [FeatureBit1]",
			conformance: "[FeatureBit1]",
			valid:       false,
			errContains: "conformance is tied to a feature bit but is optional",
		},
		{
			name:        "Event8: FeatureBit1 | [FeatureBit2]",
			conformance: "FeatureBit1 | [FeatureBit2]",
			valid:       false,
			errContains: "conformance expression is not valid",
		},
		{
			name:        "Provisional with mandatory",
			conformance: "P, M",
			valid:       true,
		},
		{
			name:        "Provisional with feature bit",
			conformance: "P, FeatureBit1",
			valid:       true,
		},
		{
			name:        "Provisional with optional",
			conformance: "P, O",
			valid:       false,
			errContains: "conformance is purely optional",
		},
		{
			name:        "Deprecated with feature bit",
			conformance: "FeatureBit1, D",
			valid:       true,
		},
		{
			name:        "Feature bit containing letter O (ON)",
			conformance: "ON",
			valid:       true,
		},
		{
			name:        "Feature bit containing letter O (OTHER)",
			conformance: "OTHER",
			valid:       true,
		},
		{
			name:        "Blank conformance",
			conformance: "",
			valid:       false,
			errContains: "conformance cannot be blank",
		},
		{
			name:        "Global event with M and nil lookup",
			conformance: "M",
			valid:       true,
			lookup:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			con := ParseConformance(tt.conformance)
			lookup := isFeature
			if tt.name == "Global event with M and nil lookup" {
				lookup = nil
			}
			err := ValidateEventConformance(con, lookup)
			if tt.valid && err != nil {
				t.Errorf("ValidateEventConformance(%q) unexpected error: %v", tt.conformance, err)
			} else if !tt.valid {
				if err == nil {
					t.Errorf("ValidateEventConformance(%q) expected error but got nil", tt.conformance)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateEventConformance(%q) expected error containing %q, got %q", tt.conformance, tt.errContains, err.Error())
				}
			}
		})
	}
}

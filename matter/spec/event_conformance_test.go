package spec

import (
	"strings"
	"testing"

	"github.com/project-chip/alchemy/matter"
	"github.com/project-chip/alchemy/matter/conformance"
)

func TestValidateEventConformances(t *testing.T) {
	tests := []struct {
		name        string
		conformance string
		valid       bool
		errContains string
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
			valid:       false,
			errContains: "conformance is not tied to a feature bit",
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
			name:        "Event9: [Attribute1 & FeatureBit1]",
			conformance: "[Attribute1 & FeatureBit1]",
			valid:       false,
			errContains: "conformance is tied to a feature bit but is optional",
		},
		{
			name:        "Event10: Parentheses (FeatureBit1)",
			conformance: "(FeatureBit1)",
			valid:       true,
		},
		{
			name:        "Event11: Logical NOT !FeatureBit1",
			conformance: "!FeatureBit1",
			valid:       true,
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
	}

	spec := &Specification{
		Clusters: make(map[*matter.Cluster]struct{}),
	}
	cluster := matter.NewCluster(nil)
	cluster.Name = "TestCluster"
	cluster.ID = matter.NewNumber(0x0080)

	features := matter.NewFeatures(nil, cluster)
	features.AddFeatureBit(matter.NewFeature(nil, "0", "FeatureBit1", "FeatureBit1", "Feature 1", conformance.Set{&conformance.Optional{}}))
	features.AddFeatureBit(matter.NewFeature(nil, "1", "FeatureBit2", "FeatureBit2", "Feature 2", conformance.Set{&conformance.Optional{}}))
	features.AddFeatureBit(matter.NewFeature(nil, "2", "OnFeature", "ON", "Feature ON", conformance.Set{&conformance.Optional{}}))
	features.AddFeatureBit(matter.NewFeature(nil, "3", "OtherFeature", "OTHER", "Feature OTHER", conformance.Set{&conformance.Optional{}}))
	cluster.Features = features

	for i, tt := range tests {
		event := matter.NewEvent(nil, cluster)
		event.Name = tt.name
		event.ID = matter.NewNumber(uint64(i + 1))
		event.Conformance = conformance.ParseConformance(tt.conformance)
		cluster.Events = append(cluster.Events, event)
	}
	spec.Clusters[cluster] = struct{}{}

	violations := ValidateEventConformances(spec)
	violationMap := make(map[string]Violation)
	for _, vs := range violations {
		for _, v := range vs {
			if ev, ok := v.Entity.(*matter.Event); ok {
				violationMap[ev.Name] = v
			}
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, hasViolation := violationMap[tt.name]
			if tt.valid && hasViolation {
				t.Errorf("ValidateEventConformances(%q) unexpected violation: %s", tt.conformance, v.Text)
			} else if !tt.valid {
				if !hasViolation {
					t.Errorf("ValidateEventConformances(%q) expected violation but got none", tt.conformance)
				} else if tt.errContains != "" && !strings.Contains(v.Text, tt.errContains) {
					t.Errorf("ValidateEventConformances(%q) expected error containing %q, got %q", tt.conformance, tt.errContains, v.Text)
				}
			}
		})
	}
}

func TestProcessEventConformanceComparison(t *testing.T) {
	baseSpec := &Specification{
		Clusters: make(map[*matter.Cluster]struct{}),
	}
	baseCluster := matter.NewCluster(nil)
	baseCluster.Name = "TestCluster"
	baseCluster.ID = matter.NewNumber(0x0080)
	baseFeatures := matter.NewFeatures(nil, baseCluster)
	baseFeatures.AddFeatureBit(matter.NewFeature(nil, "0", "FeatureBit1", "FeatureBit1", "Feature 1", conformance.Set{&conformance.Optional{}}))
	baseCluster.Features = baseFeatures

	preExistingInvalidEvent := matter.NewEvent(nil, baseCluster)
	preExistingInvalidEvent.Name = "PreExistingInvalid"
	preExistingInvalidEvent.ID = matter.NewNumber(0x01)
	preExistingInvalidEvent.Conformance = conformance.ParseConformance("O")
	baseCluster.Events = append(baseCluster.Events, preExistingInvalidEvent)
	baseSpec.Clusters[baseCluster] = struct{}{}

	headSpec := &Specification{
		Clusters: make(map[*matter.Cluster]struct{}),
	}
	headCluster := matter.NewCluster(nil)
	headCluster.Name = "TestCluster"
	headCluster.ID = matter.NewNumber(0x0080)
	headFeatures := matter.NewFeatures(nil, headCluster)
	headFeatures.AddFeatureBit(matter.NewFeature(nil, "0", "FeatureBit1", "FeatureBit1", "Feature 1", conformance.Set{&conformance.Optional{}}))
	headCluster.Features = headFeatures

	headEvent1 := matter.NewEvent(nil, headCluster)
	headEvent1.Name = "PreExistingInvalid"
	headEvent1.ID = matter.NewNumber(0x01)
	headEvent1.Conformance = conformance.ParseConformance("O")

	headEvent2 := matter.NewEvent(nil, headCluster)
	headEvent2.Name = "ValidProvisionalFeature"
	headEvent2.ID = matter.NewNumber(0x02)
	headEvent2.Conformance = conformance.ParseConformance("P, FeatureBit1")

	headEvent3 := matter.NewEvent(nil, headCluster)
	headEvent3.Name = "NewInvalidEvent"
	headEvent3.ID = matter.NewNumber(0x03)
	headEvent3.Conformance = conformance.ParseConformance("P, [FeatureBit1]")

	headCluster.Events = append(headCluster.Events, headEvent1, headEvent2, headEvent3)
	headSpec.Clusters[headCluster] = struct{}{}

	specs := &SpecPullRequest{
		Base: baseSpec,
		Head: headSpec,
	}

	violations := ProcessEventConformanceComparison(specs)

	var violationEventNames []string
	for _, vs := range violations {
		for _, v := range vs {
			violationEventNames = append(violationEventNames, v.Entity.(*matter.Event).Name)
		}
	}

	if len(violationEventNames) != 1 {
		t.Fatalf("expected exactly 1 new violation, got %d: %v", len(violationEventNames), violationEventNames)
	}

	if violationEventNames[0] != "NewInvalidEvent" {
		t.Errorf("expected violation on NewInvalidEvent, got %s", violationEventNames[0])
	}
}

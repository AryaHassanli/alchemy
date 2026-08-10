package spec

import (
	"fmt"

	"github.com/project-chip/alchemy/matter"
	"github.com/project-chip/alchemy/matter/conformance"
	"github.com/project-chip/alchemy/matter/types"
)

func ValidateEventConformances(spec *Specification) (violations map[string][]Violation) {
	return compareEventConformances(nil, spec)
}

func ProcessEventConformanceComparison(specs *SpecPullRequest) (violations map[string][]Violation) {
	if specs == nil {
		return
	}
	v1 := compareEventConformances(specs.Base, specs.Head)
	v2 := compareEventConformances(specs.BaseInProgress, specs.HeadInProgress)
	violations = MergeViolations(v1, v2)
	return
}

func compareEventConformances(base *Specification, head *Specification) (violations map[string][]Violation) {
	violations = make(map[string][]Violation)
	if head == nil {
		return
	}
	for c := range head.Clusters {
		var isFeature func(id string) bool
		if c.Features != nil {
			isFeature = func(id string) bool {
				ent, ok := c.Features.Identifier(id)
				if !ok || ent == nil {
					return false
				}
				_, isFeat := ent.(*matter.Feature)
				return isFeat || ent.EntityType() == types.EntityTypeFeature
			}
		}
		baseCluster := findBaseCluster(base, c)
		for _, e := range c.Events {
			if err := conformance.ValidateEventConformance(e.Conformance, isFeature); err != nil {
				baseEvent := findBaseEvent(baseCluster, e)
				if baseEventHadViolation(baseCluster, baseEvent) {
					continue
				}
				v := Violation{
					Type:   ViolationEventConformance,
					Entity: e,
					Text:   fmt.Sprintf("%s (%s)", e.Conformance.ASCIIDocString(), err.Error()),
				}
				if e != nil {
					v.Path, v.Line = e.Origin()
				}
				violations[v.Path] = append(violations[v.Path], v)
			}
		}
	}
	for obj := range head.GlobalObjects {
		if e, ok := obj.(*matter.Event); ok {
			if err := conformance.ValidateEventConformance(e.Conformance, nil); err != nil {
				baseEvent := findBaseGlobalEvent(base, e)
				if baseEventHadViolation(nil, baseEvent) {
					continue
				}
				v := Violation{
					Type:   ViolationEventConformance,
					Entity: e,
					Text:   fmt.Sprintf("%s (%s)", e.Conformance.ASCIIDocString(), err.Error()),
				}
				if e != nil {
					v.Path, v.Line = e.Origin()
				}
				violations[v.Path] = append(violations[v.Path], v)
			}
		}
	}
	return
}

func findBaseCluster(base *Specification, headCluster *matter.Cluster) *matter.Cluster {
	if base == nil || headCluster == nil {
		return nil
	}
	for bc := range base.Clusters {
		if headCluster.ID.Valid() && bc.ID.Valid() && headCluster.ID.Equals(bc.ID) {
			return bc
		}
		if headCluster.Name == bc.Name {
			return bc
		}
	}
	return nil
}

func findBaseEvent(baseCluster *matter.Cluster, headEvent *matter.Event) *matter.Event {
	if baseCluster == nil || headEvent == nil {
		return nil
	}
	for _, be := range baseCluster.Events {
		if headEvent.ID.Valid() && be.ID.Valid() && headEvent.ID.Equals(be.ID) {
			return be
		}
		if headEvent.Name == be.Name {
			return be
		}
	}
	return nil
}

func findBaseGlobalEvent(base *Specification, headEvent *matter.Event) *matter.Event {
	if base == nil || headEvent == nil {
		return nil
	}
	for obj := range base.GlobalObjects {
		if be, ok := obj.(*matter.Event); ok {
			if headEvent.ID.Valid() && be.ID.Valid() && headEvent.ID.Equals(be.ID) {
				return be
			}
			if headEvent.Name == be.Name {
				return be
			}
		}
	}
	return nil
}

func baseEventHadViolation(baseCluster *matter.Cluster, baseEvent *matter.Event) bool {
	if baseEvent == nil {
		return false
	}
	var isFeature func(id string) bool
	if baseCluster != nil && baseCluster.Features != nil {
		isFeature = func(id string) bool {
			ent, ok := baseCluster.Features.Identifier(id)
			if !ok || ent == nil {
				return false
			}
			_, isFeat := ent.(*matter.Feature)
			return isFeat || ent.EntityType() == types.EntityTypeFeature
		}
	}
	return conformance.ValidateEventConformance(baseEvent.Conformance, isFeature) != nil
}

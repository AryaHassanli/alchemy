package spec

import (
	"errors"
	"fmt"

	"github.com/project-chip/alchemy/matter"
	"github.com/project-chip/alchemy/matter/conformance"
	"github.com/project-chip/alchemy/matter/types"
)

// ValidateEventConformances validates that all events in a Specification satisfy the Matter event conformance rules:
//  1. An event SHALL be mandatory ("M") or mandated by the use of one or more FeatureMap bits (e.g. "VIS", "VIS & AUD", "VIS | AUD").
//  2. Events SHALL be discoverable and therefore SHALL NOT be purely optional ("O") or conditionally optional (e.g. "[VIS]").
//  3. Conformance annotations (such as Provisional "P", Deprecated "D", Disallowed "X", Obsolete, Described) are ignored during validation.
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
			if err := validateEventConformance(e.Conformance, isFeature); err != nil {
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
			if err := validateEventConformance(e.Conformance, nil); err != nil {
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

// validateEventConformance checks a single Conformance object against the event conformance rules:
//   - Blank conformance is rejected.
//   - Conformance annotations (Provisional "P", Deprecated "D", Disallowed "X", Obsolete, Described) are filtered out.
//   - If no conformance elements remain after filtering annotations, the conformance is rejected.
//   - Purely optional ("O") or conditionally optional ("[...]") conformances are rejected.
//   - Mandatory without expression ("M") is accepted.
//   - Mandatory with expression (e.g. "VIS", "VIS & AUD") is recursively inspected to ensure every referenced
//     identifier is a valid FeatureMap bit and no non-feature entities (attributes, numbers, etc.) are present.
//   - Generic or unrecognized conformance structures are rejected.
func validateEventConformance(con conformance.Conformance, isFeature func(id string) bool) error {
	if con == nil || conformance.IsBlank(con) {
		return errors.New("conformance cannot be blank")
	}

	var nonAnnotations []conformance.Conformance
	switch c := con.(type) {
	case conformance.Set:
		nonAnnotations = make([]conformance.Conformance, 0, len(c))
		for _, el := range c {
			if !isAnnotation(el) {
				nonAnnotations = append(nonAnnotations, el)
			}
		}
	default:
		if !isAnnotation(con) {
			nonAnnotations = append(nonAnnotations, con)
		}
	}

	if len(nonAnnotations) == 0 {
		return errors.New("conformance must be mandatory or mandated by a feature bit")
	}

	for _, el := range nonAnnotations {
		switch c := el.(type) {
		case *conformance.Optional:
			if c.Expression == nil && c.Choice == nil {
				return errors.New("conformance is purely optional")
			}
			hasFeature, _ := inspectFeatures(c.Expression, isFeature)
			if hasFeature {
				return errors.New("conformance is tied to a feature bit but is optional")
			}
			return errors.New("conformance could be optional")

		case *conformance.Mandatory:
			if c.Expression == nil {
				continue
			}
			hasFeature, hasNonFeature := inspectFeatures(c.Expression, isFeature)
			if !hasFeature || hasNonFeature {
				return errors.New("conformance is not tied to a feature bit")
			}

		case *conformance.Generic:
			return fmt.Errorf("conformance expression is not valid: %s", c.ASCIIDocString())

		default:
			return errors.New("conformance is not tied to a feature bit")
		}
	}

	return nil
}

func isAnnotation(con conformance.Conformance) bool {
	switch con.(type) {
	case *conformance.Provisional, *conformance.Deprecated, *conformance.Disallowed, *conformance.Obsolete, *conformance.Described:
		return true
	}
	return false
}

func inspectFeatures(node any, isFeature func(id string) bool) (hasFeature, hasNonFeature bool) {
	if node == nil {
		return false, false
	}
	switch n := node.(type) {
	case *conformance.IdentifierExpression:
		return checkIdentifier(n.Entity, n.ID, n.Field, isFeature)
	case *conformance.ReferenceExpression:
		return checkIdentifier(n.Entity, n.Reference, n.Field, isFeature)
	case *conformance.IdentifierValue:
		return checkIdentifier(n.Entity, n.ID, n.Field, isFeature)
	case *conformance.ReferenceValue:
		return checkIdentifier(n.Entity, n.Reference, n.Field, isFeature)
	case *conformance.LogicalExpression:
		hasFeature, hasNonFeature = inspectFeatures(n.Left, isFeature)
		for _, r := range n.Right {
			rFeat, rNon := inspectFeatures(r, isFeature)
			hasFeature = hasFeature || rFeat
			hasNonFeature = hasNonFeature || rNon
		}
	case *conformance.EqualityExpression:
		lFeat, lNon := inspectFeatures(n.Left, isFeature)
		rFeat, rNon := inspectFeatures(n.Right, isFeature)
		return lFeat || rFeat, lNon || rNon
	case *conformance.ComparisonExpression:
		lFeat, lNon := inspectFeatures(n.Left, isFeature)
		rFeat, rNon := inspectFeatures(n.Right, isFeature)
		return lFeat || rFeat, lNon || rNon
	case *conformance.MathOperation:
		lFeat, lNon := inspectFeatures(n.Left, isFeature)
		rFeat, rNon := inspectFeatures(n.Right, isFeature)
		return lFeat || rFeat, lNon || rNon
	}
	return hasFeature, hasNonFeature
}

func checkIdentifier(entity types.Entity, id string, field conformance.ComparisonValue, isFeature func(string) bool) (hasFeature, hasNonFeature bool) {
	if (entity != nil && entity.EntityType() == types.EntityTypeFeature) || (isFeature != nil && isFeature(id)) {
		hasFeature = true
	} else {
		hasNonFeature = true
	}
	if field != nil {
		fFeat, fNon := inspectFeatures(field, isFeature)
		hasFeature = hasFeature || fFeat
		hasNonFeature = hasNonFeature || fNon
	}
	return hasFeature, hasNonFeature
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
	return validateEventConformance(baseEvent.Conformance, isFeature) != nil
}

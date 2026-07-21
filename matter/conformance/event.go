package conformance

import (
	"errors"
	"fmt"

	"github.com/project-chip/alchemy/matter/types"
)

// ValidateEventConformance validates that an event conformance expression meets the criteria:
// 1. The conformance for an event SHALL be mandatory or mandated by the use of one or more bits from a FeatureMap
//    which results in a valid FeatureCode conformance expression which can never be optional.
// 2. Events SHALL be discoverable and therefore SHALL NOT be purely or conditionally optional.
func ValidateEventConformance(con Conformance, isFeature func(id string) bool) error {
	if con == nil || IsBlank(con) {
		return errors.New("conformance cannot be blank")
	}

	var nonAnnotations []Conformance
	switch c := con.(type) {
	case Set:
		nonAnnotations = make([]Conformance, 0, len(c))
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
		case *Optional:
			if c.Expression == nil && c.Choice == nil {
				return errors.New("conformance is purely optional")
			}
			hasFeature, _ := inspectFeatures(c.Expression, isFeature)
			if hasFeature {
				return errors.New("conformance is tied to a feature bit but is optional")
			}
			return errors.New("conformance could be optional")

		case *Mandatory:
			if c.Expression == nil {
				continue
			}
			hasFeature, hasNonFeature := inspectFeatures(c.Expression, isFeature)
			if !hasFeature || hasNonFeature {
				return errors.New("conformance is not tied to a feature bit")
			}

		case *Generic:
			return fmt.Errorf("conformance expression is not valid: %s", c.ASCIIDocString())

		default:
			return errors.New("conformance is not tied to a feature bit")
		}
	}

	return nil
}

func isAnnotation(con Conformance) bool {
	switch con.(type) {
	case *Provisional, *Deprecated, *Disallowed, *Obsolete, *Described:
		return true
	}
	return false
}

func ExpressionReferencesFeature(exp Expression, lookup func(id string) bool) bool {
	hasFeature, _ := inspectFeatures(exp, lookup)
	return hasFeature
}

func inspectFeatures(node any, isFeature func(id string) bool) (hasFeature, hasNonFeature bool) {
	if node == nil {
		return false, false
	}
	switch n := node.(type) {
	case *IdentifierExpression:
		return checkIdentifier(n.Entity, n.ID, n.Field, isFeature)
	case *ReferenceExpression:
		return checkIdentifier(n.Entity, n.Reference, n.Field, isFeature)
	case *IdentifierValue:
		return checkIdentifier(n.Entity, n.ID, n.Field, isFeature)
	case *ReferenceValue:
		return checkIdentifier(n.Entity, n.Reference, n.Field, isFeature)
	case *LogicalExpression:
		hasFeature, hasNonFeature = inspectFeatures(n.Left, isFeature)
		for _, r := range n.Right {
			rFeat, rNon := inspectFeatures(r, isFeature)
			hasFeature = hasFeature || rFeat
			hasNonFeature = hasNonFeature || rNon
		}
	case *EqualityExpression:
		lFeat, lNon := inspectFeatures(n.Left, isFeature)
		rFeat, rNon := inspectFeatures(n.Right, isFeature)
		return lFeat || rFeat, lNon || rNon
	case *ComparisonExpression:
		lFeat, lNon := inspectFeatures(n.Left, isFeature)
		rFeat, rNon := inspectFeatures(n.Right, isFeature)
		return lFeat || rFeat, lNon || rNon
	case *MathOperation:
		lFeat, lNon := inspectFeatures(n.Left, isFeature)
		rFeat, rNon := inspectFeatures(n.Right, isFeature)
		return lFeat || rFeat, lNon || rNon
	}
	return hasFeature, hasNonFeature
}

func checkIdentifier(entity types.Entity, id string, field ComparisonValue, isFeature func(string) bool) (hasFeature, hasNonFeature bool) {
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

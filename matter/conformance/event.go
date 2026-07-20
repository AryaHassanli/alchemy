package conformance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/project-chip/alchemy/matter/types"
)

// ValidateEventConformance validates that an event conformance expression meets the criteria:
// 1. The conformance for an event SHALL be mandatory or mandated by the use of one or more bits from a FeatureMap
//    which results in a valid FeatureCode conformance expression which can never be optional.
// 2. Events SHALL be discoverable and therefore SHALL NOT be purely or conditionally optional.
func ValidateEventConformance(con Conformance, isFeature ...func(id string) bool) error {
	if con == nil || IsBlank(con) {
		return errors.New("conformance cannot be blank")
	}

	var lookup func(id string) bool
	if len(isFeature) > 0 {
		lookup = isFeature[0]
	}

	if isOptionalConformance(con) {
		if isPurelyOptional(con) {
			return errors.New("conformance is purely optional")
		}
		if hasFeatureReference(con, lookup) {
			return errors.New("conformance is tied to a feature bit but is optional")
		}
		return errors.New("conformance could be optional")
	}

	if IsMandatory(con) {
		return nil
	}

	var nonAnnotations []Conformance
	switch c := con.(type) {
	case Set:
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
		case *Mandatory:
			if c.Expression != nil {
				if !ExpressionReferencesFeature(c.Expression, lookup) {
					return errors.New("conformance is not tied to a feature bit")
				}
			}
		case *Optional:
			if c.Expression == nil {
				return errors.New("conformance is purely optional")
			}
			if ExpressionReferencesFeature(c.Expression, lookup) {
				return errors.New("conformance is tied to a feature bit but is optional")
			}
			return errors.New("conformance could be optional")
		case *Generic:
			return fmt.Errorf("conformance expression is not valid: %s", c.raw)
		default:
			if !hasFeatureReference(el, lookup) && !IsMandatory(el) {
				return errors.New("conformance is not tied to a feature bit")
			}
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

func isPurelyOptional(con Conformance) bool {
	if con == nil {
		return false
	}
	switch c := con.(type) {
	case *Optional:
		return c.Expression == nil && c.Choice == nil
	case Set:
		for _, el := range c {
			switch el := el.(type) {
			case *Provisional, *Deprecated, *Disallowed, *Obsolete, *Described:
				continue
			case *Optional:
				return el.Expression == nil && el.Choice == nil
			default:
				return false
			}
		}
	}
	return false
}

func isOptionalConformance(con Conformance) bool {
	switch c := con.(type) {
	case *Optional:
		return true
	case *Generic:
		if strings.Contains(c.raw, "[") || (strings.Contains(c.raw, "O") && !strings.Contains(c.raw, "Optional")) {
			return true
		}
	case Set:
		for _, el := range c {
			if isOptionalConformance(el) {
				return true
			}
		}
	}
	return false
}

func hasFeatureReference(con Conformance, lookup func(id string) bool) bool {
	switch c := con.(type) {
	case *Mandatory:
		if c.Expression != nil {
			return ExpressionReferencesFeature(c.Expression, lookup)
		}
	case *Optional:
		if c.Expression != nil {
			return ExpressionReferencesFeature(c.Expression, lookup)
		}
	case *Generic:
		if lookup != nil {
			for _, word := range strings.FieldsFunc(c.raw, func(r rune) bool {
				return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_')
			}) {
				if lookup(word) {
					return true
				}
			}
		}
	case Set:
		for _, el := range c {
			if hasFeatureReference(el, lookup) {
				return true
			}
		}
	}
	return false
}

func ExpressionReferencesFeature(exp Expression, lookup func(id string) bool) bool {
	if exp == nil {
		return false
	}
	switch e := exp.(type) {
	case *IdentifierExpression:
		if e.Entity != nil && e.Entity.EntityType() == types.EntityTypeFeature {
			return true
		}
		if lookup != nil && lookup(e.ID) {
			return true
		}
		if e.Field != nil {
			return ComparisonValueReferencesFeature(e.Field, lookup)
		}
	case *ReferenceExpression:
		if e.Entity != nil && e.Entity.EntityType() == types.EntityTypeFeature {
			return true
		}
		if lookup != nil && lookup(e.Reference) {
			return true
		}
		if e.Field != nil {
			return ComparisonValueReferencesFeature(e.Field, lookup)
		}
	case *LogicalExpression:
		if ExpressionReferencesFeature(e.Left, lookup) {
			return true
		}
		for _, r := range e.Right {
			if ExpressionReferencesFeature(r, lookup) {
				return true
			}
		}
	case *EqualityExpression:
		return ExpressionReferencesFeature(e.Left, lookup) || ExpressionReferencesFeature(e.Right, lookup)
	case *ComparisonExpression:
		return ComparisonValueReferencesFeature(e.Left, lookup) || ComparisonValueReferencesFeature(e.Right, lookup)
	}
	return false
}

func ComparisonValueReferencesFeature(val ComparisonValue, lookup func(id string) bool) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case *IdentifierValue:
		if v.Entity != nil && v.Entity.EntityType() == types.EntityTypeFeature {
			return true
		}
		if lookup != nil && lookup(v.ID) {
			return true
		}
	case *ReferenceValue:
		if v.Entity != nil && v.Entity.EntityType() == types.EntityTypeFeature {
			return true
		}
		if lookup != nil && lookup(v.Reference) {
			return true
		}
	case *MathOperation:
		return ComparisonValueReferencesFeature(v.Left, lookup) || ComparisonValueReferencesFeature(v.Right, lookup)
	}
	return false
}

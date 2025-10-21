package mattergen

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/project-chip/alchemy/internal/text"
	"github.com/project-chip/alchemy/matter"
	"github.com/project-chip/alchemy/matter/constraint"

	"github.com/project-chip/alchemy/zap"
)

func writeClusterStructs(sb *strings.Builder, cluster *matter.Cluster) {
	if len(cluster.Structs) == 0 {
		return
	}

	sb.WriteString("\n")
	for _, s := range cluster.Structs {
		writeStruct(sb, cluster, s)
	}
}

func writeStruct(sb *strings.Builder, cluster *matter.Cluster, s *matter.Struct) {
	if s == nil {
		slog.Warn("Skipping nil struct in cluster", "clusterName", cluster.Name)
		return
	}

	structNameAlphanumeric := text.Alphanumeric(s.Name)
	if structNameAlphanumeric == "" {
		slog.Warn("Skipping struct: Sanitized name is empty", "originalName", s.Name, "clusterName", cluster.Name)
		return
	}

	sb.WriteString(fmt.Sprintf("\t struct %s {\n", structNameAlphanumeric))

	for _, field := range s.Fields {
		writeField(sb, cluster, s, field)
	}
	sb.WriteString("\t }\n")

}

func writeField(sb *strings.Builder, cluster *matter.Cluster, s *matter.Struct, field *matter.Field) {
	if field == nil || field.ID == nil || field.Type == nil {
		slog.Warn("Skipping invalid field in struct", "structName", s.Name, "clusterName", cluster.Name)
		return
	}

	fieldName := field.Name
	fieldID := field.ID.IntString()
	fieldType := getMatterFieldType(s, field)

	arrayMarker := ""
	if field.Type.IsArray() {
		arrayMarker = "[]"
	}

	sb.WriteString(fmt.Sprintf("\t\t %s %s%s = %s; \n", fieldType, fieldName, arrayMarker, fieldID))

}

func getMatterFieldType(s *matter.Struct, field *matter.Field) string {
	if s.Name == "NOCStruct" {
		fmt.Println("d")
	}

	fieldType := field.Type.Name
	lowerFieldType := strings.ToLower(fieldType)
	if zapType, ok := zap.MatterToZapMap[lowerFieldType]; ok {
		fieldType = zapType
	}
	if field.Type.HasLength() {
		if cset, ok := field.Constraint.(constraint.Set); ok {
			if len(cset) > 1 {
				slog.Warn("This field has more than one constraint", "field", field.Name)
				return fieldType
			}
			cc := matter.NewConstraintContext(field, s.Fields)
			fieldType = fmt.Sprintf("%s<%d>", fieldType, cset[0].Max(cc).Int64)
		}
	}
	return fieldType
}

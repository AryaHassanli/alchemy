package mattergen

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/project-chip/alchemy/internal/text"
	"github.com/project-chip/alchemy/matter"
)

func writeClusterEnums(sb *strings.Builder, cluster *matter.Cluster) {
	if len(cluster.Enums) == 0 {
		return
	}

	sb.WriteString("\n")
	for _, enum := range cluster.Enums {
		if enum == nil || enum.Type == nil {
			slog.Warn("Skipping invalid enum in cluster", "clusterName", cluster.Name)
			continue
		}

		enumNameAlphanumeric := text.Alphanumeric(enum.Name)
		if enumNameAlphanumeric == "" {
			slog.Warn("Skipping enum: Sanitized name is empty", "originalName", enum.Name, "clusterName", cluster.Name)
			continue
		}

		sb.WriteString(fmt.Sprintf("\t enum %s : %s {\n", enumNameAlphanumeric, enum.Type.Name))

		for _, val := range enum.Values {
			writeClusterEnumValue(sb, cluster, enum, val)
		}
		sb.WriteString("\t }\n")
	}
}

func writeClusterEnumValue(sb *strings.Builder, cluster *matter.Cluster, enum *matter.Enum, val *matter.EnumValue) {
	if val == nil || val.Value == nil {
		slog.Warn("Skipping invalid enum value in enum", "enumName", enum.Name, "clusterName", cluster.Name)
		return
	}

	valNameAlphanumeric := text.Alphanumeric(val.Name)
	if valNameAlphanumeric == "" {
		slog.Warn("Skipping enum value: Sanitized name is empty", "originalName", val.Name, "enumName", enum.Name)
		return
	}

	kName := "k" + valNameAlphanumeric
	valueStr := val.Value.IntString()

	if valNameAlphanumeric != val.Name {
		sb.WriteString(fmt.Sprintf("\t\t %s = %s [spec_name = \"%s\"];\n", kName, valueStr, val.Name))
	} else {
		sb.WriteString(fmt.Sprintf("\t\t %s = %s;\n", kName, valueStr))
	}

}

package mattergen

import (
	"fmt"
	"strings"

	"github.com/project-chip/alchemy/matter"
)

func writeClusterHeader(sb *strings.Builder, cluster *matter.Cluster, nameAlphanumeric string) {
	sb.WriteString(fmt.Sprintf("cluster %s = %s {\n", nameAlphanumeric, cluster.ID.IntString()))
}

func writeClusterFooter(sb *strings.Builder) {
	sb.WriteString("}\n")
}

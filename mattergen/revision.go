package mattergen

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/project-chip/alchemy/matter"
)

func writeClusterRevision(sb *strings.Builder, cluster *matter.Cluster) {
	revCount := len(cluster.Revisions)
	if revCount > 0 && cluster.Revisions[revCount-1] != nil && cluster.Revisions[revCount-1].Number != "" {
		sb.WriteString(fmt.Sprintf("\t revision %s;\n", cluster.Revisions[revCount-1].Number))
	} else {
		slog.Debug("Cluster has no valid last revision number, skipping revision line.", "clusterName", cluster.Name)
	}
}

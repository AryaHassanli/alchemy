package mattergen

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/project-chip/alchemy/asciidoc"
	"github.com/project-chip/alchemy/internal/pipeline"
	"github.com/project-chip/alchemy/internal/text"
	"github.com/project-chip/alchemy/matter/spec"
)

func Pipeline(cxt context.Context, specRoot string, pipelineOptions pipeline.ProcessingOptions) (err error) {
	var specification *spec.Specification

	parserOptions := spec.ParserOptions{Root: specRoot}
	pipelineOptions.Serial = true // TODO: Remove
	specification, _, err = spec.Parse(cxt, parserOptions, pipelineOptions, nil, []asciidoc.AttributeName{})

	if err != nil {
		return
	}

	// to be changed with an output folder arg in cli
	const outDir = "Z_TEST_OUT"
	if err = os.MkdirAll(outDir, 0755); err != nil {
		slog.Error("Failed to create output directory", "directory", outDir, "error", err)
		return
	}

	for cluster := range specification.Clusters {
		nameAlphanumeric := text.Alphanumeric(cluster.Name)
		filename := filepath.Join(outDir, nameAlphanumeric+".matter")

		var sb strings.Builder

		writeClusterHeader(&sb, cluster, nameAlphanumeric)
		writeClusterRevision(&sb, cluster)
		writeClusterEnums(&sb, cluster)
		writeClusterBitmaps(&sb, cluster)
		writeClusterStructs(&sb, cluster)

		// You can add calls to writeAttributes, writeCommands, etc. here

		writeClusterFooter(&sb)

		err := os.WriteFile(filename, []byte(sb.String()), 0644)
		if err != nil {
			slog.Error("Failed to write .matter file", "clusterName", cluster.Name, "filename", filename, "error", err)
			continue
		} else {
			slog.Info("Successfully wrote .matter file", "filename", filename)
		}
	}
	return
}

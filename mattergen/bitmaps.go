package mattergen

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/project-chip/alchemy/internal/text"
	"github.com/project-chip/alchemy/matter"
	"github.com/project-chip/alchemy/zap"
)

func writeClusterBitmaps(sb *strings.Builder, cluster *matter.Cluster) {
	if len(cluster.Bitmaps) == 0 {
		return
	}

	sb.WriteString("\n")
	for _, bitmap := range cluster.Bitmaps {
		writeBitmap(sb, cluster, bitmap)
	}
}

func writeBitmap(sb *strings.Builder, cluster *matter.Cluster, bitmap *matter.Bitmap) {
	bitmapNameAlphanumeric := text.Alphanumeric(bitmap.Name)
	if bitmapNameAlphanumeric == "" {
		slog.Warn("Skipping bitmap: Sanitized name is empty", "originalName", bitmap.Name, "clusterName", cluster.Name)
		return
	}

	sb.WriteString(fmt.Sprintf("\t bitmap %s : %s {\n", bitmapNameAlphanumeric, zap.MatterToZapMap[bitmap.Type.Name]))

	for _, bit := range bitmap.Bits {
		writeBit(sb, cluster, bitmap, bit)
	}
	sb.WriteString("\t }\n")
}

func writeBit(sb *strings.Builder, cluster *matter.Cluster, bitmap *matter.Bitmap, bit matter.Bit) {
	bitNameAlphanumeric := text.Alphanumeric(bit.Name())
	if bitNameAlphanumeric == "" {
		slog.Warn("Skipping bitmap value: Sanitized name is empty", "originalName", bit.Name(), "bitmapName", bitmap.Name)
		return
	}

	kName := "k" + bitNameAlphanumeric
	mask, err := bit.Mask()
	if err != nil {
		slog.Warn("Skipping bitmap value: Mask is not valid for ", "bitmapName", bitmap.Name, "bitName", bit.Name())
		return
	}

	valueStr := fmt.Sprintf("0x%X", mask)

	if bitNameAlphanumeric != bit.Name() {
		sb.WriteString(fmt.Sprintf("\t\t %s = %s [spec_name = \"%s\"];\n", kName, valueStr, bit.Name()))
	} else {
		sb.WriteString(fmt.Sprintf("\t\t %s = %s;\n", kName, valueStr))
	}

}

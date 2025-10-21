package cli

import (
	"github.com/project-chip/alchemy/internal/pipeline"
	"github.com/project-chip/alchemy/mattergen"
)

type MatterGen struct {
	pipeline.ProcessingOptions `embed:""`

	SpecRoot string `name:"spec-root" default:"connectedhomeip-spec" help:"the src root of your clone of CHIP-Specifications/connectedhomeip-spec"`
}

func (c *MatterGen) Run(cc *Context) (err error) {
	err = mattergen.Pipeline(cc, c.SpecRoot, c.ProcessingOptions)
	return
}

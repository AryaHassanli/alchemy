package dm

import (
	"fmt"
	"slices"
	"strings"

	"github.com/beevik/etree"
	"github.com/project-chip/alchemy/matter"
	"github.com/project-chip/alchemy/matter/conformance"
)

func renderEvents(es matter.EventSet, c *etree.Element) (err error) {
	if len(es) == 0 {
		return
	}
	evs := make([]*matter.Event, 0, len(es))
	for _, e := range es {
		if conformance.IsZigbee(e.Conformance) {
			continue
		}
		evs = append(evs, e)
	}

	/* TODO: Remove. Temporary workaround to output optional events */
	for _, e := range es {
		if conformance.IsOptional(e.Conformance) {
			fmt.Printf("<%s,%s,%s,%s>\n",
				e.Parent().(*matter.Cluster).ID.HexString(),
				e.Parent().(*matter.Cluster).Name,
				e.ID.HexString(),
				e.Name)
		}
	}
	/* END OF TODO: Remove. Temporary workaround to output optional events */

	slices.SortStableFunc(evs, func(a, b *matter.Event) int {
		return a.ID.Compare(b.ID)
	})
	events := c.CreateElement("events")
	for _, e := range evs {

		cx := events.CreateElement("event")
		cx.CreateAttr("id", e.ID.ShortHexString())
		cx.CreateAttr("name", e.Name)
		if len(e.Priority) > 0 {
			cx.CreateAttr("priority", strings.ToLower(e.Priority))
		}

		if e.Access.Read != matter.PrivilegeUnknown || e.Access.IsFabricSensitive() {
			a := cx.CreateElement("access")
			a.CreateAttr("readPrivilege", strings.ToLower(matter.PrivilegeNamesShort[e.Access.Read]))
			if e.Access.IsFabricSensitive() {
				a.CreateAttr("fabricSensitive", "true")
			}
		}
		err = renderConformanceElement(e.Conformance, cx, e)
		if err != nil {
			return
		}

		err = renderFields(e.Fields, cx, e)
		if err != nil {
			return
		}
	}

	return
}

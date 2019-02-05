package zepository

import (
	"fmt"
	"manala/pkg2/zemplate"
)

/***********/
/* Manager */
/***********/

type manager struct {
	references []ReferenceInterface
}

func (m *manager) LoadTemplate(refs ...zemplate.ReferenceInterface) {
	for _, ref := range refs {
		fmt.Printf("%#v\n", ref.GetTemplateName())
	}
}

func NewManager(refs ...ReferenceInterface) *manager {
	r := references

	for _, ref := range refs {
		r = append(r, ref)
	}

	return &manager{
		references: r,
	}
}

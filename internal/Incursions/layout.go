package incursions

type IncursionLayout struct {
	StagingSystem   NamedItem
	HQSystem        NamedItem
	VanguardSystems []NamedItem
	AssaultSystems  []NamedItem
}

func (layout *IncursionLayout) IsComplete() bool {
	return (layout.StagingSystem.Name != "" &&
		layout.HQSystem.Name != "" &&
		len(layout.VanguardSystems) > 0 &&
		len(layout.AssaultSystems) > 0)
}

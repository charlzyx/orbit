package workbench

// UsageModels returns the source's first-seen model order. Legacy or manually
// constructed indexes have no catalog; retain their available record order as
// a deterministic fallback, never invent lexicographic source ordinals.
func (i UsageIndex) UsageModels() []UsageModelRef {
	var out []UsageModelRef
	seen := map[UsageModelRef]bool{}
	add := func(ref UsageModelRef) {
		if !seen[ref] {
			seen[ref] = true
			out = append(out, ref)
		}
	}
	for _, ref := range i.ModelCatalog {
		add(ref)
	}
	for _, r := range i.Records {
		add(UsageModelRef{Provider: r.Provider, Model: r.Model})
	}
	for _, r := range i.Errors {
		add(UsageModelRef{Provider: r.Provider, Model: r.Model})
	}
	for _, r := range i.Tools {
		add(UsageModelRef{Provider: r.Provider, Model: r.Model})
	}
	return out
}

func (i UsageIndex) UsageModelOrder() map[UsageModelRef]int {
	out := map[UsageModelRef]int{}
	for ordinal, ref := range i.UsageModels() {
		out[ref] = ordinal
	}
	return out
}

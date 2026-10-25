package orbitui

func (x *Controller) setRPCPatches(enabled bool) {
	if x.UI.rpcPatchSaving {
		return
	}
	x.UI.rpcPatchSaving = true
	x.UI.Async(func() func() {
		err := x.Agent.SetRPCPatchesEnabled(enabled)
		return func() {
			x.UI.rpcPatchSaving = false
			if err != nil {
				x.UI.save(err)
				return
			}
			x.UI.Model.Extensions.Patches.Enabled = enabled
			if enabled {
				x.UI.Notify("RPC patches enabled — restart the agent to load them", false)
			} else {
				x.UI.Notify("RPC patches disabled — pi keeps the current patches until it updates", false)
			}
		}
	})
}

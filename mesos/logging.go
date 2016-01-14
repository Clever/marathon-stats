package mesos

import (
	kv "gopkg.in/Clever/kayvee-go.v2/logger"
)

var kvlog = kv.New("marathon-stats")

func LogState(state State) {
	stateLog := kv.M{
		"activated_slaves":   state.ActivatedSlaves,
		"deactivated_slaves": state.DeactivatedSlaves,
		"finished_tasks":     state.FinishedTasks,
		"hostname":           state.Hostname,
		"id":                 state.ID,
		"killed_tasks":       state.KilledTasks,
		"lost_tasks":         state.LostTasks,
		"staged_tasks":       state.StagedTasks,
		"started_tasks":      state.StartedTasks,
		"version":            state.Version,
		"start_time":         state.StartTime,
	}
	// add up resource counts from slaves
	totalCPU := 0.0
	totalMem := 0.0
	for _, slave := range state.Slaves {
		r := slave.Resources
		totalCPU += r.CPUs
		totalMem += r.Mem
	}

	// actual used/offer amounts come from the framework(s)
	usedCPU := 0.0
	usedMem := 0.0
	offeredCPU := 0.0
	offeredMem := 0.0
	for _, framework := range state.Frameworks {
		for _, offer := range framework.Offers {
			offeredCPU += offer.CPUs
			offeredMem += offer.Mem
		}
		usedCPU += framework.Resources.CPUs
		usedMem += framework.Resources.Mem
	}
	stateLog["used_cpu"] = usedCPU
	stateLog["used_mem"] = usedMem
	stateLog["offered_cpu"] = offeredCPU
	stateLog["offered_mem"] = offeredMem
	stateLog["total_cpu"] = totalCPU
	stateLog["total_mem"] = totalMem

	kvlog.GaugeIntD("mesos-state", state.ActivatedSlaves, stateLog)
}

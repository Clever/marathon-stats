package mesos

import (
	"log"

	"gopkg.in/clever/kayvee-go.v2"
)

type m map[string]interface{}

func LogState(state State) {
	stateLog := m{
		"source":             "marathon-stats",
		"title":              "mesos-state",
		"level":              kayvee.Info,
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
		"type":               "gauge", // This is to auto load metric into influx
	}
	// add up resource counts from slaves
	totalCPU := 0.0
	totalMem := int64(0)
	for _, slave := range state.Slaves {
		r := slave.Resources
		totalCPU += r.CPUs
		totalMem += r.Mem
	}

	// actual used/offer amounts come from the framework(s)
	usedCPU := 0.0
	usedMem := int64(0)
	offeredCPU := 0.0
	offeredMem := int64(0)
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

	log.Println(kayvee.Format(stateLog))
}

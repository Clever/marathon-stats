package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Clever/marathon-stats/marathon"
	"github.com/Clever/marathon-stats/mesos"
	"gopkg.in/clever/kayvee-go.v2"
)

type m map[string]interface{}

var (
	mesosMasterPort    string
	mesosMasterHost    string
	marathonMasterPort string
	marathonMasterHost string
	logMarathonTasks   bool
	pollInterval       time.Duration
	version            string
)

func init() {
	mesosMasterHost = getEnv("MESOS_HOST")
	mesosMasterPort = getEnv("MESOS_PORT")
	marathonMasterHost = getEnv("MARATHON_HOST")
	marathonMasterPort = getEnv("MARATHON_PORT")
	var err error
	logMarathonTasks, err = strconv.ParseBool(getEnv("LOG_MARATHON_TASKS"))
	if err != nil {
		log.Fatalf("Could not parse bool from LOG_MARATHON_TASKS: %s", err)
	}
	pollIntervalRaw := getEnv("POLL_INTERVAL")
	if pollIntervalRaw == "" {
		pollIntervalRaw = "5s"
	}
	pollInterval, err = time.ParseDuration(pollIntervalRaw)
	if err != nil {
		log.Fatalf("Could not parse duration from POLL_INTERVAL: %s", err)
	}
}

func main() {
	marathonClient := marathon.NewClient(fmt.Sprintf("%s:%s", marathonMasterHost, marathonMasterPort))
	mesosClient := mesos.NewClient(fmt.Sprintf("%s:%s", mesosMasterHost, mesosMasterPort))
	log.Println(kayvee.Format(m{
		"source":             "marathon-stats",
		"title":              "startup",
		"level":              kayvee.Info,
		"version":            version,
		"mesosMasterPort":    mesosMasterPort,
		"mesosMasterHost":    mesosMasterHost,
		"marathonMasterPort": marathonMasterPort,
		"marathonMasterHost": marathonMasterHost,
		"logMarathonTasks":   logMarathonTasks,
		"pollInterval":       pollInterval,
	}))

	ticker := time.Tick(pollInterval)
	for _ = range ticker {
		apps, err := marathonClient.GetApps()
		if err != nil {
			log.Fatal(err)
		}
		logMarathonState(apps.Apps, logMarathonTasks)

		state, err := mesosClient.GetState()
		if err != nil {
			log.Fatal(err)
		}
		logMesosState(state)
	}
}

func getEnv(envVar string) string {
	val := os.Getenv(envVar)
	if val == "" {
		log.Fatalf("Must specify env variable %s", envVar)
	}
	return val
}

func logMesosState(state mesos.State) {
	stateLog := m{
		"source":         "marathon-stats",
		"title":          "mesos-state",
		"level":          kayvee.Info,
		"finished_tasks": state.FinishedTasks,
		"hostname":       state.Hostname,
		"id":             state.ID,
		"killed_tasks":   state.KilledTasks,
		"lost_tasks":     state.LostTasks,
		"staged_tasks":   state.StagedTasks,
		"started_tasks":  state.StartedTasks,
		"version":        state.Version,
		"start_time":     state.StartTime,
		"type":           "gauge", // This is to auto load metric into influx
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

func logMarathonState(apps []marathon.App, logMarathonTasks bool) {
	var summary struct {
		appCount     int64
		runningCount int64
		stagedCount  int64
	}
	for _, app := range apps {
		summary.appCount++
		summary.runningCount += app.TasksRunning
		summary.stagedCount += app.TasksStaged
		logMarathonApp(app, logMarathonTasks)
	}
	summaryLog := m{
		"source":       "marathon-stats",
		"title":        "marathon-summary",
		"totalApps":    summary.appCount,
		"runningTasks": summary.runningCount,
		"stagedTasks":  summary.stagedCount,
		"type":         "gauge",
	}
	log.Println(kayvee.Format(summaryLog))
}

func logMarathonApp(app marathon.App, logMarathonTasks bool) {
	normalizedAppID := normalizeMarathonAppID(app.ID)
	appLog := m{
		"source":         "marathon-stats",
		"title":          fmt.Sprintf("marathon-apps.%s", normalizedAppID),
		"level":          kayvee.Info,
		"appId":          app.ID,
		"instances":      app.Instances,
		"cpus":           app.Cpus,
		"mem":            app.Mem,
		"disk":           app.Disk,
		"image":          app.Container.Docker.Image,
		"version":        app.Version,
		"tasksStaged":    app.TasksStaged,
		"tasksHealthy":   app.TasksHealthy,
		"tasksRunning":   app.TasksRunning,
		"tasksUnhealthy": app.TasksUnhealthy,
		"type":           "gauge", // This is to auto load metric into influx
	}
	log.Println(kayvee.Format(appLog))

	if !logMarathonTasks {
		return
	}

	for _, task := range app.Tasks {
		taskLog := m{
			"source":    "marathon-stats",
			"title":     fmt.Sprintf("marathon-tasks.%s", task.ID),
			"level":     kayvee.Info,
			"appId":     task.AppID,
			"id":        task.ID,
			"startedAt": task.StartedAt,
			"stagedAt":  task.StagedAt,
			"version":   task.Version,
			"type":      "gauge", // This is to auto load metric into influx
		}
		log.Println(kayvee.Format(taskLog))
	}
}

func normalizeMarathonAppID(appID string) string {
	return strings.Replace(appID[1:], "/", ".", -1)
}

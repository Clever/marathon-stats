package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
		marathon.LogState(apps.Apps, logMarathonTasks)

		state, err := mesosClient.GetState()
		if err != nil {
			log.Fatal(err)
		}
		mesos.LogState(state)
	}
}

func getEnv(envVar string) string {
	val := os.Getenv(envVar)
	if val == "" {
		log.Fatalf("Must specify env variable %s", envVar)
	}
	return val
}

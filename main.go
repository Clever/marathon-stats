package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
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
	machineHost        string
	pushListenerPort   string
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

	machineHost = getEnv("HOST")
	pushListenerPort = getEnv("PUSH_LISTENER_PORT")
}

func main() {
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
		"machineHost":        machineHost,
		"pushListenerPort":   pushListenerPort,
	}))

	var wg sync.WaitGroup
	wg.Add(2)

	go initPollClients(&wg)
	go initPushListeners(&wg)

	wg.Wait()
}

func initPollClients(wg *sync.WaitGroup) {
	marathonClient := marathon.NewClient(fmt.Sprintf("%s:%s", marathonMasterHost, marathonMasterPort))
	mesosClient := mesos.NewClient(fmt.Sprintf("%s:%s", mesosMasterHost, mesosMasterPort))

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

	wg.Done()
}

func initPushListeners(wg *sync.WaitGroup) {
	listener := marathon.NewListener(machineHost, pushListenerPort)
	err := listener.Subscribe(fmt.Sprintf("%s:%s", marathonMasterHost, marathonMasterPort))
	if err != nil {
		log.Fatal(err)
	}

	for event := range listener.Events() {
		//marathon.LogEvent(event)
		fmt.Println("event: ", event)
	}

	wg.Done()
}

func getEnv(envVar string) string {
	val := os.Getenv(envVar)
	if val == "" {
		log.Fatalf("Must specify env variable %s", envVar)
	}
	return val
}

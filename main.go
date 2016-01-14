package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Clever/aws-cost-notifier/pricer"
	"github.com/Clever/pathio"
	kv "gopkg.in/Clever/kayvee-go.v2/logger"

	"github.com/Clever/marathon-stats/cost"
	"github.com/Clever/marathon-stats/marathon"
	"github.com/Clever/marathon-stats/mesos"
)

var kvlog = kv.New("marathon-stats")

var (
	mesosMasterPort    string
	mesosMasterHost    string
	marathonMasterPort string
	marathonMasterHost string
	lastRanS3Path      string
	logMarathonTasks   bool
	pollInterval       time.Duration
	containerHost      string
	externalPort       string // Internal/external ports are relative
	internalPort       string // to the container this process runs in.
	version            string
)

func init() {
	mesosMasterHost = getEnv("MESOS_HOST")
	mesosMasterPort = getEnv("MESOS_PORT")
	marathonMasterHost = getEnv("MARATHON_HOST")
	marathonMasterPort = getEnv("MARATHON_PORT")
	lastRanS3Path = getEnv("LAST_RAN_S3_PATH")
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

	containerHost = getEnv("HOST")
	externalPort = getEnv("PORT")
	internalPort = "80" // Matches value in launch/marathon-stat.yml
}

func main() {
	kvlog.InfoD("startup", kv.M{
		"version":            version,
		"mesosMasterPort":    mesosMasterPort,
		"mesosMasterHost":    mesosMasterHost,
		"marathonMasterPort": marathonMasterPort,
		"marathonMasterHost": marathonMasterHost,
		"logMarathonTasks":   logMarathonTasks,
		"pollInterval":       pollInterval,
		"containerHost":      containerHost,
		"externalPort":       externalPort,
		"internalPort":       internalPort,
	})

	var wg sync.WaitGroup
	wg.Add(2)

	go initPollClients(&wg)
	go initPushListeners(&wg)

	wg.Wait()
}

func fetchLastRunTime(s3Path string) (time.Time, error) {
	reader, err := pathio.Reader(lastRanS3Path)
	if err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(reader)

	return time.Parse(time.RFC3339, buf.String())
}

func initPollClients(wg *sync.WaitGroup) {
	lastRan, err := fetchLastRunTime(lastRanS3Path)
	if err != nil {
		log.Fatal(err)
	}

	prices, err := pricer.FromFile("./data/aws_prices.csv")
	if err != nil {
		log.Fatal(err)
	}
	costLogger := cost.NewLogger(prices, lastRan)

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

		costLogger.LogCost(apps.Apps, state)
		pathio.Write(lastRanS3Path, []byte(costLogger.LastRan().Format(time.RFC3339)))
	}

	wg.Done()
}

func initPushListeners(wg *sync.WaitGroup) {
	listener := marathon.NewListener(containerHost, internalPort, externalPort)
	err := listener.Subscribe(fmt.Sprintf("%s:%s", marathonMasterHost, marathonMasterPort))
	if err != nil {
		log.Fatal(err)
	}

	for event := range listener.Events() {
		marathon.LogEvent(event)
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

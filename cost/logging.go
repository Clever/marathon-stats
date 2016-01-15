package cost

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Clever/aws-cost-notifier/pricer"
	kv "gopkg.in/Clever/kayvee-go.v2/logger"

	"github.com/Clever/marathon-stats/marathon"
	"github.com/Clever/marathon-stats/mesos"
)

var kvlog = kv.New("marathon-stats")

type m map[string]interface{}

type logger struct {
	prices  pricer.Pricer
	lastRan time.Time
}

type CostLogger interface {
	LogCost(apps []marathon.App, state mesos.State)
	LastRan() time.Time
}

type clusterStats struct {
	MemberCount int
	Members     map[string]int
	PricePerHr  float64
	Mem         float64
	CPUs        float64
	Disk        int
}

type memberStats struct {
	InstanceType string
	PricePerHr   float64
	mesos.Resources
}

func NewLogger(prices pricer.Pricer, lastRan time.Time) CostLogger {
	return &logger{prices: prices, lastRan: lastRan}
}

func (l *logger) LastRan() time.Time {
	return l.lastRan
}

func (l *logger) LogCost(apps []marathon.App, state mesos.State) {
	cluster, memStats := l.generateClusterStats(state)

	now := time.Now()
	billableDuration := now.Sub(l.LastRan())
	for _, app := range apps {
		for _, task := range app.Tasks {
			host := memStats[task.Host]
			l.logContainer(app, task, host, cluster, billableDuration)
		}
	}
	l.lastRan = now
}

func (l *logger) logContainer(
	app marathon.App, task marathon.Task, host memberStats, cluster clusterStats,
	billableDuration time.Duration,
) {
	members, err := json.Marshal(cluster.Members)
	if err != nil {
		kvlog.Error(fmt.Sprintf("Failed to marshal members json: %+#v", cluster.Members))
		members = []byte{}
	}

	startTime, err := time.Parse(time.RFC3339, task.StartedAt)
	if err != nil {
		kvlog.Error("Failed to parse time: " + task.StartedAt)
		startTime = time.Now()
	}

	kvlog.InfoD("container-cost", kv.M{
		"task_id":            task.ID,
		"creator":            app.Labels["creator"],
		"name":               app.ID,
		"environment":        app.Labels["env"],
		"version":            app.Labels["version"],
		"application":        app.Labels["application"],
		"new_billable_hours": billableDuration.Hours(),

		"container_created": startTime.Format("2006-01-02 15:04:05"),
		"container_mem":     app.Mem,
		"container_cpus":    app.Cpus,
		"container_disk":    app.Disk,

		"host_instance_type": host.InstanceType,
		"host_price_per_hr":  host.PricePerHr,
		"host_cpus":          host.CPUs,
		"host_mem":           host.Mem,
		"host_disk":          host.Disk,

		"cluster":              string(members),
		"cluster_price_per_hr": cluster.PricePerHr,
		"cluster_mem":          cluster.Mem,
		"cluster_cpus":         cluster.CPUs,
		"cluster_disk":         cluster.Disk,
		"cluster_member_count": cluster.MemberCount,
	})
}

func (l *logger) generateClusterStats(state mesos.State) (clusterStats, map[string]memberStats) {
	cluStats := clusterStats{
		MemberCount: state.ActivatedSlaves,
		Members:     map[string]int{},
		PricePerHr:  0.0,
		Mem:         0.0,
		CPUs:        0.0,
		Disk:        0,
	}
	memStats := map[string]memberStats{}

	for _, member := range state.Slaves {
		if !member.Active {
			continue
		}

		instanceType := strings.TrimPrefix(member.Attributes["instance_type"], "mesos.")
		cluStats.Members[instanceType] += 1

		perHr, err := l.prices.PerHr(instanceType)
		if err != nil {
			kvlog.Error(err.Error())
		}
		cluStats.PricePerHr += perHr

		cluStats.Mem += member.Resources.Mem
		cluStats.CPUs += member.Resources.CPUs
		cluStats.Disk += member.Resources.Disk

		memStats[member.Hostname] = memberStats{instanceType, perHr, member.Resources}
	}

	return cluStats, memStats
}

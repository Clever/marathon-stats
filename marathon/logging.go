package marathon

import (
	"fmt"
	"strings"

	kv "gopkg.in/Clever/kayvee-go.v2/logger"
)

var kvlog = kv.New("marathon-stats")

func LogState(apps []App, logMarathonTasks bool) {
	var summary struct {
		appCount     int
		runningCount int
		stagedCount  int
	}
	for _, app := range apps {
		summary.appCount++
		summary.runningCount += app.TasksRunning
		summary.stagedCount += app.TasksStaged
		logApp(app, logMarathonTasks)
	}
	summaryLog := kv.M{
		"totalApps":    summary.appCount,
		"runningTasks": summary.runningCount,
		"stagedTasks":  summary.stagedCount,
	}
	kvlog.GaugeIntD("marathon-summary", summary.appCount, summaryLog)
}

func logApp(app App, logMarathonTasks bool) {
	appLog := kv.M{
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
	}
	title := fmt.Sprintf("marathon-apps.%s", normalizeAppID(app.ID))
	kvlog.GaugeIntD(title, app.Instances, appLog)

	if !logMarathonTasks {
		return
	}

	for _, task := range app.Tasks {
		taskLog := kv.M{
			"appId":     task.AppID,
			"id":        task.ID,
			"startedAt": task.StartedAt,
			"stagedAt":  task.StagedAt,
			"version":   task.Version,
		}
		title := fmt.Sprintf("marathon-tasks.%s", task.ID)
		kvlog.CounterD(title, 1, taskLog)
	}
}

func normalizeAppID(appID string) string {
	return strings.Replace(appID[1:], "/", ".", -1)
}

func LogEvent(event Event) {
	eventLog := kv.M{
		"appId":     event.AppID,
		"timestamp": event.Timestamp,
		"slaveId":   event.SlaveID,
		"taskId":    event.TaskID,
		"host":      event.Host,
		"ports":     event.Ports,
		"version":   event.Version,
	}
	kvlog.CounterD("marathon-event", 1, eventLog)
}

package marathon

import (
	"fmt"
	"log"
	"strings"

	"gopkg.in/clever/kayvee-go.v2"
)

type m map[string]interface{}

func LogState(apps []App, logMarathonTasks bool) {
	var summary struct {
		appCount     int64
		runningCount int64
		stagedCount  int64
	}
	for _, app := range apps {
		summary.appCount++
		summary.runningCount += app.TasksRunning
		summary.stagedCount += app.TasksStaged
		logApp(app, logMarathonTasks)
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

func logApp(app App, logMarathonTasks bool) {
	normalizedAppID := normalizeAppID(app.ID)
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

func normalizeAppID(appID string) string {
	return strings.Replace(appID[1:], "/", ".", -1)
}

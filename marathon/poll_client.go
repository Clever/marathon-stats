package marathon

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type App struct {
	ID                    string            `json:"id"`
	Env                   map[string]string `json:"env"`
	Instances             int               `json:"instances"`
	Cpus                  float64           `json:"cpus"`
	Mem                   float64           `json:"mem"`
	Disk                  float64           `json:"disk"`
	Constraints           []Constraint      `json:"constraints"`
	Ports                 []int             `json:"ports"`
	RequirePorts          bool              `json:"requirePorts"`
	BackoffSeconds        float64           `json:"backoffSeconds"`
	BackoffFactor         float64           `json:"backoffFactor"`
	MaxLaunchDelaySeconds int               `json:"maxLaunchDelaySeconds"`
	Container             Container         `json:"container"`
	HealthChecks          []HealthCheck     `json:"healthChecks"`
	Labels                map[string]string `json:"labels"`
	Version               string            `json:"version"`
	TasksStaged           int               `json:"tasksStaged"`
	TasksRunning          int               `json:"tasksRunning"`
	TasksHealthy          int               `json:"tasksHealthy"`
	TasksUnhealthy        int               `json:"tasksUnhealthy"`
	Deployments           []Deployment      `json:"deployments"`
	Tasks                 []Task            `json:"tasks"`
}

type Constraint []string

type Container struct {
	Type    string   `json:"type"`
	Volumes []Volume `json:"volumes"`
	Docker  Docker   `json:"docker"`
}

type Volume struct {
	ContainerPath string `json:"containerPath"`
	HostPath      string `json:"hostPath"`
	Mode          string `json:"mode"`
}

type Deployment struct {
	AffectedApps   []string       `json:"affectedApps"`
	ID             string         `json:"id"`
	Steps          []DeployAction `json:"steps"`
	CurrentActions []DeployAction `json:"currentActions"`
	Version        string         `json:"version"`
	CurrentStep    int            `json:"currentStep"`
	TotalSteps     int            `json:"totalSteps"`
}

type DeployAction struct {
	Action string `json:"action"`
	App    string `json:"app"`
}

type HealthCheck struct {
	Path                   string `json:"path"`
	Protocol               string `json:"protocol"`
	PortIndex              int    `json:"portIndex"`
	GracePeriodSeconds     int    `json:"gracePeriodSeconds"`
	IntervalSeconds        int    `json:"intervalSeconds"`
	TimeoutSeconds         int    `json:"timeoutSeconds"`
	MaxConsecutiveFailures int    `json:"maxConsecutiveFailures"`
}

type Docker struct {
	Image        string            `json:"image"`
	Network      string            `json:"network"`
	PortMappings []PortMapping     `json:"portMappings"`
	Privileged   bool              `json:"privileged"`
	Parameters   []DockerParameter `json:"parameters"`
}

type DockerParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PortMapping struct {
	ContainerPort int `json:"containerPort"`
	HostPort      int `json:"hostPort"`
	ServicePort   int `json:"servicePort"`
}

type Task struct {
	AppID              string              `json:"appId"`
	ID                 string              `json:"id"`
	Host               string              `json:"host"`
	Ports              []int               `json:"ports"`
	StartedAt          string              `json:"startedAt"`
	StagedAt           string              `json:"stagedAt"`
	Version            string              `json:"version"`
	ServicePorts       []int               `json:"servicePorts"`
	HealthCheckResults []HealthCheckResult `json:"healthCheckResults"`
}

type HealthCheckResult struct {
	TaskID              string `json:"taskId"`
	FirstSuccess        string `json:"firstSuccess"`
	LastSuccess         string `json:"lastSuccess"`
	LastFailure         string `json:"lastFailure"`
	LastFailureCause    string `json:"lastFailureCause"`
	ConsecutiveFailures int    `json:"consecutiveFailures"`
	Alive               bool   `json:"Alive"`
}

type AppsResponse struct {
	Apps []App `json:"apps"`
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

type Client struct {
	client *http.Client
	host   string
}

func NewClient(marathonHost string) *Client {
	return &Client{
		client: &http.Client{},
		host:   marathonHost,
	}
}

func (c *Client) GetApps() (AppsResponse, error) {
	appsURL := url.URL{Scheme: "http", Host: c.host, Path: "/v2/apps"}
	q := appsURL.Query()
	q.Set("embed", "apps.tasks")
	appsURL.RawQuery = q.Encode()
	var decodedResponse AppsResponse
	err := c.do(appsURL, &decodedResponse)
	return decodedResponse, err
}

func (c *Client) GetTasks(onlyRunning bool) (TasksResponse, error) {
	tasksURL := url.URL{Scheme: "http", Host: c.host, Path: "/v2/tasks"}
	if onlyRunning {
		q := tasksURL.Query()
		q.Set("status", "running")
		tasksURL.RawQuery = q.Encode()
	}
	var decodedResponse TasksResponse
	err := c.do(tasksURL, &decodedResponse)
	return decodedResponse, err
}

func (c *Client) do(endpoint url.URL, response interface{}) error {
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	jsonDecoder := json.NewDecoder(resp.Body)
	if err := jsonDecoder.Decode(&response); err != nil {
		return err
	}
	return nil
}

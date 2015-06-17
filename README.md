# marathon-stats

## Overview

`marathon-stats` is a simple container which queries marathon and mesos for stats about their current state, and logs these stats to stderr.

## Config

`marathon-stats` requires the following environment variables:

- MESOS_HOST - The host for the mesos master API
- MESOS_PORT - The port for the mesos master API
- MARATHON_HOST - The host for the marathon API
- MARATHON_PORT - The port for the marathon API
- LOG_MARATHON_TASKS - Whether or not to log marathon tasks on their own line
- POLL_INTERVAL - The interval to wait between polling for stats (defaults to 5s)

## Building and running locally

```bash
$ make build
$ ./marathon-stats
```

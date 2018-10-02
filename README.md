[![GoDoc](https://godoc.org/github.com/alex-leonhardt/gocd-seeder?status.svg)](https://godoc.org/github.com/alex-leonhardt/gocd-seeder)

# GOCD-SEEDER
A GoCD-Seeder scans a GitHub org for repositories that contain a pre-specified "topic" (default: ci-gocd), if a repo is found, it will create a GoCD config repo, which will make GoCD poll the repository for the file "ci.gocd.yaml" and create (a) new pipeline/s basaed on the config in that file.

## WIP

- clean up
  - write tests
  - make use of the interfaces etc. etc.
  - write logging and metric decorators and wrap func calls where appropriate

## NTHs
_(nice to haves)_

- metric (counter) of repos created, deleted

## BUILD

```
go build -ldflags "-X main.versionString=`git rev-list --max-count=1 --branches master --abbrev-commit`"
```

### Docker

#### Docker Hub

You can now use the below command to just pull down the latest version (master)
```
docker pull alexleonhardt/gocd-seeder:latest
```

Or alternatively, by tags:
```
docker pull alexleonhardt/gocd-seeder:<tag>
```

#### Build 
Multi-stage docker build, it's using alpine as the base OS so we can get a shell if we needed to debug the app for some reason.

```
docker build --no-cache -t local/gocd-seeder:latest .
```

To check version and help you can use: 

```
docker run --rm -ti local/gocd-seeder:latest version
docker run --rm -ti local/gocd-seeder:latest help
```

## METRICS

A metrics endpoint is running by default on port `:9090` and is reachable via `http://<IP|localhost>:9090/debug/vars`; metrics are provided via `expvar` - you can use things like

- [Datadog](https://docs.datadoghq.com/integrations/go_expvar/)
- [expvarmon](https://github.com/divan/expvarmon)
  ```shell
  expvarmon -ports="9090" -vars "Goroutines,Uptime,mem:memstats.Alloc,mem:memstats.Sys,mem:memstats.HeapAlloc,mem:memstats.HeapInuse,duration:memstats.PauseNs,duration:memstats.PauseTotalNs"
  ```
- Prometheus (maybe coming soon at some point if it is useful, `expvar` seems quick and cheap right now)

to monitor the app's memory, gc, goroutines & uptime

### Config

Use env variables to set the stats port and IP to listen on:
```
export HTTP_STATS_PORT=<PORT>
export HTTP_STATS_IP=<IP|localhost>
```

the defaults are: 

```
HTTP_STATS_PORT=9090
HTTP_STATS_IP=""
```

which will make the endpoint available on all interfaces, on port 9090

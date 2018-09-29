[![GoDoc](https://godoc.org/github.com/alex-leonhardt/gocd-seeder?status.svg)](https://godoc.org/github.com/alex-leonhardt/gocd-seeder)

# gocd-seeder
A tiny app to create GoCD pipelines using the yaml config plugin, this would run in paralllel with your GoCD server

## WIP

- remove repo when repo present and `topic` was removed
- output a metric of repos created/deleted (or make a http endpoint available)
- clean up
  - write tests
  - make use of the interfaces etc. etc.
  - write logging and metric decorators and wrap func calls where appropriate
  - ensure to only output on startup once and when errors occur with more helpful messages
- make docker image avail in docker hub


## BUILD

```
go build -ldflags "-X main.versionString=`git rev-list --max-count=1 --branches master --abbrev-commit`"
```

### Docker

Multi-stage docker build, it's using alpine as the base OS so we can get a shell if we needed to debug the app for some reason.

```
docker build --no-cache -t local/gocd-seeder:latest .
```

To check version and help you can use: 

```
docker run --rm -ti local/gocd-seeder:latest version
docker run --rm -ti local/gocd-seeder:latest help
```

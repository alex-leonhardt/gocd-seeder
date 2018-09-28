# gocd-seeder
A tiny app to create GoCD pipelines using the yaml config plugin, this would run in paralllel with your GoCD server

## WIP

- remove repo when repo present and `topic` was removed
- use structured logging go-kit/log
- output a metric of repos created/deleted (or make a http endpoint available)
- clean up
  - write tests
  - make use of the interfaces etc. etc.
  - write logging and metric decorators and wrap func calls where appropriate
  - ensure to only output on startup once and when errors occur with more helpful messages
- provide a Dockerfile to create image, make it avail in docker hub


## BUILD

```
go build -ldflags "-X main.versionString=`git rev-list --max-count=1 --branches master --abbrev-commit`"
```

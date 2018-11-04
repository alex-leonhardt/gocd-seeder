TAG = latest
DOCKER_USERNAME = alexleonhardt
APP = gocd-seeder
CMD = version

login:
	echo "${DOCKER_PASSWORD}" | docker login -u ${DOCKER_USERNAME} --password-stdin

build:
	docker build -t ${DOCKER_USERNAME}/${APP}:${TAG} .

push: login
	docker push ${DOCKER_USERNAME}/${APP}:${TAG}

run: build
	docker run --rm ${DOCKER_USERNAME}/gocd-seeder ${CMD}

# --------------------

test:
	go test -v ./...

testcover:
	go test -v ./... -cover

.PHONY: login build push run test testcover

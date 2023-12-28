NAME := payroll-app
OPENAPI-PATH := openapi/payroll.yaml
VERSION              := v0.1
PLATFORMS            := darwin linux
ARCHITECTURES        := amd64
IPATH                := github.com/joshinjohnson/wave-exercise
OPTS=""

setup:
	go install github.com/discord-gophers/goapi-gen@latest

gen: setup
	goapi-gen -generate server -o handler/server.go -package handler $(OPENAPI-PATH)
	goapi-gen -generate types -o handler/types.go -package handler $(OPENAPI-PATH)

build: gen test
	go build -o $(NAME)

test:
	go test -count=1 ./... 

test-all:
	go test -count=1 -p 1 -tags=integration ./...

.PHONY: run
run:
	go run main.go

clean:
	rm -r build/

.PHONY: dist
dist: build
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 \
	     go build -ldflags "\
        -X $(IPATH)/version.Version=${VERSION} \
        -X $(IPATH)/version.Revision=${REVISION} \
        -X $(IPATH)/version.Branch=${BRANCH_NAME} \
        -X $(IPATH)/version.BuildDate=${DATE} \
         -extldflags -static \
        " -o ./build/$(GOOS)-$(GOARCH)-$(VERSION)/$(NAME) .)))
	@echo "Done building binaries"

.PHONY: start-dev-env
start-dev-env: build
	( cd docker ; docker-compose -p "payroll-dev-stack" up -d --build --force-recreate --no-deps)

.PHONY: stop-dev-env
stop-dev-env:
	( cd docker ; docker-compose -p "payroll-dev-stack" down -v )
	docker volume rm payroll-dev-stack_db

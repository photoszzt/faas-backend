GO_FILES?=$$(find . -name '*.go' |grep -v third_party)
TAG?=latest
SQUASH?=false

default: lint gen-proto build-fcdaemon build-resmngr fc-go-sdk build-cpuhog

.PHONY: fc-go-sdk
fc-go-sdk: 
	make -C third_party/firecracker-go-sdk

.PHONY: faas-provider
faas-provider: 
	make build -C third_party/faas-provider

.PHONY: test
test: goimportscheck
	go test -v . .

.PHONY: testacc
testacc: goimportscheck
	go test -count=1 -v . -run="TestAcc" -timeout 20m


.PHONY: build-resmngr-docker
build-resmngr-docker:
	docker build --rm -f Dockerfile.resmngr -t functions/resmngr:$(TAG) .

.PHONY: build-fcdaemon-docker
build-fcdaemon-docker:
	docker build --rm -f Dockerfile.fcdaemon -t functions/fcdaemon:$(TAG) .

.PHONY: tc-redirect-tap
tc-redirect-tap:
	cd third_party/firecracker-go-sdk/cni/ && make clean && make && sudo cp tc-redirect-tap /opt/cni/bin && cd -

.PHONY: build-fcdaemon
build-fcdaemon:
	go build -o fcdaemon cmd/fcdaemon/fcdaemon.go

.PHONY: build-fcclient
build-fcclient:
	go build -o fcclient cmd/fcdaemon-test-client/client.go

.PHONY: build-resmngr
build-resmngr:
	go build --ldflags "\
		-X github.com/openfaas-incubator/faas-memory/version.GitCommitSHA=${GIT_COMMIT_SHA} \
		-X \"github.com/openfaas-incubator/faas-memory/version.GitCommitMessage=${GIT_COMMIT_MESSAGE}\" \
		-X github.com/openfaas-incubator/faas-memory/version.Version=${VERSION}" \
		-o resmngr cmd/resmngr/resmngr.go

.PHONY: build-cpuhog
build-cpuhog:
	go build -o cpuhog cmd/cpuhog/cpuhog.go

.PHONY: gen-proto
gen-proto:
	python3 -m grpc_tools.protoc -I pb/fcdaemon pb/fcdaemon/fc_daemon_service.proto --go_out=plugins=grpc:pb/fcdaemon
	python3 -m grpc_tools.protoc -I pb/gpuresc pb/gpuresc/gpuresc.proto --go_out=plugins=grpc:pb/gpuresc
	python3 -m grpc_tools.protoc -I pb/resmngr pb/resmngr/resource_manager.proto  --go_out=plugins=grpc:pb/resmngr
	python3 -m grpc_tools.protoc -I pb/mngrinfo pb/mngrinfo/mngrinfo.proto  --go_out=plugins=grpc:pb/mngrinfo

.PHONY: gen-fbs
gen-fbs:
	# flatc --go --grpc -I fbs fbs/common/empty.fbs
	# flatc --go --grpc -I fbs fbs/fcdaemon/fc_daemon_service.fbs
	# flatc --go --grpc -I fbs fbs/resmngr/resource_manager.fbs
	rm -f fbs/mngrsvc/*.go
	flatc --go --grpc -I fbs --gen-onefile -o fbs/mngrsvc --go-namespace mngrsvc fbs/mngrsvc/manager_service.fbs

.PHONY: up-local-deps
up-local-deps:
	docker-compose -f./docker-compose.local.yml up -d

.PHONY: up-local
up-local: build-local
	-pkill faas-memory
	docker-compose -f ./docker-compose.local.yml up -d
	env port=8081 ./faas-memory

.PHONY: release
release:
	go get github.com/goreleaser/goreleaser; \
	goreleaser; \

.PHONY: clean
clean:
	rm -rf faas-memory || true
	rm -rf fcdaemon || true
	rm -rf fcclient || true
	rm -rf resmngr || true
	rm -rf cpuhog || true

.PHONY: goimports
goimports:
	goimports -w $(GO_FILES)

.PHONY: goimportscheck
goimportscheck:
	@sh -c "'$(CURDIR)/scripts/goimportscheck.sh'"

.PHONY: vet
vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v third_party/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

.PHONY: lint
lint:
	@echo "golint ."
	@go get golang.org/x/tools/cmd/goimports
	@golint -set_exit_status handlers/ pkg/ types/ ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Lint found errors in the source code. Please check the reported errors"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

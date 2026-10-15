# Starts up the platform, casino and a mines engine from scratch
quick-start:
	./scripts/start.sh

tidy:
	@dirs=(./casino ./pkg ./engines/mines); \
	currentDir=$$(pwd); \
	for dir in $${dirs[@]}; do \
		cd "$$dir"; \
		go mod tidy; \
		cd "$$currentDir"; \
	done

engine:=

# Set NO_CACHE=1 to build without the docker build cache.
# SET PROFILE=1 to build with profiling server enabled.
build-engine:
	./scripts/build/engine.sh ${engine}

engine-up:
	@ENGINE_NAME=${engine} ENGINE_VERSION=local docker compose -f "./engines/docker-compose.yaml" up -d

port ?= 6060

pprof-cpu:
	@go tool pprof -http : http://localhost:${port}/debug/pprof/profile?seconds=15

pprof-block:
	@go tool pprof -http : http://localhost:${port}/debug/pprof/block

pprof-mutex:
	@go tool pprof -http : http://localhost:${port}/debug/pprof/mutex

pprof-routines:
	@curl -o routines.txt http://localhost:${port}/debug/pprof/goroutine?debug=2

# net|sync|syscall|sched
traceType :=
pprof-trace:
	@tt=${traceType}; \
	if [[ -z "$$tt" ]]; then \
		echo "traceType is missing"; \
		exit 1; \
	fi
	@curl -s -o trace.out http://localhost:${port}/debug/pprof/trace?debug=1
	@go tool trace -pprof=${traceType} trace.out > ${traceType}.pprof
	@go tool pprof -http : ${traceType}.pprof; rm trace.out ${traceType}.pprof

trace: 
	@curl -s -o trace.out http://localhost:${port}/debug/pprof/trace?debug=1
	@go tool trace trace.out; rm trace.out
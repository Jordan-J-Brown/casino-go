tidy:
	@dirs=(./casino ./pkg ./engines/mines); \
	currentDir=$$(pwd); \
	for dir in $${dirs[@]}; do \
		cd "$$dir"; \
		go mod tidy; \
		cd "$$currentDir"; \
	done

engine:=
version:=

build-engine-local:
	./scripts/build/engine.sh ${engine}

build-engine-prod:
	./scripts/build/engine.sh ${engine} ${version} prod

compose-engine:
	ENGINE_NAME=${engine} \
	ENGINE_VERSION=${version} \
	docker compose -f "./engines/docker-compose.yaml" up -d

pprof-cpu:
	@go tool pprof -http : http://localhost:6060/debug/pprof/profile?seconds=15

pprof-block:
	@go tool pprof -http : http://localhost:6060/debug/pprof/block

pprof-mutex:
	@go tool pprof -http : http://localhost:6060/debug/pprof/mutex

pprof-routines:
	@curl -o routines.txt http://localhost:6060/debug/pprof/goroutine?debug=2

# net|sync|syscall|sched
traceType :=
pprof-trace:
	@tt=${traceType}; \
	if [[ -z "$$tt" ]]; then \
		echo "traceType is missing"; \
		exit 1; \
	fi
	@curl -s -o trace.out http://localhost:6060/debug/pprof/trace?debug=1
	@go tool trace -pprof=${traceType} trace.out > ${traceType}.pprof
	@go tool pprof -http : ${traceType}.pprof; rm trace.out ${traceType}.pprof

trace: 
	@curl -s -o trace.out http://localhost:6060/debug/pprof/trace?debug=1
	@go tool trace trace.out; rm trace.out
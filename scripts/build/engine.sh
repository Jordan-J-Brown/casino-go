# ! /bin/bash
readonly pkgPath="$(realpath ./pkg)"
readonly originalDir="$(pwd)"

engine="$1"
if [[ -z "$engine" ]]; then
  echo "engine arg is missing"
  exit 1
fi

tag="${2:-local}"

env="${3:-local}"

registry="$4"
if [[ -n "$registry" ]]; then 
  registry="${registry}/"
fi

cd "./engines/${engine}" || exit 1

configPath=''
if [[ "$env" == 'local' ]]; then
  configPath="config.yaml"
else
  configPath="configs/config.${env}.yaml"
fi
if [[ ! -f "$configPath" ]]; then 
  echo "Config path is not valid."
  exit 1
fi

dockerTag="${registry}${engine}:${tag}"
echo "Building engine as: ${dockerTag}"

flags=(
  "--tag=${dockerTag}"
	"--build-context=pkg=${pkgPath}"
  "--build-context=gocache=$(go env GOCACHE)"	
  "--build-context=gomodcache=$(go env GOMODCACHE)"	
	"--build-arg=CONFIG_PATH=${configPath}"
  '--file=../local.Dockerfile'
)
if [[ -n "$NO_CACHE" ]]; then 
  flags+=('--no-cache')
fi
if [[ -n "$PROFILE" ]]; then 
  flags+=('--build-arg=BUILD_TAGS=profile')
fi

docker build "${flags[@]}" .

cd "$originalDir"
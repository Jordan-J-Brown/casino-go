# ! /bin/bash
currentDir="$(pwd)"

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

GITHUB_TOKEN=$GITHUB_TOKEN docker build --tag "${dockerTag}" \
		--secret id=GITHUB_TOKEN \
		--build-arg CONFIG_PATH="${configPath}" \
    --file ../Dockerfile .

cd "$currentDir"
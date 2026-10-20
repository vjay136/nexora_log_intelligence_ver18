#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
ARTIFACT_DIR=${1:?usage: run-matrix.sh ARTIFACT_DIR PREVIOUS_DIR}
PREVIOUS_DIR=${2:?real previous-release RPM directory is required}
# shellcheck disable=SC1091
source "$ROOT/packaging/docker/images.env"
[[ -d $ARTIFACT_DIR && -d $PREVIOUS_DIR ]] || { echo 'artifact directories must exist' >&2; exit 1; }
ARTIFACT_DIR=$(cd "$ARTIFACT_DIR" && pwd)
PREVIOUS_DIR=$(cd "$PREVIOUS_DIR" && pwd)
require_previous=${REQUIRE_PREVIOUS:-1}
[[ $require_previous == 0 || $require_previous == 1 ]] || { echo 'REQUIRE_PREVIOUS must be 0 or 1' >&2; exit 1; }

# Git Bash rewrites slash-prefixed arguments before invoking Windows programs.
# That corrupts Docker's comma-delimited --mount values (for example, the
# container destination can become a truncated `C`).  Give Docker native host
# paths for bind sources and disable MSYS argument rewriting for Docker calls.
docker_path() {
  if [[ -n ${MSYSTEM:-} ]] && command -v cygpath >/dev/null 2>&1; then
    cygpath -am "$1"
  else
    printf '%s\n' "$1"
  fi
}

run_docker() {
  MSYS_NO_PATHCONV=1 docker "$@"
}

DOCKER_ARTIFACT_DIR=$(docker_path "$ARTIFACT_DIR")
DOCKER_PREVIOUS_DIR=$(docker_path "$PREVIOUS_DIR")
DOCKER_TESTS_DIR=$(docker_path "$ROOT/tests")
DOCKER_ROCKY_DIR=$(docker_path "$ROOT/tests/rocky")
DOCKER_ROCKYFILE=$(docker_path "$ROOT/tests/rocky/Dockerfile")

run_one() {
  local label=$1
  local base=$2
  local image="nli-rocky-test:$label"
  local container="nli-rocky-${label}-$$"
  run_docker build --pull=false --build-arg "BASE_IMAGE=$base" -t "$image" -f "$DOCKER_ROCKYFILE" "$DOCKER_ROCKY_DIR"
  run_docker run -d --name "$container" --privileged --cgroupns=host \
    --tmpfs /run --tmpfs /run/lock \
    --mount "type=bind,src=$DOCKER_ARTIFACT_DIR,dst=/artifacts,readonly" \
    --mount "type=bind,src=$DOCKER_PREVIOUS_DIR,dst=/previous,readonly" \
    --mount "type=bind,src=$DOCKER_TESTS_DIR,dst=/tests,readonly" \
    "$image" >/dev/null
  trap "MSYS_NO_PATHCONV=1 docker rm -f '$container' >/dev/null 2>&1 || true" EXIT
  for _ in $(seq 1 60); do run_docker exec "$container" systemctl is-system-running --wait >/dev/null 2>&1 && break; sleep 1; done
  run_docker exec -e "REQUIRE_PREVIOUS=$require_previous" "$container" bash /tests/rocky/validate.sh
  run_docker rm -f "$container" >/dev/null
  trap - EXIT
}

run_one rocky9 "$ROCKY9_IMAGE"
run_one rocky10 "$ROCKY10_IMAGE"
if [[ ${RUN_TWO_NODE_ACCEPTANCE:-1} == 1 ]]; then
  bash "$ROOT/tests/rocky/validate-two-node-webview.sh" "$ARTIFACT_DIR"
  echo 'Rocky 9/10.2 lifecycle, two-node TLS, and packaged WebView: PASS'
else
  echo 'Rocky 9/10.2 synthetic package lifecycle: PASS (real-binary two-node acceptance deferred)'
fi

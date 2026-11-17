#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

DEFAULT_IMAGE="${DEFAULT_IMAGE:-$(basename "$(pwd)")}"
DEFAULT_DOCKERFILE="${DEFAULT_DOCKERFILE:-Dockerfile}"
DEFAULT_REGISTRY="${DEFAULT_REGISTRY:-ghcr.io}"
CRED_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/go-backend-deploy"
CRED_FILE="$CRED_DIR/credentials"

usage() {
  cat <<EOF
Build and push a Docker image for this service.

Usage: ./deploy.sh [options]

  -r REGISTRY    registry host (default: $DEFAULT_REGISTRY)
  -i IMAGE       image path in the registry, e.g. owner/service (default: $DEFAULT_IMAGE)
  -t TAG         image tag (default: <branch>-<version>)
  -u USER        registry user
  -f DOCKERFILE  Dockerfile (default: $DEFAULT_DOCKERFILE)
  -g PLATFORM    target platform(s), e.g. linux/amd64,linux/arm64 (default: local)
  --latest       also tag and push :latest
  --no-push      build only
  --forget       remove saved credentials for the registry and exit
  -y             accept defaults, never prompt
  -h             this help

Version tags like v1.2.3 are refused: those are built by CI from git tags.
Saved credentials: $CRED_FILE (mode 600, one line per registry).
EOF
}

REGISTRY="" IMAGE="" TAG="" USERNAME="" DOCKERFILE="" PLATFORM="" LATEST=0 PUSH=1 FORGET=0 YES=0
while [ $# -gt 0 ]; do
  case "$1" in
    -r) REGISTRY="$2"; shift 2 ;;
    -i) IMAGE="$2"; shift 2 ;;
    -t) TAG="$2"; shift 2 ;;
    -u) USERNAME="$2"; shift 2 ;;
    -f) DOCKERFILE="$2"; shift 2 ;;
    -g) PLATFORM="$2"; shift 2 ;;
    --latest) LATEST=1; shift ;;
    --no-push) PUSH=0; shift ;;
    --forget) FORGET=1; shift ;;
    -y) YES=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

ask() {
  local var="$1" prompt="$2" def="$3" val
  if [ -n "${!var}" ]; then return; fi
  if [ "$YES" = 1 ] || [ ! -t 0 ]; then printf -v "$var" '%s' "$def"; return; fi
  read -r -p "$prompt [$def]: " val
  printf -v "$var" '%s' "${val:-$def}"
}

saved_line() {
  [ -f "$CRED_FILE" ] && grep -m1 "^$1 " "$CRED_FILE" || true
}

forget() {
  [ -f "$CRED_FILE" ] || return 0
  local tmp
  tmp="$(mktemp)"
  grep -v "^$1 " "$CRED_FILE" > "$tmp" || true
  mv "$tmp" "$CRED_FILE"
  chmod 600 "$CRED_FILE"
}

ask REGISTRY "Registry" "$DEFAULT_REGISTRY"

if [ "$FORGET" = 1 ]; then
  forget "$REGISTRY"
  echo "Forgot credentials for $REGISTRY"
  exit 0
fi

BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo local)"
BRANCH="${BRANCH//\//-}"
VERSION="$(cat VERSION 2>/dev/null || echo 0.0.0)"

ask IMAGE "Image" "$DEFAULT_IMAGE"
ask TAG "Tag" "$BRANCH-$VERSION"
ask DOCKERFILE "Dockerfile" "$DEFAULT_DOCKERFILE"

if [[ "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  echo "Refusing tag $TAG: version tags are built by CI from git tags." >&2
  exit 1
fi
if [ "$TAG" = latest ]; then
  echo "Use --latest instead of -t latest." >&2
  exit 1
fi
[ -f "$DOCKERFILE" ] || { echo "No $DOCKERFILE here." >&2; exit 1; }

REF="$REGISTRY/$IMAGE"

if [ "$PUSH" = 1 ]; then
  line="$(saved_line "$REGISTRY")"
  if [ -n "$line" ] && [ -z "$USERNAME" ]; then
    USERNAME="$(echo "$line" | cut -d' ' -f2)"
    TOKEN="$(echo "$line" | cut -d' ' -f3-)"
    echo "Using saved credentials for $USERNAME@$REGISTRY"
  else
    ask USERNAME "Registry user" "${GITHUB_ACTOR:-}"
    if [ -t 0 ]; then
      read -r -s -p "Registry token/password: " TOKEN; echo
    else
      TOKEN="${REGISTRY_TOKEN:?set REGISTRY_TOKEN when not interactive}"
    fi
  fi
  echo "$TOKEN" | docker login "$REGISTRY" -u "$USERNAME" --password-stdin
  if [ -z "$line" ]; then
    mkdir -p "$CRED_DIR"
    touch "$CRED_FILE"
    chmod 600 "$CRED_FILE"
    forget "$REGISTRY"
    echo "$REGISTRY $USERNAME $TOKEN" >> "$CRED_FILE"
  fi
fi

echo
echo "  image      $REF:$TAG"
[ "$LATEST" = 1 ] && echo "  also       $REF:latest"
echo "  dockerfile $DOCKERFILE"
echo "  platform   ${PLATFORM:-local}"
echo "  push       $([ "$PUSH" = 1 ] && echo yes || echo no)"
if [ "$YES" != 1 ] && [ -t 0 ]; then
  read -r -p "Go? [Y/n] " go
  case "$go" in n|N) exit 1 ;; esac
fi

tags=(-t "$REF:$TAG")
[ "$LATEST" = 1 ] && tags+=(-t "$REF:latest")
args=(-f "$DOCKERFILE" "${tags[@]}")

if [ -n "$PLATFORM" ]; then
  out="--load"
  [ "$PUSH" = 1 ] && out="--push"
  docker buildx build --platform "$PLATFORM" "${args[@]}" "$out" .
else
  docker build "${args[@]}" .
  if [ "$PUSH" = 1 ]; then
    docker push "$REF:$TAG"
    [ "$LATEST" = 1 ] && docker push "$REF:latest"
  fi
fi
echo "Done: $REF:$TAG"

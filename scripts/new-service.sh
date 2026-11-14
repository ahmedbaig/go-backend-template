#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "usage: scripts/new-service.sh <service-name> [module-path]" >&2
  echo "  e.g. scripts/new-service.sh orders github.com/ahmedbaig/orders" >&2
  exit 2
fi

NAME="$1"
MODULE="${2:-github.com/ahmedbaig/$NAME}"
OLD_MODULE="github.com/ahmedbaig/go-backend-template"
OLD_NAME="go-backend-template"

if ! [[ "$NAME" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "service name must be lowercase letters, digits and dashes" >&2
  exit 2
fi

cd "$(dirname "$0")/.."

files=$(grep -rl --exclude-dir=.git --exclude-dir=bin --exclude-dir=var -e "$OLD_MODULE" -e "$OLD_NAME" . || true)
for f in $files; do
  case "$f" in ./scripts/new-service.sh) continue ;; esac
  sed -i.bak -e "s#$OLD_MODULE#$MODULE#g" -e "s#$OLD_NAME#$NAME#g" "$f"
  rm -f "$f.bak"
done

for sample in configs/*.json.sample; do
  base="$(basename "$sample")"
  newbase="${base/$OLD_NAME/$NAME}"
  [ "$base" != "$newbase" ] && git mv -f "$sample" "configs/$newbase" 2>/dev/null || mv "$sample" "configs/$newbase" 2>/dev/null || true
done

echo 0.1.0 > VERSION
go mod tidy >/dev/null 2>&1 || true
echo "Renamed to $NAME ($MODULE). Next: make test && make run"

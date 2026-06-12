#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <output-dir>" >&2
  exit 2
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_dir="$1"
module_path="github.com/lx93/typhoon-client"

case "$out_dir" in
  ""|"/")
    echo "refusing to mirror into an empty path or /" >&2
    exit 2
    ;;
esac

rm -rf "$out_dir"
mkdir -p \
  "$out_dir/cmd/typhoon-client" \
  "$out_dir/client" \
  "$out_dir/relay"

cp "$repo_root/cmd/client/main.go" "$out_dir/cmd/typhoon-client/main.go"
cp "$repo_root/internal/client/"*.go "$out_dir/client/"
cp "$repo_root/internal/relay/types.go" "$out_dir/relay/types.go"

cat > "$out_dir/go.mod" <<EOF
module $module_path

go 1.22
EOF

cat > "$out_dir/README.md" <<'EOF'
> This repository is generated from [lx93/typhoon](https://github.com/lx93/typhoon).
> Do not edit generated client files here. Make changes in the Typhoon monorepo instead.

EOF
cat "$repo_root/docs/desktop-client.md" >> "$out_dir/README.md"

find "$out_dir" -type f -name '*.go' -print0 | xargs -0 perl -0pi -e \
  's#typhoon/internal/client#github.com/lx93/typhoon-client/client#g; s#typhoon/internal/relay#github.com/lx93/typhoon-client/relay#g'

(
  cd "$out_dir"
  gofmt -w ./cmd ./client ./relay
)

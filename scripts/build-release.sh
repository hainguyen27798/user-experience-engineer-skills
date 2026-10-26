#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
version="$(tr -d '[:space:]' < "${repo_root}/VERSION")"
if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "VERSION must contain X.Y.Z; found: ${version}" >&2
  exit 1
fi

dist_dir="${repo_root}/dist"
stage_root="$(mktemp -d "${TMPDIR:-/tmp}/create-ux-engineer-release.XXXXXX")"
trap 'rm -rf -- "${stage_root}"' EXIT

for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64; do
  os="${target%/*}"
  arch="${target#*/}"
  asset="create-ux-engineer_${os}_${arch}.tar.gz"
  stage_dir="${stage_root}/${os}_${arch}"
  mkdir -p "${stage_dir}"

  CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" \
    go build -trimpath -ldflags "-s -w -X main.version=${version}" \
    -o "${stage_dir}/create-ux-engineer" "${repo_root}/cmd/create-ux-engineer"
  tar -C "${stage_dir}" -czf "${stage_root}/${asset}" create-ux-engineer
  rm -rf -- "${stage_dir}"
done

install -m 0755 "${repo_root}/scripts/install.sh" "${stage_root}/install.sh"

(
  cd "${stage_root}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum create-ux-engineer_*.tar.gz install.sh > checksums.txt
  else
    shasum -a 256 create-ux-engineer_*.tar.gz install.sh > checksums.txt
  fi
)

rm -rf -- "${dist_dir}"
mv "${stage_root}" "${dist_dir}"
trap - EXIT

echo "Built create-ux-engineer ${version} release assets in ${dist_dir}."

#!/usr/bin/env bash
set -euo pipefail

readonly RELEASE_BRANCH='main'
readonly RELEASE_REMOTE='origin'
readonly REQUIRED_ASSETS=(
  'checksums.txt'
  'create-ux-engineer_darwin_amd64.tar.gz'
  'create-ux-engineer_darwin_arm64.tar.gz'
  'create-ux-engineer_linux_amd64.tar.gz'
  'create-ux-engineer_linux_arm64.tar.gz'
  'install.sh'
)

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${repo_root}"

if [[ "$(git branch --show-current)" != "${RELEASE_BRANCH}" ]]; then
  echo "Releases must be created from ${RELEASE_BRANCH}." >&2
  exit 1
fi

if [[ -n "$(git status --short)" ]]; then
  echo 'The worktree must be clean before creating a release.' >&2
  exit 1
fi

if ! command -v gh >/dev/null 2>&1; then
  echo 'GitHub CLI (gh) is required to publish release assets.' >&2
  exit 1
fi
if ! gh auth status >/dev/null 2>&1; then
  echo 'GitHub CLI is not authenticated. Run: gh auth login' >&2
  exit 1
fi

current_version="$(tr -d '[:space:]' < VERSION)"
if [[ ! "${current_version}" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
  echo "VERSION must contain X.Y.Z; found: ${current_version}" >&2
  exit 1
fi

git fetch --prune --tags "${RELEASE_REMOTE}"

remote_branch="refs/remotes/${RELEASE_REMOTE}/${RELEASE_BRANCH}"
if git show-ref --verify --quiet "${remote_branch}" && ! git merge-base --is-ancestor "${remote_branch}" HEAD; then
  echo "Local ${RELEASE_BRANCH} is behind or diverged from ${RELEASE_REMOTE}/${RELEASE_BRANCH}." >&2
  exit 1
fi

current_tag="v${current_version}"

release_has_all_assets() {
  local tag="$1"
  local published_assets
  local asset

  if ! published_assets="$(gh release view "${tag}" --json assets --jq '.assets[].name' 2>/dev/null)"; then
    return 1
  fi
  for asset in "${REQUIRED_ASSETS[@]}"; do
    if ! grep -Fx "${asset}" <<<"${published_assets}" >/dev/null; then
      return 1
    fi
  done
}

publish_assets() {
  local tag="$1"

  if gh release view "${tag}" >/dev/null 2>&1; then
    gh release upload "${tag}" dist/* --clobber
  else
    gh release create "${tag}" dist/* --verify-tag --generate-notes --title "${tag}"
  fi
}

if ! git show-ref --verify --quiet "refs/tags/${current_tag}" && [[ "$(git log -1 --pretty=%s)" == "release: ${current_tag}" ]]; then
  git tag -a "${current_tag}" -m "${current_tag}"
fi

if git show-ref --verify --quiet "refs/tags/${current_tag}" && [[ "$(git rev-list -n 1 "${current_tag}")" == "$(git rev-parse HEAD)" ]] && ! release_has_all_assets "${current_tag}"; then
  echo "Resuming incomplete ${current_tag}..."
  ./scripts/build-release.sh
  if ! git ls-remote --exit-code --tags "${RELEASE_REMOTE}" "refs/tags/${current_tag}" >/dev/null 2>&1; then
    git push --atomic "${RELEASE_REMOTE}" "HEAD:refs/heads/${RELEASE_BRANCH}" "refs/tags/${current_tag}"
  fi
  publish_assets "${current_tag}"
  echo "Published ${current_tag}."
  exit 0
fi

next_version="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.$((BASH_REMATCH[3] + 1))"
next_tag="v${next_version}"
if git show-ref --verify --quiet "refs/tags/${next_tag}"; then
  echo "Tag ${next_tag} already exists." >&2
  exit 1
fi

build_check_dir="$(mktemp -d "${TMPDIR:-/tmp}/create-ux-engineer-build.XXXXXX")"
build_check="${build_check_dir}/create-ux-engineer"
version_changed=false
cleanup() {
  rm -rf -- "${build_check_dir}"
  if [[ "${version_changed}" == true ]]; then
    printf '%s\n' "${current_version}" > VERSION
    git restore --staged -- VERSION >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

go test ./...
go test -race ./...
go vet ./...
printf '%s\n' "${next_version}" > VERSION
version_changed=true
./scripts/build-release.sh

host_os="$(go env GOOS)"
host_arch="$(go env GOARCH)"
tar -C "${build_check_dir}" -xzf "dist/create-ux-engineer_${host_os}_${host_arch}.tar.gz" create-ux-engineer
"${build_check}" --version | grep -Fx "create-ux-engineer ${next_version}" >/dev/null

git add VERSION
git commit -m "release: ${next_tag}"
version_changed=false
git tag -a "${next_tag}" -m "${next_tag}"

if ! git push --atomic "${RELEASE_REMOTE}" "HEAD:refs/heads/${RELEASE_BRANCH}" "refs/tags/${next_tag}"; then
  echo "Push failed. The local release commit and ${next_tag} were kept; rerun ./release.sh to retry." >&2
  exit 1
fi

publish_assets "${next_tag}"
echo "Published ${next_tag} with local release binaries and install.sh."

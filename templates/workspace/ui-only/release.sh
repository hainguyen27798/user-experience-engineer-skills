#!/usr/bin/env bash

set -Eeuo pipefail

readonly RELEASE_BRANCH='release'
readonly RELEASE_REMOTE='origin'

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(git -C "${script_dir}" rev-parse --show-toplevel)"
ui_dir="${repo_root}/{{UI_DIRECTORY}}"
tmp_base="${TMPDIR:-/tmp}"
tmp_base="${tmp_base%/}"
release_tmp="$(mktemp -d "${tmp_base}/{{PROJECT_NAME}}-release.XXXXXX")"
stage_dir="${release_tmp}/stage"
release_worktree="${release_tmp}/worktree"
worktree_added=false

cleanup() {
  local exit_code=$?

  trap - EXIT INT TERM

  if [[ "${worktree_added}" == true ]]; then
    git -C "${repo_root}" worktree remove --force "${release_worktree}" >/dev/null 2>&1 || true
  fi

  case "${release_tmp}" in
    "${tmp_base}"/{{PROJECT_NAME}}-release.*)
      rm -rf -- "${release_tmp}"
      ;;
  esac

  exit "${exit_code}"
}

trap cleanup EXIT INT TERM

for required_command in git npm; do
  if ! command -v "${required_command}" >/dev/null 2>&1; then
    echo "Missing required command: ${required_command}" >&2
    exit 1
  fi
done

if [[ -n "$(git -C "${repo_root}" status --porcelain --untracked-files=all)" ]]; then
  echo 'The source worktree must be clean before creating a release.' >&2
  exit 1
fi

git -C "${repo_root}" remote get-url "${RELEASE_REMOTE}" >/dev/null

if [[ -f "${repo_root}/.gitmodules" ]]; then
  git -C "${repo_root}" submodule update --init --recursive
fi

mkdir -p "${stage_dir}/frontend"

echo 'Building frontend distribution...'
(
  cd "${ui_dir}"
  npm ci
  npm run build
)
cp -R "${ui_dir}/.output/." "${stage_dir}/frontend/"
find "${stage_dir}/frontend" -type f \
  \( -name '.DS_Store' -o -name '.DS_Store.br' -o -name '.DS_Store.gz' \) \
  -delete

source_commit="$(git -C "${repo_root}" rev-parse HEAD)"
remote_release_ref="$(git -C "${repo_root}" ls-remote --heads "${RELEASE_REMOTE}" "refs/heads/${RELEASE_BRANCH}")"

if [[ -n "${remote_release_ref}" ]]; then
  git -C "${repo_root}" fetch --no-tags "${RELEASE_REMOTE}" "refs/heads/${RELEASE_BRANCH}"
  if git -C "${repo_root}" merge-base "${source_commit}" FETCH_HEAD >/dev/null 2>&1; then
    echo "The remote ${RELEASE_BRANCH} branch is not orphaned from the source history." >&2
    exit 1
  fi
  git -C "${repo_root}" worktree add --detach "${release_worktree}" FETCH_HEAD
  worktree_added=true
elif git -C "${repo_root}" show-ref --verify --quiet "refs/heads/${RELEASE_BRANCH}"; then
  if git -C "${repo_root}" merge-base "${source_commit}" "refs/heads/${RELEASE_BRANCH}" >/dev/null 2>&1; then
    echo "The local ${RELEASE_BRANCH} branch is not orphaned from the source history." >&2
    exit 1
  fi
  git -C "${repo_root}" worktree add --detach "${release_worktree}" "refs/heads/${RELEASE_BRANCH}"
  worktree_added=true
else
  git -C "${repo_root}" worktree add --detach "${release_worktree}" HEAD
  worktree_added=true
  git -C "${release_worktree}" switch --orphan "${RELEASE_BRANCH}"
fi

git -C "${release_worktree}" rm -rfq --ignore-unmatch .
find "${release_worktree}" -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf -- {} +

mkdir -p "${release_worktree}/frontend"
cp -R "${stage_dir}/frontend/." "${release_worktree}/frontend/"

release_timestamp="$(TZ=UTC date '+%Y/%m/%d %H:%M')"
git -C "${release_worktree}" add --all
git -C "${release_worktree}" commit --quiet --allow-empty -m "release ${release_timestamp}"
git -C "${release_worktree}" push "${RELEASE_REMOTE}" "HEAD:refs/heads/${RELEASE_BRANCH}"

echo "Published release ${release_timestamp} from ${source_commit}."

#!/bin/sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
current_version="$(tr -d '[:space:]' < "${repo_root}/VERSION")"
next_version="${current_version%.*}.$((${current_version##*.} + 1))"
next_tag="v${next_version}"
test_root="$(mktemp -d "${TMPDIR:-/tmp}/create-ux-engineer-release-test.XXXXXX")"
trap 'rm -rf -- "${test_root}"' EXIT HUP INT TERM

work_dir="${test_root}/work"
remote_dir="${test_root}/remote.git"
tools_dir="${test_root}/tools"
state_dir="${test_root}/gh-state"
mkdir -p "${work_dir}" "${tools_dir}" "${state_dir}"
cp -R "${repo_root}/." "${work_dir}/"
rm -rf -- "${work_dir}/.git" "${work_dir}/dist"

cat > "${tools_dir}/go" <<'EOF'
#!/bin/sh
set -eu
case "${1:-}" in
  test|vet) exit 0 ;;
  env)
    case "${2:-}" in
      GOOS) echo linux ;;
      GOARCH) echo amd64 ;;
      *) exit 1 ;;
    esac
    ;;
  *) echo "unexpected go call: $*" >&2; exit 1 ;;
esac
EOF

cat > "${tools_dir}/gh" <<'EOF'
#!/bin/sh
set -eu
case "${1:-} ${2:-}" in
  'auth status') exit 0 ;;
  'release view')
    if [ -f "${TEST_GH_STATE}/released" ]; then
      case " $* " in
        *' --json assets '*) cat "${TEST_GH_STATE}/assets" ;;
      esac
      exit 0
    fi
    exit 1
    ;;
  'release create')
    if [ ! -f "${TEST_GH_STATE}/failed-once" ]; then
      touch "${TEST_GH_STATE}/failed-once"
      exit 1
    fi
    cat > "${TEST_GH_STATE}/assets" <<'ASSETS'
checksums.txt
create-ux-engineer_darwin_amd64.tar.gz
create-ux-engineer_darwin_arm64.tar.gz
create-ux-engineer_linux_amd64.tar.gz
create-ux-engineer_linux_arm64.tar.gz
install.sh
ASSETS
    touch "${TEST_GH_STATE}/released"
    exit 0
    ;;
  'release upload') exit 0 ;;
esac
echo "unexpected gh call: $*" >&2
exit 1
EOF
chmod +x "${tools_dir}/go" "${tools_dir}/gh"

cat > "${work_dir}/scripts/build-release.sh" <<'EOF'
#!/bin/sh
set -eu
repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
version="$(tr -d '[:space:]' < "${repo_root}/VERSION")"
rm -rf -- "${repo_root}/dist"
mkdir -p "${repo_root}/dist" "${repo_root}/payload"
printf '#!/bin/sh\necho "create-ux-engineer %s"\n' "${version}" > "${repo_root}/payload/create-ux-engineer"
chmod +x "${repo_root}/payload/create-ux-engineer"
tar -C "${repo_root}/payload" -czf "${repo_root}/dist/create-ux-engineer_linux_amd64.tar.gz" create-ux-engineer
touch \
  "${repo_root}/dist/create-ux-engineer_linux_arm64.tar.gz" \
  "${repo_root}/dist/create-ux-engineer_darwin_amd64.tar.gz" \
  "${repo_root}/dist/create-ux-engineer_darwin_arm64.tar.gz" \
  "${repo_root}/dist/checksums.txt"
cp "${repo_root}/scripts/install.sh" "${repo_root}/dist/install.sh"
rm -rf -- "${repo_root}/payload"
EOF

git init --bare "${remote_dir}" >/dev/null
git -C "${work_dir}" init -b main >/dev/null
git -C "${work_dir}" config user.name 'Release Test'
git -C "${work_dir}" config user.email 'release-test@example.com'
git -C "${work_dir}" add .
git -C "${work_dir}" commit -m 'initial test state' >/dev/null
git -C "${work_dir}" remote add origin "${remote_dir}"
git -C "${work_dir}" push -u origin main >/dev/null 2>&1

run_release() {
  PATH="${tools_dir}:${PATH}" \
    TEST_GH_STATE="${state_dir}" \
    "${work_dir}/release.sh"
}

if run_release >/dev/null 2>&1; then
  echo 'release unexpectedly succeeded during the simulated first upload failure' >&2
  exit 1
fi
test "$(cat "${work_dir}/VERSION")" = "${next_version}"
test "$(git -C "${work_dir}" tag --list)" = "${next_tag}"
test "$(git -C "${work_dir}" rev-list --count HEAD)" = '2'
test -z "$(git -C "${work_dir}" status --porcelain)"

run_release >/dev/null
test "$(cat "${work_dir}/VERSION")" = "${next_version}"
test "$(git -C "${work_dir}" rev-list --count HEAD)" = '2'
test -f "${state_dir}/released"
test -n "$(git --git-dir="${remote_dir}" show-ref --tags "${next_tag}")"
test -z "$(git -C "${work_dir}" status --porcelain)"

echo 'release.sh bump and failed-upload resume checks passed.'

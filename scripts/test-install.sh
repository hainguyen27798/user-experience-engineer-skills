#!/bin/sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
test_root="$(mktemp -d "${TMPDIR:-/tmp}/create-ux-engineer-install-test.XXXXXX")"
trap 'rm -rf -- "${test_root}"' EXIT HUP INT TERM

release_dir="${test_root}/release"
tools_dir="${test_root}/tools"
home_dir="${test_root}/home"
mkdir -p "${release_dir}" "${tools_dir}" "${home_dir}"

cat > "${tools_dir}/uname" <<'EOF'
#!/bin/sh
case "$1" in
  -s) echo Linux ;;
  -m) echo x86_64 ;;
  *) exit 1 ;;
esac
EOF

cat > "${tools_dir}/curl" <<'EOF'
#!/bin/sh
set -eu
output=''
url=''
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o) output="$2"; shift 2 ;;
    -*) shift ;;
    *) url="$1"; shift ;;
  esac
done
cp "${TEST_RELEASE_DIR}/${url##*/}" "${output}"
EOF
chmod +x "${tools_dir}/uname" "${tools_dir}/curl"

write_checksum() {
  (
    cd "${release_dir}"
    if command -v sha256sum >/dev/null 2>&1; then
      sha256sum create-ux-engineer_linux_amd64.tar.gz > checksums.txt
    else
      shasum -a 256 create-ux-engineer_linux_amd64.tar.gz > checksums.txt
    fi
  )
}

publish_fixture() {
  fixture_version="$1"
  payload_dir="${test_root}/payload"
  rm -rf -- "${payload_dir}"
  mkdir -p "${payload_dir}"
  printf '%s\n' "${fixture_version}" > "${payload_dir}/create-ux-engineer"
  chmod +x "${payload_dir}/create-ux-engineer"
  tar -C "${payload_dir}" -czf "${release_dir}/create-ux-engineer_linux_amd64.tar.gz" create-ux-engineer
  write_checksum
}

run_installer() {
  PATH="${tools_dir}:${PATH}" \
    HOME="${home_dir}" \
    TEST_RELEASE_DIR="${release_dir}" \
    sh "${repo_root}/scripts/install.sh" >/dev/null
}

publish_fixture 'version-one'
run_installer
test "$(cat "${home_dir}/.local/bin/create-ux-engineer")" = 'version-one'
test -x "${home_dir}/.local/bin/create-ux-engineer"

publish_fixture 'version-two'
run_installer
test "$(cat "${home_dir}/.local/bin/create-ux-engineer")" = 'version-two'

printf 'unexpected\n' > "${test_root}/payload/unexpected.txt"
tar -C "${test_root}/payload" -czf "${release_dir}/create-ux-engineer_linux_amd64.tar.gz" create-ux-engineer unexpected.txt
write_checksum
if run_installer 2>/dev/null; then
  echo 'Installer accepted an archive with an unexpected path.' >&2
  exit 1
fi
test "$(cat "${home_dir}/.local/bin/create-ux-engineer")" = 'version-two'

rm -rf -- "${test_root}/payload"
mkdir -p "${test_root}/payload"
ln -s /bin/sh "${test_root}/payload/create-ux-engineer"
tar -C "${test_root}/payload" -czf "${release_dir}/create-ux-engineer_linux_amd64.tar.gz" create-ux-engineer
write_checksum
if run_installer 2>/dev/null; then
  echo 'Installer accepted a symlink in place of the binary.' >&2
  exit 1
fi
test "$(cat "${home_dir}/.local/bin/create-ux-engineer")" = 'version-two'

printf 'corrupt archive\n' >> "${release_dir}/create-ux-engineer_linux_amd64.tar.gz"
if run_installer 2>/dev/null; then
  echo 'Installer accepted an archive with a mismatched checksum.' >&2
  exit 1
fi
test "$(cat "${home_dir}/.local/bin/create-ux-engineer")" = 'version-two'

echo 'install.sh install, upgrade, archive validation, checksum, and atomic replacement checks passed.'

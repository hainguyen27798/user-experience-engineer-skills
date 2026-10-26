#!/bin/sh
set -eu

readonly_repository='hainguyen27798/user-experience-engineer-skills'
readonly_base_url="https://github.com/${readonly_repository}/releases/latest/download"

case "$(uname -s)" in
  Linux) platform_os='linux' ;;
  Darwin) platform_os='darwin' ;;
  *)
    echo "Unsupported operating system: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64|amd64) platform_arch='amd64' ;;
  arm64|aarch64) platform_arch='arm64' ;;
  *)
    echo "Unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

archive="create-ux-engineer_${platform_os}_${platform_arch}.tar.gz"
temp_dir="$(mktemp -d "${TMPDIR:-/tmp}/create-ux-engineer-install.XXXXXX")"
staged_binary=''

cleanup() {
  if [ -n "${staged_binary}" ]; then
    rm -f -- "${staged_binary}"
  fi
  rm -rf -- "${temp_dir}"
}
trap cleanup EXIT HUP INT TERM

curl -fsSL "${readonly_base_url}/${archive}" -o "${temp_dir}/${archive}"
curl -fsSL "${readonly_base_url}/checksums.txt" -o "${temp_dir}/checksums.txt"

checksum_line="$(grep "  ${archive}$" "${temp_dir}/checksums.txt" || true)"
if [ -z "${checksum_line}" ]; then
  echo "No checksum published for ${archive}." >&2
  exit 1
fi
printf '%s\n' "${checksum_line}" > "${temp_dir}/expected-checksum.txt"

(
  cd "${temp_dir}"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c expected-checksum.txt
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c expected-checksum.txt
  else
    echo 'SHA-256 verification requires sha256sum or shasum.' >&2
    exit 1
  fi
)

archive_entries="$(tar -tzf "${temp_dir}/${archive}")"
if [ "${archive_entries}" != 'create-ux-engineer' ]; then
  echo 'The release archive contains unexpected paths.' >&2
  exit 1
fi

tar -C "${temp_dir}" -xzf "${temp_dir}/${archive}"
if [ ! -f "${temp_dir}/create-ux-engineer" ] || [ -L "${temp_dir}/create-ux-engineer" ]; then
  echo 'The release archive does not contain a regular create-ux-engineer file.' >&2
  exit 1
fi

install_dir="${CREATE_UX_ENGINEER_INSTALL_DIR:-${HOME}/.local/bin}"
mkdir -p "${install_dir}"
staged_binary="$(mktemp "${install_dir}/.create-ux-engineer.XXXXXX")"
install -m 0755 "${temp_dir}/create-ux-engineer" "${staged_binary}"
mv -f -- "${staged_binary}" "${install_dir}/create-ux-engineer"
staged_binary=''

echo "Installed create-ux-engineer to ${install_dir}/create-ux-engineer"
case ":${PATH}:" in
  *":${install_dir}:"*) ;;
  *) echo "Add ${install_dir} to PATH to run create-ux-engineer." ;;
esac

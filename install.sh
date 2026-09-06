#!/bin/sh

set -eu

repo="nikhil25803/wordly"
install_dir="${WORDLY_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s)" in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  MINGW*|MSYS*|CYGWIN*) os=windows ;;
  *) echo "Unsupported operating system: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

extension=tar.gz
binary=wordly
if [ "$os" = windows ]; then
  extension=zip
  binary=wordly.exe
fi

archive="wordly_${os}_${arch}.${extension}"
base_url="${WORDLY_RELEASE_URL:-https://github.com/${repo}/releases/latest/download}"
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

curl -fsSL "${base_url}/${archive}" -o "${tmp_dir}/${archive}"
curl -fsSL "${base_url}/checksums.txt" -o "${tmp_dir}/checksums.txt"

expected=$(awk -v file="$archive" '$2 == file { print $1 }' "${tmp_dir}/checksums.txt")
if [ -z "$expected" ]; then
  echo "No checksum found for ${archive}" >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "${tmp_dir}/${archive}" | awk '{ print $1 }')
else
  actual=$(shasum -a 256 "${tmp_dir}/${archive}" | awk '{ print $1 }')
fi
[ "$actual" = "$expected" ] || { echo "Checksum verification failed" >&2; exit 1; }

if [ "$os" = windows ]; then
  unzip -q "${tmp_dir}/${archive}" -d "$tmp_dir"
else
  tar -xzf "${tmp_dir}/${archive}" -C "$tmp_dir"
fi

mkdir -p "$install_dir"
install "${tmp_dir}/${binary}" "${install_dir}/${binary}"
echo "Installed ${binary} to ${install_dir}"

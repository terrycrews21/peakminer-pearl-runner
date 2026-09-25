#!/usr/bin/env bash
set -euo pipefail

readonly version='2.17.1'
readonly asset_url='https://github.com/peakminer/peakminer/releases/download/v2.17.1/peakminer-2.17.1-linux-x86_64'
readonly expected_sha256='abdc8f915c149e5ca265a40dbafa59b92159a1d40478ced199e4e384060cbe3f'
readonly repo_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
readonly cache_home="${XDG_CACHE_HOME:-${HOME:?HOME must be set}/.cache}"
readonly cache_dir="$cache_home/peakminer/$version"
readonly binary="$cache_dir/peakminer"

if ! command -v curl >/dev/null 2>&1; then
  printf 'Required command not found: curl\n' >&2
  exit 127
fi
if ! command -v sha256sum >/dev/null 2>&1; then
  printf 'Required command not found: sha256sum\n' >&2
  exit 127
fi

case "$#" in
  0) mode='run' ;;
  1)
    if [[ "$1" != '--version' ]]; then
      printf 'Usage: %s [--version]\n' "${BASH_SOURCE[0]}" >&2
      exit 64
    fi
    mode='version'
    ;;
  *)
    printf 'Usage: %s [--version]\n' "${BASH_SOURCE[0]}" >&2
    exit 64
    ;;
esac

mkdir -p -- "$cache_dir"
verified=0
if [[ -f "$binary" ]] && printf '%s  %s\n' "$expected_sha256" "$binary" | sha256sum --check --status; then
  verified=1
fi

if [[ "$verified" -eq 0 ]]; then
  temporary="$(mktemp "$cache_dir/peakminer.XXXXXX")"
  trap 'rm -f -- "$temporary"' EXIT
  curl --fail --location --silent --show-error "$asset_url" --output "$temporary"
  printf '%s  %s\n' "$expected_sha256" "$temporary" | sha256sum --check
  chmod 0755 "$temporary"
  mv -f -- "$temporary" "$binary"
  trap - EXIT
fi
chmod 0755 "$binary"

if [[ "$mode" == 'version' ]]; then
  exec "$binary" --version
fi

if [[ -f "$repo_dir/.env" ]]; then
  set -a
  # .env is a trusted local shell-assignment file; never use an untrusted copy.
  source "$repo_dir/.env"
  set +a
fi

: "${PEAK_COIN:?Set PEAK_COIN in .env or the environment}"
: "${PEAK_POOL:?Set PEAK_POOL in .env or the environment}"
: "${PEAK_WALLET:?Set PEAK_WALLET in .env or the environment}"

exec "$binary"

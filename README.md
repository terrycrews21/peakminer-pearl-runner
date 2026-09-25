# PeakMiner Pearl Runner

A small Linux x86_64 launcher for running the official PeakMiner v2.17.1 binary on your own machine. PeakMiner is a GPU cryptocurrency miner; this repository contains only the launcher and setup instructions, not the miner binary.

The launcher downloads the official asset directly from [PeakMiner's v2.17.1 release](https://github.com/peakminer/peakminer/releases/tag/v2.17.1), verifies its published SHA-256, caches it under `$XDG_CACHE_HOME/peakminer/2.17.1/` (or `$HOME/.cache/peakminer/2.17.1/`), and executes it. It does not mirror or attach the binary to this repository.

## Requirements

- Linux x86_64
- Bash, `curl`, and `sha256sum`
- An NVIDIA or AMD GPU supported by PeakMiner

## Configure once

```sh
cp .env.example .env
$EDITOR .env
chmod 600 .env
```

Set the wallet, pool, and any proxy values in `.env`. The variable names are PeakMiner's documented environment aliases. Use `PEAK_PROXY`, `PEAK_PROXY_USER`, and `PEAK_PROXY_PASS` rather than putting proxy credentials in command-line options. Do not commit `.env`; it is ignored by Git.

`.env` is sourced as a local shell-assignment file, so only use a file you trust. Environment values stay out of the process argument list, but may still be visible to privileged or same-user processes with access to process environments.

## Run

```sh
./run_peakminer.sh
```

The normal launch takes no command-line arguments. For a non-mining smoke check, run:

```sh
./run_peakminer.sh --version
```

`--version` downloads and verifies the same upstream binary, prints its version, then exits without starting a mining session. Other command-line arguments are rejected so credentials are not passed through argv.

## Upstream

- [PeakMiner v2.17.1 release and license](https://github.com/peakminer/peakminer/releases/tag/v2.17.1)
- [PeakMiner CLI and environment-variable documentation](https://github.com/peakminer/peakminer#cli-reference)
- [PeakMiner license](https://github.com/peakminer/peakminer/blob/main/LICENSE)

Review and accept PeakMiner's license before use. This launcher does not alter or redistribute PeakMiner.

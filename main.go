// migrate‑pulumi‑config.go
//
// Usage:
//   go run migrate-pulumi-config.go \
/*      --src dev            \  # source stack name
        --dst new-dev        \  # destination stack name
        --src-pass oldSecret \  # passphrase for src stack
        --dst-pass newSecret */// passphrase for dst stack
//
// Prerequisites: Pulumi CLI in $PATH.

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// oneConfig models the JSON shape returned by
//
//	pulumi config --show-secrets --json
type oneConfig struct {
	Value  string `json:"value"`
	Secret bool   `json:"secret,omitempty"`
}

func main() {
	var (
		srcStack, dstStack string
		srcPass, dstPass   string
		workDir            string
	)

	flag.StringVar(&srcStack, "src", "", "source stack name")
	flag.StringVar(&dstStack, "dst", "", "destination stack name (created if absent)")
	flag.StringVar(&srcPass, "src-pass", "", "source stack passphrase")
	flag.StringVar(&dstPass, "dst-pass", "", "destination stack passphrase")
	flag.StringVar(&workDir, "dir", ".", "Pulumi project directory (default=.)")
	flag.Parse()

	if srcStack == "" || dstStack == "" || srcPass == "" || dstPass == "" {
		log.Fatalln("src, dst, src-pass and dst-pass are required")
	}

	// -----------------------------
	// 1️⃣  Fetch config from source
	// -----------------------------
	srcCfg := pullConfig(workDir, srcStack, srcPass)

	// -----------------------------
	// 2️⃣  Ensure destination stack
	// -----------------------------
	if !stackExists(workDir, dstStack, dstPass) {
		log.Printf("Destination stack %q not found – creating it\n", dstStack)
		runPulumi(workDir, dstPass, "stack", "init", dstStack)
	}

	// -----------------------------
	// 3️⃣  Set values in destination
	// -----------------------------
	for key, entry := range srcCfg {
		args := []string{"config", "set", "--stack", dstStack}
		if entry.Secret {
			args = append(args, "--secret")
		}
		args = append(args, key, entry.Value)
		fmt.Printf("→ %s\n", key)
		runPulumi(workDir, dstPass, args...)
	}

	fmt.Printf("\n✅ Migrated %d config values from %q → %q\n",
		len(srcCfg), srcStack, dstStack)
}

// pullConfig returns all key/value pairs (decrypted) from a stack
func pullConfig(dir, stack, pass string) map[string]oneConfig {
	out := runPulumi(dir, pass, "config", "--show-secrets", "--json", "--stack", stack)
	cfg := make(map[string]oneConfig)
	if err := json.Unmarshal(out, &cfg); err != nil {
		log.Fatalf("decoding JSON: %v\n", err)
	}
	return cfg
}

// stackExists checks if the stack is already present
func stackExists(dir, pass, stack string, extra ...string) bool {
	cmd := exec.Command("pulumi", append([]string{"stack", "ls", "--json"}, extra...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PULUMI_CONFIG_PASSPHRASE="+pass)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("pulumi stack ls: %v\n%s", err, out)
	}
	var stacks []struct{ Name string }
	if err := json.Unmarshal(out, &stacks); err != nil {
		log.Fatalf("decode stack ls JSON: %v", err)
	}
	for _, s := range stacks {
		if s.Name == stack {
			return true
		}
	}
	return false
}

// runPulumi wraps a pulumi CLI invocation, returning STDOUT+STDERR or exiting on error
func runPulumi(dir, pass string, args ...string) []byte {
	cmd := exec.Command("pulumi", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"PULUMI_CONFIG_PASSPHRASE="+pass,
		"PULUMI_SKIP_UPDATE_CHECK=1",
	)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	if err := cmd.Run(); err != nil {
		log.Fatalf("pulumi %v: %v\n%s", args, err, buf.String())
	}
	return buf.Bytes()
}

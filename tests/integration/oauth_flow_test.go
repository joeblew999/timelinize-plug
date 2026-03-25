package integration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testTenantID   = "gedw99"
	testAccount    = "gedw99@gmail.com"
	serverBaseURL  = "http://127.0.0.1:12002"
	oauthEnvToggle = "E2E_REAL"
)

// TestGoogleOAuthFlow drives the real OAuth flow using the local server.
// It requires manual interaction: when run, it will launch `tplug auth` which
// opens a browser window. Complete the Google sign-in using the gedw99 account.
func TestGoogleOAuthFlow(t *testing.T) {
	if os.Getenv(oauthEnvToggle) != "1" {
		t.Skipf("skipping real OAuth flow; set %s=1 to run (requires manual interaction with %s)", oauthEnvToggle, testAccount)
	}

	requireEnv(t, "GOOGLE_CLIENT_ID")
	requireEnv(t, "GOOGLE_CLIENT_SECRET")

	repoRoot := findRepoRoot(t)
	binary := buildTplugBinary(t, repoRoot)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverCmd := exec.CommandContext(ctx, binary, "serve")
	serverCmd.Dir = repoRoot
	serverCmd.Env = append(os.Environ(), fmt.Sprintf("TPLUG_TENANT_ID=%s", testTenantID))
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	if err := serverCmd.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	t.Cleanup(func() {
		cancel()
		_ = serverCmd.Process.Kill()
		_ = serverCmd.Wait()
	})

	t.Log("Starting embedded server…")
	if err := waitForServer(serverBaseURL, 30*time.Second); err != nil {
		t.Fatalf("server did not become ready: %v", err)
	}
	t.Logf("Server is ready at %s", serverBaseURL)

	t.Log("Launching OAuth flow. A browser window will open – sign in using", testAccount)
	authCmd := exec.Command(binary, "auth", "--provider", "google", "--tenant", testTenantID)
	authCmd.Dir = repoRoot
	authCmd.Env = os.Environ()
	authCmd.Stdout = os.Stdout
	authCmd.Stderr = os.Stderr

	if err := authCmd.Run(); err != nil {
		t.Fatalf("tplug auth command failed: %v", err)
	}
	t.Log("OAuth command completed")

	dataFile := filepath.Join(repoRoot, ".data", "pb", "data.db")
	info, err := os.Stat(dataFile)
	if err != nil {
		t.Fatalf("expected PocketBase data at %s: %v", dataFile, err)
	}
	if info.Size() == 0 {
		t.Fatalf("PocketBase data DB appears empty: %s", dataFile)
	}

	t.Logf("PocketBase token store located at %s (size: %d bytes)", dataFile, info.Size())
	t.Log("OAuth flow completed successfully.")
}

func requireEnv(t *testing.T, name string) {
	t.Helper()
	if strings.TrimSpace(os.Getenv(name)) == "" {
		t.Fatalf("environment variable %s must be set", name)
	}
}

func buildTplugBinary(t *testing.T, repoRoot string) string {
	t.Helper()
	output := filepath.Join(t.TempDir(), "tplug")
	cmd := exec.Command("go", "build", "-o", output, "./cmd/tplug")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build tplug binary: %v", err)
	}
	return output
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repo root (go.mod)")
		}
		dir = parent
	}
}

func waitForServer(url string, timeout time.Duration) error {
	client := http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return errors.New("timeout waiting for server readiness")
}

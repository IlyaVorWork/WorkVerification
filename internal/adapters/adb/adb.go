package adb

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ADB struct {
	emulatorAddress string
}

func NewADB(host, port string) *ADB {
	return &ADB{
		emulatorAddress: host + ":" + port,
	}
}

func (adb *ADB) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "adb", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (adb *ADB) withTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, d)
}

func (adb *ADB) connect(ctx context.Context) error {
	ctx, cancel := adb.withTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err := adb.run(ctx, "connect", adb.emulatorAddress)
	if err != nil {
		return fmt.Errorf("adb connect failed: %s", out)
	}
	return nil
}

func (adb *ADB) clearLogcat(ctx context.Context) error {
	ctx, cancel := adb.withTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := adb.run(ctx, "logcat", "-c")
	return err
}

func (adb *ADB) install(ctx context.Context, apkPath string) error {
	ctx, cancel := adb.withTimeout(ctx, 30*time.Second)
	defer cancel()

	out, err := adb.run(ctx, "install", "-r", apkPath)
	if err != nil {
		return fmt.Errorf("install failed: %s", out)
	}

	// adb иногда возвращает Success в stdout
	if !strings.Contains(out, "Success") {
		return fmt.Errorf("install not successful: %s", out)
	}

	return nil
}

func (adb *ADB) launch(ctx context.Context, pkg string) error {
	ctx, cancel := adb.withTimeout(ctx, 10*time.Second)
	defer cancel()

	out, err := adb.run(ctx, "shell", "monkey", "-p", pkg, "-c", "android.intent.category.LAUNCHER", "1")
	if err != nil {
		return fmt.Errorf("launch failed: %s", out)
	}

	return nil
}

func (adb *ADB) getLogcat(ctx context.Context) (string, error) {
	ctx, cancel := adb.withTimeout(ctx, 5*time.Second)
	defer cancel()

	out, err := adb.run(ctx, "logcat", "-d")
	return out, err
}

func (adb *ADB) hasCrash(logs string) bool {
	crashMarkers := []string{
		"FATAL EXCEPTION",
		"Process: ",
		"has crashed",
		"ANR in",
	}

	for _, m := range crashMarkers {
		if strings.Contains(logs, m) {
			return true
		}
	}
	return false
}

func (adb *ADB) clearApp(ctx context.Context, pkg string) {
	ctx, cancel := adb.withTimeout(ctx, 5*time.Second)
	defer cancel()

	adb.run(ctx, "shell", "pm", "clear", pkg)
}

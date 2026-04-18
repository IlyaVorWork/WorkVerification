package adb

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/shogo82148/androidbinary"
	"io"
	"log"
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
	fmt.Println(string(out))
	return string(out), err
}

func (adb *ADB) withTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, d)
}

func (adb *ADB) connect(ctx context.Context) error {
	ctx, cancel := adb.withTimeout(ctx, 30*time.Second)
	defer cancel()

	fmt.Println("Starting adb server...")
	out, err := adb.run(ctx, "start-server")
	if err != nil {
		return fmt.Errorf("adb server start failed: %s", out)
	}

	fmt.Println("Connecting to adb...")
	out, err = adb.run(ctx, "connect", adb.emulatorAddress)
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

type Manifest struct {
	Package string `xml:"package,attr"`
}

func (adb *ADB) getPkgName(apkPath string) (string, error) {
	r, err := zip.OpenReader(apkPath)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "AndroidManifest.xml" {

			rc, err := f.Open()
			if err != nil {
				log.Fatal(err)
			}

			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				log.Fatal(err)
			}

			reader := bytes.NewReader(data)

			xmlFile, err := androidbinary.NewXMLFile(reader)
			if err != nil {
				log.Fatal(err)
			}

			decoder := xml.NewDecoder(xmlFile.Reader())

			var manifest Manifest
			if err := decoder.Decode(&manifest); err != nil {
				log.Fatal(err)
			}

			fmt.Println("Package:", manifest.Package)
			return manifest.Package, nil
		}
	}

	return "", errors.New("manifest not found")
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

func (adb *ADB) hasCrash(logs string, pkg string) bool {
	lines := strings.Split(logs, "\n")

	for _, line := range lines {

		// только ошибки рантайма
		if !strings.Contains(line, "AndroidRuntime") {
			continue
		}

		// только наш пакет
		if !strings.Contains(line, pkg) {
			continue
		}

		if strings.Contains(line, "FATAL EXCEPTION") ||
			strings.Contains(line, "has crashed") ||
			strings.Contains(line, "ANR in") {

			return true
		}
	}

	return false
}

func (adb *ADB) stopAndDelete(ctx context.Context, pkg string) error {
	ctx, cancel := adb.withTimeout(ctx, 10*time.Second)
	defer cancel()

	out, err := adb.run(ctx, "shell", "am", "force-stop", pkg)
	if err != nil {
		return fmt.Errorf("application stop failed: %s", out)
	}

	out, err = adb.run(ctx, "uninstall", pkg)
	if err != nil {
		return fmt.Errorf("uninstall failed: %s", out)
	}

	return nil
}

func (adb *ADB) kill(ctx context.Context) {
	ctx, cancel := adb.withTimeout(ctx, 5*time.Second)
	defer cancel()

	adb.run(ctx, "kill-server")
}

func (adb *ADB) Verify(ctx context.Context, apkPath string) error {

	// 1. connect
	if err := adb.connect(ctx); err != nil {
		return err
	}

	defer adb.kill(ctx)

	// 2. очистить лог
	if err := adb.clearLogcat(ctx); err != nil {
		return err
	}

	// 3. install
	if err := adb.install(ctx, apkPath); err != nil {
		return err
	}

	pkg, err := adb.getPkgName(apkPath)
	if err != nil {
		return err
	}

	// 4. launch
	if err = adb.launch(ctx, pkg); err != nil {
		return err
	}

	// 5. подождать
	time.Sleep(20 * time.Second)

	// 6. logcat
	logs, err := adb.getLogcat(ctx)
	if err != nil {
		return err
	}

	// 7. проверка
	if adb.hasCrash(logs, pkg) {
		return fmt.Errorf("app crashed during startup")
	}

	// 8. закрытие и удалени
	if err = adb.stopAndDelete(ctx, pkg); err != nil {
		return err
	}

	return nil
}

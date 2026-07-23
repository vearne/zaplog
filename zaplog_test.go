package zaplog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestDefaultLoggerNeverNil(t *testing.T) {
	DefaultLogger = otelzap.New(zap.NewNop())
	Info("before init should not panic")
}

func TestInitLoggerDefaults(t *testing.T) {
	if err := InitLogger(); err != nil {
		t.Fatal(err)
	}
	if Level() != zapcore.InfoLevel {
		t.Fatalf("default level: got %v want info", Level())
	}
	if !zapLogConfig.AlsoStdout {
		t.Fatal("expected default stdout output")
	}
	if zapLogConfig.FilePath != "" {
		t.Fatal("expected no file by default")
	}
}

func TestInitLoggerEmptyFilePath(t *testing.T) {
	if err := InitLogger(WithFile("")); err == nil {
		t.Fatal("expected error for empty file path")
	}
	if err := InitLogger(WithFile("   ")); err == nil {
		t.Fatal("expected error for blank file path")
	}
}

func TestInitLoggerCreatesParentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "logs")
	path := filepath.Join(dir, "app.log")
	if err := InitLogger(WithFile(path), WithLevel("debug")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected log dir to exist: %v", err)
	}
	Info("mkdir ok")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	if zapLogConfig.AlsoStdout {
		t.Fatal("WithFile alone should not enable stdout")
	}
}

func TestInitLoggerFileAndStdout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := InitLogger(WithFile(path), WithStdout()); err != nil {
		t.Fatal(err)
	}
	if !zapLogConfig.AlsoStdout || zapLogConfig.FilePath == "" {
		t.Fatal("expected both file and stdout")
	}
}

func TestSetLevel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := InitLogger(WithFile(path), WithLevel("info")); err != nil {
		t.Fatal(err)
	}
	Debug("should be dropped")
	if err := SetLevel("debug"); err != nil {
		t.Fatal(err)
	}
	Debug("should be kept")
	_ = Sync()

	got := mustRead(t, path)
	if strings.Contains(got, "should be dropped") {
		t.Fatalf("debug before SetLevel should be dropped, got:\n%s", got)
	}
	if !strings.Contains(got, "should be kept") {
		t.Fatalf("debug after SetLevel should be kept, got:\n%s", got)
	}
}

func TestWithJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := InitLogger(WithFile(path), WithJSON(), WithLevel("info")); err != nil {
		t.Fatal(err)
	}
	Info("json line")
	_ = Sync()
	got := mustRead(t, path)
	if !strings.Contains(got, `"msg":"json line"`) && !strings.Contains(got, `"message":"json line"`) {
		// production encoder uses "msg"
		if !strings.Contains(got, "json line") || !strings.Contains(got, "{") {
			t.Fatalf("expected JSON log, got:\n%s", got)
		}
	}
}

func TestCallerSkipDefaultReportsCallSite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := InitLogger(WithFile(path), WithLevel("debug")); err != nil {
		t.Fatal(err)
	}

	_, _, line, _ := runtime.Caller(0)
	Info("caller skip default")
	wantLine := line + 1

	got := mustRead(t, path)
	want := fmt.Sprintf("zaplog_test.go:%d", wantLine)
	if !strings.Contains(got, want) {
		t.Fatalf("want %q in log, got:\n%s", want, got)
	}
	if strings.Contains(got, "zaplog.go:") {
		t.Fatalf("caller should skip zaplog wrapper, got:\n%s", got)
	}
}

func TestCallerSkipExtraWrapper(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := InitLogger(WithFile(path), WithLevel("debug"), WithCallerSkip(2)); err != nil {
		t.Fatal(err)
	}

	_, _, line, _ := runtime.Caller(0)
	extraWrapperInfo()
	wantLine := line + 1

	got := mustRead(t, path)
	want := fmt.Sprintf("zaplog_test.go:%d", wantLine)
	if !strings.Contains(got, want) {
		t.Fatalf("want %q in log with skip=2, got:\n%s", want, got)
	}
	if strings.Contains(got, "zaplog.go:") {
		t.Fatalf("caller should skip both wrappers, got:\n%s", got)
	}
}

func TestUnknownLevel(t *testing.T) {
	if err := InitLogger(WithStdout(), WithLevel("verbose")); err == nil {
		t.Fatal("expected error for unknown level")
	}
}

func extraWrapperInfo() {
	Info("caller skip with extra wrapper")
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

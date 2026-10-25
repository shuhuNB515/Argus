// Command argus runs the Argus backend: the dual SQLite graph stores,
// the event-driven exploration engine, and the JSON HTTP API consumed by the
// shadcn/ui frontend.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/shuhuNB515/Argus/agent"
	"github.com/shuhuNB515/Argus/config"
	"github.com/shuhuNB515/Argus/selfupdate"
	"github.com/shuhuNB515/Argus/server"
)

// version is the build version, injected at release time via
// -ldflags "-X main.version=<tag>". Defaults to "dev" for local builds.
var version = "dev"

const banner = "\n  Argus\n"

// printBanner writes the startup banner + version/runtime info to stdout.
func printBanner(addr string) {
	fmt.Print(banner)
	fmt.Println("  Argus 授权安全研究工作台")
	fmt.Printf("  Version %s  ·  %s/%s  ·  %s  ·  Starting at %s\n\n",
		version, runtime.GOOS, runtime.GOARCH, runtime.Version(), addr)
}

// main only maps run's result onto the process exit code. The exit code is part
// of the update protocol — the supervising start script reads it to decide
// whether to relaunch us (see selfupdate.ExitRestart) — so the body has to live
// in a function that can *return* rather than os.Exit past its own defers.
func main() {
	os.Exit(run())
}

func run() int {
	var (
		addr        = flag.String("addr", "127.0.0.1:8787", "HTTP listen address (default accepts local connections only; use :8787 to serve other hosts)")
		dataDir     = flag.String("data", filepath.Join(config.BaseDir(), "data"), "persistent data directory (default: per-user Argus directory)")
		proxy       = flag.String("proxy", "127.0.0.1:8788", "traffic recording proxy address (empty to disable)")
		openBrowser = flag.Bool("open-browser", os.Getenv("ARGUS_NO_BROWSER") != "1", "open the workbench after startup")
	)
	flag.Parse()

	// hand the build version to the server package so GET /api/health can report it
	// to the frontend top bar.
	server.BuildVersion = version

	printBanner(*addr)

	// capture backend logs into the in-memory sink (still to stderr) so the /logs
	// page can show a live log stream. Do this first, to catch startup logs too.
	server.StartLogCapture()
	{
		if err := os.MkdirAll(config.BaseDir(), 0o700); err != nil {
			log.Printf("无法创建用户数据目录：%v", err)
			return 1
		}
		if f, err := os.OpenFile(filepath.Join(config.BaseDir(), "startup.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
			defer f.Close()
			log.SetOutput(io.MultiWriter(log.Writer(), f))
		}
	}

	// Self-update bootstrap: swap in a staged binary, or count a post-swap boot
	// attempt and roll back if the new build keeps dying. Must run before we open
	// the stores or bind a port — this may end with "exit and let the start script
	// relaunch me", and there is no point paying for either first.
	action, upState := selfupdate.Bootstrap()
	server.SetBootUpdateState(upState)
	if action == selfupdate.Restart {
		return selfupdate.ExitRestart
	}

	// surface which config file the binary reads (absolute, so `go run`'s relative
	// "config.json" — resolved against the CWD — is unambiguous).
	cfgPath := config.Path()
	if abs, e := filepath.Abs(cfgPath); e == nil {
		cfgPath = abs
	}
	if _, e := os.Stat(cfgPath); e == nil {
		log.Printf("[config] 配置文件: %s", cfgPath)
	} else {
		log.Printf("[config] 配置文件: %s（文件不存在 · 使用 Argus 内置数据库）", cfgPath)
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, shutdown := shutdownContext(sigCtx)
	defer shutdown(agent.AbortShutdown)

	mgr, err := server.NewManager(*dataDir, *proxy)
	if err != nil {
		log.Printf("启动失败：%v", err)
		return 1
	}
	defer mgr.Close()

	// Surviving this long means a freshly swapped-in build actually works, so drop
	// the upgrade marker and stop counting attempts. Until it fires, every boot
	// increments the count and a build that keeps dying gets rolled back.
	settle := time.AfterFunc(selfupdate.SettleDelay, selfupdate.Settle)
	defer settle.Stop()

	skillDir := config.SkillDir()
	if abs, err := filepath.Abs(skillDir); err == nil {
		skillDir = abs
	}
	log.Printf("[config] skill 目录: %s", skillDir)
	srv := server.New(ctx, mgr, skillDir, *dataDir, config.BaseDir())
	srv.SetListenAddress(*addr)
	srv.WarnIfExposed(*addr)
	httpSrv := &http.Server{
		Addr:              *addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Printf("无法打开工作台端口 %s：%v", *addr, err)
		return 1
	}
	go func() {
		log.Printf("Argus %s backend listening on %s (data=%s, workers=%d)", version, *addr, *dataDir, mgr.Workers())
		if err := httpSrv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()
	if *openBrowser {
		url := workbenchURL(*addr)
		if err := launchBrowser(url); err != nil {
			log.Printf("请在浏览器打开 %s（自动打开失败：%v）", url, err)
		}
	}

	// Two ways out: a signal (normal stop → exit 0, the start script stops looping)
	// or a staged update / rollback (→ exit 75, the script relaunches us and the
	// bootstrap above installs the new build).
	code := 0
	select {
	case <-ctx.Done():
	case <-server.RestartRequested():
		code = selfupdate.ExitRestart
		shutdown(agent.AbortShutdown)
	}

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	return code
}

func workbenchURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func launchBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

// shutdownContext deliberately does not derive from signalCtx. If it did, the
// parent's plain context.Canceled could win the race before AbortShutdown was
// attached to the child, losing the diagnostic cause in every running Agent.
func shutdownContext(signalCtx context.Context) (context.Context, context.CancelCauseFunc) {
	ctx, shutdown := context.WithCancelCause(context.Background())
	go func() {
		select {
		case <-signalCtx.Done():
			shutdown(agent.AbortShutdown)
		case <-ctx.Done():
		}
	}()
	return ctx, shutdown
}

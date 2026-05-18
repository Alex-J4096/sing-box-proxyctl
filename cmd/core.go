package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var coreCmd = &cobra.Command{
	Use:   "core",
	Short: "Inspect local sing-box core status.",
}

var (
	coreConfigPath string
	corePath       string
	corePidFile    string
	coreLogFile    string
)

type managedSingboxStatus struct {
	PID        int
	PidFile    string
	LogFile    string
	Running    bool
	Stale      bool
	Unexpected bool
}

var coreStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"info"},
	Short:   "Check sing-box core and managed process status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		coreFound, err := renderCoreAvailability()
		if err != nil {
			return err
		}

		if err := renderManagedStatus(coreConfigPath, corePidFile, coreLogFile); err != nil {
			return err
		}

		if !coreFound {
			return commandError("sing-box core not found")
		}
		return nil
	},
}

var coreStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a managed sing-box process.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := startManagedSingbox(coreConfigPath, corePath, corePidFile, coreLogFile); err != nil {
			return commandError("failed to start sing-box: %w", err)
		}
		return nil
	},
}

var coreStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the managed sing-box process.",
	RunE: func(cmd *cobra.Command, args []string) error {
		status := inspectManagedSingbox(coreConfigPath, corePidFile, coreLogFile)
		if err := stopManagedSingbox(defaultPidFile(coreConfigPath, corePidFile)); err != nil {
			return commandError("failed to stop sing-box: %w", err)
		}
		if status.Running {
			pterm.Success.Println("sing-box stopped")
		} else {
			pterm.Warning.Println("managed sing-box is not running")
		}
		return nil
	},
}

var corePsCmd = &cobra.Command{
	Use:   "ps",
	Short: "Show managed sing-box process status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return renderManagedStatus(coreConfigPath, corePidFile, coreLogFile)
	},
}

var coreRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart sing-box using the current config.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := restartSingbox(coreConfigPath, corePath, corePidFile, coreLogFile); err != nil {
			return commandError("failed to restart sing-box: %w", err)
		}
		return nil
	},
}

func renderCoreAvailability() (bool, error) {
	rows := pterm.TableData{
		{"source", "path", "status", "version"},
	}

	found := false

	if path, err := exec.LookPath("sing-box"); err == nil {
		found = true
		rows = append(rows, []string{"PATH", path, "found", singboxVersion(path)})
	} else {
		rows = append(rows, []string{"PATH", "-", "missing", "-"})
	}

	localPath := localSingboxPath()
	localStatus := inspectLocalCore(localPath)
	if localStatus == "found" {
		found = true
	}
	version := "-"
	if localStatus == "found" {
		version = singboxVersion(localPath)
	}
	rows = append(rows, []string{"local", localPath, localStatus, version})

	if err := pterm.DefaultTable.
		WithHasHeader().
		WithHeaderRowSeparator("-").
		WithData(rows).
		Render(); err != nil {
		return false, commandError("failed to render core status: %w", err)
	}

	if found {
		pterm.Success.Println("sing-box core is available")
		return true, nil
	}

	return false, nil
}

func renderManagedStatus(configPath, pidFile, logFile string) error {
	status := inspectManagedSingbox(configPath, pidFile, logFile)
	state := "stopped"
	switch {
	case status.Running:
		state = "running"
	case status.Stale:
		state = "stale"
	case status.Unexpected:
		state = "unexpected"
	}

	pid := "-"
	if status.PID > 0 {
		pid = strconv.Itoa(status.PID)
	}

	rows := pterm.TableData{
		{"state", "pid", "pid_file", "log_file"},
		{state, pid, status.PidFile, status.LogFile},
	}

	if err := pterm.DefaultTable.
		WithHasHeader().
		WithHeaderRowSeparator("-").
		WithData(rows).
		Render(); err != nil {
		return commandError("failed to render process status: %w", err)
	}

	switch {
	case status.Running:
		pterm.Success.Println(fmt.Sprintf("managed sing-box is running: pid %d", status.PID))
	case status.Stale:
		pterm.Warning.Println("managed sing-box pid file is stale")
	case status.Unexpected:
		pterm.Warning.Println("pid file points to a non sing-box process")
	default:
		pterm.Warning.Println("managed sing-box is not running")
	}
	return nil
}

func inspectManagedSingbox(configPath, pidFile, logFile string) managedSingboxStatus {
	status := managedSingboxStatus{
		PidFile: defaultPidFile(configPath, pidFile),
		LogFile: defaultLogFile(configPath, logFile),
	}

	pid, err := readPidFile(status.PidFile)
	if err != nil {
		return status
	}
	status.PID = pid
	if pid <= 0 || !processRunning(pid) {
		status.Stale = true
		return status
	}
	if !looksLikeSingboxProcess(pid) {
		status.Unexpected = true
		return status
	}
	status.Running = true
	return status
}

var coreCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check whether sing-box core is available locally.",
	RunE: func(cmd *cobra.Command, args []string) error {
		found, err := renderCoreAvailability()
		if err != nil {
			return err
		}
		if !found {
			return commandError("sing-box core not found")
		}
		return nil
	},
}

func localSingboxPath() string {
	name := "sing-box"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(".", "bin", "sing-box", name)
}

func inspectLocalCore(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "missing"
		}
		return "unreadable"
	}
	if info.IsDir() {
		return "not executable"
	}
	if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
		return "not executable"
	}
	return "found"
}

func singboxVersion(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output, err := exec.CommandContext(ctx, path, "version").CombinedOutput()
	if err != nil {
		return "unknown"
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return "unknown"
}

func restartSingbox(configPath, explicitCorePath, pidFile, logFile string) error {
	if err := stopManagedSingbox(defaultPidFile(configPath, pidFile)); err != nil {
		return err
	}
	if err := startManagedSingbox(configPath, explicitCorePath, pidFile, logFile); err != nil {
		return err
	}
	pterm.Success.Println("sing-box restart completed")
	return nil
}

func startManagedSingbox(configPath, explicitCorePath, pidFile, logFile string) error {
	current := inspectManagedSingbox(configPath, pidFile, logFile)
	if current.Running {
		pterm.Success.Println(fmt.Sprintf("sing-box already running: pid %d", current.PID))
		pterm.Info.Println("pid-file: " + current.PidFile)
		pterm.Info.Println("log: " + current.LogFile)
		return nil
	}
	if current.Stale || current.Unexpected {
		_ = os.Remove(current.PidFile)
	}

	resolvedCorePath, err := resolveSingboxCore(explicitCorePath)
	if err != nil {
		return err
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}
	if _, err := os.Stat(absConfigPath); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	pidPath := defaultPidFile(configPath, pidFile)
	logPath := defaultLogFile(configPath, logFile)
	if err := os.MkdirAll(filepath.Dir(pidPath), 0755); err != nil {
		return fmt.Errorf("create pid dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	logWriter, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logWriter.Close()

	command := exec.Command(resolvedCorePath, "run", "-c", absConfigPath)
	command.Stdout = logWriter
	command.Stderr = logWriter

	if err := command.Start(); err != nil {
		return fmt.Errorf("start sing-box: %w", err)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- command.Wait()
	}()

	select {
	case err := <-waitCh:
		detail := strings.TrimSpace(readLogTail(logPath, 2048))
		if detail != "" {
			return fmt.Errorf("sing-box exited after start: %s", detail)
		}
		if err != nil {
			return fmt.Errorf("sing-box exited after start: %w", err)
		}
		return fmt.Errorf("sing-box exited after start, check log: %s", logPath)
	case <-time.After(500 * time.Millisecond):
	}

	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(command.Process.Pid)+"\n"), 0644); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}

	pterm.Success.Println(fmt.Sprintf("sing-box started: pid %d", command.Process.Pid))
	pterm.Info.Println("core: " + resolvedCorePath)
	pterm.Info.Println("config: " + absConfigPath)
	pterm.Info.Println("log: " + logPath)
	return nil
}

func stopManagedSingbox(pidFile string) error {
	pid, err := readPidFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if pid <= 0 {
		return nil
	}
	if !processRunning(pid) {
		_ = os.Remove(pidFile)
		return nil
	}
	if !looksLikeSingboxProcess(pid) {
		_ = os.Remove(pidFile)
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find sing-box process: %w", err)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop sing-box process: %w", err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !processRunning(pid) {
			_ = os.Remove(pidFile)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("kill sing-box process: %w", err)
	}
	_ = os.Remove(pidFile)
	return nil
}

func resolveSingboxCore(explicitPath string) (string, error) {
	if explicitPath != "" {
		if inspectLocalCore(explicitPath) != "found" {
			return "", fmt.Errorf("core is not executable: %s", explicitPath)
		}
		return explicitPath, nil
	}
	if path, err := exec.LookPath("sing-box"); err == nil {
		return path, nil
	}
	localPath := localSingboxPath()
	if inspectLocalCore(localPath) == "found" {
		return localPath, nil
	}
	return "", fmt.Errorf("sing-box core not found")
}

func defaultPidFile(configPath, pidFile string) string {
	if pidFile != "" {
		return pidFile
	}
	return filepath.Join(filepath.Dir(configPath), ".proxyctl-sing-box.pid")
}

func defaultLogFile(configPath, logFile string) string {
	if logFile != "" {
		return logFile
	}
	return filepath.Join(filepath.Dir(configPath), ".proxyctl-sing-box.log")
}

func readPidFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse pid file: %w", err)
	}
	return pid, nil
}

func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func readLogTail(path string, limit int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(data) > limit {
		data = data[len(data)-limit:]
	}
	lines := strings.Split(string(data), "\n")
	start := len(lines) - 6
	if start < 0 {
		start = 0
	}
	return strings.Join(lines[start:], "\n")
}

func looksLikeSingboxProcess(pid int) bool {
	if runtime.GOOS != "linux" {
		return true
	}
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "sing-box")
}

func init() {
	rootCmd.AddCommand(coreCmd)
	coreCmd.AddCommand(coreStatusCmd)
	coreCmd.AddCommand(coreCheckCmd)
	coreCmd.AddCommand(coreStartCmd)
	coreCmd.AddCommand(coreStopCmd)
	coreCmd.AddCommand(corePsCmd)
	coreCmd.AddCommand(coreRestartCmd)

	coreCmd.PersistentFlags().StringVarP(&coreConfigPath, "config", "c", util.DefaultConfigPath(), "sing-box config path")
	coreCmd.PersistentFlags().StringVar(&corePath, "core", "", "sing-box core path")
	coreCmd.PersistentFlags().StringVar(&corePidFile, "pid-file", "", "managed sing-box pid file")
	coreCmd.PersistentFlags().StringVar(&coreLogFile, "log-file", "", "managed sing-box log file")
}

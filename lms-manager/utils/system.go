package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GetCPUUsage returns the current CPU usage percentage
func GetCPUUsage() (float64, error) {
	// Read /proc/stat
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()

	fields := strings.Fields(line)
	if len(fields) < 8 {
		return 0, fmt.Errorf("invalid /proc/stat format")
	}

	// Parse CPU times
	var times []int64
	for i := 1; i < 8; i++ {
		val, err := strconv.ParseInt(fields[i], 10, 64)
		if err != nil {
			return 0, err
		}
		times = append(times, val)
	}

	// Calculate CPU usage
	idle := times[3]
	total := int64(0)
	for _, t := range times {
		total += t
	}

	// Simple calculation (not 100% accurate but lightweight)
	usage := float64(total-idle) / float64(total) * 100
	if usage > 100 {
		usage = 100
	}
	if usage < 0 {
		usage = 0
	}

	return usage, nil
}

// GetMemoryUsage returns the current memory usage percentage
func GetMemoryUsage() (float64, error) {
	// Read /proc/meminfo
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var total, available int64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) >= 2 {
			switch fields[0] {
			case "MemTotal:":
				total, _ = strconv.ParseInt(fields[1], 10, 64)
			case "MemAvailable:":
				available, _ = strconv.ParseInt(fields[1], 10, 64)
			}
		}
	}

	if total == 0 {
		return 0, fmt.Errorf("could not read memory info")
	}

	used := total - available
	usage := float64(used) / float64(total) * 100

	return usage, nil
}

// GetDiskUsage returns the current disk usage percentage
func GetDiskUsage(path string) (float64, error) {
	// Use df command
	cmd := exec.Command("df", "-h", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("invalid df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return 0, fmt.Errorf("invalid df output format")
	}

	// Parse usage percentage (e.g., "85%")
	usageStr := fields[4]
	usageStr = strings.TrimSuffix(usageStr, "%")
	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		return 0, err
	}

	return usage, nil
}

// GetUptime returns the system uptime in seconds
func GetUptime() (int64, error) {
	// Read /proc/uptime
	file, err := os.Open("/proc/uptime")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()

	fields := strings.Fields(line)
	if len(fields) < 1 {
		return 0, fmt.Errorf("invalid /proc/uptime format")
	}

	uptime, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, err
	}

	return int64(uptime), nil
}

// GetLoadAverage returns the system load average
func GetLoadAverage() ([]float64, error) {
	// Read /proc/loadavg
	file, err := os.Open("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	line := scanner.Text()

	fields := strings.Fields(line)
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid /proc/loadavg format")
	}

	var loadAvg []float64
	for i := 0; i < 3; i++ {
		val, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return nil, err
		}
		loadAvg = append(loadAvg, val)
	}

	return loadAvg, nil
}

// GetNetworkStats returns network I/O statistics
func GetNetworkStats() (map[string]interface{}, error) {
	// Read /proc/net/dev
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := make(map[string]interface{})
	scanner := bufio.NewScanner(file)

	// Skip header lines
	for i := 0; i < 2; i++ {
		scanner.Scan()
	}

	var totalBytesReceived, totalBytesSent uint64

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) >= 10 {
			interfaceName := strings.TrimSuffix(fields[0], ":")
			
			// Skip loopback interface
			if interfaceName == "lo" {
				continue
			}

			bytesReceived, _ := strconv.ParseUint(fields[1], 10, 64)
			bytesSent, _ := strconv.ParseUint(fields[9], 10, 64)

			totalBytesReceived += bytesReceived
			totalBytesSent += bytesSent
		}
	}

	stats["bytes_received"] = totalBytesReceived
	stats["bytes_sent"] = totalBytesSent

	return stats, nil
}

// IsProcessRunning checks if a process is running
func IsProcessRunning(processName string) (bool, error) {
	cmd := exec.Command("pgrep", processName)
	err := cmd.Run()
	if err != nil {
		// Process not found
		return false, nil
	}
	return true, nil
}

// GetProcessPID returns the PID of a process
func GetProcessPID(processName string) (int, error) {
	cmd := exec.Command("pgrep", processName)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	pidStr := strings.TrimSpace(string(output))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0, err
	}

	return pid, nil
}

// ExecuteCommand executes a shell command
func ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetFileSize returns the size of a file
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetDirectorySize returns the size of a directory
func GetDirectorySize(path string) (int64, error) {
	var size int64
	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}

// FormatBytes formats bytes into human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration into human readable format
func FormatDuration(seconds int64) string {
	duration := time.Duration(seconds) * time.Second
	
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// ParseInt parses a string to int
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/B3ns44d/go-ssh-key-copier/internal/utils"
)

// ParseVMFile reads the file containing a list of VMs (hostnames or IPs).
func ParseVMFile(filePath string) (*AppConfig, error) {
	expandedFilePath := utils.ExpandPath(filePath)
	file, err := os.Open(expandedFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not open VM file %s: %w", expandedFilePath, err)
	}
	defer file.Close()

	appConfig := &AppConfig{
		VMs:        make([]VMInfo, 0),
		SSHKeyPath: utils.GetDefaultSSHKeyPath(), // Initialize with default, can be overridden
	}

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") { // Skip empty lines and comments
			continue
		}

		if strings.ContainsAny(line, " \t") {
			log.Printf("Warning: VM file line %d: Entry '%s' contains whitespace. Ensure each line has only one hostname/IP.", lineNumber, line)
		}

		vm := VMInfo{Name: line, LineNumber: lineNumber}
		appConfig.VMs = append(appConfig.VMs, vm)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading VM file %s: %w", expandedFilePath, err)
	}

	return appConfig, nil
}

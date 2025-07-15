package copier

import (
	"fmt"
	"github.com/B3ns44d/go-ssh-key-copier/internal/config"
	"log"
	"os"
	"os/exec"
)

// ExecutionResult holds the outcome of an operation on a VM
type ExecutionResult struct {
	VMName  string
	Message string
	Success bool
	Error   error
}

// CheckSSHPass checks if sshpass is available in PATH
func CheckSSHPass() (string, bool) {
	path, err := exec.LookPath("sshpass")
	if err != nil {
		return "", false
	}
	return path, true
}

// ProcessVM handles the SSH key copying logic for a single VM.
func ProcessVM(
	vm config.VMInfo,
	username string,
	password string,
	sshKeyPath string,
	dryRun bool,
	sshpassPath string,
	useSSHPass bool,
	acceptHostKeys bool,
) ExecutionResult {

	if username == "" { // Should be guaranteed by main, but good to check
		return ExecutionResult{
			VMName:  vm.Name,
			Message: "Username is missing for processing.",
			Success: false,
			Error:   fmt.Errorf("internal error: username not provided to ProcessVM"),
		}
	}
	if sshKeyPath == "" {
		return ExecutionResult{
			VMName:  vm.Name,
			Message: fmt.Sprintf("SSH key path is not configured for VM %s (from config line: %d). Specify via --ssh-key or ensure default key exists", vm.Name, vm.LineNumber),
			Success: false,
			Error:   fmt.Errorf("ssh key path missing"),
		}
	}

	// SSH Key Path should already be expanded by main.go
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		return ExecutionResult{
			VMName:  vm.Name,
			Message: fmt.Sprintf("SSH key path '%s' does not exist for VM %s (from line: %d)", sshKeyPath, vm.Name, vm.LineNumber),
			Success: false,
			Error:   fmt.Errorf("ssh key file not found: %s", sshKeyPath),
		}
	}

	target := fmt.Sprintf("%s@%s", username, vm.Name)
	var cmdArgs []string
	var commandToRun string

	var baseArgs []string
	// Use sshpass if enabled, available, and password is provided (password will always be provided if prompt was successful)
	if useSSHPass && password != "" && sshpassPath != "" {
		commandToRun = sshpassPath
		baseArgs = []string{"-p", password, "ssh-copy-id"}
	} else {
		commandToRun = "ssh-copy-id"
		if password != "" {
			log.Printf("Warning: [VM: %s] Password was provided, but sshpass is not found/enabled. ssh-copy-id will attempt to use its interactive prompt.", vm.Name)
		}
	}

	if acceptHostKeys {
		baseArgs = append(baseArgs, "-o", "StrictHostKeyChecking=accept-new")
	}

	cmdArgs = append(baseArgs, "-i", sshKeyPath, target)

	// Construct the full command string for logging, obscuring password
	logCmdStr := commandToRun
	for i, arg := range cmdArgs {
		isPasswordArg := false
		if commandToRun == sshpassPath && len(cmdArgs) > i && i > 0 && cmdArgs[i-1] == "-p" {
			isPasswordArg = true
		}

		if isPasswordArg {
			logCmdStr += " '********'"
		} else {
			logCmdStr += " " + arg
		}
	}

	if dryRun {
		log.Printf("[VM: %s] Dry Run: Would execute: %s\n", vm.Name, logCmdStr)
		return ExecutionResult{
			VMName:  vm.Name,
			Message: fmt.Sprintf("Dry run: Command for %s: %s", target, logCmdStr),
			Success: true,
		}
	}

	log.Printf("[VM: %s] Attempting: %s\n", vm.Name, logCmdStr)
	if useSSHPass && password != "" && sshpassPath != "" {
		log.Printf("[VM: %s] Using sshpass for password authentication.\n", vm.Name)
	}

	cmd := exec.Command(commandToRun, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Only set Stdin if not using sshpass or if password is empty,
	// to allow ssh-copy-id to prompt if sshpass fails or isn't used.
	if !(useSSHPass && password != "" && sshpassPath != "") {
		cmd.Stdin = os.Stdin
	}

	err := cmd.Run()
	if err != nil {
		return ExecutionResult{
			VMName:  vm.Name,
			Message: fmt.Sprintf("Failed to copy SSH key to %s", target),
			Success: false,
			Error:   err,
		}
	}

	return ExecutionResult{
		VMName:  vm.Name,
		Message: fmt.Sprintf("Successfully copied SSH key to %s", target),
		Success: true,
	}
}

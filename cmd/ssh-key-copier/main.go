package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/term"

	"github.com/B3ns44d/go-ssh-key-copier/internal/config"
	"github.com/B3ns44d/go-ssh-key-copier/internal/copier"
	"github.com/B3ns44d/go-ssh-key-copier/internal/utils"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	vmListFile := flag.String("vmsfile", "~/.config/go-ssh-key-copier/vms.list", "Path to the file containing a list of VM hostnames or IPs, one per line.")
	userName := flag.String("user", "", "Username for connecting to all VMs (required).")
	sshKeyPathFlag := flag.String("ssh-key", "", "Path to the SSH public key to copy. Overrides default key discovery.")
	dryRun := flag.Bool("dry-run", false, "Print commands that would be executed without running them.")
	parallelism := flag.Int("parallelism", 5, "Number of VMs to process in parallel (0 or 1 for sequential).")

	flag.Parse()

	if *userName == "" {
		log.Println("Error: -user flag is required.")
		flag.Usage()
		os.Exit(1)
	}

	// Prompt for password securely
	fmt.Print("Enter SSH password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	password := strings.TrimSpace(string(passwordBytes))
	// Clear passwordBytes from memory as soon as possible, though Go's GC makes this tricky.
	// For truly sensitive scenarios, more complex memory handling would be needed.
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}

	appConfig, err := config.ParseVMFile(*vmListFile)
	if err != nil {
		log.Fatalf("Error loading VM list from %s: %v", *vmListFile, err)
	}

	// Determine final SSH key path
	finalSSHKeyPath := appConfig.SSHKeyPath // Path from default discovery
	if *sshKeyPathFlag != "" {
		finalSSHKeyPath = utils.ExpandPath(*sshKeyPathFlag) // Override with flag
		log.Printf("Using SSH key path from command line flag: %s", finalSSHKeyPath)
	} else if finalSSHKeyPath == "" { // If still empty (no default found)
		log.Fatalf("SSH key path could not be determined. Please use --ssh-key flag or ensure a default key (e.g. ~/.ssh/id_ed25519.pub) exists.")
	}
	// Note: appConfig.SSHKeyPath is not explicitly updated here, finalSSHKeyPath is passed directly.

	if len(appConfig.VMs) == 0 {
		log.Println("No VMs found in the VM list file:", *vmListFile)
		return
	}

	sshpassPath, sshpassFound := copier.CheckSSHPass()
	useSSHPass := sshpassFound
	// if *noSSHPassFlag {
	// 	useSSHPass = false
	// 	log.Println("sshpass usage has been explicitly disabled via flag.")
	// }

	if !sshpassFound && password != "" {
		log.Println("sshpass command not found in PATH. Will rely on ssh-copy-id's interactive password prompt if password is used.")
	} else if sshpassFound && password != "" {
		log.Println("sshpass found at:", sshpassPath, ". Will be used for password authentication.")
	} else if password == "" {
		log.Println("No password entered. ssh-copy-id might prompt if needed or fail if password is required by servers.")
	}

	fmt.Printf("\n--- Execution Plan ---\n")
	fmt.Printf("VM List File: %s\n", *vmListFile)
	fmt.Printf("Target Username: %s\n", *userName)
	fmt.Printf("Found %d VMs to process.\n", len(appConfig.VMs))
	fmt.Printf("SSH Key to use: %s\n", finalSSHKeyPath)
	fmt.Printf("Dry Run: %t\n", *dryRun)
	fmt.Printf("Parallelism: %d\n", *parallelism)
	fmt.Printf("Attempting to use sshpass if available: %t\n", useSSHPass && password != "" && sshpassPath != "")
	fmt.Println("----------------------\n")

	actualParallelism := *parallelism
	if actualParallelism <= 0 {
		actualParallelism = 1
	}
	if actualParallelism > len(appConfig.VMs) && len(appConfig.VMs) > 0 {
		actualParallelism = len(appConfig.VMs)
	}

	var wg sync.WaitGroup
	jobs := make(chan config.VMInfo, len(appConfig.VMs))
	resultsChan := make(chan copier.ExecutionResult, len(appConfig.VMs))

	for w := 1; w <= actualParallelism; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for vm := range jobs {
				log.Printf("[Worker %d] Processing VM: %s (from file line %d)\n", workerID, vm.Name, vm.LineNumber)
				result := copier.ProcessVM(vm, *userName, password, finalSSHKeyPath, *dryRun, sshpassPath, useSSHPass)
				resultsChan <- result
			}
		}(w)
	}

	for _, vm := range appConfig.VMs {
		jobs <- vm
	}
	close(jobs)

	wg.Wait()
	close(resultsChan)
	password = "" // Attempt to clear password from memory, best effort

	fmt.Println("\n--- Execution Results ---")
	successCount := 0
	errorCount := 0
	for result := range resultsChan {
		if result.Success {
			successCount++
			log.Printf("SUCCESS: %s - %s", result.VMName, result.Message)
		} else {
			errorCount++
			errMsg := fmt.Sprintf("ERROR: %s - %s", result.VMName, result.Message)
			if result.Error != nil {
				errMsg += fmt.Sprintf(" (Details: %v)", result.Error)
			}
			log.Println(errMsg)
		}
	}
	fmt.Println("-------------------------")
	log.Printf("SSH key copying process finished. Succeeded: %d, Failed: %d\n", successCount, errorCount)

	if errorCount > 0 {
		os.Exit(1)
	}
}

# Go SSH Key Copier

# Description
`go-ssh-key-copier` is a command-line utility written in Go to copy SSH public keys to multiple remote Virtual Machines (VMs). It reads a list of VMs from a configuration file and can use `sshpass` for password authentication if available, or fall back to `ssh-copy-id`'s interactive password prompt.

# Features
* Copies SSH keys to multiple VMs.
* Supports parallel execution for faster processing.
* Dry run mode to preview commands without execution.
* Reads VM hostnames/IPs from a configuration file.
* Allows specifying a custom SSH key path.
* Prompts for SSH password securely.
* Can utilize `sshpass` if available for automated password input.
* Skips empty lines and comments in the VM list file.

# Usage
The basic command structure is:
```bash
go-ssh-key-copier -vmsfile <path_to_vm_list> -user <username> [options]
```

Command-line flags:
*   `-vmsfile`: Path to the file containing VM hostnames or IPs (default: `~/.config/go-ssh-key-copier/vms.list`).
*   `-user`: Username for connecting to all VMs (required).
*   `-ssh-key`: Path to the SSH public key to copy. Overrides default key discovery (e.g., `~/.ssh/id_ed25519.pub`).
*   `-dry-run`: Print commands that would be executed without running them (default: `false`).
*   `-parallelism`: Number of VMs to process in parallel (0 or 1 for sequential, default: `5`).
*   `--accept-host-keys`: Automatically accept new SSH host keys. **Security Warning:** This option can be a security risk. By automatically accepting new host keys, you are vulnerable to man-in-the-middle attacks. Only use this option in trusted networks.

# Configuration
The VM list file should contain one VM hostname or IP address per line.
Lines starting with `#` are treated as comments and are ignored.
Empty lines are also skipped.

Example `vms.list` file:
```
# Production Servers
server1.example.com
192.168.1.100

# Staging Servers
staging-vm
```

# Installation
Prerequisites:
* Go (latest stable version recommended, or version compatible with `go.mod` if present).

Using `go install`:
```bash
go install github.com/B3ns44d/go-ssh-key-copier/cmd/ssh-key-copier@latest
```

Building from source:
```bash
git clone https://github.com/B3ns44d/go-ssh-key-copier.git
cd go-ssh-key-copier
make build
```

Optional Dependency:
* `sshpass`: For automated password input without interactive prompts. Install it using your system's package manager (e.g., `sudo apt-get install sshpass` on Debian/Ubuntu).

# Contributing
Contributions are welcome! Please open an issue or submit a pull request.
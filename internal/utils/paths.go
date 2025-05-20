package utils

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// GetDefaultSSHKeyPath tries to find a common default SSH public key.
func GetDefaultSSHKeyPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Println("Warning: Could not get current user to determine default SSH key path:", err)
		return ""
	}
	keyPaths := []string{
		filepath.Join(usr.HomeDir, ".ssh", "id_ed25519.pub"),
		filepath.Join(usr.HomeDir, ".ssh", "id_rsa.pub"),
	}
	for _, p := range keyPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	log.Println("Warning: Could not find a default SSH key (id_ed25519.pub or id_rsa.pub) in ~/.ssh/")
	return ""
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			log.Printf("Warning: Could not expand home directory for path %s: %v", path, err)
			return path // Return original path if expansion fails
		}
		return filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}

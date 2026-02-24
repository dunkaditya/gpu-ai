package wireguard

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

//go:embed templates/bootstrap.sh.tmpl
var bootstrapTmplStr string

var bootstrapTmpl = template.Must(
	template.New("bootstrap").Parse(bootstrapTmplStr),
)

// BootstrapData holds all parameters for rendering the cloud-init bootstrap script.
type BootstrapData struct {
	InstanceID         string // e.g., "gpu-4a7f"
	ProxyEndpoint      string // e.g., "203.0.113.1:51820"
	ProxyPublicKey     string // base64 WireGuard public key
	InstancePrivateKey string // base64 WireGuard private key (decrypted, NOT the encrypted form)
	InstanceAddress    string // e.g., "10.0.0.5/16" (includes CIDR mask)
	AllowedIPs         string // e.g., "10.0.0.0/16" (subnet CIDR for WireGuard AllowedIPs)
	SSHAuthorizedKeys  string // newline-separated SSH public keys
	InternalToken      string // per-instance callback auth token
	Hostname           string // e.g., "gpu-4a7f.gpu.ai"
	CallbackURL        string // full URL: "https://api.gpu.ai/internal/instances/{id}/ready"
}

var (
	// instanceIDRegex allows only alphanumeric characters and hyphens.
	instanceIDRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	// hostnameRegex allows only alphanumeric characters, hyphens, and dots.
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

	// sshKeyRegex validates SSH public key format: type base64 comment
	sshKeyRegex = regexp.MustCompile(`^(ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp(256|384|521)) [A-Za-z0-9+/=]+ .*$`)

	// shellInjectionChars are characters that could enable shell injection in SSH key comments.
	shellInjectionChars = []string{"$(", "`", ";", "|", "&"}
)

// ValidateBootstrapData validates all fields of BootstrapData, returning a combined
// error listing all validation failures.
func ValidateBootstrapData(data BootstrapData) error {
	var errs []string

	if data.InstanceID == "" {
		errs = append(errs, "InstanceID is required")
	} else if !instanceIDRegex.MatchString(data.InstanceID) {
		errs = append(errs, "InstanceID contains invalid characters (only alphanumeric and hyphens allowed)")
	}

	if data.ProxyEndpoint == "" {
		errs = append(errs, "ProxyEndpoint is required")
	}

	if data.ProxyPublicKey == "" {
		errs = append(errs, "ProxyPublicKey is required")
	} else if len(data.ProxyPublicKey) != 44 || data.ProxyPublicKey[43] != '=' {
		errs = append(errs, "ProxyPublicKey must be a 44-character base64 string ending with '='")
	}

	if data.InstancePrivateKey == "" {
		errs = append(errs, "InstancePrivateKey is required")
	}

	if data.InstanceAddress == "" {
		errs = append(errs, "InstanceAddress is required")
	} else if !strings.Contains(data.InstanceAddress, "/") {
		errs = append(errs, "InstanceAddress must be in CIDR notation (contain '/')")
	}

	if data.AllowedIPs == "" {
		errs = append(errs, "AllowedIPs is required")
	}

	if data.SSHAuthorizedKeys == "" {
		errs = append(errs, "SSHAuthorizedKeys is required")
	} else {
		// Validate each SSH key line.
		lines := strings.Split(data.SSHAuthorizedKeys, "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Check for shell injection characters first.
			for _, ch := range shellInjectionChars {
				if strings.Contains(line, ch) {
					errs = append(errs, fmt.Sprintf("SSHAuthorizedKeys line %d contains forbidden character %q", i+1, ch))
					break
				}
			}
			// Validate SSH key format.
			if !sshKeyRegex.MatchString(line) {
				errs = append(errs, fmt.Sprintf("SSHAuthorizedKeys line %d is not a valid SSH public key", i+1))
			}
		}
	}

	if data.InternalToken == "" {
		errs = append(errs, "InternalToken is required")
	}

	if data.Hostname == "" {
		errs = append(errs, "Hostname is required")
	} else if !hostnameRegex.MatchString(data.Hostname) {
		errs = append(errs, "Hostname contains invalid characters (only alphanumeric, hyphens, and dots allowed)")
	}

	if data.CallbackURL == "" {
		errs = append(errs, "CallbackURL is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("bootstrap data validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// RenderBootstrap renders the cloud-init bootstrap script with the provided data.
// It validates inputs before rendering and returns the rendered script as a string.
func RenderBootstrap(data BootstrapData) (string, error) {
	if err := ValidateBootstrapData(data); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := bootstrapTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("wireguard: render bootstrap template: %w", err)
	}

	return buf.String(), nil
}

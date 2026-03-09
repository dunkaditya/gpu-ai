package tunnel

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

var (
	// instanceIDRegex allows only alphanumeric characters and hyphens.
	instanceIDRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	// hostnameRegex allows only alphanumeric characters, hyphens, and dots.
	hostnameRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

	// shellInjectionChars are characters that could enable shell injection.
	shellInjectionChars = []string{"$(", "`", ";", "|", "&"}
)

// ValidateBootstrapData validates all fields of BootstrapData, returning a
// combined error listing all validation failures.
func ValidateBootstrapData(data BootstrapData) error {
	var errs []string

	if data.InstanceID == "" {
		errs = append(errs, "InstanceID is required")
	} else if !instanceIDRegex.MatchString(data.InstanceID) {
		errs = append(errs, "InstanceID contains invalid characters")
	}

	if data.ProxyHost == "" {
		errs = append(errs, "ProxyHost is required")
	}

	if data.FRPServerPort <= 0 {
		errs = append(errs, "FRPServerPort must be positive")
	}

	if data.FRPToken == "" {
		errs = append(errs, "FRPToken is required")
	}

	if data.RemotePort < MinPort || data.RemotePort > MaxPort {
		errs = append(errs, fmt.Sprintf("RemotePort must be in range [%d, %d]", MinPort, MaxPort))
	}

	if data.SSHAuthorizedKeys == "" {
		errs = append(errs, "SSHAuthorizedKeys is required")
	} else {
		lines := strings.Split(data.SSHAuthorizedKeys, "\n")
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			for _, ch := range shellInjectionChars {
				if strings.Contains(line, ch) {
					errs = append(errs, fmt.Sprintf("SSHAuthorizedKeys line %d contains forbidden character %q", i+1, ch))
					break
				}
			}
		}
	}

	if data.InternalToken == "" {
		errs = append(errs, "InternalToken is required")
	}

	if data.Hostname == "" {
		errs = append(errs, "Hostname is required")
	} else if !hostnameRegex.MatchString(data.Hostname) {
		errs = append(errs, "Hostname contains invalid characters")
	}

	if data.CallbackURL == "" {
		errs = append(errs, "CallbackURL is required")
	}

	if len(errs) > 0 {
		return fmt.Errorf("bootstrap data validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// RenderBootstrap renders the instance bootstrap script with the provided data.
// It validates inputs before rendering and returns the rendered script as a string.
func RenderBootstrap(data BootstrapData) (string, error) {
	if err := ValidateBootstrapData(data); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := bootstrapTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("tunnel: render bootstrap template: %w", err)
	}

	return buf.String(), nil
}

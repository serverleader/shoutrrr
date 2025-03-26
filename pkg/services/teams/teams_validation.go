package teams

import (
	"fmt"
	"regexp"
)

const (
	UUID4Length        = 36
	HashLength         = 32
	WebhookDomain      = ".webhook.office.com"
	ExpectedComponents = 7 // 1 match + 6 captures
	Path               = "webhookb2"
	ProviderName       = "IncomingWebhook"
)

// parseAndVerifyWebhookURL extracts and validates webhook components from a URL.
func parseAndVerifyWebhookURL(webhookURL string) ([5]string, error) {
	pattern, err := regexp.Compile(
		`https://([^.]+)` + WebhookDomain + `/` + Path + `/([0-9a-f-]{36})@([0-9a-f-]{36})/` + ProviderName + `/([0-9a-f]{32})/([0-9a-f-]{36})/([^/]+)`,
	)
	if err != nil {
		return [5]string{}, err
	}
	groups := pattern.FindStringSubmatch(webhookURL)
	if len(groups) != ExpectedComponents {
		return [5]string{}, fmt.Errorf(
			"invalid webhook URL format: expected %d components, got %d",
			ExpectedComponents,
			len(groups),
		)
	}
	parts := [5]string{groups[2], groups[3], groups[4], groups[5], groups[6]}
	if err := verifyWebhookParts(parts); err != nil {
		return [5]string{}, err
	}
	return parts, nil
}

// verifyWebhookParts ensures webhook components meet format requirements.
func verifyWebhookParts(parts [5]string) error {
	if len(parts[0]) != UUID4Length && parts[0] != "" {
		return fmt.Errorf("group ID must be %d characters, got %d", UUID4Length, len(parts[0]))
	}
	if len(parts[1]) != UUID4Length && parts[1] != "" {
		return fmt.Errorf("tenant ID must be %d characters, got %d", UUID4Length, len(parts[1]))
	}
	if len(parts[2]) != HashLength && parts[2] != "" {
		return fmt.Errorf("altID must be %d characters, got %d", HashLength, len(parts[2]))
	}
	if len(parts[3]) != UUID4Length && parts[3] != "" {
		return fmt.Errorf("groupOwner must be %d characters, got %d", UUID4Length, len(parts[3]))
	}
	if parts[4] == "" {
		return fmt.Errorf("extraID is required")
	}
	return nil
}

// BuildWebhookURL constructs a Teams webhook URL from components.
func BuildWebhookURL(host, group, tenant, altID, groupOwner, extraID string) string {
	return fmt.Sprintf("https://%s/%s/%s@%s/%s/%s/%s/%s",
		host, Path, group, tenant, ProviderName, altID, groupOwner, extraID)
}

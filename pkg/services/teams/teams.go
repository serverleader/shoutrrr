package teams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/services/standard"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// Constants for summary truncation and webhook parsing.
const (
	MaxSummaryLength    = 20 // Maximum length of summary before truncation
	TruncatedSummaryLen = 21 // Length of truncated summary including ellipsis
	UUID4Length         = 36 // Length of a UUIDv4 string with hyphens
	HashLength          = 32 // Length of the hexadecimal hash
	Scheme              = "teams"
	Path                = "webhookb2"
	ProviderName        = "IncomingWebhook"
	WebhookDomainSuffix = ".webhook.office.com"
	ExpectedRegexGroups = 7 // Number of regex groups (1 match + 6 captures)
)

// Service providing teams as a notification service.
type Service struct {
	standard.Standard
	Config *Config
	pkr    format.PropKeyResolver
}

// Send a notification message to Microsoft Teams.
func (service *Service) Send(message string, params *types.Params) error {
	config := service.Config

	if err := service.pkr.UpdateConfigFromParams(config, params); err != nil {
		service.Logf("Failed to update params: %v", err)
	}

	return service.doSend(config, message)
}

// Initialize loads ServiceConfig from configURL and sets logger for this Service.
func (service *Service) Initialize(configURL *url.URL, logger types.StdLogger) error {
	service.Logger.SetLogger(logger)
	service.Config = &Config{}

	service.pkr = format.NewPropKeyResolver(service.Config)

	return service.Config.setURL(&service.pkr, configURL)
}

// GetID returns the service identifier.
func (service *Service) GetID() string {
	return Scheme
}

// GetConfigURLFromCustom creates a regular service URL from one with a custom host.
func (*Service) GetConfigURLFromCustom(customURL *url.URL) (serviceURL *url.URL, err error) {
	config, err := ConfigFromWebhookURL(*customURL)
	if err != nil {
		return nil, err
	}

	resolver := format.NewPropKeyResolver(config)
	for key, vals := range customURL.Query() {
		if err := resolver.Set(key, vals[0]); err != nil {
			return nil, err
		}
	}

	return config.getURL(&resolver), nil
}

func (service *Service) doSend(config *Config, message string) error {
	lines := strings.Split(message, "\n")
	sections := make([]section, 0, len(lines)) // Pre-allocate capacity based on number of lines

	for _, line := range lines {
		sections = append(sections, section{
			Text: line,
		})
	}

	// Teams need a summary for the webhook, use title or first (truncated) row
	summary := config.Title
	if summary == "" && len(sections) > 0 {
		summary = sections[0].Text
		if len(summary) > MaxSummaryLength {
			summary = summary[:TruncatedSummaryLen]
		}
	}

	payload, err := json.Marshal(payload{
		CardType:   "MessageCard",
		Context:    "http://schema.org/extensions",
		Markdown:   true,
		Title:      config.Title,
		ThemeColor: config.Color,
		Summary:    summary,
		Sections:   sections,
	})
	if err != nil {
		return err
	}

	if config.Host == "" {
		return fmt.Errorf("host is required but not specified in the configuration")
	}

	postURL := buildWebhookURL(config.Host, config.Group, config.Tenant, config.AltID, config.GroupOwner, config.ExtraID)

	res, err := http.Post(postURL, "application/json", bytes.NewBuffer(payload))
	if err == nil && res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send notification to teams, response status code %s", res.Status)
	}

	if err != nil {
		return fmt.Errorf("an error occurred while sending notification to teams: %s", err.Error())
	}

	return nil
}

// buildWebhookURL creates a webhook URL from the configuration.
func buildWebhookURL(host, group, tenant, altID, groupOwner, extraID string) string {
	return fmt.Sprintf(
		"https://%s/%s/%s@%s/%s/%s/%s/%s",
		host,
		Path,
		group,
		tenant,
		ProviderName,
		altID,
		groupOwner,
		extraID)
}

// parseAndVerifyWebhookURL parses and verifies a webhook URL.
func parseAndVerifyWebhookURL(webhookURL string) (parts [5]string, err error) {
	// Support the format with organization.webhook.office.com and required extraID
	pattern, err := regexp.Compile(`https://([^.]+)` + WebhookDomainSuffix + `/webhookb2/([0-9a-f-]{36})@([0-9a-f-]{36})/IncomingWebhook/([0-9a-f]{32})/([0-9a-f-]{36})/([^/]+)`)
	if err != nil {
		return parts, err
	}

	groups := pattern.FindStringSubmatch(webhookURL)
	if len(groups) != ExpectedRegexGroups {
		return parts, fmt.Errorf("invalid webhook URL format: expected %d regex groups, got %d", ExpectedRegexGroups, len(groups))
	}

	// Extract the parts: [group, tenant, altID, groupOwner, extraID]
	return [5]string{groups[2], groups[3], groups[4], groups[5], groups[6]}, nil
}

// verifyWebhookParts checks if the webhook parts are valid
func verifyWebhookParts(parts [5]string) error {
	if len(parts[0]) != UUID4Length {
		return fmt.Errorf("group ID should be %d characters, got %d", UUID4Length, len(parts[0]))
	}
	if len(parts[1]) != UUID4Length {
		return fmt.Errorf("tenant ID should be %d characters, got %d", UUID4Length, len(parts[1]))
	}
	if len(parts[2]) != HashLength {
		return fmt.Errorf("altID should be %d characters, got %d", HashLength, len(parts[2]))
	}
	if len(parts[3]) != UUID4Length {
		return fmt.Errorf("groupOwner should be %d characters, got %d", UUID4Length, len(parts[3]))
	}
	// We don't validate ExtraID length as it might vary

	return nil
}

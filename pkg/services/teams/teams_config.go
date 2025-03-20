package teams

import (
	"fmt"
	"github.com/containrrr/shoutrrr/pkg/format"
	"github.com/containrrr/shoutrrr/pkg/types"
	"net/url"
	"regexp"
	"strings"

	"github.com/containrrr/shoutrrr/pkg/services/standard"
)

// Config for use within the teams plugin
type Config struct {
	standard.EnumlessConfig
	Group      string `url:"user" optional:""`
	Tenant     string `url:"host" optional:""`
	AltID      string `url:"path1" optional:""`
	GroupOwner string `url:"path2" optional:""`
	ExtraID    string `url:"path3" optional:""`

	Title string `key:"title" optional:""`
	Color string `key:"color" optional:""`
	Host  string `key:"host" optional:""`
}

func (config *Config) webhookParts() [5]string {
	return [5]string{config.Group, config.Tenant, config.AltID, config.GroupOwner, config.ExtraID}
}

// SetFromWebhookURL updates the config WebhookParts from a teams webhook URL
func (config *Config) SetFromWebhookURL(webhookURL string) error {
	// Extract the organization domain
	orgPattern, err := regexp.Compile(`https://([^.]+)` + WebhookDomainSuffix + `/`)
	if err == nil {
		orgGroups := orgPattern.FindStringSubmatch(webhookURL)
		if len(orgGroups) == 2 {
			// Set the organization domain as the host
			config.Host = orgGroups[1] + WebhookDomainSuffix
		} else {
			return fmt.Errorf("invalid webhook URL format - must contain organization domain")
		}
	}

	parts, err := parseAndVerifyWebhookURL(webhookURL)
	if err != nil {
		return err
	}

	config.setFromWebhookParts(parts)
	return nil
}

// ConfigFromWebhookURL creates a new Config from a parsed Teams Webhook URL
func ConfigFromWebhookURL(webhookURL url.URL) (*Config, error) {
	config := &Config{
		Host: webhookURL.Host,
	}

	if err := config.SetFromWebhookURL(webhookURL.String()); err != nil {
		return nil, err
	}

	return config, nil
}

// GetURL returns a URL representation of it's current field values
func (config *Config) GetURL() *url.URL {
	resolver := format.NewPropKeyResolver(config)
	return config.getURL(&resolver)
}

// SetURL updates a ServiceConfig from a URL representation of it's field values
func (config *Config) SetURL(url *url.URL) error {
	resolver := format.NewPropKeyResolver(config)
	return config.setURL(&resolver, url)
}

func (config *Config) getURL(resolver types.ConfigQueryResolver) *url.URL {
	path := "/" + config.AltID + "/" + config.GroupOwner + "/" + config.ExtraID
	
	return &url.URL{
		User:       url.User(config.Group),
		Host:       config.Tenant,
		Path:       path,
		Scheme:     Scheme,
		ForceQuery: false,
		RawQuery:   format.BuildQuery(resolver),
	}
}

func (config *Config) setURL(resolver types.ConfigQueryResolver, url *url.URL) error {
	var webhookParts [5]string

	parts := strings.Split(url.Path, "/")
	if parts[0] == "" {
		parts = parts[1:]
	}
	
	if len(parts) < 3 {
		return fmt.Errorf("invalid URL format: missing extraId component")
	}
	
	extraID := parts[2]
	
	webhookParts = [5]string{url.User.Username(), url.Hostname(), parts[0], parts[1], extraID}

	if err := verifyWebhookParts(webhookParts); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	config.setFromWebhookParts(webhookParts)

	// Set the organization domain as the host if provided in query parameters
	hostFound := false
	for key, vals := range url.Query() {
		if key == "host" && vals[0] != "" {
			config.Host = vals[0]
			hostFound = true
		}
		if err := resolver.Set(key, vals[0]); err != nil {
			return err
		}
	}
	
	// Require host parameter
	if !hostFound {
		return fmt.Errorf("missing required host parameter (organization.webhook.office.com)")
	}

	return nil
}

func (config *Config) setFromWebhookParts(parts [5]string) {
	config.Group = parts[0]
	config.Tenant = parts[1]
	config.AltID = parts[2]
	config.GroupOwner = parts[3]
	config.ExtraID = parts[4]
}

func buildWebhookURL(host, group, tenant, altID, groupOwner, extraID string) string {
	// config.Group, config.Tenant, config.AltID, config.GroupOwner, config.ExtraID
	
	// Build URL with required extraID
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

func parseAndVerifyWebhookURL(webhookURL string) (parts [5]string, err error) {
	// Only support the new format with organization.webhook.office.com and required extraID
	pattern, err := regexp.Compile(`https://([^.]+)` + WebhookDomainSuffix + `/webhookb2/([0-9a-f-]{36})@([0-9a-f-]{36})/IncomingWebhook/([0-9a-f]{32})/([0-9a-f-]{36})/([^/]+)`)
	if err != nil {
		return parts, err
	}

	groups := pattern.FindStringSubmatch(webhookURL)
	if len(groups) < 7 {
		return parts, fmt.Errorf("invalid webhook URL format - must include the extra component at the end")
	}

	// Extract the parts: [group, tenant, altID, groupOwner, extraID]
	return [5]string{groups[2], groups[3], groups[4], groups[5], groups[6]}, nil
}

const (
	// Scheme is the identifying part of this service's configuration URL
	Scheme = "teams"
	// Path is the initial path of the webhook URL for domain-scoped webhook requests
	Path = "webhookb2"
	// ProviderName is the name of the Teams integration provider
	ProviderName = "IncomingWebhook"
	// WebhookDomainSuffix is the suffix for organization-specific webhook domains
	WebhookDomainSuffix = ".webhook.office.com"
)

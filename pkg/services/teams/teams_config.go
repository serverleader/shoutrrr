package teams

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/services/standard"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// Constants for URL parsing.
const (
	ExpectedUsernameParts = 2 // Number of parts expected in username split for legacy format
	ExpectedRegexGroups   = 5 // Number of regex groups (1 match + 4 captures)
)

// Config for use within the teams plugin.
type Config struct {
	standard.EnumlessConfig
	Group      string `optional:"" url:"user"`
	Tenant     string `optional:"" url:"host"`
	AltID      string `optional:"" url:"path1"`
	GroupOwner string `optional:"" url:"path2"`

	Title string `key:"title"                  optional:""`
	Color string `key:"color"                  optional:""`
	Host  string `default:"outlook.office.com" key:"host"  optional:""`
}

func (config *Config) webhookParts() [4]string {
	return [4]string{config.Group, config.Tenant, config.AltID, config.GroupOwner}
}

// SetFromWebhookURL updates the config WebhookParts from a teams webhook URL.
func (config *Config) SetFromWebhookURL(webhookURL string) error {
	parts, err := parseAndVerifyWebhookURL(webhookURL)
	if err != nil {
		return err
	}

	config.setFromWebhookParts(parts)

	return nil
}

// ConfigFromWebhookURL creates a new Config from a parsed Teams Webhook URL.
func ConfigFromWebhookURL(webhookURL url.URL) (*Config, error) {
	config := &Config{
		Host: webhookURL.Host,
	}

	if err := config.SetFromWebhookURL(webhookURL.String()); err != nil {
		return nil, err
	}

	return config, nil
}

// GetURL returns a URL representation of its current field values.
func (config *Config) GetURL() *url.URL {
	resolver := format.NewPropKeyResolver(config)

	return config.getURL(&resolver)
}

// SetURL updates a ServiceConfig from a URL representation of its field values.
func (config *Config) SetURL(url *url.URL) error {
	resolver := format.NewPropKeyResolver(config)

	return config.setURL(&resolver, url)
}

func (config *Config) getURL(resolver types.ConfigQueryResolver) *url.URL {
	return &url.URL{
		User:       url.User(config.Group),
		Host:       config.Tenant,
		Path:       "/" + config.AltID + "/" + config.GroupOwner,
		Scheme:     Scheme,
		ForceQuery: false,
		RawQuery:   format.BuildQuery(resolver),
	}
}

func (config *Config) setURL(resolver types.ConfigQueryResolver, url *url.URL) error {
	var webhookParts [4]string

	if pass, legacyFormat := url.User.Password(); legacyFormat {
		parts := strings.Split(url.User.Username(), "@")
		if len(parts) != ExpectedUsernameParts {
			return fmt.Errorf("invalid URL format: expected %d parts in username, got %d", ExpectedUsernameParts, len(parts))
		}

		webhookParts = [4]string{parts[0], parts[1], pass, url.Hostname()}
	} else {
		parts := strings.Split(url.Path, "/")
		if parts[0] == "" {
			parts = parts[1:]
		}

		webhookParts = [4]string{url.User.Username(), url.Hostname(), parts[0], parts[1]}
	}

	if err := verifyWebhookParts(webhookParts); err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	config.setFromWebhookParts(webhookParts)

	for key, vals := range url.Query() {
		if err := resolver.Set(key, vals[0]); err != nil {
			return err
		}
	}

	return nil
}

func (config *Config) setFromWebhookParts(parts [4]string) {
	config.Group = parts[0]
	config.Tenant = parts[1]
	config.AltID = parts[2]
	config.GroupOwner = parts[3]
}

func buildWebhookURL(host, group, tenant, altID, groupOwner string) string {
	path := Path
	if host == LegacyHost {
		path = LegacyPath
	}

	return fmt.Sprintf(
		"https://%s/%s/%s@%s/%s/%s/%s",
		host,
		path,
		group,
		tenant,
		ProviderName,
		altID,
		groupOwner)
}

func parseAndVerifyWebhookURL(webhookURL string) (parts [4]string, err error) {
	pattern, err := regexp.Compile(fmt.Sprintf(`([0-9a-f-]{%d})@([0-9a-f-]{%d})/[^/]+/([0-9a-f]{%d})/([0-9a-f-]{%d})`, UUID4Length, UUID4Length, HashLength, UUID4Length))
	if err != nil {
		return parts, err
	}

	groups := pattern.FindStringSubmatch(webhookURL)
	if len(groups) != ExpectedRegexGroups {
		return parts, fmt.Errorf("invalid webhook URL format: expected %d regex groups, got %d", ExpectedRegexGroups, len(groups))
	}

	copy(parts[:], groups[1:])

	return parts, nil
}

const (
	Scheme       = "teams"
	LegacyHost   = "outlook.office.com"
	LegacyPath   = "webhook"
	Path         = "webhookb2"
	ProviderName = "IncomingWebhook"
)

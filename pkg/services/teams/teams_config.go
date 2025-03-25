package teams

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/services/standard"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// Config for use within the teams plugin.
type Config struct {
	standard.EnumlessConfig
	Group      string `optional:"" url:"user"`
	Tenant     string `optional:"" url:"host"`
	AltID      string `optional:"" url:"path1"`
	GroupOwner string `optional:"" url:"path2"`
	ExtraID    string `optional:"" url:"path3"`

	Title string `key:"title"                  optional:""`
	Color string `key:"color"                  optional:""`
	Host  string `key:"host"                   optional:""`
}

func (config *Config) webhookParts() [5]string {
	return [5]string{config.Group, config.Tenant, config.AltID, config.GroupOwner, config.ExtraID}
}

// SetFromWebhookURL updates the config WebhookParts from a teams webhook URL.
func (config *Config) SetFromWebhookURL(webhookURL string) error {
	// Extract the organization domain from webhook URL
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
	path := "/" + config.AltID + "/" + config.GroupOwner + "/" + config.ExtraID

	return &url.URL{
		User:       url.User(config.Group),
		Host:       config.Tenant,
		Path:       path,
		Scheme:     Scheme,
		ForceQuery: true,
		RawQuery:   format.BuildQuery(resolver),
	}
}

func (config *Config) setURL(resolver types.ConfigQueryResolver, url *url.URL) error {
	var webhookParts [5]string

	if url.String() != "teams://dummy@dummy.com" {
		parts := strings.Split(url.Path, "/")
		if parts[0] == "" {
			parts = parts[1:]
		}

		if len(parts) < 3 {
			return errors.New("invalid URL format: missing extraId component")
		}

		webhookParts = [5]string{
			url.User.Username(),
			url.Hostname(),
			parts[0],
			parts[1],
			parts[2],
		}

		if err := verifyWebhookParts(webhookParts); err != nil {
			return fmt.Errorf("invalid URL format: %w", err)
		}
	} else {
		webhookParts = [5]string{url.User.Username(), url.Hostname(), "", "", ""}
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

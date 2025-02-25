package mattermost

import (
	"errors"
	"net/url"
	"strings"

	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/services/standard"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// Config object holding all information.
type Config struct {
	standard.EnumlessConfig
	UserName string `desc:"Override webhook user"    optional:""                                                              url:"user"`
	Icon     string `default:""                      desc:"Use emoji or URL as icon (based on presence of http(s):// prefix)" key:"icon,icon_emoji,icon_url" optional:""`
	Title    string `default:""                      desc:"Notification title, optionally set by the sender (not used)"       key:"title"`
	Channel  string `desc:"Override webhook channel" optional:""                                                              url:"path2"`
	Host     string `desc:"Mattermost server host"   url:"host,port"`
	Token    string `desc:"Webhook token"            url:"path1"`
}

// GetURL returns a URL representation of it's current field values.
func (config *Config) GetURL() *url.URL {
	paths := []string{"", config.Token, config.Channel}
	if config.Channel == "" {
		paths = paths[:2]
	}

	var user *url.Userinfo
	if config.UserName != "" {
		user = url.User(config.UserName)
	}

	resolver := format.NewPropKeyResolver(config)

	return &url.URL{
		User:       user,
		Host:       config.Host,
		Path:       strings.Join(paths, "/"),
		Scheme:     Scheme,
		ForceQuery: false,
		RawQuery:   format.BuildQuery(&resolver),
	}
}

// SetURL updates a ServiceConfig from a URL representation of it's field values.
func (config *Config) SetURL(url *url.URL) error {
	resolver := format.NewPropKeyResolver(config)

	return config.setURL(&resolver, url)
}

func (config *Config) setURL(resolver types.ConfigQueryResolver, serviceURL *url.URL) error {
	config.Host = serviceURL.Hostname()

	if serviceURL.Path == "" || serviceURL.Path == "/" {
		return errors.New(string(NotEnoughArguments))
	}

	config.UserName = serviceURL.User.Username()
	path := strings.Split(serviceURL.Path[1:], "/")

	for key, vals := range serviceURL.Query() {
		if err := resolver.Set(key, vals[0]); err != nil {
			return err
		}
	}

	if len(path) < 1 {
		return errors.New(string(NotEnoughArguments))
	}

	config.Token = path[0]

	if len(path) > 1 {
		if path[1] != "" {
			config.Channel = path[1]
		}
	}

	return nil
}

// ErrorMessage for error events within the mattermost service.
type ErrorMessage string

const (
	// Scheme is the identifying part of this service's configuration URL.
	Scheme = "mattermost"
	// NotEnoughArguments provided in the service URL.
	NotEnoughArguments ErrorMessage = "the apiURL does not include enough arguments, either provide 1 or 3 arguments (they may be empty)"
)

// CreateConfigFromURL to use within the mattermost service.
func CreateConfigFromURL(serviceURL *url.URL) (*Config, error) {
	config := Config{}
	err := config.SetURL(serviceURL)

	return &config, err
}

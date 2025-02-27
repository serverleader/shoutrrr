package mattermost

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/services/standard"

	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// Service sends notifications to a pre-configured channel or user.
type Service struct {
	standard.Standard
	Config *Config
	pkr    format.PropKeyResolver
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

// Send a notification message to Mattermost.
func (service *Service) Send(message string, params *types.Params) error {
	config := service.Config
	apiURL := buildURL(config)

	if err := service.pkr.UpdateConfigFromParams(config, params); err != nil {
		return err
	}

	json, _ := CreateJSONPayload(config, message, params)

	res, err := http.Post(apiURL, "application/json", bytes.NewReader(json))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send notification to service, response status code %s", res.Status)
	}

	return err
}

// Builds the actual URL the request should go to.
func buildURL(config *Config) string {
	return fmt.Sprintf("https://%s/hooks/%s", config.Host, config.Token)
}

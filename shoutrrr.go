package shoutrrr

import (
	"github.com/serverleader/shoutrrr/internal/meta"
	"github.com/serverleader/shoutrrr/pkg/router"
	"github.com/serverleader/shoutrrr/pkg/types"
)

// Version of the library
var Version = meta.Version

// SendAsync sends notifications to all services in parallel to all the specified URLs.
// It returns a channel that the errors for each service will be sent on.
func SendAsync(message string, urls ...string) chan error {
	r, _ := router.New(nil, urls...)
	return r.SendAsync(message, nil)
}

// Send sends notifications to all services specified in the URLs.
// It returns the first error that occurs or nil if no errors occurred.
func Send(message string, urls ...string) error {
	errs := SendAsync(message, urls...)
	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// SetLogger creates a router with the specified logger
func SetLogger(logger types.StdLogger) *router.ServiceRouter {
	r, _ := router.New(logger)
	return r
}

// CreateSender returns a notification sender configured according to the supplied URL
func CreateSender(rawURLs ...string) (*router.ServiceRouter, error) {
	return router.New(nil, rawURLs...)
}

// NewSender returns a notification sender, writing any log output to logger and configured
// to send to the services indicated by the supplied URLs
func NewSender(logger types.StdLogger, serviceURLs ...string) (*router.ServiceRouter, error) {
	return router.New(logger, serviceURLs...)
}

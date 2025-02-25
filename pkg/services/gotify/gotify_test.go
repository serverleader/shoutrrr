package gotify

import (
	"log"
	"net/url"
	"testing"

	"github.com/nicholas-fedor/shoutrrr/internal/testutils"

	"github.com/jarcoal/httpmock"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestGotify(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Shoutrrr Gotify Suite")
}

var logger *log.Logger

var _ = ginkgo.Describe("the Gotify plugin URL building and token validation functions", func() {
	ginkgo.It("should build a valid gotify URL", func() {
		config := Config{
			Token: "Aaa.bbb.ccc.ddd",
			Host:  "my.gotify.tld",
		}
		url, err := buildURL(&config)
		gomega.Expect(err).To(gomega.BeNil())
		expectedURL := "https://my.gotify.tld/message?token=Aaa.bbb.ccc.ddd"
		gomega.Expect(url).To(gomega.Equal(expectedURL))
	})

	ginkgo.When("TLS is disabled", func() {
		ginkgo.It("should use http schema", func() {
			config := Config{
				Token:      "Aaa.bbb.ccc.ddd",
				Host:       "my.gotify.tld",
				DisableTLS: true,
			}
			url, err := buildURL(&config)
			gomega.Expect(err).To(gomega.BeNil())
			expectedURL := "http://my.gotify.tld/message?token=Aaa.bbb.ccc.ddd"
			gomega.Expect(url).To(gomega.Equal(expectedURL))
		})
	})

	ginkgo.When("a custom path is provided", func() {
		ginkgo.It("should add it to the URL", func() {
			config := Config{
				Token: "Aaa.bbb.ccc.ddd",
				Host:  "my.gotify.tld",
				Path:  "/gotify",
			}
			url, err := buildURL(&config)
			gomega.Expect(err).To(gomega.BeNil())
			expectedURL := "https://my.gotify.tld/gotify/message?token=Aaa.bbb.ccc.ddd"
			gomega.Expect(url).To(gomega.Equal(expectedURL))
		})
	})

	ginkgo.When("provided a valid token", func() {
		ginkgo.It("should return true", func() {
			token := "Ahwbsdyhwwgarxd"
			gomega.Expect(isTokenValid(token)).To(gomega.BeTrue())
		})
	})
	ginkgo.When("provided a token with an invalid prefix", func() {
		ginkgo.It("should return false", func() {
			token := "Chwbsdyhwwgarxd"
			gomega.Expect(isTokenValid(token)).To(gomega.BeFalse())
		})
	})
	ginkgo.When("provided a token with an invalid length", func() {
		ginkgo.It("should return false", func() {
			token := "Chwbsdyhwwga"
			gomega.Expect(isTokenValid(token)).To(gomega.BeFalse())
		})
	})
	ginkgo.Describe("creating the API URL", func() {
		ginkgo.When("the token is invalid", func() {
			ginkgo.It("should return an error", func() {
				config := Config{
					Token: "invalid",
				}
				_, err := buildURL(&config)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
	})
	ginkgo.Describe("creating a config", func() {
		ginkgo.When("parsing the configuration URL", func() {
			ginkgo.It("should be identical after de-/serialization (with path)", func() {
				testURL := "gotify://my.gotify.tld/gotify/Aaa.bbb.ccc.ddd?title=Test+title"

				config := &Config{}
				gomega.Expect(config.SetURL(testutils.URLMust(testURL))).To(gomega.Succeed())
				gomega.Expect(config.GetURL().String()).To(gomega.Equal(testURL))
			})
			ginkgo.It("should be identical after de-/serialization (without path)", func() {
				testURL := "gotify://my.gotify.tld/Aaa.bbb.ccc.ddd?disabletls=Yes&priority=1&title=Test+title"

				config := &Config{}
				gomega.Expect(config.SetURL(testutils.URLMust(testURL))).To(gomega.Succeed())
				gomega.Expect(config.GetURL().String()).To(gomega.Equal(testURL))
			})
			ginkgo.It("should allow slash at the end of the token", func() {
				url := testutils.URLMust("gotify://my.gotify.tld/Aaa.bbb.ccc.ddd/")

				config := &Config{}
				gomega.Expect(config.SetURL(url)).To(gomega.Succeed())
				gomega.Expect(config.Token).To(gomega.Equal("Aaa.bbb.ccc.ddd"))
			})
			ginkgo.It("should allow slash at the end of the token, with additional path", func() {
				url := testutils.URLMust("gotify://my.gotify.tld/path/to/gotify/Aaa.bbb.ccc.ddd/")

				config := &Config{}
				gomega.Expect(config.SetURL(url)).To(gomega.Succeed())
				gomega.Expect(config.Token).To(gomega.Equal("Aaa.bbb.ccc.ddd"))
			})
			ginkgo.It("should not crash on empty token or path slash at the end of the token", func() {
				config := &Config{}
				gomega.Expect(config.SetURL(testutils.URLMust("gotify://my.gotify.tld//"))).To(gomega.Succeed())
				gomega.Expect(config.SetURL(testutils.URLMust("gotify://my.gotify.tld/"))).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Describe("sending the payload", func() {
		var err error
		var service Service
		ginkgo.AfterEach(func() {
			httpmock.DeactivateAndReset()
		})
		ginkgo.It("should not report an error if the server accepts the payload", func() {
			serviceURL, _ := url.Parse("gotify://my.gotify.tld/Aaa.bbb.ccc.ddd")
			err = service.Initialize(serviceURL, logger)
			httpmock.ActivateNonDefault(service.GetHTTPClient())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			targetURL := "https://my.gotify.tld/message?token=Aaa.bbb.ccc.ddd"
			httpmock.RegisterResponder("POST", targetURL, testutils.JSONRespondMust(200, messageResponse{}))

			err = service.Send("Message", nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
		ginkgo.It("should not panic if an error occurs when sending the payload", func() {
			serviceURL, _ := url.Parse("gotify://my.gotify.tld/Aaa.bbb.ccc.ddd")
			err = service.Initialize(serviceURL, logger)
			httpmock.ActivateNonDefault(service.GetHTTPClient())
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			targetURL := "https://my.gotify.tld/message?token=Aaa.bbb.ccc.ddd"
			httpmock.RegisterResponder("POST", targetURL, testutils.JSONRespondMust(401, errorResponse{
				Name:        "Unauthorized",
				Code:        401,
				Description: "you need to provide a valid access token or user credentials to access this api",
			}))

			err = service.Send("Message", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})

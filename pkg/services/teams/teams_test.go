package teams

import (
	"errors"
	"log"
	"net/url"
	"testing"

	"github.com/jarcoal/httpmock"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const (
	legacyWebhookURL = "https://outlook.office.com/webhook/11111111-4444-4444-8444-cccccccccccc@22222222-4444-4444-8444-cccccccccccc/IncomingWebhook/33333333012222222222333333333344/44444444-4444-4444-8444-cccccccccccc"
	scopedWebhookURL = "https://test.webhook.office.com/webhookb2/11111111-4444-4444-8444-cccccccccccc@22222222-4444-4444-8444-cccccccccccc/IncomingWebhook/33333333012222222222333333333344/44444444-4444-4444-8444-cccccccccccc"
	scopedDomainHost = "test.webhook.office.com"
	testURLBase      = "teams://11111111-4444-4444-8444-cccccccccccc@22222222-4444-4444-8444-cccccccccccc/33333333012222222222333333333344/44444444-4444-4444-8444-cccccccccccc"
	scopedURLBase    = testURLBase + `?host=` + scopedDomainHost
)

var logger = log.New(ginkgo.GinkgoWriter, "Test", log.LstdFlags)

func TestTeams(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Shoutrrr Teams Suite")
}

var _ = ginkgo.Describe("the teams service", func() {
	ginkgo.When("creating the webhook URL", func() {
		ginkgo.It("should match the expected output for legacy URLs", func() {
			config := Config{}
			config.setFromWebhookParts([4]string{
				"11111111-4444-4444-8444-cccccccccccc",
				"22222222-4444-4444-8444-cccccccccccc",
				"33333333012222222222333333333344",
				"44444444-4444-4444-8444-cccccccccccc",
			})
			apiURL := buildWebhookURL(LegacyHost, config.Group, config.Tenant, config.AltID, config.GroupOwner)
			gomega.Expect(apiURL).To(gomega.Equal(legacyWebhookURL))

			parts, err := parseAndVerifyWebhookURL(apiURL)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(parts).To(gomega.Equal(config.webhookParts()))
		})
		ginkgo.It("should match the expected output for custom URLs", func() {
			config := Config{}
			config.setFromWebhookParts([4]string{
				"11111111-4444-4444-8444-cccccccccccc",
				"22222222-4444-4444-8444-cccccccccccc",
				"33333333012222222222333333333344",
				"44444444-4444-4444-8444-cccccccccccc",
			})
			apiURL := buildWebhookURL(scopedDomainHost, config.Group, config.Tenant, config.AltID, config.GroupOwner)
			gomega.Expect(apiURL).To(gomega.Equal(scopedWebhookURL))

			parts, err := parseAndVerifyWebhookURL(apiURL)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(parts).To(gomega.Equal(config.webhookParts()))
		})
	})

	ginkgo.Describe("creating a config", func() {
		ginkgo.When("parsing the configuration URL", func() {
			ginkgo.It("should be identical after de-/serialization", func() {
				testURL := testURLBase + "?color=aabbcc&host=test.outlook.office.com&title=Test+title"

				url, err := url.Parse(testURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "parsing")

				config := &Config{Host: LegacyHost}
				err = config.SetURL(url)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "verifying")

				outputURL := config.GetURL()
				gomega.Expect(outputURL.String()).To(gomega.Equal(testURL))
			})
		})
	})

	ginkgo.Describe("converting custom URL to service URL", func() {
		ginkgo.When("an invalid custom URL is provided", func() {
			ginkgo.It("should return an error", func() {
				service := Service{}
				testURL := "teams+https://google.com/search?q=what+is+love"

				customURL, err := url.Parse(testURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "parsing")

				_, err = service.GetConfigURLFromCustom(customURL)
				gomega.Expect(err).To(gomega.HaveOccurred(), "converting")
			})
		})
		ginkgo.When("a valid custom URL is provided", func() {
			ginkgo.It("should set the host field from the custom URL", func() {
				service := Service{}
				testURL := `teams+` + scopedWebhookURL

				customURL, err := url.Parse(testURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "parsing")

				serviceURL, err := service.GetConfigURLFromCustom(customURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "converting")

				gomega.Expect(serviceURL.String()).To(gomega.Equal(scopedURLBase))
			})
			ginkgo.It("should preserve the query params in the generated service URL", func() {
				service := Service{}
				testURL := "teams+" + legacyWebhookURL + "?color=f008c1&title=TheTitle"

				customURL, err := url.Parse(testURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "parsing")

				serviceURL, err := service.GetConfigURLFromCustom(customURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "converting")

				gomega.Expect(serviceURL.String()).To(gomega.Equal(testURLBase + "?color=f008c1&title=TheTitle"))
			})
		})
	})

	ginkgo.Describe("sending the payload", func() {
		var err error
		var service Service
		ginkgo.BeforeEach(func() {
			httpmock.Activate()
		})
		ginkgo.AfterEach(func() {
			httpmock.DeactivateAndReset()
		})
		ginkgo.It("should not report an error if the server accepts the payload", func() {
			serviceURL, _ := url.Parse(scopedURLBase)
			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			httpmock.RegisterResponder("POST", scopedWebhookURL, httpmock.NewStringResponder(200, ""))

			err = service.Send("Message", nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
		ginkgo.It("should not panic if an error occurs when sending the payload", func() {
			serviceURL, _ := url.Parse(testURLBase)
			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			httpmock.RegisterResponder("POST", legacyWebhookURL, httpmock.NewErrorResponder(errors.New("dummy error")))

			err = service.Send("Message", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})

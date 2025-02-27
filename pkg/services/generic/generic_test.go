package generic

import (
	"errors"
	"io"
	"log"
	"net/url"
	"testing"

	"github.com/nicholas-fedor/shoutrrr/internal/testutils"

	"github.com/jarcoal/httpmock"
	"github.com/nicholas-fedor/shoutrrr/pkg/format"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
)

// Test constants.
const (
	TestWebhookURL = "https://host.tld/webhook" // Default test webhook URL
)

func TestGeneric(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Shoutrrr Generic Webhook Suite")
}

var (
	logger  = log.New(ginkgo.GinkgoWriter, "Test", log.LstdFlags)
	service *Service
)

var _ = ginkgo.Describe("the Generic service", func() {
	ginkgo.BeforeEach(func() {
		service = &Service{}
		service.SetLogger(logger)
	})
	ginkgo.When("parsing a custom URL", func() {
		ginkgo.It("should strip generic prefix before parsing", func() {
			customURL, err := url.Parse("generic+https://test.tld")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			actualURL, err := service.GetConfigURLFromCustom(customURL)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			_, expectedURL := testCustomURL("https://test.tld")
			gomega.Expect(actualURL.String()).To(gomega.Equal(expectedURL.String()))
		})

		ginkgo.When("a HTTP URL is provided", func() {
			ginkgo.It("should disable TLS", func() {
				config, _ := testCustomURL("http://example.com")
				gomega.Expect(config.DisableTLS).To(gomega.BeTrue())
			})
		})
		ginkgo.When("a HTTPS URL is provided", func() {
			ginkgo.It("should enable TLS", func() {
				config, _ := testCustomURL("https://example.com")
				gomega.Expect(config.DisableTLS).To(gomega.BeFalse())
			})
		})
		ginkgo.It("should escape conflicting custom query keys", func() {
			expectedURL := "generic://example.com/?__template=passed"
			config, srvURL := testCustomURL("https://example.com/?template=passed")
			gomega.Expect(config.Template).NotTo(gomega.Equal("passed")) // captured
			whURL := config.WebhookURL().String()
			gomega.Expect(whURL).To(gomega.Equal("https://example.com/?template=passed"))
			gomega.Expect(srvURL.String()).To(gomega.Equal(expectedURL))
		})
		ginkgo.It("should handle both escaped and service prop version of keys", func() {
			config, _ := testServiceURL("generic://example.com/?__template=passed&template=captured")
			gomega.Expect(config.Template).To(gomega.Equal("captured"))
			whURL := config.WebhookURL().String()
			gomega.Expect(whURL).To(gomega.Equal("https://example.com/?template=passed"))
		})

		ginkgo.When("the URL includes custom headers", func() {
			ginkgo.It("should strip the headers from the webhook query", func() {
				config, _ := testServiceURL("generic://example.com/?@authorization=frend")
				gomega.Expect(config.WebhookURL().Query()).NotTo(gomega.HaveKey("@authorization"))
				gomega.Expect(config.WebhookURL().Query()).NotTo(gomega.HaveKey("authorization"))
			})
			ginkgo.It("should add the headers to the config custom header map", func() {
				config, _ := testServiceURL("generic://example.com/?@authorization=frend")
				gomega.Expect(config.headers).To(gomega.HaveKeyWithValue("Authorization", "frend"))
			})
			ginkgo.When("header keys are in camelCase", func() {
				ginkgo.It("should add headers with kebab-case keys", func() {
					config, _ := testServiceURL("generic://example.com/?@userAgent=gozilla+1.0")
					gomega.Expect(config.headers).To(gomega.HaveKeyWithValue("User-Agent", "gozilla 1.0"))
				})
			})
		})

		ginkgo.When("the URL includes extra data", func() {
			ginkgo.It("should strip the extra data from the webhook query", func() {
				config, _ := testServiceURL("generic://example.com/?$context=inside+joke")
				gomega.Expect(config.WebhookURL().Query()).NotTo(gomega.HaveKey("$context"))
				gomega.Expect(config.WebhookURL().Query()).NotTo(gomega.HaveKey("context"))
			})
			ginkgo.It("should add the extra data to the config extra data map", func() {
				config, _ := testServiceURL("generic://example.com/?$context=inside+joke")
				gomega.Expect(config.extraData).To(gomega.HaveKeyWithValue("context", "inside joke"))
			})
		})
	})
	ginkgo.When("retrieving the webhook URL", func() {
		ginkgo.It("should build a valid webhook URL", func() {
			expectedURL := "https://example.com/path?foo=bar"
			config, _ := testServiceURL("generic://example.com/path?foo=bar")
			gomega.Expect(config.WebhookURL().String()).To(gomega.Equal(expectedURL))
		})

		ginkgo.When("TLS is disabled", func() {
			ginkgo.It("should use http schema", func() {
				config := Config{
					webhookURL: &url.URL{
						Host: "test.tld",
					},
					DisableTLS: true,
				}
				gomega.Expect(config.WebhookURL().Scheme).To(gomega.Equal("http"))
			})
		})
		ginkgo.When("TLS is not disabled", func() {
			ginkgo.It("should use https schema", func() {
				config := Config{
					webhookURL: &url.URL{
						Host: "test.tld",
					},
					DisableTLS: false,
				}
				gomega.Expect(config.WebhookURL().Scheme).To(gomega.Equal("https"))
			})
		})
	})

	ginkgo.Describe("creating a config", func() {
		ginkgo.When("creating a default config", func() {
			ginkgo.It("should not return an error", func() {
				config := &Config{}
				pkr := format.NewPropKeyResolver(config)
				err := pkr.SetDefaultProps(config)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})
		ginkgo.When("parsing the configuration URL", func() {
			ginkgo.It("should be identical after de-/serialization", func() {
				testURL := "generic://user:pass@host.tld/api/v1/webhook?%24context=inside-joke&%40Authorization=frend&__title=w&contenttype=a%2Fb&template=f&title=t"

				url, err := url.Parse(testURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "parsing")

				config := &Config{}
				pkr := format.NewPropKeyResolver(config)
				gomega.Expect(pkr.SetDefaultProps(config)).To(gomega.Succeed())
				err = config.SetURL(url)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "verifying")

				outputURL := config.GetURL()
				gomega.Expect(outputURL.String()).To(gomega.Equal(testURL))
			})
		})
	})

	ginkgo.Describe("building the payload", func() {
		var service Service
		var config Config
		ginkgo.BeforeEach(func() {
			service = Service{}
			config = Config{
				MessageKey: "message",
				TitleKey:   "title",
			}
		})
		ginkgo.When("no template is specified", func() {
			ginkgo.It("should use the message as payload", func() {
				payload, err := service.getPayload(&config, types.Params{"message": "test message"})
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				contents, err := io.ReadAll(payload)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(string(contents)).To(gomega.Equal("test message"))
			})
		})
		ginkgo.When("template is specified as `JSON`", func() {
			ginkgo.It("should create a JSON object as the payload", func() {
				config.Template = "JSON"
				params := types.Params{"title": "test title"}
				sendParams := createSendParams(&config, params, "test message")
				payload, err := service.getPayload(&config, sendParams)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				contents, err := io.ReadAll(payload)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(string(contents)).To(gomega.MatchJSON(`{
					"title":   "test title",
					"message": "test message"
				}`))
			})
			ginkgo.When("alternate keys are specified", func() {
				ginkgo.It("should create a JSON object using the specified keys", func() {
					config.Template = "JSON"
					config.MessageKey = "body"
					config.TitleKey = "header"
					params := types.Params{"title": "test title"}
					sendParams := createSendParams(&config, params, "test message")
					payload, err := service.getPayload(&config, sendParams)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					contents, err := io.ReadAll(payload)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					gomega.Expect(string(contents)).To(gomega.MatchJSON(`{
						"header":   "test title",
						"body": "test message"
					}`))
				})
			})
		})
		ginkgo.When("a valid template is specified", func() {
			ginkgo.It("should apply the template to the message payload", func() {
				err := service.SetTemplateString("news", `{{.title}} ==> {{.message}}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				params := types.Params{}
				params.SetTitle("BREAKING NEWS")
				params.SetMessage("it's today!")
				config.Template = "news"
				payload, err := service.getPayload(&config, params)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				contents, err := io.ReadAll(payload)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(string(contents)).To(gomega.Equal("BREAKING NEWS ==> it's today!"))
			})
			ginkgo.When("given nil params", func() {
				ginkgo.It("should apply template with message data", func() {
					err := service.SetTemplateString("arrows", `==> {{.message}} <==`)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					config.Template = "arrows"
					payload, err := service.getPayload(&config, types.Params{"message": "LOOK AT ME"})
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					contents, err := io.ReadAll(payload)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					gomega.Expect(string(contents)).To(gomega.Equal("==> LOOK AT ME <=="))
				})
			})
		})
		ginkgo.When("an unknown template is specified", func() {
			ginkgo.It("should return an error", func() {
				_, err := service.getPayload(&Config{Template: "missing"}, nil)
				gomega.Expect(err).To(gomega.HaveOccurred())
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
			serviceURL, _ := url.Parse("generic://host.tld/webhook")
			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			targetURL := TestWebhookURL
			httpmock.RegisterResponder("POST", targetURL, httpmock.NewStringResponder(200, ""))

			err = service.Send("Message", nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
		ginkgo.It("should not panic if an error occurs when sending the payload", func() {
			serviceURL, _ := url.Parse("generic://host.tld/webhook")
			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			targetURL := TestWebhookURL
			httpmock.RegisterResponder("POST", targetURL, httpmock.NewErrorResponder(errors.New("dummy error")))

			err = service.Send("Message", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should not return an error when an unknown param is encountered", func() {
			serviceURL, _ := url.Parse("generic://host.tld/webhook")
			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			targetURL := TestWebhookURL
			httpmock.RegisterResponder("POST", targetURL, httpmock.NewStringResponder(200, ""))

			err = service.Send("Message", &types.Params{"unknown": "param"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
		ginkgo.It("should use the configured HTTP method", func() {
			serviceURL, _ := url.Parse("generic://host.tld/webhook?method=GET")
			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			targetURL := TestWebhookURL
			httpmock.RegisterResponder("GET", targetURL, httpmock.NewStringResponder(200, ""))

			err = service.Send("Message", nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
		ginkgo.It("should not mutate the given params", func() {
			serviceURL, _ := url.Parse("generic://host.tld/webhook?method=GET")
			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			targetURL := TestWebhookURL
			httpmock.RegisterResponder("GET", targetURL, httpmock.NewStringResponder(200, ""))

			params := types.Params{"title": "TITLE"}

			err = service.Send("Message", &params)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(params).To(gomega.Equal(types.Params{"title": "TITLE"}))
		})
	})
	ginkgo.Describe("the service upstream client", func() {
		var server *ghttp.Server
		var serverHost string
		ginkgo.BeforeEach(func() {
			server = ghttp.NewServer()
			serverHost = testutils.URLMust(server.URL()).Host
		})
		ginkgo.AfterEach(func() {
			server.Close()
		})

		ginkgo.When("custom headers are configured", func() {
			ginkgo.It("should add those headers to the request", func() {
				serviceURL := testutils.URLMust("generic://host.tld/webhook?disabletls=yes&@authorization=frend&@userAgent=gozilla+1.0")
				serviceURL.Host = serverHost
				gomega.Expect(service.Initialize(serviceURL, logger)).NotTo(gomega.HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/webhook"),
						ghttp.VerifyHeaderKV("authorization", "frend"),
						ghttp.VerifyHeaderKV("user-agent", "gozilla 1.0"),
					),
				)

				gomega.Expect(service.Send("Message", nil)).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("extra data is configured", func() {
			ginkgo.When("json template is used", func() {
				ginkgo.It("should add those extra data fields to the request", func() {
					serviceURL := testutils.URLMust("generic://host.tld/webhook?disabletls=yes&template=json&$context=inside+joke")
					serviceURL.Host = serverHost
					gomega.Expect(service.Initialize(serviceURL, logger)).NotTo(gomega.HaveOccurred())

					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/webhook"),
							ghttp.VerifyJSONRepresenting(map[string]string{
								"message": "Message",
								"context": "inside joke",
							}),
						),
					)

					gomega.Expect(service.Send("Message", nil)).NotTo(gomega.HaveOccurred())
				})
			})
		})
	})
	ginkgo.Describe("the normalized header key format", func() {
		ginkgo.It("should match the format", func() {
			gomega.Expect(normalizedHeaderKey("content-type")).To(gomega.Equal("Content-Type"))
		})
		ginkgo.It("should match the format", func() {
			gomega.Expect(normalizedHeaderKey("contentType")).To(gomega.Equal("Content-Type"))
		})
		ginkgo.It("should match the format", func() {
			gomega.Expect(normalizedHeaderKey("ContentType")).To(gomega.Equal("Content-Type"))
		})
		ginkgo.It("should match the format", func() {
			gomega.Expect(normalizedHeaderKey("Content-Type")).To(gomega.Equal("Content-Type"))
		})
	})
})

func testCustomURL(testURL string) (*Config, *url.URL) {
	customURL, err := url.Parse(testURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	config, pkr, err := ConfigFromWebhookURL(*customURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return config, config.getURL(&pkr)
}

func testServiceURL(testURL string) (*Config, *url.URL) {
	serviceURL, err := url.Parse(testURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	config, pkr := DefaultConfig()
	err = config.setURL(&pkr, serviceURL)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return config, config.getURL(&pkr)
}

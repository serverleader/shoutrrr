package bark

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/nicholas-fedor/shoutrrr/internal/testutils"
	"github.com/nicholas-fedor/shoutrrr/pkg/format"

	"github.com/jarcoal/httpmock"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	gomegaformat "github.com/onsi/gomega/format"
)

func TestBark(t *testing.T) {
	gomegaformat.CharactersAroundMismatchToInclude = 20

	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Shoutrrr Bark Suite")
}

var (
	service    *Service = &Service{}
	envBarkURL *url.URL
	logger     *log.Logger = testutils.TestLogger()
	_                      = ginkgo.BeforeSuite(func() {
		envBarkURL, _ = url.Parse(os.Getenv("SHOUTRRR_BARK_URL"))
	})
)

var _ = ginkgo.Describe("the bark service", func() {
	ginkgo.When("running integration tests", func() {
		ginkgo.It("should not error out", func() {
			if envBarkURL.String() == "" {
				ginkgo.Skip("No integration test ENV URL was set")

				return
			}

			configURL := testutils.URLMust(envBarkURL.String())
			gomega.Expect(service.Initialize(configURL, logger)).To(gomega.Succeed())
			gomega.Expect(service.Send("This is an integration test message", nil)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("the config", func() {
		ginkgo.When("getting a API URL", func() {
			ginkgo.It("should return the expected URL", func() {
				gomega.Expect(getAPIForPath("path")).To(gomega.Equal("https://host/path/endpoint"))
				gomega.Expect(getAPIForPath("/path")).To(gomega.Equal("https://host/path/endpoint"))
				gomega.Expect(getAPIForPath("/path/")).To(gomega.Equal("https://host/path/endpoint"))
				gomega.Expect(getAPIForPath("path/")).To(gomega.Equal("https://host/path/endpoint"))
				gomega.Expect(getAPIForPath("/")).To(gomega.Equal("https://host/endpoint"))
				gomega.Expect(getAPIForPath("")).To(gomega.Equal("https://host/endpoint"))
			})
		})
		ginkgo.When("only required fields are set", func() {
			ginkgo.It("should set the optional fields to the defaults", func() {
				serviceURL := testutils.URLMust("bark://:devicekey@hostname")
				gomega.Expect(service.Initialize(serviceURL, logger)).To(gomega.Succeed())

				gomega.Expect(*service.config).To(gomega.Equal(Config{
					Host:      "hostname",
					DeviceKey: "devicekey",
					Scheme:    "https",
				}))
			})
		})
		ginkgo.When("parsing the configuration URL", func() {
			ginkgo.It("should be identical after de-/serialization", func() {
				testURL := "bark://:device-key@example.com:2225/?badge=5&category=CAT&group=GROUP&scheme=http&title=TITLE&url=URL"
				config := &Config{}
				pkr := format.NewPropKeyResolver(config)
				gomega.Expect(config.setURL(&pkr, testutils.URLMust(testURL))).To(gomega.Succeed(), "verifying")
				gomega.Expect(config.GetURL().String()).To(gomega.Equal(testURL))
			})
		})
	})

	ginkgo.When("sending the push payload", func() {
		ginkgo.BeforeEach(func() {
			httpmock.Activate()
		})
		ginkgo.AfterEach(func() {
			httpmock.DeactivateAndReset()
		})

		ginkgo.It("should not report an error if the server accepts the payload", func() {
			serviceURL := testutils.URLMust("bark://:devicekey@hostname")
			gomega.Expect(service.Initialize(serviceURL, logger)).To(gomega.Succeed())

			httpmock.RegisterResponder("POST", service.config.GetAPIURL("push"), testutils.JSONRespondMust(200, apiResponse{
				Code:    http.StatusOK,
				Message: "OK",
			}))

			gomega.Expect(service.Send("Message", nil)).To(gomega.Succeed())
		})
		ginkgo.It("should not panic if a server error occurs", func() {
			serviceURL := testutils.URLMust("bark://:devicekey@hostname")
			gomega.Expect(service.Initialize(serviceURL, logger)).To(gomega.Succeed())

			httpmock.RegisterResponder("POST", service.config.GetAPIURL("push"), testutils.JSONRespondMust(500, apiResponse{
				Code:    500,
				Message: "someone turned off the internet",
			}))

			gomega.Expect(service.Send("Message", nil)).To(gomega.HaveOccurred())
		})
		ginkgo.It("should not panic if a server responds with an unknown message", func() {
			serviceURL := testutils.URLMust("bark://:devicekey@hostname")
			gomega.Expect(service.Initialize(serviceURL, logger)).To(gomega.Succeed())

			httpmock.RegisterResponder("POST", service.config.GetAPIURL("push"), testutils.JSONRespondMust(200, apiResponse{
				Code:    500,
				Message: "For some reason, the response code and HTTP code is different?",
			}))

			gomega.Expect(service.Send("Message", nil)).To(gomega.HaveOccurred())
		})
		ginkgo.It("should not panic if a communication error occurs", func() {
			httpmock.DeactivateAndReset()
			serviceURL := testutils.URLMust("bark://:devicekey@nonresolvablehostname")
			gomega.Expect(service.Initialize(serviceURL, logger)).To(gomega.Succeed())
			gomega.Expect(service.Send("Message", nil)).To(gomega.HaveOccurred())
		})
	})

	ginkgo.Describe("the basic service API", func() {
		ginkgo.Describe("the service config", func() {
			ginkgo.It("should implement basic service config API methods correctly", func() {
				testutils.TestConfigGetInvalidQueryValue(&Config{})
				testutils.TestConfigSetInvalidQueryValue(&Config{}, "bark://:mock-device@host/?foo=bar")

				testutils.TestConfigSetDefaultValues(&Config{})

				testutils.TestConfigGetEnumsCount(&Config{}, 0)
				testutils.TestConfigGetFieldsCount(&Config{}, 9)
			})
		})
		ginkgo.Describe("the service instance", func() {
			ginkgo.BeforeEach(func() {
				httpmock.Activate()
			})
			ginkgo.AfterEach(func() {
				httpmock.DeactivateAndReset()
			})
			ginkgo.It("should implement basic service API methods correctly", func() {
				serviceURL := testutils.URLMust("bark://:devicekey@hostname")
				gomega.Expect(service.Initialize(serviceURL, logger)).To(gomega.Succeed())
				testutils.TestServiceSetInvalidParamValue(service, "foo", "bar")
			})
		})
	})
})

func getAPIForPath(path string) string {
	c := Config{Host: "host", Path: path, Scheme: "https"}

	return c.GetAPIURL("endpoint")
}

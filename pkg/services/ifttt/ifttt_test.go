package ifttt

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/jarcoal/httpmock"
	"github.com/nicholas-fedor/shoutrrr/internal/testutils"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

func TestIFTTT(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Shoutrrr IFTTT Suite")
}

var (
	service    *Service
	logger     *log.Logger
	envTestURL string
	_          = ginkgo.BeforeSuite(func() {
		envTestURL = os.Getenv("SHOUTRRR_IFTTT_URL")
		logger = testutils.TestLogger()
	})
)

var _ = ginkgo.Describe("the ifttt package", func() {
	ginkgo.BeforeEach(func() {
		service = &Service{}
	})
	ginkgo.When("running integration tests", func() {
		ginkgo.It("should work without errors", func() {
			if envTestURL == "" {
				return
			}

			serviceURL, err := url.Parse(envTestURL)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = service.Initialize(serviceURL, logger)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = service.Send(
				"this is an integration test",
				nil,
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
	ginkgo.When("creating a config", func() {
		ginkgo.When("given an url", func() {
			ginkgo.It("should return an error if no arguments where supplied", func() {
				serviceURL, _ := url.Parse("ifttt://")
				err := service.Initialize(serviceURL, logger)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
			ginkgo.It("should return an error if no webhook ID is given", func() {
				serviceURL, _ := url.Parse("ifttt:///?events=event1")
				err := service.Initialize(serviceURL, logger)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
			ginkgo.It("should return an error no events are given", func() {
				serviceURL, _ := url.Parse("ifttt://dummyID")
				err := service.Initialize(serviceURL, logger)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
			ginkgo.It("should return an error when an invalid query key is given", func() {
				serviceURL, _ := url.Parse("ifttt://dummyID/?events=event1&badquery=foo")
				err := service.Initialize(serviceURL, logger)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
			ginkgo.It("should return an error if message value is above 3", func() {
				serviceURL, _ := url.Parse("ifttt://dummyID/?events=event1&messagevalue=8")
				config := Config{}
				err := config.SetURL(serviceURL)
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
			ginkgo.It("should not return an error if webhook ID and at least one event is given", func() {
				serviceURL, _ := url.Parse("ifttt://dummyID/?events=event1")
				err := service.Initialize(serviceURL, logger)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
			ginkgo.It("should set value1, value2 and value3", func() {
				serviceURL, _ := url.Parse("ifttt://dummyID/?events=dummyevent&value3=three&value2=two&value1=one")
				config := Config{}
				err := config.SetURL(serviceURL)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Expect(config.Value1).To(gomega.Equal("one"))
				gomega.Expect(config.Value2).To(gomega.Equal("two"))
				gomega.Expect(config.Value3).To(gomega.Equal("three"))
			})
		})
	})
	ginkgo.When("serializing a config to URL", func() {
		ginkgo.When("given multiple events", func() {
			ginkgo.It("should return an URL with all the events comma-separated", func() {
				expectedURL := "ifttt://dummyID/?events=foo%2Cbar%2Cbaz&messagevalue=0"
				config := Config{
					Events:            []string{"foo", "bar", "baz"},
					WebHookID:         "dummyID",
					UseMessageAsValue: 0,
				}
				resultURL := config.GetURL().String()
				gomega.Expect(resultURL).To(gomega.Equal(expectedURL))
			})
		})

		ginkgo.When("given values", func() {
			ginkgo.It("should return an URL with all the values", func() {
				expectedURL := "ifttt://dummyID/?messagevalue=0&value1=v1&value2=v2&value3=v3"
				config := Config{
					WebHookID: "dummyID",
					Value1:    "v1",
					Value2:    "v2",
					Value3:    "v3",
				}
				resultURL := config.GetURL().String()
				gomega.Expect(resultURL).To(gomega.Equal(expectedURL))
			})
		})
	})
	ginkgo.When("sending a message", func() {
		ginkgo.It("should error if the response code is not 204 no content", func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			setupResponder("foo", "dummy", 404, "")

			URL, _ := url.Parse("ifttt://dummy/?events=foo")

			if err := service.Initialize(URL, logger); err != nil {
				ginkgo.Fail("errored during initialization")
			}

			err := service.Send("hello", nil)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
		ginkgo.It("should not error if the response code is 204", func() {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			setupResponder("foo", "dummy", 204, "")

			URL, _ := url.Parse("ifttt://dummy/?events=foo")

			if err := service.Initialize(URL, logger); err != nil {
				ginkgo.Fail("errored during initialization")
			}

			err := service.Send("hello", nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
	ginkgo.When("creating a json payload", func() {
		ginkgo.When("given config values \"a\", \"b\" and \"c\"", func() {
			ginkgo.It("should return a valid jsonPayload string with values \"a\", \"b\" and \"c\"", func() {
				bytes, err := createJSONToSend(&Config{
					Value1:            "a",
					Value2:            "b",
					Value3:            "c",
					UseMessageAsValue: 0,
				}, "d", nil)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				payload := jsonPayload{}
				err = json.Unmarshal(bytes, &payload)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				gomega.Expect(payload.Value1).To(gomega.Equal("a"))
				gomega.Expect(payload.Value2).To(gomega.Equal("b"))
				gomega.Expect(payload.Value3).To(gomega.Equal("c"))
			})
		})
		ginkgo.When("message value is set to 3", func() {
			ginkgo.It("should return a jsonPayload string with value2 set to message", func() {
				config := &Config{
					Value1: "a",
					Value2: "b",
					Value3: "c",
				}

				for i := 1; i <= 3; i++ {
					config.UseMessageAsValue = uint8(i)
					bytes, err := createJSONToSend(config, "d", nil)
					gomega.Expect(err).ToNot(gomega.HaveOccurred())

					payload := jsonPayload{}
					err = json.Unmarshal(bytes, &payload)
					gomega.Expect(err).ToNot(gomega.HaveOccurred())

					if i == 1 {
						gomega.Expect(payload.Value1).To(gomega.Equal("d"))
					} else if i == 2 {
						gomega.Expect(payload.Value2).To(gomega.Equal("d"))
					} else if i == 3 {
						gomega.Expect(payload.Value3).To(gomega.Equal("d"))
					}

				}
			})
		})
		ginkgo.When("given a param overrides for value1, value2 and value3", func() {
			ginkgo.It("should return a jsonPayload string with value1, value2 and value3 overridden", func() {
				bytes, err := createJSONToSend(&Config{
					Value1:            "a",
					Value2:            "b",
					Value3:            "c",
					UseMessageAsValue: 0,
				}, "d", (*types.Params)(&map[string]string{
					"value1": "e",
					"value2": "f",
					"value3": "g",
				}))
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				payload := &jsonPayload{}
				err = json.Unmarshal(bytes, payload)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				gomega.Expect(payload.Value1).To(gomega.Equal("e"))
				gomega.Expect(payload.Value2).To(gomega.Equal("f"))
				gomega.Expect(payload.Value3).To(gomega.Equal("g"))
			})
		})
	})
})

func setupResponder(event string, key string, code int, body string) {
	targetURL := fmt.Sprintf("https://maker.ifttt.com/trigger/%s/with/key/%s", event, key)
	httpmock.RegisterResponder("POST", targetURL, httpmock.NewStringResponder(code, body))
}

package rocketchat

import (
	"net/url"
	"os"
	"testing"

	"github.com/nicholas-fedor/shoutrrr/internal/testutils"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var (
	service          *Service
	envRocketchatURL *url.URL
	_                = ginkgo.BeforeSuite(func() {
		service = &Service{}
		envRocketchatURL, _ = url.Parse(os.Getenv("SHOUTRRR_ROCKETCHAT_URL"))
	})
)

func TestRocketchat(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Shoutrrr Rocketchat Suite")
}

var _ = ginkgo.Describe("the rocketchat service", func() {
	ginkgo.When("running integration tests", func() {
		ginkgo.It("should work without errors", func() {
			if envRocketchatURL.String() == "" {
				return
			}
			serviceURL, _ := url.Parse(envRocketchatURL.String())
			service.Initialize(serviceURL, testutils.TestLogger())
			err := service.Send(
				"this is an integration test",
				nil,
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
	ginkgo.Describe("the rocketchat config", func() {
		ginkgo.When("generating a config object", func() {
			rocketchatURL, _ := url.Parse("rocketchat://rocketchat.my-domain.com/tokenA/tokenB")
			config := &Config{}
			err := config.SetURL(rocketchatURL)
			ginkgo.It("should not have caused an error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
			ginkgo.It("should set host", func() {
				gomega.Expect(config.Host).To(gomega.Equal("rocketchat.my-domain.com"))
			})
			ginkgo.It("should set token A", func() {
				gomega.Expect(config.TokenA).To(gomega.Equal("tokenA"))
			})
			ginkgo.It("should set token B", func() {
				gomega.Expect(config.TokenB).To(gomega.Equal("tokenB"))
			})
			ginkgo.It("should not set channel or username", func() {
				gomega.Expect(config.Channel).To(gomega.BeEmpty())
				gomega.Expect(config.UserName).To(gomega.BeEmpty())
			})
		})
		ginkgo.When("generating a new config with url, that has no token", func() {
			rocketchatURL, _ := url.Parse("rocketchat://rocketchat.my-domain.com")
			config := &Config{}
			err := config.SetURL(rocketchatURL)
			ginkgo.It("should return an error", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})
		ginkgo.When("generating a config object with username only", func() {
			rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com/tokenA/tokenB")
			config := &Config{}
			err := config.SetURL(rocketchatURL)
			ginkgo.It("should not have caused an error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
			ginkgo.It("should set username", func() {
				gomega.Expect(config.UserName).To(gomega.Equal("testUserName"))
			})
			ginkgo.It("should not set channel", func() {
				gomega.Expect(config.Channel).To(gomega.BeEmpty())
			})
		})
		ginkgo.When("generating a config object with channel only", func() {
			rocketchatURL, _ := url.Parse("rocketchat://rocketchat.my-domain.com/tokenA/tokenB/testChannel")
			config := &Config{}
			err := config.SetURL(rocketchatURL)
			ginkgo.It("should not hav caused an error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
			ginkgo.It("should set channel", func() {
				gomega.Expect(config.Channel).To(gomega.Equal("#testChannel"))
			})
			ginkgo.It("should not set username", func() {
				gomega.Expect(config.UserName).To(gomega.BeEmpty())
			})
		})
		ginkgo.When("generating a config object with channel and userName", func() {
			rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com/tokenA/tokenB/testChannel")
			config := &Config{}
			err := config.SetURL(rocketchatURL)
			ginkgo.It("should not hav caused an error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
			ginkgo.It("should set channel", func() {
				gomega.Expect(config.Channel).To(gomega.Equal("#testChannel"))
			})
			ginkgo.It("should set username", func() {
				gomega.Expect(config.UserName).To(gomega.Equal("testUserName"))
			})
		})
		ginkgo.When("generating a config object with user and userName", func() {
			rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com/tokenA/tokenB/@user")
			config := &Config{}
			err := config.SetURL(rocketchatURL)
			ginkgo.It("should not hav caused an error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
			ginkgo.It("should set channel", func() {
				gomega.Expect(config.Channel).To(gomega.Equal("@user"))
			})
			ginkgo.It("should set username", func() {
				gomega.Expect(config.UserName).To(gomega.Equal("testUserName"))
			})
		})
	})
	ginkgo.Describe("Sending messages", func() {
		ginkgo.When("sending a message completely without parameters", func() {
			rocketchatURL, _ := url.Parse("rocketchat://rocketchat.my-domain.com/tokenA/tokenB")
			config := &Config{}
			config.SetURL(rocketchatURL)
			ginkgo.It("should generate the correct url to call", func() {
				generatedURL := buildURL(config)
				gomega.Expect(generatedURL).To(gomega.Equal("https://rocketchat.my-domain.com/hooks/tokenA/tokenB"))
			})
			ginkgo.It("should generate the correct JSON body", func() {
				json, err := CreateJSONPayload(config, "this is a message", nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(string(json)).To(gomega.Equal("{\"text\":\"this is a message\"}"))
			})
		})
		ginkgo.When("sending a message with pre set username and channel", func() {
			rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com/tokenA/tokenB/testChannel")
			config := &Config{}
			config.SetURL(rocketchatURL)
			ginkgo.It("should generate the correct JSON body", func() {
				json, err := CreateJSONPayload(config, "this is a message", nil)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(string(json)).To(gomega.Equal("{\"text\":\"this is a message\",\"username\":\"testUserName\",\"channel\":\"#testChannel\"}"))
			})
		})
		ginkgo.When("sending a message with pre set username and channel but overwriting them with parameters", func() {
			rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com/tokenA/tokenB/testChannel")
			config := &Config{}
			config.SetURL(rocketchatURL)
			ginkgo.It("should generate the correct JSON body", func() {
				params := (*types.Params)(&map[string]string{"username": "overwriteUserName", "channel": "overwriteChannel"})
				json, err := CreateJSONPayload(config, "this is a message", params)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(string(json)).To(gomega.Equal("{\"text\":\"this is a message\",\"username\":\"overwriteUserName\",\"channel\":\"overwriteChannel\"}"))
			})
		})
		ginkgo.When("sending to an URL which contains HOST:PORT", func() {
			rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com:5055/tokenA/tokenB/testChannel")
			config := &Config{}
			config.SetURL(rocketchatURL)
			ginkgo.It("should generate a correct hook URL https://HOST:PORT", func() {
				hookURL := buildURL(config)
				gomega.Expect(hookURL).To(gomega.ContainSubstring("my-domain.com:5055"))
			})
		})
		ginkgo.When("sending to an URL with badly syntaxed #channel name", func() {
			ginkgo.It("should properly parse the Channel", func() {
				rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com:5055/tokenA/tokenB/###########################testChannel")
				config := &Config{}
				config.SetURL(rocketchatURL)
				gomega.Expect(config.Channel).To(gomega.ContainSubstring("###########################testChannel"))
			})
			ginkgo.It("should properly parse the Channel", func() {
				rocketchatURL, _ := url.Parse("rocketchat://testUserName@rocketchat.my-domain.com:5055/tokenA/tokenB/#testChannel")
				config := &Config{}
				config.SetURL(rocketchatURL)
				gomega.Expect(config.Channel).To(gomega.ContainSubstring("#testChannel"))
			})
		})
	})
})

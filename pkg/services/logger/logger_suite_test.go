package logger_test

import (
	"log"
	"testing"

	"github.com/nicholas-fedor/shoutrrr/internal/testutils"
	"github.com/nicholas-fedor/shoutrrr/pkg/services/logger"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

func TestLogger(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Logger Suite")
}

var _ = ginkgo.Describe("the logger service", func() {
	ginkgo.When("sending a notification", func() {
		ginkgo.It("should output the message to the log", func() {
			logbuf := gbytes.NewBuffer()
			service := &logger.Service{}
			_ = service.Initialize(testutils.URLMust(`logger://`), log.New(logbuf, "", 0))

			err := service.Send(`Failed - Requires Toaster Repair Level 10`, nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Eventually(logbuf).Should(gbytes.Say("Failed - Requires Toaster Repair Level 10"))
		})

		ginkgo.It("should not mutate the passed params", func() {
			service := &logger.Service{}
			_ = service.Initialize(testutils.URLMust(`logger://`), nil)
			params := types.Params{}
			err := service.Send(`Failed - Requires Toaster Repair Level 10`, &params)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			gomega.Expect(params).To(gomega.BeEmpty())
		})

		ginkgo.When("when a template has been added", func() {
			ginkgo.It("should render template with params", func() {
				logbuf := gbytes.NewBuffer()
				service := &logger.Service{}
				_ = service.Initialize(testutils.URLMust(`logger://`), log.New(logbuf, "", 0))
				err := service.SetTemplateString(`message`, `{{.level}}: {{.message}}`)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				params := types.Params{
					"level": "warning",
				}
				err = service.Send(`Requires Toaster Repair Level 10`, &params)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Eventually(logbuf).Should(gbytes.Say("warning: Requires Toaster Repair Level 10"))
			})
		})
	})
})

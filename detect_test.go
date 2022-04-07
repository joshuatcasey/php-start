package phpstart_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/paketo-buildpacks/packit/v2"
	phpstart "github.com/paketo-buildpacks/php-start"
	"github.com/sclevine/spec"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		detect     packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		detect = phpstart.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Detect", func() {
		it("requires php, php-fpm, httpd, httpd-conf, and httpd-start and provides httpd-start", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "php",
						Metadata: phpstart.BuildPlanMetadata{
							Build: true,
						},
					},
					{
						Name: "php-fpm",
						Metadata: phpstart.BuildPlanMetadata{
							Build:  true,
							Launch: true,
						},
					},
					{
						Name: "httpd",
						Metadata: phpstart.BuildPlanMetadata{
							Launch: true,
						},
					},
					{
						Name: "httpd-config",
						Metadata: phpstart.BuildPlanMetadata{
							Launch: true,
							Build:  true,
						},
					},
				},
			}))
		})

		context("with composer", func() {
			context("with composer.json", func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(workingDir, "composer.json"), []byte(""), os.ModePerm)).To(Succeed())
				})

				it("requires composer-packages", func() {
					result, err := detect(packit.DetectContext{
						WorkingDir: workingDir,
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(result.Plan.Requires).To(ContainElements(packit.BuildPlanRequirement{
						Name: "composer-packages",
						Metadata: phpstart.BuildPlanMetadata{
							Build: true,
						},
					}))
				})
			})

			context("with $COMPOSER", func() {
				it.Before(func() {
					Expect(os.Setenv("COMPOSER", "some/other-file.json")).To(Succeed())
				})

				it.After(func() {
					Expect(os.Unsetenv("COMPOSER")).To(Succeed())
				})

				context("that points to an existing file", func() {
					it.Before(func() {
						Expect(os.Mkdir(filepath.Join(workingDir, "some"), os.ModeDir|os.ModePerm)).To(Succeed())
						Expect(os.WriteFile(filepath.Join(workingDir, "some", "other-file.json"), []byte(""), os.ModePerm)).To(Succeed())
					})

					it("requires composer-packages", func() {
						result, err := detect(packit.DetectContext{
							WorkingDir: workingDir,
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(result.Plan.Requires).To(ContainElements(packit.BuildPlanRequirement{
							Name: "composer-packages",
							Metadata: phpstart.BuildPlanMetadata{
								Build: true,
							},
						}))
					})
				})

				context("that points to a non existing file", func() {
					it("requires composer-packages", func() {
						result, err := detect(packit.DetectContext{
							WorkingDir: workingDir,
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(result.Plan.Requires).ToNot(ContainElements(MatchFields(IgnoreExtras, Fields{
							"Name": Equal("composer-packages"),
						})))
					})
				})
			})
		}, spec.Sequential())
	})
}

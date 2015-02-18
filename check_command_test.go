package resource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/google/go-github/github"

	"github.com/concourse/github-release-resource"
	"github.com/concourse/github-release-resource/fakes"
)

var _ = Describe("Check Command", func() {
	var (
		githubClient *fakes.FakeGitHub
		command      *resource.CheckCommand

		returnedReleases []github.RepositoryRelease
	)

	BeforeEach(func() {
		githubClient = &fakes.FakeGitHub{}
		command = resource.NewCheckCommand(githubClient)

		returnedReleases = []github.RepositoryRelease{}
	})

	JustBeforeEach(func() {
		githubClient.ListReleasesReturns(returnedReleases, nil)
	})

	Context("when this is the first time that the resource has been run", func() {
		Context("when there are no releases", func() {
			BeforeEach(func() {
				returnedReleases = []github.RepositoryRelease{}
			})

			It("returns an error when there are no releases", func() {
				_, err := command.Run(resource.CheckRequest{})
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when there are releases", func() {
			BeforeEach(func() {
				returnedReleases = []github.RepositoryRelease{
					{TagName: github.String("v0.4.0")},
					{TagName: github.String("0.1.3")},
					{TagName: github.String("v0.1.2")},
				}
			})

			It("outputs the most recent version only", func() {
				command := resource.NewCheckCommand(githubClient)

				response, err := command.Run(resource.CheckRequest{})
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(HaveLen(1))
				Ω(response[0]).Should(Equal(resource.Version{
					Tag: "v0.4.0",
				}))
			})
		})
	})

	Context("when there are prior versions", func() {
		Context("when there are no releases", func() {
			BeforeEach(func() {
				returnedReleases = []github.RepositoryRelease{}
			})

			It("returns an error when there are no releases", func() {
				_, err := command.Run(resource.CheckRequest{})
				Ω(err).Should(HaveOccurred())
			})
		})

		Context("when there are releases", func() {
			BeforeEach(func() {
				returnedReleases = []github.RepositoryRelease{
					{TagName: github.String("v0.1.4")},
					{TagName: github.String("0.4.0")},
					{TagName: github.String("v0.1.3")},
					{TagName: github.String("0.1.2")},
				}
			})

			It("returns an empty list if the lastet version has been checked", func() {
				command := resource.NewCheckCommand(githubClient)

				response, err := command.Run(resource.CheckRequest{
					Version: resource.Version{
						Tag: "0.4.0",
					},
				})
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(BeEmpty())
			})

			It("returns all of the versions that are newer", func() {
				command := resource.NewCheckCommand(githubClient)

				response, err := command.Run(resource.CheckRequest{
					Version: resource.Version{
						Tag: "v0.1.3",
					},
				})
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response).Should(Equal([]resource.Version{
					{Tag: "v0.1.4"},
					{Tag: "0.4.0"},
				}))
			})
		})
	})
})

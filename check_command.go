package resource

import (
	"sort"
	"strconv"

	"github.com/blang/semver"
	"github.com/zachgersh/go-github/github"
)

type CheckCommand struct {
	github GitHub
}

func NewCheckCommand(github GitHub) *CheckCommand {
	return &CheckCommand{
		github: github,
	}
}

func (c *CheckCommand) Run(request CheckRequest) ([]Version, error) {
	releases, err := c.github.ListReleases()
	if err != nil {
		return []Version{}, err
	}

	if len(releases) == 0 {
		return []Version{}, nil
	}

	var filteredReleases []github.RepositoryRelease

	for _, release := range releases {
		if request.Source.Drafts != *release.Draft {
			continue
		}
		if release.TagName == nil {
			continue
		}
		if _, err := semver.New(determineVersionFromTag(*release.TagName)); err != nil {
			continue
		}

		filteredReleases = append(filteredReleases, release)
	}

	sort.Sort(byVersion(filteredReleases))

	if len(filteredReleases) == 0 {
		return []Version{}, nil
	}
	latestRelease := filteredReleases[len(filteredReleases)-1]

	if (request.Version == Version{}) {
		return []Version{
			versionFromDraft(&latestRelease),
		}, nil
	}

	if *latestRelease.TagName == request.Version.Tag {
		return []Version{}, nil
	}

	upToLatest := false
	reversedVersions := []Version{}
	for _, release := range filteredReleases {

		if upToLatest {
			reversedVersions = append(reversedVersions, versionFromDraft(&release))
		} else {
			if *release.Draft {
				id := *release.ID
				upToLatest = request.Version.ID == strconv.Itoa(id)
			} else {
				version := *release.TagName
				upToLatest = request.Version.Tag == version
			}
		}
	}

	if !upToLatest {
		// current version was removed; start over from latest
		reversedVersions = append(
			reversedVersions,
			versionFromDraft(&filteredReleases[len(filteredReleases)-1]),
		)
	}

	return reversedVersions, nil
}

type byVersion []github.RepositoryRelease

func (r byVersion) Len() int {
	return len(r)
}

func (r byVersion) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r byVersion) Less(i, j int) bool {
	first, err := semver.New(determineVersionFromTag(*r[i].TagName))
	if err != nil {
		return true
	}

	second, err := semver.New(determineVersionFromTag(*r[j].TagName))
	if err != nil {
		return false
	}

	return first.LT(*second)
}

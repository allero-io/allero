package githubConnector

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
)

func ListByOwnerWithPagination(client *github.Client, owner string, ownerType string) (chan *github.Repository, error) {
	repositoriesChan := make(chan *github.Repository)
	var err error

	listOptions := github.ListOptions{
		PerPage: 100,
		Page:    1,
	}

	go func() {
		defer close(repositoriesChan)
		fmt.Printf("Start fetching all repositories of owner %s\n", owner)
		totalFetchedForOwner := 0
		for {
			var repos []*github.Repository
			var resp *github.Response

			if ownerType == "Organization" {
				opt := &github.RepositoryListByOrgOptions{
					ListOptions: listOptions,
				}
				repos, resp, err = client.Repositories.ListByOrg(context.Background(), owner, opt)
			} else {
				opt := &github.RepositoryListOptions{
					ListOptions: listOptions,
				}
				repos, resp, err = client.Repositories.List(context.Background(), owner, opt)
			}

			for _, repo := range repos {
				repositoriesChan <- repo
			}
			fetchedInPage := min(listOptions.Page*listOptions.PerPage, len(repos))
			totalFetchedForOwner = totalFetchedForOwner + fetchedInPage
			fmt.Printf("Fetched %d repositories for owner %s\n", totalFetchedForOwner, owner)
			listOptions.Page = resp.NextPage

			if resp.NextPage == 0 {
				break
			}
		}
		fmt.Printf("Finished fetching all repositories of owner %s\n", owner)
	}()

	return repositoriesChan, err
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

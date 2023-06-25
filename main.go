package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MateSousa/create-release/initializers"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

const (
	prLabelPending string = "createrelease:pending"
	prLabelmerged  string = "createrelease:merged"
)

func main() {
	env, err := initializers.LoadEnv()
	if err != nil {
		fmt.Printf("error loading env: %v", err)
		os.Exit(1)
	}

	client, err := CreateGithubClient(env)
	if err != nil {
		fmt.Printf("error creating github client: %v", err)
		os.Exit(1)
	}

	// Create a release PR from base to target branch
	pr, err := CreateReleasePR(client, env)
	if err != nil {
		fmt.Printf("error creating release PR: %v", err)
		os.Exit(1)
	}

	fmt.Printf("PR created: %v", pr)

	os.Exit(0)
}

// Create a github client with a token
func CreateGithubClient(env initializers.Env) (*github.Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: env.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return client, nil
}

// Create a release PR from base to target branch
func CreateReleasePR(client *github.Client, env initializers.Env) (*github.PullRequest, error) {

	// Check if exist a PR with a label "createrelease:pending"
	prExist, err := CheckIfPRExist(client, env)
	if err != nil {
		return nil, err
	}
	if prExist {
		fmt.Println("PR already exist, exiting...")
		os.Exit(0)
	}

	// new release tag
	newTag, err := GetLatestReleaseTag(client, env)
	if err != nil {
		return nil, err
	}

	// Create a release PR from base to target branch
	newPR := &github.NewPullRequest{
		Title: github.String("Release " + newTag),
		Head:  github.String(env.BaseBranch),
		Base:  github.String(env.TargetBranch),
		Body:  github.String("This is an automated PR to create a new release"),
	}
	pr, _, err := client.PullRequests.Create(context.Background(), env.RepoOwner, env.RepoName, newPR)
	if err != nil {
		return nil, err
	}

	// Add label "createrelease:pending" to PR
	err = AddPendingLabel(client, env, pr)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// Check if exist a PR with a label "createrelease:pending"
// If exist, return nil
func CheckIfPRExist(client *github.Client, env initializers.Env) (bool, error) {
	prList, _, err := client.PullRequests.List(context.Background(), env.RepoOwner, env.RepoName, nil)
	if err != nil {
		return false, err
	}

	for _, pr := range prList {
		for _, label := range pr.Labels {
			if *label.Name == prLabelPending {
				return true, nil
			}
		}
	}
	return false, nil
}

// Create a label "createrelease:pending" and add to PR
func AddPendingLabel(client *github.Client, env initializers.Env, pr *github.PullRequest) error {
	_, _, err := client.Issues.AddLabelsToIssue(context.Background(), env.RepoOwner, env.RepoName, *pr.Number, []string{prLabelPending})
	if err != nil {
		return err
	}

	return nil
}

// Get the latest release tag and increment the minor version
func GetLatestReleaseTag(client *github.Client, env initializers.Env) (string, error) {
	var releaseTag string

	releaseList, _, err := client.Repositories.ListReleases(context.Background(), env.RepoOwner, env.RepoName, nil)
	if err != nil {
		return "", err
	}

	noReleaseTag := len(releaseList) == 0
	if noReleaseTag {
		releaseTag = "v0.0.1"
	} else {
		latestReleaseTag := releaseList[0].GetTagName()
		latestReleaseTagSplit := strings.Split(latestReleaseTag, ".")

		latestReleaseTagMajorVersion, err := strconv.Atoi(latestReleaseTagSplit[0][1:])
		if err != nil {
			return "", err
		}

		latestReleaseTagMinorVersion, err := strconv.Atoi(latestReleaseTagSplit[1])
		if err != nil {
			return "", err
		}
		latestReleaseTagPatchVersion, err := strconv.Atoi(latestReleaseTagSplit[2])
		if err != nil {
			return "", err
		}

		if latestReleaseTagPatchVersion == 9 {
			latestReleaseTagMinorVersion = latestReleaseTagMinorVersion + 1
			latestReleaseTagPatchVersion = 0
		} else {
			latestReleaseTagPatchVersion = latestReleaseTagPatchVersion + 1
		}
		if latestReleaseTagMinorVersion == 9 && latestReleaseTagPatchVersion == 9 {
			latestReleaseTagMajorVersion = latestReleaseTagMajorVersion + 1
			latestReleaseTagMinorVersion = 0
			latestReleaseTagPatchVersion = 0
		}

		releaseTag = fmt.Sprintf("v%d.%d.%d", latestReleaseTagMajorVersion, latestReleaseTagMinorVersion, latestReleaseTagPatchVersion)
	}

	return releaseTag, nil
}

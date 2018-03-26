package vsts

import (
	"fmt"
	"log"
	"strings"
)

const (
	botCommentPrefix = "[BOT]\n"
)

var config *Config

func init() {
	var err error
	config, err = GetConfig()
	if err != nil {
		log.Fatal(err)
	}
}

// ReviewGoTest review if all changed Go files have test updated.
func ReviewGoTest(pr *PullRequest) error {
	goSuffix := ".go"
	goTestSuffix := "_test.go"
	commentMsg := fmt.Sprintf("%s\nPlease update test.", botCommentPrefix)

	diffs, err := getDiffsBetweenBranches(getBranchNameFromRefName(pr.Resource.TargetRefName), getBranchNameFromRefName(pr.Resource.SourceRefName))
	if err != nil {
		return err
	}

	var changedGoFiles []string
	var changedGoTestFiles []string

	for _, change := range diffs.Changes {
		if strings.HasSuffix(change.Item.Path, goSuffix) {
			changedGoFiles = append(changedGoFiles, change.Item.Path)
		} else if strings.HasSuffix(change.Item.Path, goTestSuffix) {
			changedGoTestFiles = append(changedGoTestFiles, change.Item.Path)
		}
	}

	var missingTestGoFiles []string
	for _, changedGoFile := range changedGoFiles {
		for _, changedGoTestFile := range changedGoTestFiles {
			if strings.EqualFold(strings.TrimSuffix(changedGoTestFile, goTestSuffix), strings.TrimSuffix(changedGoFile, goSuffix)) {
				log.Printf("%s has test update: %s", changedGoFile, changedGoTestFile)
				break
			}
		}
		missingTestGoFiles = append(missingTestGoFiles, changedGoFile)
	}

	commentThreads, err := getCommentThreads(pr.Resource.PullRequestID)
	log.Printf("threads: %v", commentThreads.Count)

	if err != nil {
		return err
	}

	for _, goFile := range missingTestGoFiles {
		commentThread := commentThread{}
		for _, thread := range commentThreads.Value {
			if !thread.IsDeleted && strings.EqualFold(thread.ThreadContext.FilePath, goFile) {
				for _, comment := range thread.Comments {
					if comment.ID == 1 && comment.Author.ID == config.UserID && strings.HasPrefix(comment.Content, botCommentPrefix) {
						commentThread = thread
						break
					}
				}
			}
		}

		if commentThread.Status == "" {
			// create thread
			err := createCommentThread(pr.Resource.PullRequestID, goFile, commentMsg)
			if err != nil {
				return err
			}
		} else {
			// add comment
			err := addComment(pr.Resource.PullRequestID, commentThread, commentMsg)
			if err != nil {
				return err
			}
			// set thread active
			err = setCommentThreadStatus(pr.Resource.PullRequestID, commentThread, 1)
			if err != nil {
				return err
			}
		}
	}

	// vote
	if len(missingTestGoFiles) == 0 {
		for _, reviewer := range pr.Resource.Reviewers {
			if strings.EqualFold(reviewer.ID, config.UserID) {
				if reviewer.Vote < 0 {
					// reset
					err := votePullRequest(pr.Resource.PullRequestID, 0)
					if err != nil {
						return err
					}
				}
				break
			}
		}
	} else {
		// wait
		err := votePullRequest(pr.Resource.PullRequestID, -5)
		if err != nil {
			return err
		}
	}

	return nil
}

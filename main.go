package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/wenwu449/vsts-pr-review-sample/vsts"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	config, err := vsts.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	pr, err := vsts.ParsePullRequest()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Got PR update: %s\n", pr.ID)

	if !strings.EqualFold(pr.Resource.TargetRefName, fmt.Sprintf("%s/%s", "refs/heads", config.MasterBranch)) {
		log.Fatal(fmt.Errorf("unexpected target branch: %s", pr.Resource.TargetRefName))
	}

	err = vsts.ReviewGoTest(pr)
	if err != nil {
		log.Fatal(err)
	}
}

package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GitlabService struct {
	accessToken               string
	host                      string
	mergeRequestApprovalRules []MergeRequestApprovalRules
	teams                     map[Team][]Username
}

type MergeRequestApprovalRules struct {
	projectID int
	steps     []Step
}

type Step struct {
	Order             int
	ApprovalsRequired int
	Teams             []Team
}

type Team string
type Username string

func CreateGitlabService(accessToken string, host string) *GitlabService {
	return &GitlabService{
		accessToken: accessToken,
		host:        host,
		mergeRequestApprovalRules: []MergeRequestApprovalRules{
			{
				projectID: 154,
				steps: []Step{
					{
						Order:             1,
						ApprovalsRequired: 1,
						Teams: []Team{
							"sickbustardsclub",
						},
					},
				},
			},
		},
		teams: map[Team][]Username{
			"sickbustardsclub": {"denis.ulyanov"},
		},
	}
}

type MergeRequest struct {
	ID        int    `json:"id"`
	Iid       int    `json:"iid"`
	ProjectID int    `json:"project_id"`
	Title     string `json:"title"`
	State     string `json:"state"`
	Author    User   `json:"author"`
	Assignees []User `json:"assignees"`
	Assignee  User   `json:"assignee"`
	WIP       bool   `json:"work_in_progress"`
}

func (gl *GitlabService) AssignMergeRequest(ctx context.Context) {
	for ruleIndex := range gl.mergeRequestApprovalRules {
		mergeRequests := gl.MergeRequestList(ctx, gl.mergeRequestApprovalRules[ruleIndex].projectID)
		for mergeRequestIndex := range mergeRequests {
			gl.CheckMR(&mergeRequests[mergeRequestIndex], &gl.mergeRequestApprovalRules[ruleIndex])
		}
	}
}

func (gl *GitlabService) callMethod(method string) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v4/%s", gl.host, method), nil)
	req.Header.Add("PRIVATE-TOKEN", gl.accessToken)
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (gl *GitlabService) MergeRequestList(ctx context.Context, projectID int) []MergeRequest {
	body, _ := gl.callMethod(fmt.Sprintf("projects/%d/merge_requests?state=opened", projectID))

	var mergeRequestList []MergeRequest
	err := json.Unmarshal(body, &mergeRequestList)
	if err != nil {
		fmt.Println(err)
		return []MergeRequest{}
	}

	return mergeRequestList
}

package gitlab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

type Approvals struct {
	ApprovedBy []struct {
		User User `json:"user"`
	} `json:"approved_by"`
	SuggestedApprovers []User `json:"suggested_approvers"`
}

type MergeRequestApprovalRule struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	RuleType             string `json:"rule_type"`
	EligibleApprovers    []User `json:"eligible_approvers"`
	ApprovalsRequired    int    `json:"approvals_required"`
	Users                []User `json:"users"`
	ContainsHiddenGroups bool   `json:"contains_hidden_groups"`
	ApprovedBy           []User `json:"approved_by"`
	Approved             bool   `json:"approved"`
}

func (gl *GitlabService) CheckMR(mr *MergeRequest, rules *MergeRequestApprovalRules) {
	var approvals Approvals
	body, err := gl.callMethod(fmt.Sprintf("projects/%d/merge_requests/%d/approvals", mr.ProjectID, mr.Iid))
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &approvals)
	if err != nil {
		return
	}

	sort.SliceStable(rules.steps, func(i, j int) bool {
		return rules.steps[i].Order < rules.steps[j].Order
	})

	for _, step := range rules.steps {
		approvedCount := 0
		requireUser := map[Username]struct{}{}
		for _, team := range step.Teams {
			for _, user := range gl.teams[team] {
				requireUser[user] = struct{}{}
			}
		}

		for _, user := range approvals.ApprovedBy {
			_, ok := requireUser[user.User.Username]
			if ok {
				delete(requireUser, user.User.Username)
				approvedCount++
			}
		}

		if !(approvedCount >= step.ApprovalsRequired && len(approvals.SuggestedApprovers) == 0) {
			needYet := step.ApprovalsRequired - approvedCount

			for _, user := range approvals.SuggestedApprovers {
				_, ok := requireUser[user.Username]
				if ok {
					delete(requireUser, user.Username)
					needYet--
				}
			}

			approverUsername := []Username{}
			for _, user := range approvals.ApprovedBy {
				approverUsername = append(approverUsername, user.User.Username)
			}

			for _, user := range approvals.SuggestedApprovers {
				approverUsername = append(approverUsername, user.Username)
			}

			// approverUsername = append(approverUsername)

			gl.ChangeAllowedApproversForMergeRequest(mr, fmt.Sprintf("step%d", step.Order), approverUsername)
			return
		}
	}
}

func (gl *GitlabService) ChangeAllowedApproversForMergeRequest(mr *MergeRequest, ruleName string, approverUsername []Username) {
	mergeRequestApprovalRule := struct {
		Rules []MergeRequestApprovalRule `json:"rules"`
	}{}
	body, err := gl.callMethod(fmt.Sprintf("projects/%d/merge_requests/%d/approval_state", mr.ProjectID, mr.ID))
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &mergeRequestApprovalRule)
	if err != nil {
		return
	}

	approverIds := func(approverUsername []Username) []int {
		result := []int{}
		for _, username := range approverUsername {
			t := gl.FindUser(username)
			if t != nil {
				result = append(result, t.ID)
			}
		}
		return result
	}(approverUsername)

	for _, rule := range mergeRequestApprovalRule.Rules {
		if rule.Name == ruleName {
			gl.updateApprovalRules(mr, rule, approverIds)
			return
		}
	}

	gl.createApprovalRules(mr, ruleName, approverIds)
}

func (gl *GitlabService) updateApprovalRules(mr *MergeRequest, rule MergeRequestApprovalRule, approverIds []int) *MergeRequestApprovalRule {
	client := &http.Client{}
	body, _ := json.Marshal(map[string]interface{}{
		"name":               rule.Name,
		"user_ids":           approverIds,
		"approvals_required": len(approverIds),
	})
	fmt.Println(string(body))

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/approval_rules/%d", gl.host, mr.ProjectID, mr.Iid, rule.ID), strings.NewReader(string(body)))
	req.Header.Add("PRIVATE-TOKEN", gl.accessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var mergeRequestApprovalRule MergeRequestApprovalRule
	err = json.Unmarshal(body, &mergeRequestApprovalRule)
	if err != nil {
		return nil
	}

	return &mergeRequestApprovalRule
}

func (gl *GitlabService) createApprovalRules(mr *MergeRequest, ruleName string, approverIds []int) *MergeRequestApprovalRule {
	client := &http.Client{}
	body, _ := json.Marshal(map[string]interface{}{
		"name":               ruleName,
		"user_ids":           approverIds,
		"approvals_required": len(approverIds),
	})
	fmt.Println(string(body))

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/approval_rules", gl.host, mr.ProjectID, mr.Iid), strings.NewReader(string(body)))
	req.Header.Add("PRIVATE-TOKEN", gl.accessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var mergeRequestApprovalRule MergeRequestApprovalRule
	err = json.Unmarshal(body, &mergeRequestApprovalRule)
	if err != nil {
		return nil
	}

	return &mergeRequestApprovalRule
}

package main

import "testing"

func TestRunPRGateReopensOnlyGateClosedPullRequests(t *testing.T) {
	basePR := PullRequest{
		Number: 87, State: "closed", Base: Branch{Ref: "main"}, Body: "Fixes #73",
		User: User{Login: "contributor", Type: "User"},
	}

	tests := []struct {
		name       string
		pr         PullRequest
		gateClosed bool
		wantState  string
	}{
		{"reopens a gate-closed pull request", basePR, true, "open"},
		{"does not reopen a manually closed pull request", basePR, false, ""},
		{"does not reopen a merged pull request", PullRequest{Number: 87, State: "closed", Merged: true, Base: Branch{Ref: "main"}, Body: "Fixes #73", User: basePR.User}, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &fakeGitHub{
				pullRequest: tt.pr,
				issues: map[IssueReference]Issue{
					{Repository: "flock", Number: 73}: {Number: 73, State: "open", Labels: []Label{{Name: openToAnyone}}},
				},
				gateClosed: tt.gateClosed,
			}
			err := RunPRGate(api, "flock-php", "flock", PullRequestEvent{Action: "edited", Number: 87, BodyEdited: true})
			if err != nil {
				t.Fatal(err)
			}
			if api.state != tt.wantState {
				t.Fatalf("state = %q, want %q", api.state, tt.wantState)
			}
		})
	}
}

type fakeGitHub struct {
	pullRequest PullRequest
	issues      map[IssueReference]Issue
	gateClosed  bool
	state       string
}

func (f *fakeGitHub) PullRequest(string, string, int) (PullRequest, error) { return f.pullRequest, nil }
func (f *fakeGitHub) Permission(string, string, string) (string, error)    { return "read", nil }
func (f *fakeGitHub) Files(string, string, int) ([]File, error) {
	return []File{{Filename: "internal/site.go"}}, nil
}
func (f *fakeGitHub) Issue(_ string, repository string, number int) (Issue, error) {
	return f.issues[IssueReference{Repository: repository, Number: number}], nil
}
func (f *fakeGitHub) GateClosed(string, string, int) (bool, error) { return f.gateClosed, nil }
func (f *fakeGitHub) Comment(string, string, int, string) error    { return nil }
func (f *fakeGitHub) SetPullRequestState(_ string, _ string, _ int, state string) error {
	f.state = state
	return nil
}
func (f *fakeGitHub) AddLabel(string, string, int, string) error { return nil }

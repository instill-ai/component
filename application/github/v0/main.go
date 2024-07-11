//go:generate compogen readme ./config ./README.mdx
package github

import (
	"context"
	_ "embed"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/base"
	"github.com/instill-ai/x/errmsg"
)

const (
	taskListPRs             = "TASK_LIST_PULL_REQUESTS"
	taskGetPR               = "TASK_GET_PULL_REQUEST"
	taskGetCommit           = "TASK_GET_COMMIT"
	taskGetReviewComments   = "TASK_LIST_REVIEW_COMMENTS"
	taskCreateReviewComment = "TASK_CREATE_REVIEW_COMMENT"
	taskListIssues          = "TASK_LIST_ISSUES"
	taskGetIssue            = "TASK_GET_ISSUE"
	taskCreateIssue         = "TASK_CREATE_ISSUE"
	taskCreateWebhook       = "TASK_CREATE_WEBHOOK"
)

var (
	//go:embed config/definition.json
	definitionJSON []byte
	//go:embed config/setup.json
	setupJSON []byte
	//go:embed config/tasks.json
	tasksJSON []byte

	once sync.Once
	comp *component
)

type component struct {
	base.Component
}

type execution struct {
	base.ComponentExecution
	execute func(context.Context, *structpb.Struct) (*structpb.Struct, error)
	client  Client
}

// Init returns an implementation of IConnector that interacts with Slack.
func Init(bc base.Component) *component {
	once.Do(func() {
		comp = &component{Component: bc}
		err := comp.LoadDefinition(definitionJSON, setupJSON, tasksJSON, nil)
		if err != nil {
			panic(err)
		}
	})

	return comp
}

func (c *component) CreateExecution(sysVars map[string]any, setup *structpb.Struct, task string) (*base.ExecutionWrapper, error) {
	ctx := context.Background()
	githubClient := newClient(ctx, setup)
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
		client:             githubClient,
	}
	switch task {
	case taskListPRs:
		e.execute = e.client.listPullRequestsTask
	case taskGetPR:
		e.execute = e.client.getPullRequestTask
	case taskGetReviewComments:
		e.execute = e.client.listReviewCommentsTask
	case taskCreateReviewComment:
		e.execute = e.client.createReviewCommentTask
	case taskGetCommit:
		e.execute = e.client.getCommitTask
	case taskListIssues:
		e.execute = e.client.listIssuesTask
	case taskGetIssue:
		e.execute = e.client.getIssueTask
	case taskCreateIssue:
		e.execute = e.client.createIssueTask
	case taskCreateWebhook:
		e.execute = e.client.createWebhookTask
	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(ctx context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		if err := e.FillInDefaultValues(input); err != nil {
			return nil, err
		}
		output, err := e.execute(ctx, input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

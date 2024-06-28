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
	taskGetAllPRs = "TASK_GET_ALL_PULL_REQUESTS"
	taskGetPR  = "TASK_GET_PULL_REQUEST"
	taskGetCommit  = "TASK_GET_COMMIT"
	taskGetReviewComments = "TASK_GET_REVIEW_COMMENTS"
	taskCreateReviewComment = "TASK_CREATE_REVIEW_COMMENT"
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
	execute func(*structpb.Struct) (*structpb.Struct, error)
	client  GitHubClient
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
	client := newClient(ctx, setup)

	githubClient := GitHubClient{client: client}
	e := &execution{
		ComponentExecution: base.ComponentExecution{Component: c, SystemVariables: sysVars, Setup: setup, Task: task},
		client:             githubClient,
	}

	switch task {
	case taskGetAllPRs:
		e.execute = e.client.getAllPullRequestsTask
	case taskGetPR:
		e.execute = e.client.getPullRequestTask
	case taskGetReviewComments:
		e.execute = e.client.getAllReviewCommentsTask
	case taskCreateReviewComment:
		e.execute = e.client.createReviewCommentTask
		// e.execute = e.client.getAllPullRequests
	case taskGetCommit:
		e.execute = e.client.getCommitTask

	default:
		return nil, errmsg.AddMessage(
			fmt.Errorf("not supported task: %s", task),
			fmt.Sprintf("%s task is not supported.", task),
		)
	}

	return &base.ExecutionWrapper{Execution: e}, nil
}

func (e *execution) Execute(_ context.Context, inputs []*structpb.Struct) ([]*structpb.Struct, error) {
	outputs := make([]*structpb.Struct, len(inputs))

	for i, input := range inputs {
		output, err := e.execute(input)
		if err != nil {
			return nil, err
		}

		outputs[i] = output
	}

	return outputs, nil
}

func (c *component) Test(sysVars map[string]any, setup *structpb.Struct) error {

	return nil
}

package onlineapi

import (
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"testing"
)

func getOnlineAPITaskInfo() execute.ExecutorTaskInfo {
	executorConfig := execute.ExecutorConfig{
		OnlineAPI: map[string]execute.OnlineAPIConfig{
			"fofa":   {},
			"hunter": {},
			"quake":  {},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "",
			OrgId:          "test1",
			MainTaskId:     "onlineapi_test",
			ExecutorConfig: executorConfig,
		},
	}
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig

	return taskInfo
}

func TestResult_ParseResult(t *testing.T) {
	taskInfo := getOnlineAPITaskInfo()
	taskInfo.Executor = "fofa"
	result := Do(taskInfo)
	docs := ParseResult(taskInfo, result)
	for _, doc := range docs {
		t.Log(doc)
	}
}

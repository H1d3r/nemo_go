package icp

import (
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
)

func getOnlineAPITaskInfo() execute.ExecutorTaskInfo {
	executorConfig := execute.ExecutorConfig{
		ICP: map[string]execute.ICPConfig{
			"icp": {
				APIName: []string{"chinaz", "beianx"},
			},
			"icpPlus2": {
				APIName: []string{"chinaz", "beianx"},
			},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			OrgId:          "test1",
			MainTaskId:     "onlineapi_test",
			ExecutorConfig: executorConfig,
		},
	}
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig

	return taskInfo
}

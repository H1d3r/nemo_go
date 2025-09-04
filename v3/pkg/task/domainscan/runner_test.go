package domainscan

import (
	"testing"

	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
)

func TestResult_ParseResult(t *testing.T) {
	executorConfig := execute.ExecutorConfig{
		DomainScan: map[string]execute.DomainscanConfig{
			"massdns":   {},
			"subfinder": {},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			TargetMap:      map[string]string{execute.TargetRootDomain: "test.com"},
			OrgId:          "test1",
			MainTaskId:     "domaintest",
			ExecutorConfig: executorConfig,
		},
		Executor: "subfinder",
	}

	taskInfo.MainTaskInfo.ExecutorConfig.DomainScan = executorConfig.DomainScan

	result := Do(taskInfo)
	docs := result.ParseResult(taskInfo)
	for _, doc := range docs {
		t.Log(doc)
	}
}

func TestResovleSubdomain(t *testing.T) {
	domain := "yewxt.elane.cn"
	cname, resolvedSubdomains := ResolveDomain(domain)
	t.Log(cname)
	t.Log(resolvedSubdomains)
}

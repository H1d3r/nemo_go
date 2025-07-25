package interal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/core"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/mark3labs/mcp-go/mcp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type AssetData struct {
	Authority  string   `json:"authority"`
	Host       string   `json:"host"`
	Domain     string   `json:"domain,,omitempty"`
	IP         []string `json:"ip"`
	Port       string   `json:"port"`
	Status     string   `json:"status,omitempty"`
	Location   []string `json:"location"`
	Service    string   `json:"service,omitempty"`
	Title      string   `json:"title,omitempty"`
	Header     string   `json:"header,omitempty"`
	Cert       string   `json:"cert,omitempty"`
	Banner     string   `json:"banner,omitempty"`
	App        []string `json:"app,omitempty"`
	IconHash   string   `json:"icon_hash,omitempty"`
	IsCDN      bool     `json:"cdn,omitempty"`
	IsHoneypot bool     `json:"honeypot,omitempty"`
	Vul        []string `json:"vul,omitempty"` // 漏洞列表
	Icp        string   `json:"icp,omitempty"` // 备案信息
	IcpCompany string   `json:"icp_company,omitempty"`
	Whois      string   `json:"whois,omitempty"`
	UpdateTime string   `json:"update_time"`
}

type ResponseAssetData struct {
	Total         int         `json:"total"`
	Page          int         `json:"page"`
	PageSize      int         `json:"page_size"`
	AssetDataList []AssetData `json:"asset_data_list"`
}

func StartMCPServer() {
	s := core.NewMCPServer()
	// Add helloTool
	helloTool := mcp.NewTool("hello_nemo",
		mcp.WithDescription("NemoV3的MCP Server，返回Nemo的欢迎信息，并输出版本信息！"),
	)
	s.AddTool(helloTool, HelloHandler)
	//
	s.AddTool(mcp.NewTool("portscan",
		mcp.WithDescription("使用nemo创建一个端口扫描任务，扫描指定IP地址的存活端口，并进行指纹识别；创建任务后，可通过SSE实时获取任务状态和结果"),
		mcp.WithString("ip", mcp.Required(), mcp.Description("要扫描的IP地址，也可以是IP段，如192.168.0.0/24，一次只能指定一个IP地址或IP段")),
		mcp.WithString("port", mcp.DefaultString("--top-ports 1000"), mcp.Description("要扫描的端口，如80,443,10000-10010，或者--top-ports 1000（100或者10）；默认为--top-ports 1000")),
		mcp.WithString("rate", mcp.DefaultString("1000"), mcp.Description("扫描速率，每秒发送多少个包，默认1000")),
		mcp.WithString("bin", mcp.DefaultString("nmap"), mcp.Description("扫描工具，可选nmap、masscan、gogo")),
		mcp.WithBoolean("pocscan", mcp.DefaultBool(false), mcp.Description("是否开启POC和漏洞扫描，开启后会根据扫描的指纹自动匹配poc，扫描常见漏洞的POC，默认关闭")),
	), PortscanHandler)

	s.AddTool(mcp.NewTool("domainscan",
		mcp.WithDescription("使用nemo创建一个域名扫描任务，对域名执行子域名爆破和枚举操作，并进行指纹识别；创建任务后，可通过SSE实时获取任务状态和结果"),
		mcp.WithString("domain", mcp.Required(), mcp.Description("要扫描的域名，如example.com，一次只能指定一个域名")),
		mcp.WithBoolean("pocscan", mcp.DefaultBool(false), mcp.Description("是否开启POC和漏洞扫描，开启后会根据扫描的指纹自动匹配poc，扫描常见漏洞的POC，默认关闭")),
	), DomainscanHandler)

	s.AddTool(mcp.NewTool("onlineapi_scan",
		mcp.WithDescription("使用nemo创建一个调用在线API查询资产信息的任务，并进行指纹识别；创建任务后，可通过SSE实时获取任务状态和结果"),
		mcp.WithString("target", mcp.Required(), mcp.Description("要查询的资产信息，可以是IP地址或者域名，一次只能指定一个资产信息")),
		mcp.WithString("api", mcp.DefaultString("fofa"), mcp.Description("要调用的在线API，支持fofa、hunter、quake，请选择其中一个；一次只能指定一个API，默认为fofa")),
		mcp.WithBoolean("pocscan", mcp.DefaultBool(false), mcp.Description("是否开启POC和漏洞扫描，开启后会根据扫描的指纹自动匹配poc，扫描常见漏洞的POC，默认关闭")),
	), OnlineAPIHandler)

	s.AddTool(mcp.NewTool("query_task",
		mcp.WithDescription("查询nemo任务状态，需指定创建任务时返回的任务ID"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("nemo任务ID")),
	), QueryTaskHandler)

	s.AddTool(mcp.NewTool("query_result",
		mcp.WithDescription("查询nemo执行任务结果，需指定创建任务时返回的任务ID"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("nemo任务ID")),
		mcp.WithString("page", mcp.DefaultString("1"), mcp.Description("查询结果的页数，从1开始")),
		mcp.WithString("page_size", mcp.DefaultString("20"), mcp.Description("每次查询任务结果的最大条数，默认20")),
	), QueryResultHandler)

	s.AddTool(mcp.NewTool("query_asset",
		mcp.WithDescription("查询nemo数据库中存储的已收集到的资产信息，可根据主机名、IP、端口、标题、指纹等进行查询，可以指定多个查询条件，多个条件之间是AND关系"),
		mcp.WithString("host", mcp.DefaultString(""), mcp.Description("主机名，支持模糊查询")),
		mcp.WithString("ip", mcp.DefaultString(""), mcp.Description("IP地址，单个IP或带掩码的IP段")),
		mcp.WithString("port", mcp.DefaultString(""), mcp.Description("端口，支持单个端口或端口范围，如80,443,10000-10010")),
		mcp.WithString("title", mcp.DefaultString(""), mcp.Description("标题，支持模糊查询")),
		mcp.WithString("fingerprint", mcp.DefaultString(""), mcp.Description("指纹，支持模糊查询")),
		mcp.WithString("page", mcp.DefaultString("1"), mcp.Description("查询结果的页数，从1开始")),
		mcp.WithString("page_size", mcp.DefaultString("20"), mcp.Description("每次查询任务结果的最大条数，默认20")),
	), QueryAssetHandler)

	// Start MCP server
	serverAddr := fmt.Sprintf("%s:%d", conf.GlobalServerConfig().MCPServer.Host, conf.GlobalServerConfig().MCPServer.Port)
	logging.CLILog.Infof("start MCP Server：%s", serverAddr)
	logging.RuntimeLog.Infof("start MCP Server：%s", serverAddr)
	if conf.GlobalServerConfig().MCPServer.TLSEnable {
		if err := core.StartMCPServerWithTLS(s, serverAddr, conf.GlobalServerConfig().MCPServer.TLSCert, conf.GlobalServerConfig().MCPServer.TLSKey); err != nil {
			logging.CLILog.Error(err)
			return
		}
	} else {
		if err := core.StartMCPServer(s, serverAddr); err != nil {
			logging.CLILog.Error(err)
			return
		}
	}
}

func HelloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if ok, _ := core.CheckAuth(ctx); !ok {
		return core.FailResponse("认证失败")
	}
	if fileContent, err := os.ReadFile(filepath.Join(conf.GetRootPath(), "version.txt")); err == nil {
		return core.SuccessResponse(fmt.Sprintf("这是Nemo MCP Server，Nemo Version：%s!", strings.TrimSpace(string(fileContent))))
	}
	return core.SuccessResponse("你好，Hello!")
}

func PortscanHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if ok, _ := core.CheckAuth(ctx); !ok {
		return core.FailResponse("认证失败")
	}
	var err error
	// 处理参数
	var ip, port, bin, workspaceId string
	var rate int
	var pocscan bool
	if ip, err = core.GetRequiredArgumentString(&request, "ip"); err != nil {
		return core.FailResponseError("请指定目标IP", err)
	}
	port = core.GetDefaultArgumentString(&request, "port", "--top-ports 1000")
	rate = core.GetDefaultArgumentInt(&request, "rate", 1000)
	bin = core.GetDefaultArgumentString(&request, "bin", "nmap")
	pocscan = core.GetDefaultArgumentBool(&request, "pocscan", false)
	if workspaceId, err = core.GetWorkspaceIdFromContext(ctx); err != nil {
		return core.FailResponseError("获取工作空间失败", err)
	}
	// 任务参数
	config := execute.ExecutorConfig{
		PortScan: map[string]execute.PortscanConfig{
			bin: {
				Port:               port,
				Rate:               rate,
				Target:             ip,
				Tech:               "-sS",
				MaxOpenedPortPerIp: 50,
			}},
		FingerPrint: map[string]execute.FingerprintConfig{
			"fingerprint": {
				IsHttpx:        true,
				IsFingerprintx: true,
				IsScreenshot:   true,
				IsIconHash:     true,
			}},
	}
	if pocscan {
		config.PocScan = map[string]execute.PocscanConfig{
			"nuclei": {
				PocType: "matchFinger",
			},
		}
	}
	// 创建任务
	taskId := uuid.New().String()
	taskName := "MCP-端口扫描-指纹识别"
	if pocscan {
		taskName = "MCP-端口扫描-指纹识别-POC扫描-指纹匹配"
	}
	if err = newMainTask(config, workspaceId, ip, taskId, taskName, "来自MCP的端口扫描任务"); err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("创建任务失败，请稍后重试", err)
	}
	logging.RuntimeLog.Infof("MCP 创建任务成功，任务ID：%s", taskId)
	return core.SuccessResponse(fmt.Sprintf("创建任务成功，任务ID：%s，请稍后查询任务状态和结果", taskId))
}

func DomainscanHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if ok, _ := core.CheckAuth(ctx); !ok {
		return core.FailResponse("认证失败")
	}
	// 处理参数
	var err error
	var domain, workspaceId string
	var pocscan bool
	if domain, err = core.GetRequiredArgumentString(&request, "domain"); err != nil {
		return core.FailResponseError("请指定目标域名", err)
	}
	pocscan = core.GetDefaultArgumentBool(&request, "pocscan", false)
	if workspaceId, err = core.GetWorkspaceIdFromContext(ctx); err != nil {
		return core.FailResponseError("获取工作空间失败", err)
	}
	// 任务参数
	config := execute.ExecutorConfig{
		DomainScan: map[string]execute.DomainscanConfig{
			"subfinder": {
				IsIgnoreCDN:            true,
				IsIgnoreChinaOther:     true,
				IsIgnoreOutsideChina:   true,
				MaxResolvedDomainPerIP: 100,
			},
			"massdns": {
				WordlistFile:           "normal",
				IsIgnoreCDN:            true,
				IsIgnoreChinaOther:     true,
				IsIgnoreOutsideChina:   true,
				MaxResolvedDomainPerIP: 100,
			},
		},
		FingerPrint: map[string]execute.FingerprintConfig{
			"fingerprint": {
				IsHttpx:        true,
				IsFingerprintx: true,
				IsScreenshot:   true,
				IsIconHash:     true,
			}},
	}
	if pocscan {
		config.PocScan = map[string]execute.PocscanConfig{
			"nuclei": {
				PocType: "matchFinger",
			},
		}
	}
	// 创建任务
	taskId := uuid.New().String()
	taskName := "MCP-域名扫描-指纹识别"
	if pocscan {
		taskName = "MCP-域名扫描-指纹识别-POC扫描-指纹匹配"
	}
	if err = newMainTask(config, workspaceId, domain, taskId, taskName, "来自MCP的域名扫描任务"); err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("创建任务失败，请稍后重试", err)
	}
	logging.RuntimeLog.Infof("MCP 创建任务成功，任务ID：%s", taskId)
	return core.SuccessResponse(fmt.Sprintf("创建任务成功，任务ID：%s，请稍后查询任务状态和结果", taskId))
}

func OnlineAPIHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if ok, _ := core.CheckAuth(ctx); !ok {
		return core.FailResponse("认证失败")
	}
	// 处理参数
	var err error
	var target, api, workspaceId string
	var pocscan bool
	if target, err = core.GetRequiredArgumentString(&request, "target"); err != nil {
		return core.FailResponseError("请指定目标，可以是IP或者域名", err)
	}
	api = core.GetDefaultArgumentString(&request, "api", "fofa")
	pocscan = core.GetDefaultArgumentBool(&request, "pocscan", false)
	if workspaceId, err = core.GetWorkspaceIdFromContext(ctx); err != nil {
		return core.FailResponseError("获取工作空间失败", err)
	}
	// 任务参数
	config := execute.ExecutorConfig{
		OnlineAPI: map[string]execute.OnlineAPIConfig{
			api: {
				IsIgnoreCDN:            true,
				IsIgnoreChinaOther:     true,
				IsIgnoreOutsideChina:   true,
				SearchLimitCount:       1000,
				SearchPageSize:         100,
				MaxOpenedPortPerIp:     50,
				MaxResolvedDomainPerIP: 100,
			},
		},
		FingerPrint: map[string]execute.FingerprintConfig{
			"fingerprint": {
				IsHttpx:        true,
				IsFingerprintx: true,
				IsScreenshot:   true,
				IsIconHash:     true,
			}},
	}
	if pocscan {
		config.PocScan = map[string]execute.PocscanConfig{
			"nuclei": {
				PocType: "matchFinger",
			},
		}
	}
	// 创建任务
	taskId := uuid.New().String()
	taskName := "MCP-在线API调用"
	if pocscan {
		taskName = "MCP-在线API调用-POC扫描-指纹匹配"
	}
	if err = newMainTask(config, workspaceId, target, taskId, taskName, "来自MCP的在线API调用任务"); err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("创建任务失败，请稍后重试", err)
	}
	logging.RuntimeLog.Infof("MCP 创建任务成功，任务ID：%s", taskId)
	return core.SuccessResponse(fmt.Sprintf("创建任务成功，任务ID：%s，请稍后查询任务状态和结果", taskId))
}

func QueryTaskHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if ok, _ := core.CheckAuth(ctx); !ok {
		return core.FailResponse("认证失败")
	}
	// 处理参数
	var err error
	var taskId string
	if taskId, err = core.GetRequiredArgumentString(&request, "task_id"); err != nil {
		return core.FailResponseError("请指定任务ID", err)
	}
	// 查询任务
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponse("查询任务失败，请稍后重试")
	}
	defer db.CloseClient(mongoClient)
	doc, err := db.NewMainTask(mongoClient).GetByTaskId(taskId)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("查询任务失败，请稍后重试", err)
	}
	if doc.TaskId != taskId {
		logging.RuntimeLog.Error("任务不存在")
		return core.FailResponse("任务不存在")
	}
	var status, progress string
	switch doc.Status {
	case core.CREATED:
		status = "已创建"
	case core.STARTED:
		status = "执行中"
	case core.FAILURE:
		status = "执行失败"
	case core.SUCCESS:
		status = fmt.Sprintf("执行成功，任务结果：%s，请查看详细结果的资产", doc.Result)
	}
	progress = fmt.Sprintf("%d%%[%s]", int(doc.ProgressRate*100), doc.Progress)

	return core.SuccessResponse(fmt.Sprintf("任务状态：%s，进度：%s", status, progress))
}

func QueryResultHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if ok, _ := core.CheckAuth(ctx); !ok {
		return core.FailResponse("认证失败")
	}
	var err error
	// 处理参数
	var taskId, workspaceId string
	var page, pageSize int
	if taskId, err = core.GetRequiredArgumentString(&request, "task_id"); err != nil {
		return core.FailResponseError("请指定任务ID", err)
	}
	pageSize = core.GetDefaultArgumentInt(&request, "page_size", 20)
	if pageSize < 1 || pageSize > 100 {
		return core.FailResponse("分页大小必须在1-100之间")
	}
	page = core.GetDefaultArgumentInt(&request, "page", 1)
	if page < 1 {
		return core.FailResponse("分页页码必须从1开始")
	}
	if workspaceId, err = core.GetWorkspaceIdFromContext(ctx); err != nil {
		return core.FailResponseError("获取工作空间失败", err)
	}
	// 查询结果
	resp, err := getTaskResultData(workspaceId, taskId, bson.M{"taskId": taskId}, pageSize, page)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("查询结果失败，请稍后重试", err)
	}
	resultJson, err := json.Marshal(resp)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("查询结果失败，请稍后重试", err)
	}
	return mcp.NewToolResultText(string(resultJson)), nil
}

func QueryAssetHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if ok, _ := core.CheckAuth(ctx); !ok {
		return core.FailResponse("认证失败")
	}
	// 处理参数
	var err error
	var host, ip, port, title, fingerprint, workspaceId string
	var page, pageSize int
	host = core.GetDefaultArgumentString(&request, "host", "")
	ip = core.GetDefaultArgumentString(&request, "ip", "")
	port = core.GetDefaultArgumentString(&request, "port", "")
	title = core.GetDefaultArgumentString(&request, "title", "")
	fingerprint = core.GetDefaultArgumentString(&request, "fingerprint", "")
	pageSize = core.GetDefaultArgumentInt(&request, "page_size", 20)
	if pageSize < 1 || pageSize > 100 {
		return core.FailResponse("分页大小必须在1-100之间")
	}
	page = core.GetDefaultArgumentInt(&request, "page", 1)
	if page < 1 {
		return core.FailResponse("分页页码必须从1开始")
	}
	if workspaceId, err = core.GetWorkspaceIdFromContext(ctx); err != nil {
		return core.FailResponseError("获取工作空间失败", err)
	}
	// 解析查询条件
	var queryFilter string
	var queryExprList []string
	if len(host) > 0 {
		queryExprList = append(queryExprList, fmt.Sprintf("host==\"%s\"", host))
	}
	if len(ip) > 0 {
		queryExprList = append(queryExprList, fmt.Sprintf("ip==\"%s\"", ip))
	}
	if len(port) > 0 {
		queryExprList = append(queryExprList, fmt.Sprintf("port==\"%s\"", port))
	}
	if len(title) > 0 {
		queryExprList = append(queryExprList, fmt.Sprintf("title==\"%s\"", title))
	}
	if len(fingerprint) > 0 {
		fp := fingerprint
		fingerExpr := fmt.Sprintf("(server==\"%s\" || service==\"%s\" || banner==\"%s\" || app==\"%s\" || cert==\"%s\" || header==\"%s\")", fp, fp, fp, fp, fp, fp)
		queryExprList = append(queryExprList, fingerExpr)
	}
	if len(queryExprList) > 0 {
		queryFilter = strings.Join(queryExprList, " && ")
	} else {
		return core.FailResponse("请指定指定一个查询条件")
	}
	filter, err := db.ParseQuery(queryFilter)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("查询条件解析失败", err)
	}
	// 查询结果
	resp, err := getTaskResultData(workspaceId, "", filter, pageSize, page)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("查询结果失败，请稍后重试", err)
	}
	resultJson, err := json.Marshal(resp)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return core.FailResponseError("查询结果失败，请稍后重试", err)
	}
	return mcp.NewToolResultText(string(resultJson)), nil
}

func getTaskResultData(workspaceId, taskId string, filter bson.M, pageSize, page int) (*ResponseAssetData, error) {
	var respData []AssetData
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		return nil, err
	}
	defer db.CloseClient(mongoClient)
	colName := db.GlobalAsset
	if len(taskId) > 0 {
		colName = db.TaskAsset
	}
	// 获取任务结果
	focusAsset := db.NewAsset(workspaceId, colName, taskId, mongoClient)
	results, err := focusAsset.Find(filter, page, pageSize, true, true)
	if err != nil {
		logging.RuntimeLog.Error(err)
		return nil, err
	}
	honeypot := core.NewHoneypot(workspaceId)
	vul := db.NewVul(workspaceId, db.GlobalVul, mongoClient)
	icpStringMap := make(map[string]string)
	icpCompanyMap := make(map[string]string)
	icp := db.NewICP(mongoClient)

	for _, result := range results {
		asset := AssetData{
			Authority:  result.Authority,
			Domain:     result.Domain,
			Host:       result.Host,
			Status:     result.HttpStatus,
			Service:    result.Service,
			Title:      result.Title,
			Header:     result.HttpHeader,
			Cert:       result.Cert,
			Banner:     result.Banner,
			App:        result.App,
			IsCDN:      result.IsCDN,
			IconHash:   result.IconHash,
			IsHoneypot: honeypot.IsHoneypot(result.Host),
			UpdateTime: core.FormatDateTime(result.UpdateTime),
		}
		// 处理IP和Location
		locationMap := make(map[string]struct{})
		if len(result.Ip.IpV4) > 0 {
			for _, ip := range result.Ip.IpV4 {
				asset.IP = append(asset.IP, ip.IPName)
				// 对IP进行honeypot检测，进一步确保host为域名的时候不漏过检测
				if honeypot.IsHoneypot(ip.IPName) {
					asset.IsHoneypot = true
				}
				if len(ip.Location) > 0 {
					locationMap[ip.Location] = struct{}{}
				}
			}
		}
		if len(result.Ip.IpV6) > 0 {
			for _, ip := range result.Ip.IpV6 {
				asset.IP = append(asset.IP, ip.IPName)
				// 对IP进行honeypot检测，进一步确保host为域名的时候不漏过检测
				if honeypot.IsHoneypot(ip.IPName) {
					asset.IsHoneypot = true
				}
				if len(ip.Location) > 0 {
					locationMap[ip.Location] = struct{}{}
				}
			}
		}
		if len(locationMap) > 0 {
			for location := range locationMap {
				asset.Location = append(asset.Location, location)
			}
		}
		if result.Port > 0 {
			asset.Port = strconv.Itoa(result.Port)
		}
		// 处理漏洞列表
		vulDocs, _ := vul.Find(bson.M{"authority": result.Authority}, 0, 0)
		for _, v := range vulDocs {
			asset.Vul = append(asset.Vul, v.PocFile)
		}
		//icp:
		if result.Category == db.CategoryDomain && len(result.Domain) > 0 {
			if _, ok := icpStringMap[result.Domain]; !ok {
				icpResult, errIcp := icp.GetByDomain(result.Domain)
				if errIcp == nil && icpResult != nil && icpResult.Domain == result.Domain {
					icpStringMap[result.Domain] = icpResult.ToJSONString()
					icpCompanyMap[result.Domain] = icpResult.UnitName
				}
			}
		}
		if _, ok := icpCompanyMap[result.Domain]; ok {
			asset.IcpCompany = icpCompanyMap[result.Domain]
		}
		if _, ok := icpStringMap[result.Domain]; ok {
			asset.Icp = icpStringMap[result.Domain]
		}

		////whois
		//if result.Category == db.CategoryDomain && len(result.Domain) > 0 {
		//	if _, ok := whoisMap[result.Domain]; !ok {
		//		whoisResult, errWhois := icp.GetByDomain(result.Domain, db.QueryWhois)
		//		if errWhois == nil {
		//			whoisMap[result.Domain] = whoisResult.Content
		//		}
		//	}
		//	if _, ok := whoisMap[result.Domain]; ok {
		//		asset.Whois = whoisMap[result.Domain]
		//	}
		respData = append(respData, asset)
	}
	resp := ResponseAssetData{
		Page:          page,
		PageSize:      pageSize,
		Total:         len(results),
		AssetDataList: respData,
	}
	return &resp, nil
}

// newMainTask 创建一个新的maintask
func newMainTask(config execute.ExecutorConfig, workspaceId string, target string, taskId string, taskName string, description string) (err error) {
	configJson, _ := json.Marshal(config)
	mongoClient, err := db.GetClient()
	if err != nil {
		return err
	}
	defer db.CloseClient(mongoClient)
	doc := db.MainTaskDocument{
		WorkspaceId: workspaceId,
		TaskId:      taskId,
		TaskName:    taskName,
		Description: description,
		Target:      target,
		Args:        string(configJson),
		Status:      core.CREATED,
	}
	isSuccess, err := db.NewMainTask(mongoClient).Insert(doc)
	if err != nil {
		return err
	}
	if !isSuccess {
		return fmt.Errorf("创建任务失败")
	}
	return nil
}

package onlineapi

import (
	"fmt"
	"strconv"

	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/custom"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
)

type OnlineSearchResult struct {
	Domain  string
	Host    string
	IP      string
	Port    string
	Title   string
	Country string
	City    string
	Server  string
	Banner  string
	Service string
	Cert    string
	Source  string
	App     []string
}

type QueryDataResult struct {
	Domain   string
	Category string
	Content  string
}

type fofaQueryResult struct {
	Results      [][]string `json:"results"`
	Size         int        `json:"size"`
	Page         int        `json:"page"`
	Mode         string     `json:"mode"`
	IsError      bool       `json:"error"`
	ErrorMessage string     `json:"errmsg"`
}

var (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.125 Safari/537.36"
)

func ParseResult(config execute.ExecutorTaskInfo, searchResult []OnlineSearchResult) (docs []db.AssetDocument) {
	var onlineAPIConfig execute.OnlineAPIConfig
	for _, apiConfig := range config.OnlineAPI {
		onlineAPIConfig = apiConfig
		break
	}
	// 获取IP位置信息（注意：这里不能加载自定义IP位置信息，因为自定义位置信息需要从数据库中获取，worker无法访问数据库）
	ip4l := custom.NewIPv4Location("")
	ip6l, _ := custom.NewIPv6Location()
	cdn := custom.NewCDNCheck()
	// 解析搜索结果，生成历史文档
	tldExacter := utils.NewTldExtract()
	for _, result := range searchResult {
		host := utils.ParseHost(result.Host)
		// 根据配置过滤
		if onlineAPIConfig.IsIgnoreOutsideChina || onlineAPIConfig.IsIgnoreChinaOther {
			var isIp bool
			var location string
			if utils.CheckIPV4(host) {
				isIp = true
				location = ip4l.FindPublicIP(host)
			} else if utils.CheckIPV6(host) {
				if ip6l != nil {
					isIp = true
					location = ip6l.Find(host)
				}
			}
			if isIp && len(location) > 0 {
				if onlineAPIConfig.IsIgnoreOutsideChina && utils.CheckIPLocationOutsideChina(location) {
					logging.RuntimeLog.Warningf("忽略中国境外ip:%s ->location: %s", host, location)
					continue
				}
				if onlineAPIConfig.IsIgnoreChinaOther && utils.CheckIPLocationInChinaOther(location) {
					logging.RuntimeLog.Warningf("忽略中国港澳台ip:%s ->location: %s", host, location)
					continue
				}
			}
		}
		isCDN, cdnName, CName := cdn.Check(host)
		if onlineAPIConfig.IsIgnoreCDN && isCDN {
			logging.RuntimeLog.Warningf("忽略CDN域名：domain:%s,CDNName:%s,CName:%s", host, cdnName, CName)
			continue
		}
		// 生成文档
		doc := db.AssetDocument{
			OrgId:   config.OrgId,
			TaskId:  config.MainTaskId,
			IsCDN:   isCDN,
			CName:   CName,
			Title:   result.Title,
			Server:  result.Server,
			Banner:  result.Banner,
			Service: result.Service,
			App:     result.App,
			Cert:    result.Cert,
		}
		var domain string
		if len(result.Domain) > 0 {
			domain = tldExacter.ExtractFLD(result.Domain)
		}
		if len(domain) > 0 {
			doc.Domain = domain
			doc.Category = db.CategoryDomain
		}
		doc.Host = host
		doc.Authority = doc.Host
		if len(result.Port) > 0 {
			doc.Authority = fmt.Sprintf("%s:%s", doc.Host, result.Port)
			port, err := strconv.Atoi(result.Port)
			if err == nil {
				doc.Port = port
			}
		}
		if utils.CheckIPV4(doc.Host) {
			doc.Category = db.CategoryIPv4
		} else if utils.CheckIPV6(doc.Host) {
			doc.Category = db.CategoryIPv6
		} else {
			doc.Category = db.CategoryDomain
		}
		if len(result.IP) > 0 {
			if utils.CheckIPV4(result.IP) {
				doc.Ip.IpV4 = append(doc.Ip.IpV4, db.IPV4{
					IPName: result.IP,
				})
			} else if utils.CheckIPV6(result.IP) {
				doc.Ip.IpV6 = append(doc.Ip.IpV6, db.IPV6{
					IPName: result.IP,
				})
			}
		}
		docs = append(docs, doc)
	}

	return
}

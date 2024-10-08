package controllers

import (
	"github.com/hanc00l/nemo_go/v2/pkg/db"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"strings"
)

type KeySearchController struct {
	BaseController
}
type keyWordInitRequestParam struct {
	AddOrgId        int    `form:"add_org_id"`
	AddKeyWord      string `form:"add_key_word"`
	AddSearchTime   string `form:"add_search_time"`
	AddExcludeWords string `form:"add_exclude_words"`
	AddCheckMod     string `form:"add_check_mod"`
	AddCount        int    `form:"add_count"`
	AddFOFA         bool   `form:"fofa"`
	AddHunter       bool   `form:"hunter"`
	AddQuake        bool   `form:"quake"`
}

type keySearchRequestParam struct {
	DatableRequestParam
	OrgId        int    `form:"org_id"`
	KeyWord      string `form:"key_word"`
	SearchTime   string `form:"search_time"`
	ExcludeWords string `form:"exclude_words"`
	CheckMod     string `form:"check_mod"`
}

type KeyWordList struct {
	Id             int    `json:"id"`
	Index          int    `json:"index"`
	OrgId          string `json:"org_id"`
	KeyWord        string `json:"key_word"`
	Engine         string `json:"engine"`
	SearchTime     string `json:"search_time"`
	ExcludeWords   string `json:"exclude_words"`
	CheckMod       string `json:"check_mod"`
	IsDelete       bool   `json:"is_delete"`
	Count          int    `json:"count"`
	CreateDatetime string `json:"create_datetime"`
	UpdateDatetime string `json:"update_datetime"`
	WorkspaceId    int    `json:"workspace"`
}

type KeyWordInfo struct {
	Id             int    `json:"id"`
	OrgId          int    `json:"org_id"`
	KeyWord        string `json:"key_word"`
	IsFofa         bool   `json:"fofa"`
	IsHunter       bool   `json:"hunter"`
	IsQuake        bool   `json:"quake"`
	SearchTime     string `json:"search_time"`
	ExcludeWords   string `json:"exclude_words"`
	CheckMod       string `json:"check_mod"`
	IsDelete       bool   `json:"is_delete"`
	Count          int    `json:"count"`
	CreateDatetime string `json:"create_datetime"`
	UpdateDatetime string `json:"update_datetime"`
}

func (c *KeySearchController) IndexAction() {
	c.Layout = "base.html"
	c.TplName = "key-word-list.html"
}

func (c *KeySearchController) IndexBlackAction() {
	c.Layout = "base.html"
	c.TplName = "task-cron-list.html"
}

// AddSaveAction 保存新增的记录
func (c *KeySearchController) AddSaveAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetCurrentWorkspace()
	if workspaceId <= 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	keyWordData := keyWordInitRequestParam{}
	err := c.ParseForm(&keyWordData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	kw := db.KeyWord{}
	kw.OrgId = keyWordData.AddOrgId
	kw.KeyWord = keyWordData.AddKeyWord
	kw.SearchTime = keyWordData.AddSearchTime
	kw.ExcludeWords = keyWordData.AddExcludeWords
	kw.CheckMod = keyWordData.AddCheckMod
	kw.Count = keyWordData.AddCount
	kw.WorkspaceId = workspaceId
	var engines []string
	if keyWordData.AddFOFA {
		engines = append(engines, "xfofa")
	}
	if keyWordData.AddHunter {
		engines = append(engines, "xhunter")
	}
	if keyWordData.AddQuake {
		engines = append(engines, "xquake")
	}
	kw.Engine = strings.Join(engines, ",")
	c.MakeStatusResponse(kw.Add())
}

// validateRequestParam 校验请求的参数
func (c *KeySearchController) validateRequestParam(req *keySearchRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// ListAction IP列表
func (c *KeySearchController) ListAction() {
	defer c.ServeJSON()

	req := keySearchRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)

	resp := c.getKeyWordListData(req)
	c.Data["json"] = resp
}

// GetAction 一个记录的详细情况
func (c *KeySearchController) GetAction() {
	defer c.ServeJSON()

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	kw := db.KeyWord{Id: id}
	if kw.Get() {
		kwi := KeyWordInfo{
			Id:           kw.Id,
			OrgId:        kw.OrgId,
			KeyWord:      kw.KeyWord,
			SearchTime:   kw.SearchTime,
			ExcludeWords: kw.ExcludeWords,
			CheckMod:     kw.CheckMod,
			Count:        kw.Count,
		}
		engines := strings.Split(kw.Engine, ",")
		for _, e := range engines {
			switch e {
			case "fofa", "xfofa":
				kwi.IsFofa = true
			case "hunter", "xhunter":
				kwi.IsHunter = true
			case "quake", "xquake":
				kwi.IsQuake = true
			}
		}
		c.Data["json"] = kwi
	} else {
		c.Data["json"] = KeyWordInfo{}
	}
}

// UpdateAction 更新记录
func (c *KeySearchController) UpdateAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	kwi := keyWordInitRequestParam{}
	err = c.ParseForm(&kwi)
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		c.FailedStatus(err.Error())
		return
	}
	kw := db.KeyWord{Id: id}
	updateMap := make(map[string]interface{})
	updateMap["org_id"] = kwi.AddOrgId
	updateMap["key_word"] = kwi.AddKeyWord
	updateMap["search_time"] = kwi.AddSearchTime
	updateMap["exclude_words"] = kwi.AddExcludeWords
	updateMap["check_mod"] = kwi.AddCheckMod
	updateMap["count"] = kwi.AddCount
	var engines []string
	if kwi.AddFOFA {
		engines = append(engines, "xfofa")
	}
	if kwi.AddHunter {
		engines = append(engines, "xhunter")
	}
	if kwi.AddQuake {
		engines = append(engines, "xquake")
	}
	updateMap["engine"] = strings.Join(engines, ",")
	c.MakeStatusResponse(kw.Update(updateMap))
}

// DeleteKeyWordAction 删除一个记录
func (c *KeySearchController) DeleteKeyWordAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	ip := db.KeyWord{Id: id}
	updateMap := make(map[string]interface{})
	updateMap["is_delete"] = true
	ip.Update(updateMap)
	c.MakeStatusResponse(true)
}

// getSearchMap 根据查询参数生成查询条件
func (c *KeySearchController) getSearchMap(req keySearchRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

	workspaceId := c.GetCurrentWorkspace()
	if workspaceId > 0 {
		searchMap["workspace_id"] = workspaceId
	}
	if req.KeyWord != "" {
		searchMap["key_word"] = req.KeyWord
	}
	if req.SearchTime != "" {
		searchMap["search_time"] = req.SearchTime
	}
	if req.ExcludeWords != "" {
		searchMap["exclude_words"] = req.ExcludeWords
	}
	if req.CheckMod != "" {
		searchMap["check_mod"] = req.CheckMod
	}
	if req.OrgId > 0 {
		searchMap["org_id"] = req.OrgId
	}
	return
}

// getKeyWordListData 获取列表数据
func (c *KeySearchController) getKeyWordListData(req keySearchRequestParam) (resp DataTableResponseData) {
	keyWords := db.KeyWord{}
	searchMap := c.getSearchMap(req)
	workspaceId := c.GetCurrentWorkspace()
	if workspaceId > 0 {
		searchMap["workspace_id"] = workspaceId
	}
	startPage := req.Start/req.Length + 1
	results, total := keyWords.Gets(searchMap, startPage, req.Length)
	for i, keyWordRow := range results {
		r := KeyWordList{}
		r.Id = keyWordRow.Id
		r.KeyWord = keyWordRow.KeyWord
		r.Engine = keyWordRow.Engine
		r.Index = req.Start + i + 1
		if keyWordRow.CheckMod == "title" {
			r.CheckMod = "title"
		} else if keyWordRow.CheckMod == "body" {
			r.CheckMod = "body"
		} else if keyWordRow.CheckMod == "self" {
			r.CheckMod = "自定义"
		} else {
			r.CheckMod = "未知错误"
		}
		r.ExcludeWords = keyWordRow.ExcludeWords
		r.SearchTime = keyWordRow.SearchTime
		r.WorkspaceId = keyWordRow.WorkspaceId
		orgDb := db.Organization{}
		orgDb.Id = keyWordRow.OrgId
		if orgDb.Get() {
			r.OrgId = orgDb.OrgName
		}
		r.Count = keyWordRow.Count
		r.UpdateDatetime = FormatDateTime(keyWordRow.UpdateDatetime)
		r.CreateDatetime = FormatDateTime(keyWordRow.CreateDatetime)

		resp.Data = append(resp.Data, r)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return resp
}

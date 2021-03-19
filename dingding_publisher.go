package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/nsqio/go-nsq"
)

// LogDataInfo log data structure
type LogDataInfo struct {
	MachineName  string   `json:"machineName"`
	GamePlatform string   `json:"gamePlatform"`
	NodeName     string   `json:"nodeName"`
	FileName     string   `json:"fileName"`
	Msg          string   `json:"message"`
	IsAtAll      bool     `json:"isAtAll"`
	AtMobiles    []string `json:"atMobiles"`
}

// AlarmDataInfo alarm data structure
type AlarmDataInfo struct {
	Msg       string   `json:"message"`
	IsAtAll   bool     `json:"isAtAll"`
	AtMobiles []string `json:"atMobeles"`
}

// DingDingReqMarkdown dingding req markdown schema structure
type DingDingReqMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DingDingReqText dingding req text schema structure
type DingDingReqText struct {
	Content string `json:"content"`
}

// DingDingReqAtInfo dingding req at structure
// for dingding's bug, ony IsAtAll can be used when schema is markdown
type DingDingReqAtInfo struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

// DingDingReqBodyInfo dingding req body structure
type DingDingReqBodyInfo struct {
	MsgType  string              `json:"msgtype"`
	Markdown DingDingReqMarkdown `json:"markdown"`
	Text     DingDingReqText     `json:"text"`
	At       DingDingReqAtInfo   `json:"at"`
}

// DingDingPublisher dingding publisher structure
type DingDingPublisher struct {
	client     *http.Client
	tokenIndex int
	schema     string
	filter     *MsgFilterConfig
	mutex      sync.RWMutex
}

// NewDingDingPublisher create dingding publisher
func NewDingDingPublisher(filter *MsgFilterConfig) (*DingDingPublisher, error) {
	schema := "text"
	if filter.Schema != "" {
		schema = filter.Schema
	}

	var err error
	publisher := &DingDingPublisher{
		filter:     filter,
		schema:     schema,
		tokenIndex: 0,
	}

	publisher.client = &http.Client{}

	return publisher, err
}

// generateMarkDownBody 生成markdown格式报警信息
func generateMarkDownBody(logData LogDataInfo) ([]byte, error) {
	machineStr := ""
	if logData.MachineName != "" {
		machineStr = fmt.Sprintf("机器名:**%s**\n\n", logData.MachineName)
	}
	reqBody := DingDingReqBodyInfo{
		MsgType: "markdown",
		Markdown: DingDingReqMarkdown{
			Title: fmt.Sprintf("%s\n", logData.Msg),
			Text: fmt.Sprintf("\n\n## %s渠道%s节点报错收集\n\n%s文件名:**%s**\n```lua\n%s\n```",
				logData.GamePlatform, logData.NodeName, machineStr, logData.FileName, logData.Msg),
		},
		At: DingDingReqAtInfo{
			AtMobiles: logData.AtMobiles,
			IsAtAll:   logData.IsAtAll,
		},
	}

	return json.Marshal(reqBody)
}

// generateTextBody generate text schema alarm msg
func generateTextBody(logData LogDataInfo) ([]byte, error) {
	machineStr := ""
	if logData.MachineName != "" {
		machineStr = logData.MachineName
	}
	reqBody := DingDingReqBodyInfo{
		MsgType: "text",
		Text: DingDingReqText{
			// Content: fmt.Sprintf("%s\n%s渠道%s节点报错收集\n%s\n%s\n", logData.Msg, logData.GamePlatform, logData.NodeName, machineStr, logData.FileName),
			Content: fmt.Sprintf("%s\n主题: %s(%s) 节点报错收集\n机器: %s\n文件: %s", logData.Msg, logData.GamePlatform,
				logData.NodeName, machineStr, logData.FileName),
		},
		At: DingDingReqAtInfo{
			AtMobiles: logData.AtMobiles,
			IsAtAll:   logData.IsAtAll,
		},
	}

	return json.Marshal(reqBody)
}

func generateAlarmTextBody(alarmData AlarmDataInfo) ([]byte, error) {
	reqBody := DingDingReqBodyInfo{
		MsgType: "text",
		Text: DingDingReqText{
			Content: alarmData.Msg,
		},
		At: DingDingReqAtInfo{
			AtMobiles: alarmData.AtMobiles,
			IsAtAll:   alarmData.IsAtAll,
		},
	}

	return json.Marshal(reqBody)
}

// generateAccessToken get access token by loop
func (publisher *DingDingPublisher) generateAccessToken() string {
	var accessToken string

	publisher.mutex.RLock()
	defer publisher.mutex.RUnlock()

	if len(publisher.filter.HTTPAccessTokens) == 0 {
		return accessToken
	}

	accessToken = publisher.filter.HTTPAccessTokens[publisher.tokenIndex]
	publisher.tokenIndex = publisher.tokenIndex + 1
	if publisher.tokenIndex == len(publisher.filter.HTTPAccessTokens) {
		publisher.tokenIndex = 0
	}

	return accessToken
}

func (publisher *DingDingPublisher) sendDingDingMsg(reqBodyJSON []byte, accessToken string) {
	publisher.mutex.RLock()
	req, err := http.NewRequest("POST", fmt.Sprintf("%s://%s?access_token=%s", publisher.filter.Protocol,
		publisher.filter.URL, accessToken), bytes.NewReader(reqBodyJSON))
	if err != nil {
		publisher.mutex.RUnlock()
		fmt.Printf("sendDingDingMsg fail:%v %s", err, string(reqBodyJSON))
		return
	}
	publisher.mutex.RUnlock()

	req.Header.Set("Content-Type", "application/json")
	resp, err := publisher.client.Do(req)
	if err != nil {
		fmt.Printf("sendDingDingMsg do fail:%v", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("sendDingDingMsg fail:%s\n", string(body))
		return
	}
}

// todo: 使用etcd读取配置
func (publisher *DingDingPublisher) filterMessage(machineName, gamePlatform, nodeName, fileName, msg string) {
	isIgnore := true

	publisher.mutex.RLock()
	defer publisher.mutex.RUnlock()

	// only special keys need alarm
	for _, value := range publisher.filter.FilterKeys {
		if strings.Contains(msg, value) {
			isIgnore = false
			break
		}
	}

	// some keys need ignore
	for _, value := range publisher.filter.IgnoreKeys {
		if strings.Contains(msg, value) {
			isIgnore = true
			break
		}
	}

	if isIgnore {
		return
	}

	accessToken := publisher.generateAccessToken()
	fmt.Printf("accessToken:%s %s\n", accessToken, machineName)
	if accessToken == "" {
		return
	}

	isAtAll := true
	// not at all people for some key
	for _, value := range publisher.filter.NotAtKeys {
		if strings.Contains(msg, value) {
			isAtAll = false
			break
		}
	}

	// don't at all when at mobiles exist
	if len(publisher.filter.AtMobiles) > 0 {
		isAtAll = false
	}

	logData := LogDataInfo{
		MachineName:  machineName,
		GamePlatform: gamePlatform,
		NodeName:     nodeName,
		FileName:     fileName,
		Msg:          msg,
		IsAtAll:      isAtAll,
		AtMobiles:    publisher.filter.AtMobiles,
	}

	var reqBodyJSON []byte
	var err error
	if publisher.schema == "text" {
		reqBodyJSON, err = generateTextBody(logData)
	} else {
		reqBodyJSON, err = generateMarkDownBody(logData)
	}
	if err != nil {
		fmt.Printf("filterMessage file:%v", err)
		return
	}

	go publisher.sendDingDingMsg(reqBodyJSON, accessToken)
}

func (publisher *DingDingPublisher) alarmMessage(msg string) {
	fmt.Printf("alarmMessage:%s\n", msg)
	isIgnore := true

	publisher.mutex.RLock()
	defer publisher.mutex.RUnlock()

	// only special keys need alarm
	for _, value := range publisher.filter.FilterKeys {
		if strings.Contains(msg, value) {
			isIgnore = false
			break
		}
	}

	// some keys need ignore
	for _, value := range publisher.filter.IgnoreKeys {
		if strings.Contains(msg, value) {
			isIgnore = false
			break
		}
	}

	if isIgnore {
		return
	}

	accessToken := publisher.generateAccessToken()
	fmt.Printf("accessToken:%s\n", accessToken)
	if accessToken == "" {
		return
	}

	isAtAll := true
	// not at all people for some key
	for _, value := range publisher.filter.NotAtKeys {
		if strings.Contains(msg, value) {
			isAtAll = false
			break
		}
	}

	if len(publisher.filter.AtMobiles) > 0 {
		isAtAll = false
	}

	alarmData := AlarmDataInfo{
		Msg:       msg,
		IsAtAll:   isAtAll,
		AtMobiles: publisher.filter.AtMobiles,
	}

	reqBodyJson, err := generateAlarmTextBody(alarmData)
	if err != nil {
		fmt.Printf("generateAlarmTextBody fail: %v %v", err, alarmData)
		return
	}

	go publisher.sendDingDingMsg(reqBodyJson, accessToken)
}

func (publisher *DingDingPublisher) handleMessage(m *nsq.Message) error {
	data := make(map[string]interface{})
	err := json.Unmarshal(m.Body, &data)
	if err != nil {
		// alarm text message if unmarshal fail
		message := string(m.Body)
		publisher.alarmMessage(message)
		return nil
	}

	if data["message"] == nil || data["log"] == nil {
		message := ""
		if data["message"] != nil {
			message = data["message"].(string)
		} else {
			message = string(m.Body)
		}
		publisher.alarmMessage(message)
		return err
	}

	machineName := ""
	if data["machineName"] != nil {
		machineName = data["machineName"].(string)
	}
	logData := data["log"].(map[string]interface{})
	fileData := logData["file"].(map[string]interface{})
	publisher.filterMessage(machineName, data["gamePlatform"].(string), data["nodeName"].(string),
		fileData["path"].(string), data["message"].(string))

	return err
}

func (publisher *DingDingPublisher) updateConfig(filter *MsgFilterConfig) {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()

	publisher.filter = filter
	// maybe there are fewer tokens
	publisher.tokenIndex = 0

	if filter.Schema != "" {
		publisher.schema = filter.Schema
	}
}

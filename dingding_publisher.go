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
	GamePlatform string `json:"gamePlatform"`
	NodeName     string `json:"nodeName"`
	FileName     string `json:"fileName"`
	Msg          string `json:"message"`
	IsAtAll      bool   `json:"isAtAll"`
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
	client       *http.Client
	accessTokens []string
	tokenIndex   int
	tokenMu      sync.Mutex
	schema       string
	filter       *MsgFilterConfig
}

// NewDingDingPublisher create dingding publisher
func NewDingDingPublisher(filter *MsgFilterConfig) (*DingDingPublisher, error) {
	var err error
	publisher := &DingDingPublisher{
		filter:     filter,
		schema:     "markdown",
		tokenIndex: 0,
	}

	publisher.client = &http.Client{}

	return publisher, err
}

// generateMarkDownBody 生成markdown格式报警信息
func generateMarkDownBody(logData LogDataInfo) ([]byte, error) {
	reqBody := DingDingReqBodyInfo{
		MsgType: "markdown",
		Markdown: DingDingReqMarkdown{
			Title: "报错信息",
			Text: fmt.Sprintf("\n\n## %s渠道%s节点报错收集\n\n文件名:**%s**\n\n```lua\n%s\n```",
				logData.GamePlatform, logData.NodeName, logData.FileName, logData.Msg),
		},
		At: DingDingReqAtInfo{
			AtMobiles: []string{},
			IsAtAll:   logData.IsAtAll,
		},
	}

	return json.Marshal(reqBody)
}

// generateTextBody generate text schema alarm msg
func generateTextBody(logData LogDataInfo) ([]byte, error) {
	reqBody := DingDingReqBodyInfo{
		MsgType: "text",
		Text: DingDingReqText{
			Content: fmt.Sprintf("%s渠道%s节点报错收集", logData.GamePlatform, logData.NodeName),
		},
		At: DingDingReqAtInfo{
			AtMobiles: []string{},
			IsAtAll:   logData.IsAtAll,
		},
	}

	return json.Marshal(reqBody)
}

// generateAccessToken get access token by loop
func (publisher *DingDingPublisher) generateAccessToken() string {
	var accessToken string
	if len(publisher.accessTokens) == 0 {
		return accessToken
	}

	publisher.tokenMu.Lock()
	defer publisher.tokenMu.Unlock()

	if len(publisher.accessTokens) == 0 {
		return accessToken
	}

	accessToken = publisher.accessTokens[publisher.tokenIndex]
	publisher.tokenIndex = publisher.tokenIndex + 1
	if publisher.tokenIndex == len(publisher.accessTokens) {
		publisher.tokenIndex = 0
	}

	return accessToken
}

func (publisher *DingDingPublisher) sendDingDingMsg(logData LogDataInfo, accessToken string) {
	var reqBodyJSON []byte
	var err error
	if publisher.schema == "text" {
		reqBodyJSON, err = generateTextBody(logData)
	} else {
		reqBodyJSON, err = generateMarkDownBody(logData)
	}
	if err != nil {
		fmt.Printf("generateMarkDownBody fail:%s", err)
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s://%s?access_token=%s", publisher.filter.Protocol,
		publisher.filter.URL, publisher.filter.HTTPAccessTokens), bytes.NewReader(reqBodyJSON))
	if err != nil {
		fmt.Printf("sendDingDingMsg fail:%v %s", err, string(reqBodyJSON))
		return
	}

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

	fmt.Println("sendDingDingMsg success", string(body))
}

// todo: 使用etcd读取配置
func (publisher *DingDingPublisher) filterMessage(gamePlatform, nodeName, fileName, msg string) {
	isIgnore := true

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

	logData := LogDataInfo{
		GamePlatform: gamePlatform,
		NodeName:     nodeName,
		FileName:     fileName,
		Msg:          msg,
		IsAtAll:      isAtAll,
	}
	go publisher.sendDingDingMsg(logData, accessToken)
}

func (publisher *DingDingPublisher) handleMessage(m *nsq.Message) error {
	data := make(map[string]interface{})
	err := json.Unmarshal(m.Body, &data)
	if err != nil {
		fmt.Println("handleMessage unmarshal fail:", err)
		return err
	}

	logData := data["log"].(map[string]interface{})
	fileData := logData["file"].(map[string]interface{})
	publisher.filterMessage(data["gamePlatform"].(string), data["nodeName"].(string), fileData["path"].(string),
		data["message"].(string))

	return err
}

func (publisher *DingDingPublisher) updateConfig(filter *MsgFilterConfig) {
	publisher.filter = filter
	fmt.Printf("nsqConsumer updateConfig: %v\n", filter)
}

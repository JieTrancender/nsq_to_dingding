package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/nsqio/go-nsq"
)

// LogDataInfo log data structure
type LogDataInfo struct {
	GamePlatform string `json:"gamePlatform"`
	NodeName     string `json:"nodeName"`
	FileName     string `json:"fileName"`
	Msg          string `json:"message"`
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

// DingDingReqBodyInfo dingding req body structure
type DingDingReqBodyInfo struct {
	MsgType  string              `json:"msgtype"`
	Markdown DingDingReqMarkdown `json:"markdown"`
	Text     DingDingReqText     `json:"text"`
}

// DingDingPublisher dingding publisher structure
type DingDingPublisher struct {
	client      *http.Client
	accessToken string
	url         string
	protocol    string
	schema      string
}

// NewDingDingPublisher create dingding publisher
func NewDingDingPublisher(protocol, url, accessToken string) (*DingDingPublisher, error) {
	var err error
	publisher := &DingDingPublisher{
		protocol:    protocol,
		url:         url,
		accessToken: accessToken,
		schema:      "text",
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
	}

	return json.Marshal(reqBody)
}

func (publisher *DingDingPublisher) sendDingDingMsg(logData LogDataInfo) {
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s://%s?access_token=%s", publisher.protocol, publisher.url,
		publisher.accessToken), bytes.NewReader(reqBodyJSON))
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
	// 报错调用栈、报警日志、死循环检测
	if !strings.Contains(msg, "stack traceback") &&
		!strings.Contains(msg, "alarm") &&
		!strings.Contains(msg, "maybe in an endless loop") {
		return
	}

	// 特定关键字不报错，例如聊天后台请求
	if strings.Contains(msg, "chatMsgFilter") {
		return
	}

	if publisher.accessToken == "" {
		return
	}

	logData := LogDataInfo{
		GamePlatform: gamePlatform,
		NodeName:     nodeName,
		FileName:     fileName,
		Msg:          msg,
	}
	go publisher.sendDingDingMsg(logData)
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

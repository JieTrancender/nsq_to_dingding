package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nsqio/go-nsq"
)

// DingDingPublisher dingding publisher structure
type DingDingPublisher struct {
	client      *http.Client
	accessToken string
}

// NewDingDingPublisher create dingding publisher
func NewDingDingPublisher(url, accessToken string) (*DingDingPublisher, error) {
	var err error
	publisher := &DingDingPublisher{
		accessToken: accessToken,
	}

	publisher.client = &http.Client{}

	return publisher, err
}

func (publisher *DingDingPublisher) handMessage(m *nsq.Message) error {
	data := make(map[string]interface{})
	err := json.Unmarshal(m.Body, &data)
	if err != nil {
		fmt.Println("handMessage unmarshal fail:", err)
		return err
	}

	logData := data["log"].(map[string]interface{})
	fileData := logData["file"].(map[string]interface{})
	publisher.filterMessage(data["gamePlatform"].(string), data["nodeName"].(string), fileData["path"].(string),
		data["message"].(string))

	return err
}

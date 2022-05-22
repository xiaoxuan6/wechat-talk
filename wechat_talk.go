package wechat_talk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"
)

const webhook = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="

const (
	msgTypeText     = "text"
	msgTypeMarkdown = "markdown"
	msgTypeNews     = "news"
	msgTypeImage    = "image"
	msgTypeFile     = "file"
)

type Robot struct {
	Key string `json:"key"`
}

func NewRobot(key string) *Robot {
	return &Robot{
		Key: key,
	}
}

type textMessage struct {
	MsgType string `json:"msgtype"`
	Text    text   `json:"text"`
}

type text struct {
	Content             string   `json:"content" describe:"文本内容，最长不超过2048个字节，必须是utf8编码"`
	MentionedList       []string `json:"mentioned_list" describe:"userid的列表，提醒群中的指定成员(@某个成员)，@all表示提醒所有人"`
	MentionedMobileList []string `json:"mentioned_mobile_list" describe:"手机号列表，提醒手机号对应的群成员(@某个成员)，@all表示提醒所有人"`
}

func (r *Robot) SendText(content string, mentionedList, mentionedMobileList []string) error {
	return r.send(&textMessage{
		MsgType: msgTypeText,
		Text: text{
			Content:             content,
			MentionedList:       mentionedList,
			MentionedMobileList: mentionedMobileList,
		},
	})
}

type markdownMessage struct {
	MsgType  string   `json:"msgtype"`
	Markdown markdown `json:"markdown"`
}

type markdown struct {
	Content string `json:"content" describe:"markdown内容，最长不超过4096个字节，必须是utf8编码"`
}

func (r *Robot) SendMarkdown(content string) error {
	return r.send(&markdownMessage{
		MsgType: msgTypeMarkdown,
		Markdown: markdown{
			Content: content,
		},
	})
}

type imageMessage struct {
	MsgType string `json:"msgtype"`
	Image   image  `json:"image"`
}

type image struct {
	Base64 string `json:"base64" describe:"图片内容的base64编码"`
	Md5    string `json:"md5" describe:"图片内容（base64编码前）的md5值"`
}

// SendImage 注：图片（base64编码前）最大不能超过2M，支持JPG,PNG格式
func (r *Robot) SendImage(base64, md5 string) error {
	return r.send(&imageMessage{
		MsgType: msgTypeImage,
		Image: image{
			Base64: base64,
			Md5:    md5,
		},
	})
}

type newsMessage struct {
	MsgType string `json:"msgtype"`
	News    news   `json:"news"`
}

type news struct {
	Articles []Articles `json:"articles" describe:"图文消息，一个图文消息支持1到8条图文"`
}

type Articles struct {
	Title       string `json:"title" describe:"标题，不超过128个字节，超过会自动截断"`
	Description string `json:"description" describe:"描述，不超过512个字节，超过会自动截断"`
	Url         string `json:"url" describe:"点击后跳转的链接。"`
	PicUrl      string `json:"picurl" describe:"图文消息的图片链接，支持JPG、PNG格式，较好的效果为大图 1068*455，小图150*150。"`
}

func (r *Robot) SendNews(articles []Articles) error {
	return r.send(&newsMessage{
		MsgType: msgTypeNews,
		News: news{
			Articles: articles,
		},
	})
}

type fileMessage struct {
	MsgType string `json:"msgtype"`
	File    file   `json:"file" describe:"文件id，通过下文的文件上传接口获取"`
}

type file struct {
	MediaId string `json:"media_id" describe:"文件id，通过下文的文件上传接口获取"`
}

func (r *Robot) SendFile(mediaId string) error {
	return r.send(&fileMessage{
		MsgType: msgTypeFile,
		File: file{
			MediaId: mediaId,
		},
	})
}

func (r *Robot) send(msg interface{}) (err error) {

	body, er := json.Marshal(msg)
	if er != nil {
		return errors.New("json 格式化错误")
	}

	uri := fmt.Sprintf("%s%s", webhook, r.Key)
	res, err := resty.New().R().
		SetHeader("Content-Type", "application/json;charset=utf-8").
		SetBody(string(body)).
		Post(uri)

	if err != nil {
		return err
	}

	var item = make(map[string]interface{})
	_ = json.Unmarshal(res.Body(), &item)

	if item["errcode"] == float64(0) {
		return nil
	}

	return errors.New(item["errmsg"].(string))
}

//注意client 本身是连接池，不要每次请求时创建client
var (
	HttpClient = &http.Client{
		Timeout: 3 * time.Second,
	}
)

// UploadFile 素材上传得到media_id，该media_id仅三天内有效
// media_id只能是对应上传文件的机器人可以使用
// 要求文件大小在5B~20M之间
func (r *Robot) UploadFile(filename string, file io.Reader) (string, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/upload_media?key=%s&type=file", r.Key)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	formFile, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(formFile, file)
	if err != nil {
		return "", err
	}

	err = writer.Close()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", writer.FormDataContentType())
	resp, err := HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(content, &data)
	if err != nil {
		return "", errors.New("json 解析数据失败")
	}

	if data["errcode"] != float64(0) {
		return "", errors.New(data["errmsg"].(string))
	}

	return data["media_id"].(string), nil
}

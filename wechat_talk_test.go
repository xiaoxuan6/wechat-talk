package wechat_talk_test

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	wechat_talk "github.com/xiaoxuan6/wechat-talk"
	"os"
	"strings"
	"testing"
)

const key = "8d734ce0-c4eb-40ea-aab3-acd71fbaf50d"

func TestSendText(t *testing.T) {
	err := wechat_talk.NewRobot(key).SendText("广州今日天气：29度，大部分多云，降雨概率：60%", []string{}, []string{})
	assert.Nil(t, err)
}

func TestSendMarkdown(t *testing.T) {
	err := wechat_talk.NewRobot(key).SendMarkdown("实时新增用户反馈<font color=\\\"warning\\\">132例</font>，请相关同事注意。\\n\n         >类型:<font color=\\\"comment\\\">用户反馈</font>\n         >普通用户反馈:<font color=\\\"comment\\\">117例</font>\n         >VIP用户反馈:<font color=\\\"comment\\\">15例</font>")
	assert.Nil(t, err)
}

func getImgVal() ([]byte, error) {
	filename := "./image.jpg"

	_, err := os.Stat(filename)
	if ok := os.IsNotExist(err); ok {
		return nil, errors.New("图片不存在")
	}

	b, errs := os.ReadFile(filename)

	if errs != nil {
		return nil, errors.New("读取图片内容失败")
	}

	return b, nil
}

func TestSendImage(t *testing.T) {
	imgStr, err := getImgVal()
	assert.Nil(t, err, "图片不存在或者读取内容失败")
	assert.NotEmpty(t, imgStr)

	h := md5.New()
	h.Write(imgStr)
	md5Str := fmt.Sprintf("%x", h.Sum(nil))
	base64Str := base64.StdEncoding.EncodeToString(imgStr)

	err = wechat_talk.NewRobot(key).SendImage(base64Str, md5Str)

	assert.Nil(t, err)
}

func TestSendNews(t *testing.T) {
	articles := make([]wechat_talk.Articles, 0)
	article := wechat_talk.Articles{
		Title:       "中秋节礼品领取",
		Description: "今年中秋节公司有豪礼相送",
		Url:         "www.qq.com",
		PicUrl:      "http://res.mail.qq.com/node/ww/wwopenmng/images/independent/doc/test_pic_msg1.png",
	}

	articles = append(articles, article)

	err := wechat_talk.NewRobot(key).SendNews(articles)
	assert.Nil(t, err)
}

func getMediaId() (string, error) {
	filename := "./image.jpg"

	b, _ := os.ReadFile(filename)

	file := bytes.NewBuffer(b)
	mediaId, err := wechat_talk.NewRobot(key).UploadFile(strings.Replace(filename, "./", "", -1), file)
	if err != nil {
		return "", err
	}

	return mediaId, nil
}

func TestSendFile(t *testing.T) {
	mediaId, err := getMediaId()
	assert.Nil(t, err)

	err = wechat_talk.NewRobot(key).SendFile(mediaId)
	assert.Nil(t, err)
}

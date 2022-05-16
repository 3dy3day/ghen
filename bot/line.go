package bot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/line/line-bot-sdk-go/linebot"
)

type lineBot struct {
	bot    *linebot.Client
	secret string
	token  string
}

func (l lineBot) isValidSignature(signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	hash := hmac.New(sha256.New, []byte(l.secret))
	hash.Write(body)
	return hmac.Equal(decoded, hash.Sum(nil))
}

func (l lineBot) parseRequest(r events.APIGatewayProxyRequest) ([]*linebot.Event, error) {
	if !l.isValidSignature(r.Headers["X-Line-Signature"], []byte(r.Body)) {
		return nil, linebot.ErrInvalidSignature
	}
	request := &struct {
		Events []*linebot.Event `json:"events"`
	}{}
	if err := json.Unmarshal([]byte(r.Body), request); err != nil {
		return nil, err
	}
	return request.Events, nil
}

func (l lineBot) Broadcast(str string) error {
	_, err := l.bot.BroadcastMessage(linebot.NewTextMessage(str)).Do()
	return err
}

func (l lineBot) Reply(str string, request events.APIGatewayProxyRequest) error {
	events, err := l.parseRequest(request)
	if err != nil {
		return err
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if _, err = l.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
					log.Print(err)
				}
			case *linebot.StickerMessage:
				replyMessage := fmt.Sprintf(
					"sticker id is %s, stickerResourceType is %s", message.StickerID, message.StickerResourceType)
				if _, err = l.bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
	return nil
}

func CreateLineBot(secret, token string) Bot {
	bot, err := linebot.New(secret, token)
	if err != nil {
		panic(err)
	}

	return lineBot{
		secret: secret,
		token:  token,
		bot:    bot,
	}
}

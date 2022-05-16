package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"ghen/bot"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type Request struct {
	AcquisitionTime string  `json:"acquisition_time"`
	Temperature     float64 `json:"temperature"`
	Humidity        float64 `json:"humidity"`
}

type Response struct {
	Result int `json:"result"`
}

var secret string
var token string
var folderId string
var googleSecret string

var layout = "2006-01-02"

func init() {
	secret = os.Getenv("LINE_SCECRET")
	token = os.Getenv("LINE_TOKEN")
	folderId = os.Getenv("LOG_FOLDER_ID")
	googleSecret = os.Getenv("GOOGLE_SECRET")
}

func logging(request Request) error {
	ctx := context.Background()

	decodedSecret, err := base64.StdEncoding.DecodeString(googleSecret)
	if err != nil {
		return err
	}

	config, err := google.JWTConfigFromJSON([]byte(decodedSecret), drive.DriveFileScope)
	if err != nil {
		return err
	}
	client := config.Client(ctx)

	srv, err := drive.New(client)
	if err != nil {
		return err
	}

	res, err := srv.Files.List().Q("'" + folderId + "' in parents").Do()
	if err != nil {
		return err
	}

	var (
		content string
		fileId  string
		file    *drive.File
	)
	filename := "ghen_log" + time.Now().Format(layout) + ".csv"
	for _, item := range res.Files {
		if filename != item.Name {
			continue
		}
		file = item
		break
	}

	if file != nil {
		logResponse, err := srv.Files.Get(file.Id).Download()
		if err != nil {
			return err
		}
		defer logResponse.Body.Close()

		data, err := ioutil.ReadAll(logResponse.Body)
		if err != nil {
			return err
		}
		content = string(data)
		fileId = file.Id
	} else {
		content = "acquisition_time,Temperature,Humidity\n"
		file = &drive.File{
			Name:        filename,
			Description: "ghen log file",
			Parents: []string{
				folderId,
			},
		}
		file, err = srv.Files.Create(file).Media(strings.NewReader(content)).Do()
		if err != nil {
			return err
		}
		fileId = file.Id
		fmt.Println("create: " + filename)
	}

	content += request.AcquisitionTime + "," + strconv.FormatFloat(request.Temperature, 'f', 2, 64) + "," + strconv.FormatFloat(request.Humidity, 'f', 2, 64) + "\n"

	file.Id = ""
	_, err = srv.Files.Update(fileId, file).Media(strings.NewReader(content)).Do()
	if err != nil {
		return err
	}
	return nil
}

func proc(request Request) Response {
	err := logging(request)
	if err != nil {
		panic(err)
	}

	if request.Temperature >= 40.0 {
		sendMessage := "現在、温室の室温が40度を超えています\n"
		sendMessage += request.AcquisitionTime + " の温室情報\n"
		sendMessage += "温度：" + strconv.FormatFloat(request.Temperature, 'f', 2, 64) + "\n"
		sendMessage += "湿度：" + strconv.FormatFloat(request.Humidity, 'f', 2, 64)

		lineBot := bot.CreateLineBot(secret, token)
		lineBot.Broadcast(sendMessage)
	}

	response := Response{
		Result: 0,
	}
	return response
}

func handler(requestSource events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println(requestSource.Body)

	var request Request
	json.Unmarshal([]byte(requestSource.Body), &request)

	response := proc(request)
	bytes, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		Body:       string(bytes),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}

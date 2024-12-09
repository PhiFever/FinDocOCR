package main

import (
	"FinDocOCR/config"
	"FinDocOCR/doctype"
	"FinDocOCR/proc"
	"FinDocOCR/proc/invoice/vat"
	"FinDocOCR/proc/ticket/train"
	"FinDocOCR/utils"
	"bufio"
	"context"
	"fmt"
	"github.com/carlmjohnson/requests"
	_ "github.com/joho/godotenv/autoload"
	"os"
)

type AccessResponseBody struct {
	RefreshToken  string `json:"refresh_token"`
	ExpiresIn     int    `json:"expires_in"`
	SessionKey    string `json:"session_key"`
	AccessToken   string `json:"access_token"`
	Scope         string `json:"scope"`
	SessionSecret string `json:"session_secret"`
}

func GetBaiduAccessToken(clientId string, clientSecret string) (string, error) {
	requestUrl := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?"+
		"grant_type=client_credentials&client_id=%s&client_secret=%s", clientId, clientSecret)
	//fmt.Println(requestUrl)
	var accessResponseBody AccessResponseBody
	err := requests.
		URL(requestUrl).
		ToJSON(&accessResponseBody).
		Fetch(context.Background())
	if err != nil {
		return "", err
	}

	return accessResponseBody.AccessToken, nil
}

func main() {
	BaiduClientId := os.Getenv("BAIDU_CLIENT_ID")
	BaiduClientSecret := os.Getenv("BAIDU_CLIENT_SECRET")
	docDir := os.Getenv("DOC_DIR")

	logger := config.GetLogger()
	if BaiduClientId == "" || BaiduClientSecret == "" {
		logger.Fatalln("BAIDU_CLIENT_ID or BAIDU_CLIENT_SECRET is not set")
	}

	accessToken, err := GetBaiduAccessToken(BaiduClientId, BaiduClientSecret)
	if err != nil {
		logger.Error(err)
	}
	logger.Info("Access Token: ", accessToken)

	if docDir == "" {
		logger.Fatalln("DOC_DIR is not set")
	}

	docPaths, err := utils.FindSuffixesInDir(docDir, []string{".jpg", ".jpeg", ".png", ".pdf"})
	if len(docPaths) == 0 || err != nil {
		logger.Fatalln("No docs found in the directory")
	}
	docList := make([]doctype.Document, 0)
	for _, docPath := range docPaths {
		logger.Info("Processing doc: ", docPath)
		imageBytes, err := utils.ImageResize(docPath)
		if err != nil {
			logger.Error(err)
		}
		response, err := utils.GetMultipleInvoice(imageBytes, accessToken)
		if err != nil {
			logger.Error(err)
		}

		//logger.Debug(string(response))

		finDoc, err := proc.ProcessInvoice(response)
		if err != nil {
			logger.Error(err)
		}
		docList = append(docList, finDoc)
	}

	factory := &proc.DocumentFactory{}
	collections := make(map[doctype.DocumentType]doctype.DocumentCollection)

	// 将文档分类到对应的集合中
	for _, finDoc := range docList {
		var docType doctype.DocumentType
		var doc doctype.Document

		switch d := finDoc.(type) {
		case *vat.Doc:
			docType = doctype.TypeVatInvoice
			doc = d
		case *train.Doc:
			docType = doctype.TypeTrainTicket
			doc = d
		}

		if collection, exists := collections[docType]; exists {
			collection.Add(doc)
		} else {
			collection := factory.CreateCollection(docType)
			collection.Add(doc)
			collections[docType] = collection
		}
	}

	// 统一处理所有集合的保存
	for _, collection := range collections {
		if err := collection.SaveToFile(); err != nil {
			logger.Error(err)
		}
	}

	logger.Info("处理完成，按'Enter'以继续...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

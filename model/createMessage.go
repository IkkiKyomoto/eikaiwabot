package model

//ボットサーバーで行う処理を記述
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type (
	//OpenAI APIに送信するデータを格納する構造体
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	JsonReq struct {
		Model       string    `json:"model"`
		Messages    []Message `json:"messages"`
		Temperature float32   `json:"temperature"`
	}
	//OpenAI APIから受け取ったデータを格納する構造体
	JsonRes struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
)

// createMessage関数は、ユーザーからのメッセージを受け取り、それをOpenAI APIに送信し、返却されたAIアシスタントの回答を文字列型で返す。
func CreateMessage(message string) string {
	res, err := sendRequest(message)
	if err != nil {
		fmt.Println("エラーが発生しました:", err)
	}
	return res
}

// OpenAI APIにリクエストを送信
func sendRequest(message string) (string, error) {
	const (
		//chatGFTのエンドURL
		apiUrl      = "https://api.openai.com/v1/chat/completions"
		model       = "gpt-3.5-turbo-0613"
		method      = "POST"
		temperature = 0.7
		//エラー発生時にユーザー側に返されるメッセージ
		errStr = "エラーが発生しました"
	)

	//OpenAI APIキー
	apiKey := os.Getenv("OPENAI_API_KEY")

	//OpenAI APIに送信するデータを作成
	messages := []Message{
		{
			Role:    "system",
			Content: "You are an English teacher. You talk with user kindly",
		},
		{
			Role:    "user",
			Content: message,
		},
	}

	jreq := JsonReq{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
	}

	reqData, err := json.Marshal(jreq)
	if err != nil {
		fmt.Println("JSONのエンコードでエラーが発生しました。")
		return errStr, err
	}

	// HTTP POSTリクエストを作成
	req, err := http.NewRequest(method, apiUrl, bytes.NewReader(reqData))
	if err != nil {
		fmt.Println("リクエストの作成中にエラーが発生しました:", err)
		return errStr, err
	}

	// リクエストヘッダーを設定
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// HTTPクライアントを作成
	client := &http.Client{}

	// リクエストを送信してChatGPTからの応答を取得
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("リクエストの送信中にエラーが発生しました:", err)
		return errStr, err
	}
	defer resp.Body.Close()

	// ChatGPTからの応答を読み取り
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("応答の読み取り中にエラーが発生しました:", err)
		return errStr, err
	}
	// Jsonデータをデコード
	var jres JsonRes
	if err := json.Unmarshal(responseBody, &jres); err != nil {
		fmt.Println("Jsonデータのデコードでエラーが発生しました:", err)
		return errStr, err
	}
	// APIから正常にレスポンスが返されたか確認
	if jres.Error.Message != "" {
		err := errors.New("APIリクエストエラー")
		fmt.Println("OpenAI APIのリクエストが不正です:", err)
		return errStr, err
	}
	return jres.Choices[0].Message.Content, nil
}

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

type JsonRes struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func CreateMessage(message string) string {
	res, err := sendRequest(message)
	if err != nil {
		fmt.Println("エラーが発生しました:", err)
	}

	// ChatGPTからの応答を表示
	return res
}

func sendRequest(message string) (string, error) {
	// ChatGPTのエンドポイントURL
	apiUrl := "https://api.openai.com/v1/chat/completions"

	// ChatGPT APIキー
	apiKey := os.Getenv("OPENAI_API_KEY")

	// メッセージと送信するデータを作成

	requestData := fmt.Sprintf(`{"model": "gpt-3.5-turbo-0613", "messages": [{"role": "system", "content": "You are an English teacher. You talk with user kindly"}, {"role": "user", "content": "%s"}], "temperature": 0.7}`, message)

	// HTTP POSTリクエストを作成
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBufferString(requestData))
	if err != nil {
		fmt.Println("リクエストの作成中にエラーが発生しました:", err)
		return "エラーが発生しました", err
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
		return "エラーが発生しました", err
	}
	defer resp.Body.Close()

	// ChatGPTからの応答を読み取り
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("応答の読み取り中にエラーが発生しました:", err)
		return "エラーが発生しました", err
	}
	fmt.Println(string(responseBody))
	var jres JsonRes
	if err := json.Unmarshal(responseBody, &jres); err != nil {
		fmt.Println("Jsonデータのデコードでエラーが発生しました:", err)
		return "エラーが発生しました", err
	}
	if jres.Error.Message != "" {
		err := errors.New("APIリクエストエラー")
		fmt.Println("OpenAI APIのリクエストが不正です:", err)
		return "エラーが発生しました", err
	}
	fmt.Println(jres)
	return jres.Choices[0].Message.Content, nil
}

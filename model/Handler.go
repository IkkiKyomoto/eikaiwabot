package model

//ボットサーバーで行う処理を記述
import (
	"bytes"
	"database/sql"
	"eikaiwabot/database"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

const (
	//エラー発生時にユーザー側に返されるメッセージ
	errStr = "エラーが発生しました"
)

var (
	//環境変数の読み込み
	SQL_SOURCE     = os.Getenv("SQL_SOURCE")
	SQL            = os.Getenv("SQL")
	OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")
	MODEL          = os.Getenv("MODEL")
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
func Handler(message string, userID string) string {

	db, err := sql.Open(SQL, SQL_SOURCE)
	if err != nil {
		fmt.Println("DBとのコネクションでエラーが発生しました:", err)
		return errStr
	}
	defer db.Close()

	// 受け取ったメッセージをDBに追加
	latestAct := database.Activity{
		Role:    "user",
		UserID:  userID,
		Message: message,
	}
	err = database.InsertRow(db, latestAct)
	if err != nil {
		fmt.Println("ユーザーから受け取ったメッセージの挿入でエラーが発生しました:", err)
		return errStr
	}

	//　過去のメッセージを受け取る
	acts, err := database.GetRows(db, latestAct.UserID)
	if err != nil {
		fmt.Println("アクティビティの取得でエラーが発生しました:", err)
		return errStr
	}

	//メッセージを作成
	messages := []Message{
		{
			Role:    "system",
			Content: "You are an English teacher. You talk with user kindly",
		},
	}

	for _, a := range acts {
		messages = append(messages, Message{Role: a.Role, Content: a.Message})
	}

	//OpenAI APIにデータを送信
	resStr, err := sendRequest(messages[:])
	if err != nil {
		fmt.Println("リクエストの処理でエラーが発生しました:", err)
		return errStr
	}

	err = database.InsertRow(db, database.Activity{UserID: latestAct.UserID, Role: "assistant", Message: resStr})
	if err != nil {
		fmt.Println("ユーザーから受け取ったメッセージの挿入でエラーが発生しました:", err)
		return errStr
	}

	return resStr
}

// OpenAI APIにリクエストを送信
func sendRequest(messages []Message) (string, error) {
	const (
		//chatGPTのエンドURL
		apiUrl      = "https://api.openai.com/v1/chat/completions"
		method      = "POST"
		temperature = 0.7
	)

	jreq := JsonReq{
		Model:       MODEL,
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", OPENAI_API_KEY))

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
	// JSONデータをデコード
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

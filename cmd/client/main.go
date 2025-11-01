package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

const (
	SocketPath = "/tmp/minidb.sock"
)

// Message はリクエスト/レスポンスのスキーマ (サーバーと共通)
type Message struct {
	Operation string `json:"op"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`
	Result    string `json:"result,omitempty"`
	Status    string `json:"status"`
}

func main() {
	// 1. ミニストレージサーバーへ接続
	conn, err := net.Dial("unix", SocketPath)
	if err != nil {
		log.Fatalf("サーバーに接続できませんでした: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	log.Println("✅ ミニストレージサーバーに接続しました。")

	// --- 1. SET操作 ---
	key := "user_session_token"
	value := "{\"token\":\"abc123xyz\",\"expires\":\"2024-12-31T23:59:59Z\"}"

	log.Println("\n--- 1. SET操作: データを保存 ---")

	if e := sendAndReceive(conn, reader, Message{Operation: "SET", Key: key, Value: value}); e != nil {
		log.Fatalf("SET操作に失敗しました: %v", e)
	}

	// --- 2. GET操作 ---
	log.Println("\n--- 2. GET操作: データを取得 ---")

	if e := sendAndReceive(conn, reader, Message{Operation: "GET", Key: key}); e != nil {
		log.Fatalf("GET操作に失敗しました: %v", e)
	}

	// --- 3. 存在しないGET操作 ---
	log.Println("\n--- 3. 存在しないKEYのGET ---")

	if e := sendAndReceive(conn, reader, Message{Operation: "GET", Key: "non_existent_key"}); e != nil {
		log.Fatalf("存在しないKEYのGET操作に失敗しました: %v", e)
	}

	time.Sleep(1 * time.Second)
}

func sendAndReceive(conn net.Conn, reader *bufio.Reader, req Message) error {
	// リクエストをJSONに変換し、改行を追加
	reqData, _ := json.Marshal(req)
	conn.Write(append(reqData, '\n'))

	// 応答を読み込む
	resData, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("応答読み込み失敗: %v", err)
	}

	var res Message
	if e := json.Unmarshal([]byte(resData), &res); e != nil {
		return fmt.Errorf("[エラー] 応答のJSON解析失敗: %v", e)
	}

	log.Printf(">>> 応答: ステータス=%s, 結果/メッセージ=%s\n", res.Status, res.Result)

	return nil
}

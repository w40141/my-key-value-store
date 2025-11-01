package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	// SocketPath はUnixドメインソケットのファイルパス
	SocketPath = "/tmp/minidb.sock"
)

// Message はリクエスト/レスポンスのスキーマ
type Message struct {
	Operation string `json:"op"`
	Database  string `json:"db"`
	Table     string `json:"table"`
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"`  // SETの場合のみ使用
	Result    string `json:"result,omitempty"` // 応答の場合、取得した値やステータス
	Status    string `json:"status"`           // "OK" または "ERROR"
}

// インメモリのキーバリューストア
var (
	// 例: store["app_db:users:1"] = "{...JSONデータ...}"
	store = make(map[string]string)
	// マップを並行処理から保護するためのMutex
	storeMutex sync.RWMutex
)

func main() {
	// 既存のソケットファイルがあれば削除（エラー対策）
	if _, err := os.Stat(SocketPath); err == nil {
		os.Remove(SocketPath)
	}

	log.Println("ミニストレージサーバーを起動中...")

	// 1. Unixドメインソケットで待ち受ける
	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		log.Fatalf("リスニングエラー: %v", err)
	}
	defer listener.Close()

	log.Printf("Unixソケット %s で待ち受け中...\n", SocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("接続の受け入れに失敗しました:", err.Error())

			continue
		}
		// 接続を処理
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// クライアントからのメッセージ（JSON + 改行）を読み込む
		// 実際はメッセージ長を先に送るプロトコル設計がより堅牢
		data, err := reader.ReadString('\n')
		if err != nil {
			// 接続切断または読み込みエラー
			log.Println(">>> クライアント切断またはエラー:", err)

			return
		}

		var req Message
		// JSONデコード
		if err := json.Unmarshal([]byte(data), &req); err != nil {
			sendResponse(conn, Message{Status: "ERROR", Result: "Invalid JSON"})

			continue
		}

		log.Printf("<<< 受信: OP=%s, KEY=%s\n", req.Operation, req.Key)

		// 3. 操作に基づいてデータを処理
		res := processOperation(req)

		// 4. 応答をクライアントに返す
		sendResponse(conn, res)
	}
}

func processOperation(req Message) Message {
	switch strings.ToUpper(req.Operation) {
	case "SET":
		storeMutex.Lock()
		store[req.Key] = req.Value
		storeMutex.Unlock()

		return Message{Status: "OK", Result: "Data set successfully"}
	case "GET":
		storeMutex.RLock()
		val, ok := store[req.Key]
		storeMutex.RUnlock()

		if ok {
			return Message{Status: "OK", Result: val}
		}

		return Message{Status: "ERROR", Result: "Key not found"}
	case "DELETE":
		storeMutex.Lock()
		delete(store, req.Key)
		storeMutex.Unlock()

		return Message{Status: "OK", Result: "Key deleted"}
	default:
		return Message{Status: "ERROR", Result: "Unknown operation"}
	}
}

// JSON応答を送信するヘルパー関数
func sendResponse(conn net.Conn, res Message) {
	jsonData, _ := json.Marshal(res)
	// 応答の最後に改行文字(\n)を追加して、クライアントがデータの区切りを認識できるようにする
	conn.Write(append(jsonData, '\n'))
}

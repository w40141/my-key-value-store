package main

import (
	"fmt"
	"net"
	"os"
)

const (
	Host       = "localhost"
	Port       = "8080"
	ServerType = "tcp"
)

func main() {
	fmt.Println("サーバーを起動中...")

	// 1. 指定されたホストとポートで待ち受ける
	listener, err := net.Listen(ServerType, Host+":"+Port)
	if err != nil {
		fmt.Println("エラーが発生しました:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("ポート %s でリスニング中...\n", Port)

	// 2. 永続的にクライアント接続を待ち受ける
	for {
		// 新しい接続を受け入れる
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("接続の受け入れに失敗しました:", err.Error())
			continue
		}
		// 新しい接続を処理するためのゴルーチンを起動
		go handleConnection(conn)
	}
}

// 接続を個別に処理する関数
func handleConnection(conn net.Conn) {
	// 関数が終了したら接続を閉じる
	defer conn.Close()

	fmt.Printf("新しいクライアントが接続しました: %s\n", conn.RemoteAddr().String())

	// クライアントからデータを受信するためのバッファ
	buffer := make([]byte, 1024)

	for {
		// データを読み込む
		n, err := conn.Read(buffer)
		// エラーチェック（接続終了など）
		if err != nil {
			// io.EOFはクライアントが正常に接続を閉じたことを示す
			// 通常のエラーの場合は、ログに出力してループを抜ける
			// fmt.Println("Readエラーまたは接続終了:", err.Error())
			break
		}

		// 受信したデータを文字列に変換して出力
		receivedData := string(buffer[:n])
		fmt.Printf("クライアント %s からのメッセージ: %s\n", conn.RemoteAddr().String(), receivedData)

		// （オプション）クライアントに応答を返す
		// _, err = conn.Write([]byte("サーバーがメッセージを受け取りました。\n"))
	}

	fmt.Printf("クライアント %s との接続を閉じました。\n", conn.RemoteAddr().String())
}

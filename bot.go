package stegobot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var apiURL = "https://api.ottotto.dev/api/v1"

type Bot struct {
	ID            string
	Name          string
	FirebaseToken string
	Life          int
}

func New(ctx context.Context, cfg *Config) (*Bot, error) {
	token, err := SignUp(ctx, cfg.APIKey)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		FirebaseToken: token,
	}

	return bot, nil
}

func (b *Bot) Run() error {
	if err := b.join(); err != nil {
		return err
	}

	return nil
}

func (b *Bot) join() error {
	// GET Room ID
	roomID, err := b.getRoomID()
	if err != nil {
		return err
	}

	// GET One Time Password
	otp, err := b.getOTP()
	if err != nil {
		return err
	}

	b.ws(roomID, otp)

	// GET One Time Password
	return nil
}

func (b *Bot) getRoomID() (string, error) {
	req, err := http.NewRequest("GET", apiURL+"/rooms/matching", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+b.FirebaseToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var room struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &room); err != nil {
		return "", err
	}

	return room.ID, nil
}

func (b *Bot) getOTP() (string, error) {
	req, err := http.NewRequest("POST", apiURL+"/otp", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+b.FirebaseToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var otp struct {
		Otp string `json:"otp"`
	}
	if err := json.Unmarshal(body, &otp); err != nil {
		return "", err
	}

	return otp.Otp, nil
}

func (b *Bot) ws(roomID, otp string) {
	// WebSocketサーバーのURL
	url := "wss://api.ottotto.dev/api/v1/rooms/" + roomID + "?p=" + otp
	fmt.Println(url)

	// WebSocketに接続
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	ch := make(chan string)
	done := make(chan struct{})

	type Message struct {
		Type    string      `json:"type"`
		Payload interface{} `json:"payload"`
	}

	type NextSeq struct {
		Type    string `json:"type"`
		Payload struct {
			Value string `json:"value"`
			Type  string `json:"type"`
			Level int    `json:"level"`
		} `json:"payload"`
	}

	time.Sleep(15 * time.Second)

	go func() {
		for {
			// メッセージの読み込み
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var m Message
			if err := json.Unmarshal(message, &m); err != nil {
				log.Println("unmarshal:", err)
				return
			}

			switch m.Type {
			case "NextSeq":
				var nextSeq NextSeq
				if err := json.Unmarshal(message, &nextSeq); err != nil {
					log.Println("unmarshal:", err)
					return
				}

				ch <- nextSeq.Payload.Value
			case "Result":
				close(done)
			}
		}
	}()

	// サーバーにメッセージを送信
	for {
		select {
		case msg := <-ch:
			start := time.Now()

			for i := range msg {
				time.Sleep(350 * time.Millisecond)
				key := TypingKey{
					Type: "TypingKey",
					Payload: struct {
						InputSeq string `json:"inputSeq"`
					}{
						InputSeq: msg[:i+1],
					},
				}

				if err := c.WriteJSON(key); err != nil {
					log.Println("write:", err)
					return
				}
			}

			event := FinCurrentSeq{
				Type: "FinCurrentSeq",
				Payload: struct {
					Cause string `json:"cause"`
				}{
					Cause: "failed",
				},
			}

			if after := time.Since(start); after < 5*time.Second {
				event.Payload.Cause = "succeeded"
			}

			if err := c.WriteJSON(&event); err != nil {
				log.Println("write:", err)
				return
			}

			// メッセージの送信
		case <-done:
			return
		}
	}

}

type FinCurrentSeq struct {
	Type    string `json:"type"`
	Payload struct {
		Cause string `json:"cause"`
	} `json:"payload"`
}

type TypingKey struct {
	Type    string `json:"type"`
	Payload struct {
		InputSeq string `json:"inputSeq"`
	} `json:"payload"`
}

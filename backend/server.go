package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"

	"net/http"
	"time"

	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	_ "github.com/joho/godotenv/autoload"
)


var (
	ALLOWED_DOMAIN = os.Getenv("ALLOWED_DOMAIN")
	QUEUE_URL = os.Getenv("QUEUE_URL")

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == ALLOWED_DOMAIN
		},
	}
)

func handleMessage(msg types.Message, ws *websocket.Conn, c echo.Context) error {
	data, _ := hex.DecodeString(aws.ToString(msg.Body))

	payloadLength := int(data[1])

	// Skip processing if msgId in byte stream does not match `GLOBAL_POSITION_INT=33`
	if payloadLength <= 0 || !bytes.Equal(data[7:10], []byte{33, 0, 0}) {
		err := errors.New("messageID does not match supported type")
		c.Logger().Error(err)
		return err
	}

	payload := data[10: payloadLength+10]

	latSlice, lonSlice := payload[4:8], payload[8:12]

	// Convert bytes to uint32 first
	latValue := binary.LittleEndian.Uint32(latSlice)
	lonValue := binary.LittleEndian.Uint32(lonSlice)

	lat, lon, sysId := float32(int32(latValue))/1e7, float32(int32(lonValue))/1e7, int32(data[5])

	mavlinkData := map[string]interface{}{
		"sysid":     sysId,
		"timestamp": time.Now().Unix(),
		"lat":   lat,
		"lon":   lon,
	}

	// Write JSON
	err := ws.WriteJSON(mavlinkData)

	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			c.Logger().Error("unexpected close:", err)
		} else {
			c.Logger().Warn("WriteJSON returned an error but message likely delivered:", err)
		}
		return err
	}

	return nil
}

func hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Error(err)
		return err
	}
	defer ws.Close()

	ctx := context.Background()

	// Load default AWS credentials & region
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		c.Logger().Error("config error:", err)
		return err
	}

	client := sqs.NewFromConfig(cfg)

	for {
		// Sleep for 1.5 seconds

		// Fetch MAVLINK byte data from SQS
		// Parse data into appropriate structure, filter out parsing by ensuring only packets
		// with `Message ID = 33 (GLOBAL_POSITION_INT)` get processed since we are only interested in location data
		// for this demo.

		// Receive one message
		resp, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(QUEUE_URL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			c.Logger().Error("receive error:", err)
			continue
		}

		if len(resp.Messages) == 0 {
			c.Logger().Error("No messages available")
			continue
		}

		for _, msg := range resp.Messages {
			err = handleMessage(msg, ws, c)
			if err != nil {
				c.Logger().Error(err)
			}
		}

		entries := make([]types.DeleteMessageBatchRequestEntry, 0, len(resp.Messages))

		for i, m := range resp.Messages {
			entries = append(entries, types.DeleteMessageBatchRequestEntry{
				Id:            aws.String(fmt.Sprintf("msg-%d", i)),
				ReceiptHandle: m.ReceiptHandle,
			})
		}

		_, err = client.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
			QueueUrl: aws.String(QUEUE_URL),
			Entries:  entries,
		})
		if err != nil {
			// handle
			c.Logger().Error(err)
		}
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/ws", hello)
	e.Logger.Fatal(e.Start(":1323"))
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/isayme/go-logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type windowSize struct {
	Rows uint16 `json:"rows" msgpack:"rows"`
	Cols uint16 `json:"cols" msgpack:"cols"`
	X    uint16 `json:"x" msgpack:"x"`
	Y    uint16 `json:"y" msgpack:"y"`
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "public")

	e.GET("/ws", handleWebsocket)

	e.Logger.Fatal(e.Start(":1323"))
}

func handleWebsocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	shell := exec.Command("bash")
	// shell.Env = append(os.Environ(), "TERM=xterm")

	ptmx, err := pty.Start(shell)
	if err != nil {
		logger.Errorf("pty.Start error: %v", err)
		return err
	}

	defer func() { _ = ptmx.Close() }() // Best effort.
	// Copy stdin to the pty and the pty to stdout.
	// NOTE: The goroutine will keep reading until the next keystroke before returning.
	go func() {
		for {
			messageType, reader, err := ws.NextReader()
			if err != nil {
				logger.Errorf("Unable to grab next reader: %v", err)
				return
			}

			if messageType == websocket.TextMessage {
				warnMsg := fmt.Sprintf("Unexpected text message: %d", messageType)
				logger.Warn(warnMsg)
				ws.WriteMessage(websocket.TextMessage, []byte(warnMsg))
				continue
			}

			dataTypeBuf := make([]byte, 1)
			read, err := reader.Read(dataTypeBuf)
			if err != nil {
				errorMsg := fmt.Sprintf("Unable to read message type from reader: %v", err)
				logger.Error(errorMsg)
				ws.WriteMessage(websocket.TextMessage, []byte(errorMsg))
				return
			}

			if read != 1 {
				logger.Error("Unexpected number of bytes read")
				return
			}

			switch dataTypeBuf[0] {
			case 0:
				copied, err := io.Copy(ptmx, reader)
				if err != nil {
					logger.Errorf("Error after copying %d bytes, err: %v", copied, err)
				} else {
					logger.Infof("read %d bytes", copied)
				}
			case 1:
				decoder := json.NewDecoder(reader)
				resizeMessage := windowSize{}
				err := decoder.Decode(&resizeMessage)
				if err != nil {
					// ws.WriteMessage(websocket.TextMessage, []byte("Error decoding resize message: "+err.Error()))
					continue
				}
				logger.Infof("Resizing terminal: %v", resizeMessage)
				winSize := &pty.Winsize{
					Rows: resizeMessage.Rows,
					Cols: resizeMessage.Cols,
					X:    resizeMessage.X,
					Y:    resizeMessage.Y,
				}
				pty.Setsize(ptmx, winSize)
			default:
				logger.Errorf("Unknown data type: %d", dataTypeBuf[0])
			}
		}
	}()

	for {
		readBuf := make([]byte, 1024)
		n, err := ptmx.Read(readBuf)
		logger.Infof("ptmx.Read: n: %d, err: %v", n, err)
		if n >= 0 {
			for i := 0; i < n; i++ {
				logger.Infof("ptmx.Read: %02x", readBuf[i])
			}
			err := ws.WriteMessage(websocket.BinaryMessage, readBuf[:n])
			if err != nil {
				logger.Errorf("receive error: %v", err)
				break
			}
		}

		if err != nil {
			logger.Errorf("read ptmx error: %v", err)
			break
		}
	}

	return nil
}

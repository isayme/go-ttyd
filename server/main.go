package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/isayme/go-logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	TYPE_DATA   = 0
	TYPE_RESIZE = 1
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
	shell.Env = append(shell.Environ(), "TERM=xterm")

	ptmx, err := pty.Start(shell)
	if err != nil {
		logger.Errorf("pty.Start error: %v", err)
		return err
	}

	defer func() { ptmx.Close() }()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		dataTypeBuf := make([]byte, 1)

		for {
			messageType, reader, err := ws.NextReader()
			if err != nil {
				logger.Errorf("unable to grab next reader: %v", err)
				return
			}

			if messageType == websocket.TextMessage {
				warnMsg := fmt.Sprintf("unexpected message type: %d", messageType)
				logger.Warn(warnMsg)
				return
			}

			read, err := reader.Read(dataTypeBuf)
			if err != nil {
				errorMsg := fmt.Sprintf("unable to read message type from reader: %v", err)
				logger.Error(errorMsg)
				return
			}

			if read != 1 {
				logger.Error("read data type fail")
				return
			}

			dataType := dataTypeBuf[0]
			switch dataType {
			case TYPE_DATA:
				copied, err := io.Copy(ptmx, reader)
				if err != nil {
					logger.Errorf("Error after copying %d bytes, err: %v", copied, err)
				}
			case TYPE_RESIZE:
				resizeMsgBuf, err := io.ReadAll(reader)
				if err != nil {
					logger.Warnf("Error decoding resize message: %v", err)
					continue
				}

				resizeMessage := windowSize{}
				err = msgpack.Unmarshal(resizeMsgBuf, &resizeMessage)
				if err != nil {
					logger.Warnf("Error msgpack.Unmarshal: %v", err)
					continue
				}

				logger.Infof("Resizing terminal: %+v", resizeMessage)

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

	go func() {
		defer wg.Done()

		readBuf := make([]byte, 1024)

		for {
			n, err := ptmx.Read(readBuf)
			if n >= 0 {
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
	}()

	wg.Wait()

	return nil
}

package bahamut

import (
	"net/http"

	"github.com/aporeto-inc/elemental"
	"golang.org/x/net/websocket"

	log "github.com/Sirupsen/logrus"
)

func writeWebSocketError(ws *websocket.Conn, response *elemental.Response, err error) {

	if !ws.IsServerConn() {
		return
	}

	var outError elemental.Errors

	switch e := err.(type) {
	case elemental.Error:
		outError = elemental.NewErrors(e)
	case elemental.Errors:
		outError = e
	default:
		outError = elemental.NewErrors(elemental.NewError("Internal Server Error", e.Error(), "bahamut", http.StatusInternalServerError))
	}

	response.StatusCode = outError.Code()
	response.Encode(outError)

	if e := websocket.JSON.Send(ws, response); e != nil {
		log.WithFields(log.Fields{
			"package":       "bahamut",
			"error":         e.Error(),
			"originalError": err.Error(),
		}).Error("Unable to encode error.")
	}
}

func writeWebsocketResponse(ws *websocket.Conn, response *elemental.Response, c *Context) error {

	if !ws.IsServerConn() {
		return nil
	}

	if c.StatusCode == 0 {
		switch c.Operation {
		case elemental.OperationCreate:
			c.StatusCode = http.StatusCreated
		case elemental.OperationInfo:
			c.StatusCode = http.StatusNoContent
		default:
			c.StatusCode = http.StatusOK
		}
	}

	if c.Operation == elemental.OperationRetrieveMany || c.Operation == elemental.OperationInfo {

		response.Count = c.Count.Current
		response.Total = c.Count.Total
	}

	if c.OutputData != nil {

		if err := response.Encode(c.OutputData); err != nil {
			return err
		}
	}

	response.StatusCode = c.StatusCode

	return websocket.JSON.Send(ws, response)
}
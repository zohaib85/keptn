package websockethelper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/keptn/keptn/cli/utils/credentialmanager"
)

type myCloudEvent struct {
	CloudEventsVersion string          `json:"cloudEventsVersion"`
	ContentType        string          `json:"contentType"`
	Data               json.RawMessage `json:"data"`
	EventID            string          `json:"eventID"`
	EventTime          string          `json:"eventTime"`
	EventType          string          `json:"eventType"`
	Type               string          `json:"type"`
	Source             string          `json:"source"`
}

// LogData represents log data
type LogData struct {
	Message   string `json:"message"`
	Terminate bool   `json:"terminate"`
}

// OpenWS opens a websocket
func OpenWS(token string) (*websocket.Conn, *http.Response, error) {

	endPoint, _, err := credentialmanager.GetCreds()
	if err != nil {
		return nil, nil, err
	}

	wsEndPoint := endPoint
	wsEndPoint.Scheme = "ws"

	// origin := "http://localhost/"
	// u := url.URL{Scheme: wsEndPoint.Scheme, Host: *wsEndPoint, Path: "/"}
	header := http.Header{}
	header.Add("Token", token)
	return websocket.DefaultDialer.Dial(wsEndPoint.String(), header)
	// if err != nil {
	// 	log.Fatal("dial: ", err)
	// }

	// config, err := websocket. .NewConfig(wsEndPoint.String(), origin)
	// config.Header = http.Header{}
	// config.Header.Add("Token", token)
	// return websocket.DialConfig(config)
}

// readCE reads a cloud event
func readCE(ws *websocket.Conn) (interface{}, error) {

	ws.SetReadDeadline(time.Now().Add(time.Minute))
	var msg = make([]byte, 512)
	_, message, err := ws.ReadMessage() // .Read(msg)
	if err != nil {
		log.Println("read: ", err)
		return nil, err
	}
	log.Println("received: ", message)
	return getCloudEventData(msg)
}

func getCloudEventData(data []byte) (interface{}, error) {

	ce := myCloudEvent{}
	if err := json.Unmarshal(data, &ce); err != nil {
		fmt.Println("JSON unmarshalling error. Cloud events are expected")
		return nil, err
	}

	var dst interface{}
	eventType := ce.EventType
	if eventType == "" && ce.Type != "" {
		eventType = ce.Type
	}

	switch eventType {
	case "sh.keptn.events.log":
		dst = new(LogData)
	default:
		dst = new(interface{})
	}

	if err := json.Unmarshal(ce.Data, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}

// PrintWSContent prints received cloud events
func PrintWSContent(ws *websocket.Conn) error {

	for {
		ceData, err := readCE(ws)
		if err != nil || ceData == nil {
			return err
		}
		switch ceData.(type) {
		case *LogData:
			logData := ceData.(*LogData)
			handleLogCE(*logData)
			if logData.Terminate {
				return nil
			}
		default:
			fmt.Printf("Cloud event type unknown")
			return nil
		}
	}
}

func handleLogCE(log LogData) {
	fmt.Println(log.Message)
}

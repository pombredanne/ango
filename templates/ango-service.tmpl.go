package {{.PackageName}}

import (
	"errors"
	"fmt"
	"github.com/GeertJohan/websocket"
	"encoding/json"
	"net/http"
)

var protocolVersion = "{{.ProtocolVersion}}"

var (
	ErrInvalidVersionString = errors.New("invalid version string")
	ErrInvalidMessageType   = errors.New("invalid message type")
	ErrUnkonwnProcedure     = errors.New("unknown procedure")
)

const (
	msgTypeRequest  = "req"
	msgTypeResponse = "res"
)

// root structure for incoming message json
type angoInMsg struct {
	Type       string          `json:"type"`      // "req" or "res"
	Procedure  string          `json:"procedure"` // name for the procedure 9when "req"
	CallbackID uint64          `json:"cb_id"`     // callback ID for request or response
	Data       json.RawMessage `json:"data"`      // remain raw, depends on procedure
	Error      json.RawMessage `json:"error"`     // remain raw, depens on ??
}

// root structure for outgoing message json
type angoOutMsg struct {
	Type       string      `json:"type"`      // "req" or "res"
	Procedure  string      `json:"procedure"` // name for the procedure 9when "req"
	CallbackID uint64      `json:"cb_id"`     // callback ID for request or response
	Data       interface{} `json:"data"`      // remain raw, depends on procedure
	Error      interface{} `json:"error"`     // remain raw, depens on ??
}

{{range .Service.ServerProcedures}}
type angoInData{{.CapitalizedName}} struct {
	{{range .Args}}
		{{.CapitalizedName}} {{.Type.GoTypeName}} `json:"{{.Name}}"` {{end}}
}
{{end}}

// {{.Service.CapitalizedName}}SessionInterface types all methods that can be called by the client
type {{.Service.CapitalizedName}}SessionInterface interface {
	// Stop is called when the session is about to end (websocket closed)
	Stop(err error)

	{{range .Service.ServerProcedures}}
		// {{.CapitalizedName}} is a ango procedure defined in the .ango file
		{{.CapitalizedName}} ( {{.GoArgs}} )( {{.GoRets}} )
	{{end}}
}

// New{{.Service.CapitalizedName}}SessionInterface must return a new instance implementing {{.Service.CapitalizedName}}SessionInterface
type New{{.Service.CapitalizedName}}SessionInterface func()(handler {{.Service.CapitalizedName}}SessionInterface)

// {{.Service.CapitalizedName}}Server handles incomming http requests
type {{.Service.CapitalizedName}}Server struct {
	NewSession               New{{.Service.CapitalizedName}}SessionInterface //++ inline type?
	ErrorIncommingConnection func(err error)
}

//++ TODO: what to do with errors?
//++ add fields to Server? ErrorIncommingConnection(err error)
//++ when error occurs and non-nil: call the function with the error

// ServeHTTP hijacks incomming http connections and sets up the websocket communication
func (server *{{.Service.CapitalizedName}}Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		}
		if server.ErrorIncommingConnection != nil {
			server.ErrorIncommingConnection(err)
		}
		return
	}

	receivedVersion, err := conn.ReadText()
	if err != nil {
		if server.ErrorIncommingConnection != nil {
			server.ErrorIncommingConnection(err)
		}
		return
	}
	if receivedVersion != protocolVersion {
		_ = conn.WriteText("invalid")
		fmt.Printf("err: %s\n", err)
		fmt.Printf("in: '%s'\n", receivedVersion)
		fmt.Printf("hv: '%s'\n", protocolVersion)
		if server.ErrorIncommingConnection != nil {
			server.ErrorIncommingConnection(ErrInvalidVersionString)
		}
		return
	}
	err = conn.WriteText("good")
	if err != nil {
		if server.ErrorIncommingConnection != nil {
			server.ErrorIncommingConnection(err)
		}
		return
	}

	fmt.Println("Valid protocol version detected")

	session := server.NewSession() //++ TODO: give {{.Service.CapitalizedName}}Client to NewSession()
	
	// run protocol
	err = run{{.Service.CapitalizedName}}Protocol(conn, session)
	// err can be nil, but we want to call .Stop always
	session.Stop(err)
}

func run{{.Service.CapitalizedName}}Protocol(conn *websocket.Conn, session {{.Service.CapitalizedName}}SessionInterface) error {
	for {
		inMsg := &angoInMsg{}
		err := conn.ReadJSON(inMsg)
		if err != nil {
			return err
		}

		switch inMsg.Type {
		case msgTypeRequest:
			fmt.Printf("Have request: %s\n", inMsg.Procedure)
			switch inMsg.Procedure {
			{{range .Service.ServerProcedures}}
				case "{{.Name}}":
					procArgs := &angoInData{{.CapitalizedName}}{}
					err = json.Unmarshal(inMsg.Data, procArgs)
					if err != nil {
						return err
					}
					session.{{.CapitalizedName}}( {{range $i, $arg := .Args}} {{if $i}},{{end}} procArgs.{{$arg.CapitalizedName}} {{end}} )
			{{end}}
			default:
				return ErrUnkonwnProcedure
			}
		case msgTypeResponse:
			fmt.Printf("Have response: %d\n", inMsg.CallbackID)
		default:
			return ErrInvalidMessageType
		}
	}
}
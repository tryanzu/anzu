package transmit

type Message struct {
    Room    string `json:"room"`
    Event   string `json:"event"`
    Message map[string]interface{} `json:"message"`
}
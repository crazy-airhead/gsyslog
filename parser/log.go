package parser

type Log struct {
	Data map[string]interface{} `json:"data"`
	Raw  []byte                 `json:"raw"`
}

func NewLog(header map[string]interface{}, body []byte) *Log {
	return &Log{
		Data: header,
		Raw:  body,
	}
}

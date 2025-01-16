package parser

import "time"

const (
	LogBody = "body"
)

type Log struct {
	// 结构化的数据
	Header map[string]interface{} `json:"header"`
	// 原始数据
	Body []byte `json:"body"`

	// 是否有错，有错时结构化数据不可能
	Err error

	// 辅助解析用，解析的过程中cursor会发生变化
	cursor int
	// 辅助解析用
	len int
	// 是否忽略标签
	skipTag bool
}

func NewLog(body []byte) *Log {
	return &Log{
		Header: make(map[string]interface{}),
		Body:   body,
		cursor: 0,
		len:    len(body),
	}
}

func NewLogWith(header map[string]interface{}, body []byte) *Log {
	if header == nil {
		header = make(map[string]interface{})
	}

	return &Log{
		Header: header,
		Body:   body,
	}
}

func (l *Log) Cursor() int {
	return l.cursor
}

func (l *Log) MoveCursor() {
	l.cursor++
}

func (l *Log) MoveCursorN(step int) {
	l.cursor += step
}

func (l *Log) SetCursor(cursor int) {
	l.cursor = cursor
}

func (l *Log) Len() int {
	return l.len
}

func (l *Log) SetSkipTag(skipTag bool) {
	l.skipTag = skipTag
}

func (l *Log) SkipTag() bool {
	return l.skipTag
}

func (l *Log) Set(key string, val interface{}) {
	l.Header[key] = val
}

// SetPriority rfc316 rfc5424
func (l *Log) SetPriority(priority int) {
	l.Set("priority", priority)
}

// SetFacility  rfc316 rfc5424
func (l *Log) SetFacility(facility int) {
	l.Set("facility", facility)
}

// SetSeverity  rfc316 rfc5424
func (l *Log) SetSeverity(severity int) {
	l.Set("severity", severity)
}

// SetTag  rfc316 rfc5424
func (l *Log) SetTag(tag string) {
	l.Set("tag", tag)
}

// SetClient  rfc316 rfc5424
func (l *Log) SetClient(client string) {
	l.Set("client", client)
}

// SetHostname  rfc316 rfc5424
func (l *Log) SetHostname(hostname string) {
	l.Set("hostname", hostname)
}

// SetTimestamp  rfc316 rfc5424
func (l *Log) SetTimestamp(timestamp time.Time) {
	l.Set("timestamp", timestamp)
}

// SetContent  rfc316
func (l *Log) SetContent(content string) {
	l.Set("content", content)
}

// SetVersion  rfc5424
func (l *Log) SetVersion(version int) {
	l.Set("version", version)
}

// SetAppName  rfc5424
func (l *Log) SetAppName(appName string) {
	l.Set("appName", appName)
}

// SetProcId  rfc5424
func (l *Log) SetProcId(procId string) {
	l.Set("procId", procId)
}

// SetMsgId rfc5424
func (l *Log) SetMsgId(msgId string) {
	l.Set("msgId", msgId)
}

// SetStructuredData rfc5424
func (l *Log) SetStructuredData(structuredData string) {
	l.Set("structuredData", structuredData)
}

// SetMessage rfc5424
func (l *Log) SetMessage(message string) {
	l.Set("message", message)
}

func (l *Log) Get(key string) interface{} {
	// find body first
	if key == LogBody && len(l.Body) != 0 {
		return l.Body
	}

	// then get from header
	return l.Header[key]
}

func (l *Log) GetString(key string) string {
	// find body first
	if key == LogBody && len(l.Body) != 0 {
		return string(l.Body)
	}

	// then get from header
	val := l.Header[key]
	if s, ok := (val).(string); ok {
		return s
	}

	return ""
}

package mmq

import (
	stdjson "encoding/json"
	"log"
	"os"
	"strings"

	json "github.com/bitly/go-simplejson"
)

// Message 使用的消息
type Message struct {
	json.Json
}

func NewMessage() *Message {
	return &Message{}
}

// Set 支持v是string、bool、int、map[string]interface{}、[]interface{}
func (m *Message) Set(path string, v interface{}) {
	m.SetPath(strings.Split(path, "."), v)
}

func (m *Message) Get(path string) *Message {
	if path == "" {
		return m
	}
	return &Message{*m.GetPath(strings.Split(path, ".")...)}
}

func (m *Message) String(path string, def ...string) string {
	return m.Get(path).MustString(def...)
}

func (m *Message) Bool(path string, def ...bool) bool {
	return m.Get(path).MustBool(def...)
}

func (m *Message) Int(path string, def ...int) int {
	return m.Get(path).MustInt(def...)
}

func (m *Message) Array() []Message {
	arr, err := m.Json.Array()
	if arr == nil || err != nil {
		return nil
	}
	marray := make([]Message, len(arr))
	for i, a := range arr {
		marray[i].SetPath([]string{}, a)
	}
	return marray
}

func (m *Message) Map(path string, def ...map[string]interface{}) map[string]interface{} {
	return m.Get(path).MustMap(def...)
}

// Unmarshal 把m解析到v上。类似json.Unmarshal()
func (m *Message) Unmarshal(v interface{}) error {
	b := m.ToBytes()
	return stdjson.Unmarshal(b, v)
}

// ToBytes Message转成[]byte
func (m *Message) ToBytes() []byte {
	b, err := m.EncodePretty()
	if err != nil {
		log.Printf("messagep[%v].EncodePretty() error: %v", m, err)
	}
	return b
}

// ToString Message转成string
func (m *Message) ToString() string {
	return string(m.ToBytes())
}

// MessageFromBytes 字节数组转成Message
func MessageFromBytes(data []byte) (*Message, error) {
	m, err := json.NewJson(data)
	if err != nil {
		return nil, err
	}
	return &Message{*m}, nil
}

// MessageFromString 字符串转成Message
func MessageFromString(s string) (*Message, error) {
	m, err := MessageFromBytes([]byte(s))
	return m, err
}

// MessageFromStruct 类似json.Marshal()
func MessageFromStruct(v interface{}) *Message {
	b, _ := stdjson.Marshal(v)
	m, _ := MessageFromBytes(b)
	return m
}

// MessageFromFile 从filepath读取Message
func MessageFromFile(filepath string) (*Message, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	m, err := MessageFromBytes(b)
	return m, err
}

// ToFile 把Message保存到filepath
func (m *Message) ToFile(filepath string) error {
	b, err := m.EncodePretty()
	if err != nil {
		return err
	}
	const defaultFileMode = 0644
	return os.WriteFile(filepath, b, defaultFileMode)
}

/*
TODO
func New() *Message
func NewMessage(v interface{}) *Message

func (m *Message) Set(path string, val interface{})
func (m *Message) Get(path string) *Message

func (m *Message) MessageArray() ([]Message, error)
func (m *Message) MustMessageArray(msg ...[]Message) []Message

func MessageFromFile(filepath string) (*Message, error)
func (m *Message) ToFile(filepath string) error

*/

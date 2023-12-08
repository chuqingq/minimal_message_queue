package mmq

import (
	"bytes"
	"encoding/gob"
)

type Command struct {
	Cmd   string
	Topic string
	Data  []byte
}

func (s *Command) FromBytes(b []byte) error {
	return gob.NewDecoder(bytes.NewReader(b)).Decode(s)
}

func (s *Command) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(s)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const (
	CmdPublish     = "publish"
	CmdSubscribe   = "subscribe"
	CmdUnsubscribe = "unsubscribe"
)

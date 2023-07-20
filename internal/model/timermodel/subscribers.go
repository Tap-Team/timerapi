package timermodel

import (
	"bytes"
	"encoding/gob"
)

type Subscribers map[int64]struct{}

func (s Subscribers) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode((map[int64]struct{})(s)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Subscribers) UnmarshalBinary(data []byte) error {
	buf := bytes.NewReader(data)
	enc := gob.NewDecoder(buf)
	if err := enc.Decode((*map[int64]struct{})(s)); err != nil {
		return err
	}
	return nil
}

func (s Subscribers) Array() []int64 {
	keys := make([]int64, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

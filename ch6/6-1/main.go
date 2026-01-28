package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

func encode(gl GameLog) ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(gl)
	if err != nil {
		return nil, fmt.Errorf("error encoding: %v", err)
	}

	return b.Bytes(), nil
}

func decode(data []byte) (GameLog, error) {
	var gl GameLog
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&gl)
	if err != nil {
		return GameLog{}, fmt.Errorf("error decoding: %v", err)
	}

	return gl, nil
}

// don't touch below this line

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}

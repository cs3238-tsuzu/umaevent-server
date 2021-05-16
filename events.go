package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/dop251/goja"
)

type UmaEvent struct {
	Event     string `json:"e"`
	Character string `json:"n"`
	C         string `json:"c"`
	L         string `json:"l,omitempty"`
	A         string `json:"a,omitempty"`
	K         string `json:"k"`
	Choices   []struct {
		Choice string `json:"n"`
		Result string `json:"t"`
	} `json:"choices"`
}

func fetchUmaEvents() ([]UmaEvent, error) {
	eventEndpoint := os.Getenv("EVENT_DATA_ENDPOINT")

	resp, err := client.Get(eventEndpoint)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	script, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	runtime := goja.New()
	_, err = runtime.RunString(string(script) + "\nfunction f() { return JSON.stringify(eventDatas); }")

	if err != nil {
		return nil, err
	}

	var fn func() string
	err = runtime.ExportTo(runtime.Get("f"), &fn)
	if err != nil {
		return nil, err
	}

	var results []UmaEvent
	if err := json.Unmarshal([]byte(fn()), &results); err != nil {
		return nil, err
	}

	return results, nil
}

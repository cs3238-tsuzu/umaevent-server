package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/dop251/goja"
)

type UmaEventChoice struct {
	Choice string `json:"n"`
	Result string `json:"t"`
}

type UmaEvent struct {
	Event     string           `json:"e"`
	Character string           `json:"n"`
	C         string           `json:"c"`
	L         string           `json:"l,omitempty"`
	A         string           `json:"a,omitempty"`
	K         string           `json:"k"`
	Choices   []UmaEventChoice `json:"choices"`
}

var (
	specificCondRegex = regexp.MustCompile(`(.*)\(.*[：:](.*)\)`)
)

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

func dumpUmaEvents(events []UmaEvent, name string) error {
	fp, err := os.Create(name)

	if err != nil {
		return err
	}
	defer fp.Close()

	if err = json.NewEncoder(fp).Encode(events); err != nil {
		return err
	}

	return nil
}

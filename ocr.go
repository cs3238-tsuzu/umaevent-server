package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"os"
)

var (
	ocrServerEndpoint = "http://10.20.40.14:8080/file"
)

func init() {
	if v, ok := os.LookupEnv("OCR_SERVER_ENDPOINT"); ok {
		ocrServerEndpoint = v
	}
}

func ocr(data []byte) (string, error) {
	buf := bytes.NewBuffer(nil)
	mr := multipart.NewWriter(buf)
	{
		field, err := mr.CreateFormField("languages")

		if err != nil {
			return "", err
		}

		field.Write([]byte("jpn"))
	}
	{
		file, err := mr.CreateFormFile("file", "ocr.jpg")

		if err != nil {
			return "", err
		}

		if _, err := file.Write(data); err != nil {
			return "", err
		}
	}
	mr.Close()

	req, err := http.NewRequest("POST", ocrServerEndpoint, buf)

	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mr.FormDataContentType())

	var result struct {
		Result  string `json:"result"`
		Version string `json:"version"`
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Result, nil
}

package main

import (
	"encoding/json"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

var (
	client = http.Client{}

	umaEvents []UmaEvent
)

type Choice struct {
	Choice string `json:"choice"`
	Effect string `json:"effect"`
}

type Result struct {
	OK        bool     `json:"ok"`
	EventName string   `json:"eventName"`
	Choices   []Choice `json:"choices"`
}

func main() {
	events, err := fetchUmaEvents()
	if err != nil {
		panic(err)
	}

	dumpUmaEvents(events, "events.json")

	mux := http.NewServeMux()

	mux.HandleFunc("/upload", func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			rw.WriteHeader(http.StatusNotImplemented)

			return
		}

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil {
			log.Fatal(err)
		}
		if !strings.HasPrefix(mediaType, "multipart/") {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Only multipart is supported"))

			return
		}

		mr := multipart.NewReader(req.Body, params["boundary"])

		var images postedImages
		images.Choices = make([][]byte, 3)

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err)

				return
			}
			slurp, err := io.ReadAll(p)
			if err != nil {
				log.Println(err)

				return
			}

			switch p.FormName() {
			case "title":
				images.Title = slurp
			case "choice1":
				images.Choices[0] = slurp
			case "choice2":
				images.Choices[1] = slurp
			case "choice3":
				images.Choices[2] = slurp
			}
		}

		event, err := NewHandler(events, &images).handle()

		if err != nil {
			log.Println(err)
		}

		if event == nil {
			event = &UmaEvent{}
		}
		choices := make([]Choice, 0, len(event.Choices))
		for i := range event.Choices {
			choices = append(choices, Choice{
				Choice: event.Choices[i].Choice,
				Effect: strings.ReplaceAll(event.Choices[i].Result, "[br]", "\n"),
			})
		}

		json.NewEncoder(rw).Encode(Result{
			OK:        err == nil,
			EventName: event.Event,
			Choices:   choices,
		})
	})

	addr := ":8080"
	if port, ok := os.LookupEnv("PORT"); ok {
		addr = ":" + port
	}

	http.ListenAndServe(addr, mux)
}

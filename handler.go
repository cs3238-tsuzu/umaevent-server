package main

import (
	"bytes"
	"image/color"
	"image/jpeg"
	"log"
	"math"
	"sort"
	"sync"

	"golang.org/x/sync/errgroup"
)

var (
	workerConcurrency = 4
)

type Handler struct {
	events []UmaEvent
	images *postedImages

	parsed *parsedImages
}

func NewHandler(events []UmaEvent, images *postedImages) *Handler {
	return &Handler{
		events: events,
		images: images,
	}
}

type postedImages struct {
	Title   []byte
	Choices [][]byte
}

type parsedImages struct {
	Title   string
	Choices []string
}

func (h *Handler) handle() (*UmaEvent, error) {
	ok, err := h.prepare()

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, nil
	}

	return h.find(), nil
}

func (h *Handler) find() *UmaEvent {
	if len(h.events) == 0 {
		return nil
	}

	type result struct {
		score int
		event *UmaEvent
	}

	scores := make([]*result, len(h.events))

	var wg sync.WaitGroup
	for i := 0; i < workerConcurrency; i++ {
		wg.Add(1)
		lh := len(h.events) * i / workerConcurrency
		rh := len(h.events) * (i + 1) / workerConcurrency

		go func() {
			defer wg.Done()

			for ; lh < rh; lh++ {
				scores[lh] = &result{
					score: h.calcScore(&h.events[lh]),
					event: &h.events[lh],
				}
			}
		}()
	}
	wg.Wait()

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score < scores[j].score
	})

	log.Printf("matched: %s(%d)", scores[0].event.Event, scores[0].score)
	if scores[0].score >= 15 {
		return nil
	}

	return scores[0].event
}

func (h *Handler) calcScore(event *UmaEvent) int {
	score := levenshtein(event.Event, h.parsed.Title)

	if len(event.Choices) < len(h.parsed.Choices) {
		return math.MaxInt32
	}

	for i := range h.parsed.Choices {
		eventIdx := len(event.Choices) - 1 - i
		parsedIdx := len(h.parsed.Choices) - 1 - i

		score += levenshtein(event.Choices[eventIdx].Choice, h.parsed.Choices[parsedIdx])
	}

	return score
}

func (h *Handler) prepare() (bool, error) {
	gr := errgroup.Group{}
	var parsed parsedImages
	parsed.Choices = make([]string, 3)
	availableChoiceFlags := make([]bool, 3)

	// Title
	gr.Go(func() error {
		title, err := ocr(h.images.Title)
		parsed.Title = title

		return err
	})

	for i := 0; i < 3; i++ {
		i := i

		gr.Go(func() error {
			img, err := jpeg.Decode(bytes.NewReader(h.images.Choices[i]))

			if err != nil {
				return err
			}

			availableChoiceFlags[i] = color.GrayModel.Convert(img.At(0, 0)).(color.Gray).Y >= 252 &&
				color.GrayModel.Convert(img.At(0, img.Bounds().Dy()-1)).(color.Gray).Y >= 252

			if !availableChoiceFlags[i] {
				return nil
			}

			choice1, err := ocr(h.images.Choices[i])
			parsed.Choices[i] = choice1

			return err
		})
	}

	err := gr.Wait()
	if err != nil {
		return false, err
	}

	if !availableChoiceFlags[1] || !availableChoiceFlags[2] {
		return false, nil
	}

	if !availableChoiceFlags[0] {
		parsed.Choices = parsed.Choices[1:]
	}

	h.parsed = &parsed

	return true, nil
}

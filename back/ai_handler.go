package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type ChatRequest struct {
	Message string `json:"message"`
}

var aiClient *openai.Client

type visitor struct {
	Requests []time.Time
}

var visitors = make(map[string]*visitor)

var visitorsMutex sync.Mutex

const (
	maxRequests = 10

	windowTime = time.Minute

	maxMessageLength = 500
)

func InitAI() {

	apiKey := os.Getenv("OPENROUTER_API_KEY")

	if apiKey == "" {

		log.Println("OPENROUTER KEY NOT FOUND")

		return

	}

	config := openai.DefaultConfig(apiKey)

	config.BaseURL = "https://openrouter.ai/api/v1"

	aiClient = openai.NewClientWithConfig(config)

	log.Println("AI CONNECTED")

}

func AIChatHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	if r.Method != http.MethodPost {

		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)

		return

	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1024,
	)

	ip := r.RemoteAddr

	host, _, splitErr := net.SplitHostPort(ip)

	if splitErr == nil {
		ip = host
	}

	if !checkRateLimit(ip) {

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "Слишком много запросов. Попробуйте через минуту.",
			},
		)

		return

	}

	var data ChatRequest

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "Некорректный запрос.",
			},
		)

		return

	}

	message := strings.TrimSpace(
		data.Message,
	)

	if message == "" {

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "Напишите вопрос.",
			},
		)

		return

	}

	SaveAIRequest(message)

	if len(message) > maxMessageLength {

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "Ваш вопрос слишком длинный.",
			},
		)

		return

	}

	if !isDentalQuestion(message) {

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "Я могу отвечать только на вопросы по стоматологии. Например: лечение кариеса, имплантация, боль в зубе или стоимость услуг.",
			},
		)

		return

	}

	if aiClient == nil {

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "ИИ временно недоступен.",
			},
		)

		return

	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		20*time.Second,
	)

	defer cancel()

	response, err := aiClient.CreateChatCompletion(

		ctx,

		openai.ChatCompletionRequest{

			Model: "openrouter/free",

			Messages: []openai.ChatCompletionMessage{

				{

					Role: "system",

					Content: `

Ты — ИИ-консультант стоматологической клиники.

Правила:

- отвечай только на вопросы по стоматологии;
- объясняй простыми словами;
- не ставь диагноз;
- не назначай лекарства;
- не заменяй врача;
- советуй посетить стоматолога;
- предлагай записаться на консультацию;
- отвечай 2-4 предложениями.

`,
				},

				{

					Role: "user",

					Content: message,
				},
			},

			Temperature: 0.5,
		},
	)

	if err != nil {

		log.Println("OPENROUTER ERROR:")
		log.Println(err)

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "Ошибка подключения к ИИ.",
			},
		)

		return

	}

	if len(response.Choices) == 0 {

		json.NewEncoder(w).Encode(
			map[string]string{

				"answer": "ИИ не смог сформировать ответ.",
			},
		)

		return

	}

	answer := response.Choices[0].Message.Content

	json.NewEncoder(w).Encode(
		map[string]string{

			"answer": answer,
		},
	)

}

func checkRateLimit(ip string) bool {

	visitorsMutex.Lock()

	defer visitorsMutex.Unlock()

	now := time.Now()

	if visitors[ip] == nil {

		visitors[ip] = &visitor{}

	}

	validRequests := []time.Time{}

	for _, requestTime := range visitors[ip].Requests {

		if now.Sub(requestTime) < windowTime {

			validRequests = append(
				validRequests,
				requestTime,
			)

		}

	}

	visitors[ip].Requests = validRequests

	if len(validRequests) >= maxRequests {

		return false

	}

	visitors[ip].Requests = append(
		visitors[ip].Requests,
		now,
	)

	return true

}

func isDentalQuestion(text string) bool {

	text = strings.ToLower(text)

	keywords := []string{

		"зуб",
		"зубы",
		"стомат",
		"кариес",
		"имплант",
		"имплантация",
		"брекет",
		"корон",
		"винир",
		"десн",
		"боль",
		"удал",
		"чист",
		"пломб",
		"канал",
		"прикус",
		"челюст",
		"врач",
		"клиник",
		"лечение",
		"цена",
		"стоимость",
		"стоматолог",
	}

	for _, word := range keywords {

		if strings.Contains(text, word) {

			return true

		}

	}

	return false

}

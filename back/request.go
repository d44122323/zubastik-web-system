package main
import "time"
type Request struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Comment string `json:"comment"`
	Services string `json:"services"`
	Price    int    `json:"price"`
	Source   string `json:"source"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
package main

import "fmt"

func SendEmail(cfg Config, data FormData) {
	fmt.Println("Email отправка:", data)
}

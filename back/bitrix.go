package main
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)
func SendBitrix(cfg Config, data FormData) error {
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"TITLE": "Заявка с сайта",
			"NAME": data.Name,
			"PHONE": []map[string]string{
				{
					"VALUE":      data.Phone,
					"VALUE_TYPE": "WORK",
				},
			},
			"SOURCE_ID": "WEB",
			"COMMENTS": fmt.Sprintf(
				"Услуги: %s\nСтоимость: %d ₽\nКомментарий: %s",
				data.Services,
				data.Price,
				data.Comment,
			),
		},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Println("BITRIX REQUEST:")
	fmt.Println(string(jsonData))
	req, err := http.Post(
		cfg.BitrixWebhook,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	defer req.Body.Close()
	body, _ := io.ReadAll(req.Body)
	fmt.Println("BITRIX STATUS:")
	fmt.Println(req.Status)
	fmt.Println("BITRIX RESPONSE:")
	fmt.Println(string(body))
	if req.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"bitrix error: %s",
			string(body),
		)
	}
	return nil
}
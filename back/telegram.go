package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TgMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type tgResponse struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
}

var tgConnectMu sync.Mutex

func telegramToken(cfg Config) string {
	if v := strings.TrimSpace(cfg.TelegramBotToken); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("TELEGRAM_PATIENT_BOT_TOKEN"))
}

func telegramStaffToken(cfg Config) string {
	if v := strings.TrimSpace(cfg.TelegramStaffBotToken); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("TELEGRAM_STAFF_BOT_TOKEN"))
}

func telegramStaffChatID(cfg Config) string {
	if v := strings.TrimSpace(cfg.TelegramStaffChatID); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("TELEGRAM_STAFF_CHAT_ID"))
}

func telegramStaffCall(cfg Config, method string, payload any) ([]byte, error) {
	token := telegramStaffToken(cfg)
	if token == "" {
		return nil, errors.New("TELEGRAM_STAFF_BOT_TOKEN is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.telegram.org/bot"+token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram staff http %d: %s", resp.StatusCode, string(raw))
	}
	var tr tgResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, err
	}
	if !tr.OK {
		return nil, fmt.Errorf("telegram staff api error: %s", string(raw))
	}
	return raw, nil
}

func telegramUsername(cfg Config) string {
	if v := strings.TrimSpace(cfg.TelegramBotUsername); v != "" {
		return strings.TrimPrefix(v, "@")
	}
	return strings.TrimPrefix(strings.TrimSpace(os.Getenv("TELEGRAM_PATIENT_BOT_USERNAME")), "@")
}

func telegramCall(cfg Config, method string, payload any) ([]byte, error) {
	token := telegramToken(cfg)
	if token == "" {
		return nil, errors.New("TELEGRAM_BOT_TOKEN is not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.telegram.org/bot"+token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram http %d: %s", resp.StatusCode, string(raw))
	}
	var tr tgResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return nil, err
	}
	if !tr.OK {
		return nil, fmt.Errorf("telegram api error: %s", string(raw))
	}
	return raw, nil
}

func SendTelegram(cfg Config, data FormData) error {
	chatID := telegramStaffChatID(cfg)
	if chatID == "" {
		return errors.New("TELEGRAM_STAFF_CHAT_ID is not configured")
	}
	msg := TgMessage{
		ChatID: chatID,
		Text:   fmt.Sprintf("📨 Новая заявка!\n\n👤 Имя: %s\n📞 Телефон: %s\n💬 Комментарий: %s", data.Name, data.Phone, data.Comment),
	}
	_, err := telegramStaffCall(cfg, "sendMessage", msg)
	return err
}

func InitTelegramSchema() error {
	_, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS telegram_accounts (
	id SERIAL PRIMARY KEY,
	patient_id INTEGER NOT NULL UNIQUE REFERENCES patients(id) ON DELETE CASCADE,
	telegram_user_id BIGINT NOT NULL UNIQUE,
	telegram_username VARCHAR(255) NOT NULL DEFAULT '',
	first_name VARCHAR(255) NOT NULL DEFAULT '',
	last_name VARCHAR(255) NOT NULL DEFAULT '',
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	connected_at TIMESTAMP NOT NULL DEFAULT NOW(),
	last_seen_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS telegram_connect_tokens (
	token VARCHAR(128) PRIMARY KEY,
	patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
	expires_at TIMESTAMP NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_tg_connect_tokens_patient ON telegram_connect_tokens(patient_id);
CREATE TABLE IF NOT EXISTS telegram_sent (
	id SERIAL PRIMARY KEY,
	patient_id INTEGER NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
	request_id INTEGER REFERENCES requests(id) ON DELETE CASCADE,
	kind VARCHAR(50) NOT NULL,
	sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
	UNIQUE(request_id, kind)
);
`)
	return err
}

func TelegramStatusHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	var connected bool
	var username, first, last string
	err = DB.QueryRow(`SELECT TRUE,telegram_username,first_name,last_name FROM telegram_accounts WHERE patient_id=$1 AND is_active=TRUE`, u.PatientID).Scan(&connected, &username, &first, &last)
	if err == sql.ErrNoRows {
		connected = false
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"connected": connected, "username": username, "firstName": first, "lastName": last})
}

func TelegramConnectHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	if telegramUsername(LoadConfig()) == "" || telegramToken(LoadConfig()) == "" {
		http.Error(w, "Telegram не настроен", 503)
		return
	}
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		http.Error(w, "не удалось создать токен", 500)
		return
	}
	token := hex.EncodeToString(raw)
	_, _ = DB.Exec(`DELETE FROM telegram_connect_tokens WHERE patient_id=$1 OR expires_at<NOW()`, u.PatientID)
	_, err = DB.Exec(`INSERT INTO telegram_connect_tokens(token,patient_id,expires_at) VALUES($1,$2,NOW()+INTERVAL '15 minutes')`, token, u.PatientID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	bot := telegramUsername(LoadConfig())
	writeJSON(w, map[string]string{"url": "https://t.me/" + url.PathEscape(bot) + "?start=" + url.QueryEscape(token)})
}

func TelegramDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	u, err := currentPatientUser(r)
	if err != nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	_, err = DB.Exec(`UPDATE telegram_accounts SET is_active=FALSE WHERE patient_id=$1`, u.PatientID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"status": "ok"})
}

func sendTelegramToPatient(cfg Config, patientID int, text string) error {
	var chatID int64
	var active bool
	err := DB.QueryRow(`SELECT telegram_user_id,is_active FROM telegram_accounts WHERE patient_id=$1`, patientID).Scan(&chatID, &active)
	if err != nil {
		return err
	}
	if !active {
		return errors.New("telegram не подключён")
	}
	_, err = telegramCall(cfg, "sendMessage", map[string]any{"chat_id": chatID, "text": text})
	if err == nil {
		_, _ = DB.Exec(`UPDATE telegram_accounts SET last_seen_at=NOW() WHERE patient_id=$1`, patientID)
	}
	return err
}

func TelegramSendRequestHandler(w http.ResponseWriter, r *http.Request, cfg Config, doctorOnly bool) {
	var allowed bool
	var doctorID int
	if doctorOnly {
		u, err := DoctorRequestGuard(r)
		if err != nil {
			http.Error(w, "unauthorized", 401)
			return
		}
		allowed = true
		doctorID = u.DoctorID
	} else {
		allowed = isAdminRequest(r)
	}
	if !allowed {
		http.Error(w, "unauthorized", 401)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/telegram/request/")
	if idStr == r.URL.Path {
		idStr = strings.TrimPrefix(r.URL.Path, "/api/doctor/telegram/request/")
	}
	id, err := strconv.Atoi(strings.Trim(idStr, "/"))
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", 400)
		return
	}
	var patientID int
	var name, service, date, tm, doctor string
	q := `SELECT r.patient_id,COALESCE(p.name,''),COALESCE(r.services,''),COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),COALESCE(d.name,'') FROM requests r JOIN patients p ON p.id=r.patient_id LEFT JOIN doctors d ON d.id=r.doctor_id WHERE r.id=$1`
	args := []any{id}
	if doctorOnly {
		q += ` AND r.doctor_id=$2`
		args = append(args, doctorID)
	}
	if err := DB.QueryRow(q, args...).Scan(&patientID, &name, &service, &date, &tm, &doctor); err != nil {
		http.Error(w, "Запись не найдена", 404)
		return
	}
	text := "🔔 Напоминание о приёме\n\n"
	if name != "" {
		text += "Пациент: " + name + "\n"
	}
	if date != "" {
		text += "📅 " + date + "\n"
	}
	if tm != "" {
		text += "🕐 " + tm + "\n"
	}
	if doctor != "" {
		text += "👨‍⚕️ Врач: " + doctor + "\n"
	}
	if service != "" {
		text += "🦷 Услуга: " + service + "\n"
	}
	if err := sendTelegramToPatient(cfg, patientID, text); err != nil {
		http.Error(w, "Telegram: "+err.Error(), 409)
		return
	}
	writeJSON(w, map[string]any{"status": "ok"})
}

func telegramStatusNotification(cfg Config, requestID int, status string) {
	var patientID int
	var name, service, date, tm, doctor string
	err := DB.QueryRow(`SELECT r.patient_id,COALESCE(p.name,''),COALESCE(r.services,''),COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),COALESCE(d.name,'') FROM requests r JOIN patients p ON p.id=r.patient_id LEFT JOIN doctors d ON d.id=r.doctor_id WHERE r.id=$1`, requestID).Scan(&patientID, &name, &service, &date, &tm, &doctor)
	if err != nil {
		return
	}
	var text string
	switch status {
	case "Подтверждена":
		text = "✅ Запись подтверждена"
	case "Отменена", "Отменено", "Отменено пациентом":
		text = "❌ Запись отменена"
	case "Завершена":
		text = "🦷 Приём завершён"
	default:
		return
	}
	if date != "" {
		text += "\n\n📅 " + date
	}
	if tm != "" {
		text += " · " + tm
	}
	if doctor != "" {
		text += "\n👨‍⚕️ " + doctor
	}
	if service != "" {
		text += "\n🦷 " + service
	}
	_ = sendTelegramToPatient(cfg, patientID, text)
}

func telegramRescheduleNotification(cfg Config, requestID int) {
	var patientID int
	var date, tm, doctor, service string
	if err := DB.QueryRow(`SELECT r.patient_id,COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),COALESCE(d.name,''),COALESCE(r.services,'') FROM requests r LEFT JOIN doctors d ON d.id=r.doctor_id WHERE r.id=$1`, requestID).Scan(&patientID, &date, &tm, &doctor, &service); err != nil {
		return
	}
	text := "🔄 Запись перенесена"
	if date != "" {
		text += "\n\n📅 " + date
	}
	if tm != "" {
		text += " · " + tm
	}
	if doctor != "" {
		text += "\n👨‍⚕️ " + doctor
	}
	if service != "" {
		text += "\n🦷 " + service
	}
	_ = sendTelegramToPatient(cfg, patientID, text)
}

func telegramReminderWorker(cfg Config) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		for _, item := range []struct {
			minutes int
			kind    string
			label   string
		}{{1440, "24h", "Завтра у вас приём"}, {120, "2h", "Напоминание о приёме"}} {
			rows, err := DB.Query(`SELECT r.id,r.patient_id FROM requests r WHERE r.appointment_date IS NOT NULL AND r.appointment_time IS NOT NULL AND r.status IN ('Подтверждена','Новая') AND (r.appointment_date::timestamp+r.appointment_time) BETWEEN NOW()+($1::text||' minutes')::interval-INTERVAL '1 minute' AND NOW()+($1::text||' minutes')::interval+INTERVAL '1 minute'`, item.minutes)
			if err != nil {
				continue
			}
			for rows.Next() {
				var rid, pid int
				if rows.Scan(&rid, &pid) == nil {
					var sent bool
					_ = DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM telegram_sent WHERE request_id=$1 AND kind=$2)`, rid, item.kind).Scan(&sent)
					if sent {
						continue
					}
					var date, tm, doctor, service string
					if DB.QueryRow(`SELECT COALESCE(r.appointment_date::text,''),COALESCE(TO_CHAR(r.appointment_time,'HH24:MI'),''),COALESCE(d.name,''),COALESCE(r.services,'') FROM requests r LEFT JOIN doctors d ON d.id=r.doctor_id WHERE r.id=$1`, rid).Scan(&date, &tm, &doctor, &service) == nil {
						text := item.label + "\n\n📅 " + date + " · " + tm
						if doctor != "" {
							text += "\n👨‍⚕️ " + doctor
						}
						if service != "" {
							text += "\n🦷 " + service
						}
						if sendTelegramToPatient(cfg, pid, text) == nil {
							_, _ = DB.Exec(`INSERT INTO telegram_sent(patient_id,request_id,kind) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, pid, rid, item.kind)
						}
					}
				}
			}
			rows.Close()
		}
	}
}

func telegramPoller(cfg Config) {
	token := telegramToken(cfg)
	if token == "" {
		return
	}

	// This application uses long polling. A webhook left over from a previous
	// deployment prevents getUpdates from working, so remove it once at startup.
	if _, err := telegramCall(cfg, "deleteWebhook", map[string]any{"drop_pending_updates": false}); err != nil {
		log.Printf("TELEGRAM WEBHOOK WARNING: %v", err)
	}

	if raw, err := telegramCall(cfg, "getMe", map[string]any{}); err != nil {
		log.Printf("TELEGRAM PATIENT BOT ERROR: %v", err)
	} else {
		var me struct { OK bool `json:"ok"`; Result struct { Username string `json:"username"` } `json:"result"` }
		if err := json.Unmarshal(raw, &me); err == nil && me.OK {
			log.Printf("TELEGRAM PATIENT BOT CONNECTED: @%s", me.Result.Username)
		}
	}

	var offset int64
	for {
		raw, err := telegramCall(cfg, "getUpdates", map[string]any{"offset": offset + 1, "timeout": 20, "allowed_updates": []string{"message"}})
		if err != nil {
			log.Printf("TELEGRAM POLLING ERROR: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		var resp struct {
			OK     bool `json:"ok"`
			Result []struct {
				UpdateID int64 `json:"update_id"`
				Message  *struct {
					Chat struct {
						ID                            int64  `json:"id"`
						Username, FirstName, LastName string `json:"username"`
					} `json:"chat"`
					Text string `json:"text"`
				} `json:"message"`
			} `json:"result"`
		}
		if json.Unmarshal(raw, &resp) != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		for _, u := range resp.Result {
			if u.UpdateID > offset {
				offset = u.UpdateID
			}
			if u.Message == nil {
				continue
			}
			handleTelegramMessage(cfg, u.Message.Chat.ID, u.Message.Chat.Username, u.Message.Chat.FirstName, u.Message.Chat.LastName, u.Message.Text)
		}
	}
}

func handleTelegramMessage(cfg Config, chatID int64, username, first, last, text string) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "/start") {
		parts := strings.Fields(text)
		if len(parts) > 1 {
			var patientID int
			err := DB.QueryRow(`SELECT patient_id FROM telegram_connect_tokens WHERE token=$1 AND expires_at>NOW()`, parts[1]).Scan(&patientID)
			if err == nil {
				_, _ = DB.Exec(`INSERT INTO telegram_accounts(patient_id,telegram_user_id,telegram_username,first_name,last_name) VALUES($1,$2,$3,$4,$5) ON CONFLICT(patient_id) DO UPDATE SET telegram_user_id=EXCLUDED.telegram_user_id,telegram_username=EXCLUDED.telegram_username,first_name=EXCLUDED.first_name,last_name=EXCLUDED.last_name,is_active=TRUE,last_seen_at=NOW()`, patientID, chatID, username, first, last)
				_, _ = DB.Exec(`DELETE FROM telegram_connect_tokens WHERE token=$1`, parts[1])
				_ = sendTelegramChat(cfg, chatID, "✅ Telegram подключён к вашему личному кабинету ЗУБАСТИК.\nТеперь вы будете получать уведомления о записях.")
				return
			}
		}
	}
	_ = sendTelegramChat(cfg, chatID, "Для подключения Telegram откройте личный кабинет на сайте и нажмите «Подключить Telegram».")
}

func sendTelegramChat(cfg Config, chatID int64, text string) error {
	_, err := telegramCall(cfg, "sendMessage", map[string]any{"chat_id": chatID, "text": text})
	return err
}

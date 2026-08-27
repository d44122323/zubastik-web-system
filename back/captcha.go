package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CaptchaChallenge struct {
	Token     string `json:"token"`
	Question  string `json:"question"`
	ExpiresAt int64  `json:"expiresAt"`
}

var captchaMu sync.Mutex

func InitCaptchaSchema() error {
	_, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS captcha_challenges (
    token VARCHAR(128) PRIMARY KEY,
    answer_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    ip VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_captcha_expires ON captcha_challenges(expires_at);
CREATE TABLE IF NOT EXISTS public_rate_limits (
    id SERIAL PRIMARY KEY,
    rate_key VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_public_rate_limits_key_time ON public_rate_limits(rate_key, created_at);
`)
	return err
}

func captchaRandomInt(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return time.Now().UnixNano() % max
	}
	return n.Int64()
}

func newCaptchaToken() string {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(raw)
}

func hashCaptchaAnswer(token, answer string) string {
	h := sha256.Sum256([]byte(token + ":" + strings.TrimSpace(answer)))
	return hex.EncodeToString(h[:])
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func issueCaptcha(w http.ResponseWriter, r *http.Request) {
	a := captchaRandomInt(9) + 1
	b := captchaRandomInt(9) + 1
	op := "+"
	answer := a + b
	if captchaRandomInt(2) == 1 {
		op = "−"
		if a < b {
			a, b = b, a
		}
		answer = a - b
	}
	token := newCaptchaToken()
	expires := time.Now().Add(5 * time.Minute)
	_, err := DB.Exec(`INSERT INTO captcha_challenges(token,answer_hash,expires_at,ip) VALUES($1,$2,$3,$4)`, token, hashCaptchaAnswer(token, fmt.Sprintf("%d", answer)), expires, clientIP(r))
	if err != nil {
		http.Error(w, "Не удалось создать проверку", 500)
		return
	}
	_, _ = DB.Exec(`DELETE FROM captcha_challenges WHERE expires_at < NOW() OR used=TRUE`)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CaptchaChallenge{Token: token, Question: fmt.Sprintf("Сколько будет %d %s %d?", a, op, b), ExpiresAt: expires.Unix()})
}

func validateCaptcha(r *http.Request, token, answer string) error {
	token = strings.TrimSpace(token)
	answer = strings.TrimSpace(answer)
	if token == "" || answer == "" {
		return fmt.Errorf("Пройдите проверку на бота")
	}
	var storedHash string
	var expires time.Time
	var used bool
	err := DB.QueryRow(`SELECT answer_hash,expires_at,used FROM captcha_challenges WHERE token=$1`, token).Scan(&storedHash, &expires, &used)
	if err == sql.ErrNoRows || used || time.Now().After(expires) {
		return fmt.Errorf("Проверка устарела. Обновите CAPTCHA")
	}
	if storedHash != hashCaptchaAnswer(token, answer) {
		return fmt.Errorf("Неверный ответ проверки")
	}
	_, err = DB.Exec(`UPDATE captcha_challenges SET used=TRUE WHERE token=$1 AND used=FALSE`, token)
	return err
}

func enforceRateLimit(key string, max int, window time.Duration) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	captchaMu.Lock()
	defer captchaMu.Unlock()
	cutoff := time.Now().Add(-window)
	_, _ = DB.Exec(`DELETE FROM public_rate_limits WHERE created_at < $1`, cutoff.Add(-time.Hour))
	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM public_rate_limits WHERE rate_key=$1 AND created_at >= $2`, key, cutoff).Scan(&count); err != nil {
		return err
	}
	if count >= max {
		return fmt.Errorf("Слишком много запросов. Попробуйте позже")
	}
	_, err := DB.Exec(`INSERT INTO public_rate_limits(rate_key) VALUES($1)`, key)
	return err
}

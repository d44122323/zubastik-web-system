package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

type Service struct {
	ID        int    `json:"id"`
	Category  string `json:"category"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	IsActive  bool   `json:"isActive"`
	SortOrder int    `json:"sortOrder"`
}

type ServicePayload struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	IsActive bool   `json:"isActive"`
}

func InitServicesTable() error {
	_, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    category VARCHAR(200) NOT NULL DEFAULT '',
    name VARCHAR(500) NOT NULL UNIQUE,
    price INTEGER NOT NULL DEFAULT 0 CHECK (price >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_services_category ON services(category);
CREATE INDEX IF NOT EXISTS idx_services_active ON services(is_active);
`)
	if err != nil {
		return err
	}
	return seedServices()
}

func seedServices() error {
	type seed struct {
		category, name string
		price          int
		order          int
	}
	data := []seed{
		{"Консультация и диагностика", "Первичный осмотр и консультация", 0, 10},
		{"Профессиональная гигиена", "Ультразвуковая чистка", 500, 20},
		{"Профессиональная гигиена", "Air Flow", 500, 21},
		{"Профессиональная гигиена", "Фторирование", 500, 22},
		{"Лечение зубов", "Лечение кариеса под протезирование", 0, 30},
		{"Лечение зубов", "Лечение кариеса", 600, 31},
		{"Лечение зубов", "Лечение пульпита", 600, 32},
		{"Лечение зубов", "Лечение периодонтита", 750, 33},
		{"Лечение зубов", "Реставрация зуба", 950, 34},
		{"Удаление зубов и хирургия", "Удаление зуба под протезирование", 0, 40},
		{"Удаление зубов и хирургия", "Удаление временного (молочного) зуба", 0, 41},
		{"Удаление зубов и хирургия", "Удаление постоянного зуба (простое)", 200, 42},
		{"Удаление зубов и хирургия", "Удаление постоянного зуба (сложное)", 450, 43},
		{"Удаление зубов и хирургия", "Удаление зуба мудрости (3 моляр)", 1200, 44},
		{"Удаление зубов и хирургия", "Удаление ретинированного / дистопированного зуба - I степень сложности", 900, 45},
		{"Удаление зубов и хирургия", "Удаление ретинированного / дистопированного зуба - II степень сложности", 200, 46},
		{"Удаление зубов и хирургия", "Удаление ретинированного / дистопированного зуба - III степень сложности", 400, 47},
		{"Удаление зубов и хирургия", "Удаление отломка коронковой части зуба", 600, 48},
		{"Удаление зубов и хирургия", "Вскрытие поднадкостничного абсцесса", 200, 49},
		{"Удаление зубов и хирургия", "Пластика уздечки верхней или нижней губы", 1500, 50},
		{"Удаление зубов и хирургия", "Рассечение капюшона при перикоронарите", 300, 51},
		{"Удаление зубов и хирургия", "Кюретаж лунки ранее удалённого зуба", 250, 52},
		{"Удаление зубов и хирургия", "Наложение шва", 200, 53},
		{"Коронки и виниры", "Коронка (титановый сплав + керамика NORITAKE, Япония)", 2700, 60},
		{"Коронки и виниры", "Коронка (титановый сплав + керамика VITA VM, Германия)", 2900, 61},
		{"Коронки и виниры", "Коронка (чистый титан + керамика NORITAKE, Япония)", 3100, 62},
		{"Коронки и виниры", "Коронка из диоксида циркония (Китай)", 6250, 63},
		{"Коронки и виниры", "Коронка из диоксида циркония (Япония)", 9500, 64},
		{"Коронки и виниры", "Коронка из диоксида циркония (Германия)", 11000, 65},
		{"Коронки и виниры", "Винир из диоксида циркония (Китай)", 6000, 66},
		{"Коронки и виниры", "Винир из диоксида циркония (Япония)", 6500, 67},
		{"Коронки и виниры", "Винир из диоксида циркония (Германия)", 7400, 68},
		{"Имплантация", "Имплант (Корея)", 25000, 70},
		{"Имплантация", "Имплант (Япония)", 25000, 71},
		{"Имплантация", "Имплант (Германия)", 25000, 72},
		{"Имплантация", "Имплант (Швейцария)", 25000, 73},
		{"Имплантация", "Формирователь десны", 2000, 74},
		{"Имплантация", "Абатмент", 4000, 75},
		{"Имплантация", "Синус-лифтинг (без учёта расходных материалов)", 5000, 76},
		{"Имплантация", "Забор костного трансплантата", 20000, 77},
		{"Пародонтология", "Консультация врача-пародонтолога", 1000, 80},
		{"Пародонтология", "Лечение воспаления дёсен", 3000, 81},
		{"Пародонтология", "Закрытый кюретаж", 2500, 82},
		{"Пародонтология", "Открытый кюретаж", 3000, 83},
		{"Пародонтология", "Ультразвуковое снятия налёта, зубного камня (обе челюсти)", 500, 84},
		{"Пародонтология", "Пескоструйная обработка, шлифовка (обе челюсти)", 500, 85},
		{"Пародонтология", "Нанесение защитного покрытия, фторирование", 500, 86},
		{"Металлокерамические коронки", "Коронка никель-хромовый сплав с использованием Китайской керамической массы", 1800, 90},
		{"Металлокерамические коронки", "Коронка никель-хромовый сплав с использованием Японской керамической массы NORITAKE", 2800, 91},
		{"Металлокерамические коронки", "Коронка никель-хромовый сплав с использованием Немецкой керамической массы VITA VM", 2600, 92},
	}
	for _, s := range data {
		_, err := DB.Exec(`INSERT INTO services(category,name,price,is_active,sort_order) VALUES($1,$2,$3,TRUE,$4) ON CONFLICT(name) DO NOTHING`, s.category, s.name, s.price, s.order)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetServices(activeOnly bool) ([]Service, error) {
	q := `SELECT id,category,name,price,is_active,sort_order FROM services`
	if activeOnly {
		q += ` WHERE is_active=true`
	}
	q += ` ORDER BY category, sort_order, id`
	rows, err := DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Service, 0)
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.ID, &s.Category, &s.Name, &s.Price, &s.IsActive, &s.SortOrder); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

func ResolveServices(ids []int) (string, int, error) {
	if len(ids) == 0 {
		return "", 0, nil
	}
	clean := make([]int, 0, len(ids))
	seen := map[int]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		clean = append(clean, id)
	}
	if len(clean) == 0 {
		return "", 0, nil
	}
	rows, err := DB.Query(`SELECT id,name,price FROM services WHERE id = ANY($1) AND is_active=true ORDER BY sort_order,id`, pq.Array(clean))
	if err != nil {
		return "", 0, err
	}
	defer rows.Close()
	names := make([]string, 0, len(clean))
	total := 0
	for rows.Next() {
		var id, price int
		var name string
		if err := rows.Scan(&id, &name, &price); err != nil {
			return "", 0, err
		}
		names = append(names, name)
		total += price
	}
	if err := rows.Err(); err != nil {
		return "", 0, err
	}
	if len(names) != len(clean) {
		return "", 0, &serviceError{"Одна или несколько услуг недоступны"}
	}
	return strings.Join(names, ", "), total, nil
}

func UpdateService(id int, p ServicePayload) (Service, error) {
	if strings.TrimSpace(p.Name) == "" || p.Price < 0 {
		return Service{}, &serviceError{"Некорректная услуга или цена"}
	}
	_, err := DB.Exec(`UPDATE services SET category=$1,name=$2,price=$3,is_active=$4,updated_at=NOW() WHERE id=$5`, strings.TrimSpace(p.Category), strings.TrimSpace(p.Name), p.Price, p.IsActive, id)
	if err != nil {
		return Service{}, err
	}
	var s Service
	err = DB.QueryRow(`SELECT id,category,name,price,is_active,sort_order FROM services WHERE id=$1`, id).Scan(&s.ID, &s.Category, &s.Name, &s.Price, &s.IsActive, &s.SortOrder)
	return s, err
}

type serviceError struct{ msg string }

func (e *serviceError) Error() string { return e.msg }

func ServicesPublicHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}
	list, err := GetServices(true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func ServicesAdminHandler(w http.ResponseWriter, r *http.Request) {
	if !isAdminRequest(r) {
		http.Error(w, "unauthorized", 401)
		return
	}
	switch r.Method {
	case http.MethodGet:
		list, err := GetServices(false)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	case http.MethodPut:
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/admin/services/"))
		if err != nil {
			http.Error(w, "Invalid ID", 400)
			return
		}
		var p ServicePayload
		if err = json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		s, err := UpdateService(id, p)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s)
	default:
		http.Error(w, "Method not allowed", 405)
	}
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

func InitDoctorsTable() error {
	_, err := DB.Exec(`
CREATE TABLE IF NOT EXISTS doctors (
	id SERIAL PRIMARY KEY,
	name VARCHAR(200) NOT NULL,
	photo VARCHAR(500) NOT NULL DEFAULT '',
	position VARCHAR(500) NOT NULL DEFAULT '',
	experience VARCHAR(100) NOT NULL DEFAULT '',
	specialization VARCHAR(500) NOT NULL DEFAULT '',
	education TEXT NOT NULL DEFAULT '',
	description JSONB NOT NULL DEFAULT '[]'::jsonb,
	specialization_full TEXT NOT NULL DEFAULT '',
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS doctor_schedule (
	id SERIAL PRIMARY KEY,
	doctor_id INTEGER NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
	weekday SMALLINT NOT NULL CHECK (weekday BETWEEN 1 AND 7),
	is_working BOOLEAN NOT NULL DEFAULT TRUE,
	start_time TIME NOT NULL DEFAULT '09:00',
	end_time TIME NOT NULL DEFAULT '18:00',
	UNIQUE (doctor_id, weekday),
	CHECK (end_time > start_time)
);

ALTER TABLE requests ADD COLUMN IF NOT EXISTS doctor_id INTEGER REFERENCES doctors(id) ON DELETE SET NULL;
ALTER TABLE requests ADD COLUMN IF NOT EXISTS appointment_date DATE;
ALTER TABLE requests ADD COLUMN IF NOT EXISTS appointment_time TIME;
CREATE INDEX IF NOT EXISTS idx_requests_doctor_datetime ON requests(doctor_id, appointment_date, appointment_time);
CREATE UNIQUE INDEX IF NOT EXISTS ux_requests_doctor_datetime_active
ON requests(doctor_id, appointment_date, appointment_time)
WHERE doctor_id IS NOT NULL AND appointment_date IS NOT NULL AND appointment_time IS NOT NULL
  AND status NOT IN ('Отменена', 'Отменено', 'Отменено пациентом');
CREATE INDEX IF NOT EXISTS idx_doctors_active ON doctors(is_active);
CREATE INDEX IF NOT EXISTS idx_requests_doctor_id ON requests(doctor_id);
CREATE INDEX IF NOT EXISTS idx_doctor_schedule_doctor_id ON doctor_schedule(doctor_id);
CREATE TABLE IF NOT EXISTS doctor_unavailability (
    id SERIAL PRIMARY KEY,
    doctor_id INTEGER NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (end_date >= start_date)
);
CREATE INDEX IF NOT EXISTS idx_doctor_unavailability_doctor_dates ON doctor_unavailability(doctor_id,start_date,end_date);
`)
	return err
}

func seedDoctors() []Doctor {
	return []Doctor{
		{1, "Сун Мин", "img/specimg1.png", "Главный врач-стоматолог, имплантолог, стоматолог-ортопед", "20 лет", "Имплантация и протезирование", "Профессор, выпускник Шанхайского университета, член Китайской стоматологической ассоциации.", []string{"Главный врач стоматологического центра «ЗУБАСТИК».", "Профессор, выпускник Шанхайского университета, член Китайской стоматологической ассоциации.", "Автор более 70 научных публикаций, в том числе 7 статей в журналах SCI.", "Руководил 5 исследовательскими проектами (2003–2019), удостоен трех провинциальных и университетских наград."}, "Дентальная имплантация, эстетическая реставрация, керамические коронки и вкладки, съемное и комбинированное протезирование, цифровое изготовление полных зубных протезов.", true},
		{2, "Екатерина Андреевна", "img/specimg2.png", "Заместитель главного врача, врач-стоматолог терапевт, хирург", "10 лет", "Терапевтическая и хирургическая стоматология", "Окончила Первый Московский государственный медицинский университет им. И.М. Сеченова. Врач высшей квалификационной категории. Член Стоматологической ассоциации России (СтАР).", []string{"Стаж работы — более 10 лет.", "Регулярный участник международных конгрессов по имплантологии и эндодонтии.", "Автор более 30 научных статей, соавтор методических пособий по лечению каналов корней зубов."}, "Лечение кариеса, пульпита и периодонтита с использованием микроскопа; сложное удаление зубов, резекция верхушек корней, пластика уздечек; комплексное планирование лечения.", true},
		{3, "Цао Ли", "img/specimg3.png", "Врач стоматолог-ортопед, специалист по протезированию", "8 лет", "Цифровое и ортопедическое протезирование", "Выпускница Шанхайского университета стоматологии. Сертифицированный специалист по цифровому протезированию.", []string{"Владеет современными методами восстановления зубов: от микропротезирования до сложного съемного и несъемного протезирования.", "Работает с материалами премиум-класса — цирконием и керамикой.", "Индивидуальный подход к каждому пациенту, акцент на естественную эстетику и комфорт."}, "Виниры, вкладки, коронки, съемное и несъемное протезирование, цифровое моделирование и эстетическое восстановление зубов.", true},
		{4, "Нурлан Аскарович", "img/specimg4.png", "Врач стоматолог-терапевт, пародонтолог", "7 лет", "Терапия, эндодонтия и пародонтология", "Окончил Московский государственный медико-стоматологический университет им. А.И. Евдокимова (МГМСУ). Член Стоматологической ассоциации России (СтАР).", []string{"Постоянно совершенствует навыки в области эндодонтии и пародонтологии, участвует в научно-практических конференциях.", "Стаж работы — 7 лет."}, "Терапевтическое лечение зубов любой сложности, эндодонтическое лечение, профессиональная гигиена, лечение гингивита и пародонтита, шинирование и профилактика заболеваний пародонта.", true},
		{5, "Константин Сергеевич", "img/specimg5.png", "Врач стоматолог-универсал, имплантолог, ортодонт", "10 лет", "Имплантация, ортодонтия и комплексное лечение", "Окончил Московский государственный медико-стоматологический университет им. А.И. Евдокимова (МГМСУ).", []string{"В своей работе сочетает терапевтический, хирургический и ортодонтический подходы.", "Проводит установку имплантатов с последующим протезированием, исправляет прикус брекетами и элайнерами, выполняет эстетические реставрации.", "Ведёт пациентов комплексно — от первичной консультации до полного восстановления улыбки."}, "Имплантация, протезирование, ортодонтия, брекет-системы, элайнеры и эстетическая реставрация.", true},
		{6, "Алекс Томасович", "img/specimg6.png", "Врач стоматолог-ортопед, CAD/CAM протезирование", "8 лет", "Цифровое протезирование CAD/CAM", "Окончил Московский государственный медико-стоматологический университет им. А.И. Евдокимова (МГМСУ). Член СтАР. Прошёл специализированное обучение по цифровому протезированию и CAD/CAM в Европейском центре стоматологических технологий.", []string{"Специализируется на цифровом моделировании и создании ортопедических конструкций с использованием 3D-технологий.", "Работает с материалами премиум-класса и добивается максимальной точности и естественности результата."}, "Коронки, виниры, вкладки, мосты, безметалловая керамика, диоксид циркония, E-max, полное восстановление зубного ряда и эстетическое протезирование.", true},
		{7, "Пань Ин", "img/specimg7.png", "Врач стоматолог-ортопед, эстетический стоматолог", "10 лет", "Эстетическая стоматология и протезирование", "Окончил Московский государственный медико-стоматологический университет им. А.И. Евдокимова (МГМСУ). Постоянно повышает квалификацию по эстетической реставрации и цифровой стоматологии.", []string{"Постоянно повышает квалификацию по эстетической реставрации и протезированию, посещает международные мастер-классы по цифровой стоматологии.", "Особое внимание уделяет эстетике: цвету, форме, прозрачности и естественности улыбки.", "Работает с материалами премиум-класса и создаёт реставрации, неотличимые от натуральных зубов."}, "Коронки, виниры, люминиры, вкладки, мостовидные протезы из керамики и диоксида циркония, эстетическое восстановление улыбки.", true},
	}
}

func SeedDoctors() error {
	var count int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM doctors`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, d := range seedDoctors() {
		data, err := json.Marshal(d.Description)
		if err != nil {
			return err
		}
		_, err = DB.Exec(`
INSERT INTO doctors (id,name,photo,position,experience,specialization,education,description,specialization_full,is_active)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (id) DO UPDATE SET
name=EXCLUDED.name, photo=EXCLUDED.photo, position=EXCLUDED.position,
experience=EXCLUDED.experience, specialization=EXCLUDED.specialization,
education=EXCLUDED.education, description=EXCLUDED.description,
specialization_full=EXCLUDED.specialization_full, is_active=EXCLUDED.is_active,
updated_at=NOW()`, d.ID, d.Name, d.Photo, d.Position, d.Experience, d.Specialization, d.Education, data, d.SpecializationFull, d.IsActive)
		if err != nil {
			return err
		}
	}
	_, err := DB.Exec(`SELECT setval(pg_get_serial_sequence('doctors','id'), COALESCE((SELECT MAX(id) FROM doctors),1), true)`)
	if err != nil {
		return err
	}
	// Initial demo schedule: weekdays 09:00-18:00. Existing administrator changes are preserved.
	for doctorID := 1; doctorID <= 7; doctorID++ {
		for weekday := 1; weekday <= 5; weekday++ {
			_, err = DB.Exec(`INSERT INTO doctor_schedule (doctor_id, weekday, is_working, start_time, end_time) VALUES ($1,$2,TRUE,'09:00','18:00') ON CONFLICT (doctor_id,weekday) DO NOTHING`, doctorID, weekday)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetDoctors(activeOnly bool) ([]Doctor, error) {
	query := `SELECT id,name,photo,position,experience,specialization,education,description,specialization_full,is_active FROM doctors`
	if activeOnly {
		query += ` WHERE is_active=true`
	}
	query += ` ORDER BY id`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Doctor
	for rows.Next() {
		var d Doctor
		var raw []byte
		if err := rows.Scan(&d.ID, &d.Name, &d.Photo, &d.Position, &d.Experience, &d.Specialization, &d.Education, &raw, &d.SpecializationFull, &d.IsActive); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &d.Description); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func GetDoctor(id int) (Doctor, error) {
	var d Doctor
	var raw []byte
	err := DB.QueryRow(`SELECT id,name,photo,position,experience,specialization,education,description,specialization_full,is_active FROM doctors WHERE id=$1`, id).Scan(&d.ID, &d.Name, &d.Photo, &d.Position, &d.Experience, &d.Specialization, &d.Education, &raw, &d.SpecializationFull, &d.IsActive)
	if err != nil {
		return d, err
	}
	err = json.Unmarshal(raw, &d.Description)
	return d, err
}

func CreateDoctor(p DoctorPayload) (Doctor, error) {
	data, err := json.Marshal(p.Description)
	if err != nil {
		return Doctor{}, err
	}

	var id int
	err = DB.QueryRow(`
		INSERT INTO doctors
			(name, photo, position, experience, specialization, education, description, specialization_full, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,TRUE)
		RETURNING id
	`, p.Name, p.Photo, p.Position, p.Experience, p.Specialization, p.Education, data, p.SpecializationFull).Scan(&id)
	if err != nil {
		return Doctor{}, err
	}

	return GetDoctor(id)
}

func UpdateDoctor(id int, p DoctorPayload) (Doctor, error) {
	data, err := json.Marshal(p.Description)
	if err != nil {
		return Doctor{}, err
	}

	result, err := DB.Exec(`
		UPDATE doctors
		SET name=$1,
			photo=$2,
			position=$3,
			experience=$4,
			specialization=$5,
			education=$6,
			description=$7,
			specialization_full=$8,
			updated_at=NOW()
		WHERE id=$9
	`, p.Name, p.Photo, p.Position, p.Experience, p.Specialization,
		p.Education, data, p.SpecializationFull, id)
	if err != nil {
		return Doctor{}, err
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return Doctor{}, sql.ErrNoRows
	}

	return GetDoctor(id)
}

func DeactivateDoctor(id int) error {
	result, err := DB.Exec(`
		UPDATE doctors
		SET is_active=FALSE, updated_at=NOW()
		WHERE id=$1
	`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	// Keep the linked doctor account in sync with the clinic status.
	_, _ = DB.Exec(`UPDATE users SET is_active=FALSE, updated_at=NOW() WHERE doctor_id=$1 AND role='DOCTOR'`, id)
	return nil
}

func ActivateDoctor(id int) error {
	result, err := DB.Exec(`
		UPDATE doctors
		SET is_active=TRUE, updated_at=NOW()
		WHERE id=$1
	`, id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	_, _ = DB.Exec(`UPDATE users SET is_active=TRUE, updated_at=NOW() WHERE doctor_id=$1 AND role='DOCTOR'`, id)
	return nil
}

var scheduleDayNames = map[int]string{1: "Понедельник", 2: "Вторник", 3: "Среда", 4: "Четверг", 5: "Пятница", 6: "Суббота", 7: "Воскресенье"}

func GetDoctorSchedule(doctorID int) (DoctorSchedule, error) {
	result := DoctorSchedule{DoctorID: doctorID, Days: make([]DoctorScheduleDay, 0, 7)}
	rows, err := DB.Query(`
		SELECT weekday, is_working, TO_CHAR(start_time, 'HH24:MI'), TO_CHAR(end_time, 'HH24:MI')
		FROM doctor_schedule WHERE doctor_id=$1 ORDER BY weekday
	`, doctorID)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var d DoctorScheduleDay
		if err := rows.Scan(&d.Weekday, &d.IsWorking, &d.StartTime, &d.EndTime); err != nil {
			return result, err
		}
		d.DayName = scheduleDayNames[d.Weekday]
		result.Days = append(result.Days, d)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	// Always return all seven days, including unsaved days.
	byDay := make(map[int]DoctorScheduleDay, len(result.Days))
	for _, d := range result.Days {
		byDay[d.Weekday] = d
	}
	result.Days = make([]DoctorScheduleDay, 0, 7)
	for day := 1; day <= 7; day++ {
		if d, ok := byDay[day]; ok {
			result.Days = append(result.Days, d)
			continue
		}
		working := day <= 5
		result.Days = append(result.Days, DoctorScheduleDay{Weekday: day, DayName: scheduleDayNames[day], IsWorking: working, StartTime: "09:00", EndTime: "18:00"})
	}
	return result, nil
}

func SaveDoctorSchedule(schedule DoctorSchedule) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, day := range schedule.Days {
		if day.Weekday < 1 || day.Weekday > 7 {
			return fmt.Errorf("invalid weekday")
		}
		start := day.StartTime
		end := day.EndTime
		if start == "" {
			start = "09:00"
		}
		if end == "" {
			end = "18:00"
		}
		_, err = tx.Exec(`
			INSERT INTO doctor_schedule (doctor_id, weekday, is_working, start_time, end_time)
			VALUES ($1,$2,$3,$4::time,$5::time)
			ON CONFLICT (doctor_id, weekday) DO UPDATE SET
				is_working=EXCLUDED.is_working, start_time=EXCLUDED.start_time, end_time=EXCLUDED.end_time
		`, schedule.DoctorID, day.Weekday, day.IsWorking, start, end)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func UpdateDoctorPhoto(id int, photo string) error {
	result, err := DB.Exec(`UPDATE doctors SET photo=$1, updated_at=NOW() WHERE id=$2`, photo, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

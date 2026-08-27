package main

import "errors"

func CheckAdmin(
	login string,
	password string,
) (Admin, error) {
	var admin Admin
	err := DB.QueryRow(`
		SELECT
		id,
		login
		FROM admins
		WHERE login=$1
		AND password=$2
	`,
		login,
		password,
	).Scan(
		&admin.ID,
		&admin.Login,
	)
	if err != nil {
		return admin, errors.New("Неверный логин или пароль!")
	}
	return admin, nil
}

package main

import (
	"fmt"
	"net/http"
)

func AuthMiddleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		cookie, err := r.Cookie("admin_session")
		fmt.Println("PATH:", r.URL.Path)
		if err != nil {
			fmt.Println("COOKIE ERROR:", err)
			http.Redirect(
				w,
				r,
				"/?account=admin",
				http.StatusSeeOther,
			)
			return
		}
		fmt.Println(
			"COOKIE VALUE:",
			cookie.Value,
		)
		if cookie.Value != "true" {
			fmt.Println("WRONG COOKIE")
			http.Redirect(
				w,
				r,
				"/?account=admin",
				http.StatusSeeOther,
			)
			return
		}
		next.ServeHTTP(
			w,
			r,
		)
	})
}

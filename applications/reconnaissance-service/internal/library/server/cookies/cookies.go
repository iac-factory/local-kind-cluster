package cookies

import (
	"net/http"
	"os"
	"strings"
	"time"
)

type Options struct {
	Domain string
	Secure bool
	HTTP   bool          // HTTP represents HTTP-Only Cookie settings
	Site   http.SameSite // Site represents Same-Site cookie settings
}

type Variadic func(o *Options)

func settings() *Options {
	return &Options{}
}

func Secure(w http.ResponseWriter, name, value string, options ...Variadic) {
	var o = settings()
	for _, option := range options {
		option(o)
	}

	domain := o.Domain
	if domain == "" {
		if v := os.Getenv("CI"); strings.ToLower(v) != "true" {
			domain = ""
		} else if v := os.Getenv("NAMESPACE"); v == "development" {
			domain = ""
		}
	}

	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(3 * time.Hour),
		MaxAge:   86400,
		Secure:   true,                    // Ensure the cookie is sent only over HTTPS
		HttpOnly: true,                    // Prevent JavaScript from accessing the cookie
		SameSite: http.SameSiteStrictMode, // Enforce SameSite policy
	}

	http.SetCookie(w, &cookie)
}

func New(w http.ResponseWriter, name, value string, options ...Variadic) {
	var o = settings()
	for _, option := range options {
		option(o)
	}

	domain := o.Domain
	if domain == "" {
		if v := os.Getenv("CI"); strings.ToLower(v) != "true" {
			domain = ""
		} else if v := os.Getenv("NAMESPACE"); v == "development" {
			domain = ""
		}
	}

	secure := o.Secure
	httponly := o.HTTP
	samesite := o.Site

	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(3 * time.Hour),
		MaxAge:   86400,
		Secure:   secure,
		HttpOnly: httponly,
		SameSite: samesite,
	}

	http.SetCookie(w, &cookie)
}

func Delete(w http.ResponseWriter, name string, options ...Variadic) {
	var o = settings()
	for _, option := range options {
		option(o)
	}

	domain := o.Domain
	if domain == "" {
		if v := os.Getenv("CI"); strings.ToLower(v) != "true" {
			domain = ""
		} else if v := os.Getenv("NAMESPACE"); v == "development" {
			domain = ""
		}
	}

	secure := o.Secure
	httponly := o.HTTP
	samesite := o.Site

	cookie := http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now(),
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: httponly,
		SameSite: samesite,
	}

	http.SetCookie(w, &cookie)
}

package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/stretchr/objx"

	"secret"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse()

	googleKey := secret.GoogleKey()
	googlePass := secret.GooglePass()

	gomniauth.SetSecurityKey("PUT YOUR AUTH KEY HERE")
	gomniauth.WithProviders(
		facebook.New("key", "secret", "http://localhost:8080/auth/callback/facebook"),
		github.New("key", "secret", "http://localhost:8080/auth/callback/github"),
		google.New(googleKey, googlePass, "http://localhost:8080/auth/callback/google"),
	)
	r := newRoom()
	//http.Handle("/assets/", http.StripPrefix("/asstes", http.FileServer(http.Dir("/path/to/assets"))))
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	go r.run()

	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

package handlers

import (
	"bytes"
	"html/template"
	"net/http"
)

func renderTemplate(w http.ResponseWriter, tmplFile string, data any) {
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		render500(w)
		return
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		render500(w)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	output.WriteTo(w)
}

func render404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	tmpl, err := template.ParseFiles("templates/404.html")
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	tmpl.Execute(w, nil)
}

func render500(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	tmpl, err := template.ParseFiles("templates/500.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func render405(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	tmpl, err := template.ParseFiles("templates/405.html")
	if err != nil {
		http.Error(w, "Method not Allowed", http.StatusMethodNotAllowed)
	}
	tmpl.Execute(w, nil)
}

package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PageData struct {
	Title       string
	CurrentPage string
}

// Diretório onde ficam os arquivos com conflito
const conflictDir = "conflicts"

func renderTemplate(w http.ResponseWriter, page string, data PageData) {
	t, err := template.ParseFiles(
		"templates/base.html",
		"templates/"+page+".html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template parse error:", err)
		return
	}

	err = t.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Template execute error:", err)
	}
}

func testeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Teste Widget",
		CurrentPage: "teste-widget",
	}
	renderTemplate(w, "teste-widget", data)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := PageData{
		Title:       "Home - Monaco Editor Site",
		CurrentPage: "home",
	}
	renderTemplate(w, "home", data)
}

func editorHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Editor - Monaco Editor Site",
		CurrentPage: "editor",
	}
	renderTemplate(w, "editor", data)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Sobre - Monaco Editor Site",
		CurrentPage: "about",
	}
	renderTemplate(w, "about", data)
}

func diffHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Diff - Monaco Editor Site",
		CurrentPage: "diff",
	}
	renderTemplate(w, "diff", data)
}

// -------------------------------------------------------------
// API: lista arquivos no diretório de conflitos
// -------------------------------------------------------------
func apiListFiles(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(conflictDir)
	if err != nil {
		// Se o diretório não existe, retorna lista vazia
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	if files == nil {
		files = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// -------------------------------------------------------------
// API: retorna conteúdo de um arquivo
// -------------------------------------------------------------
func apiReadFile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "parâmetro 'name' ausente", http.StatusBadRequest)
		return
	}

	// Segurança: impede travessia de diretório
	if strings.Contains(name, "..") || strings.Contains(name, "/") {
		http.Error(w, "nome inválido", http.StatusBadRequest)
		return
	}

	path := filepath.Join(conflictDir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "arquivo não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// -------------------------------------------------------------
// API: salva o resultado resolvido
// -------------------------------------------------------------
func apiSaveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.Name == "" {
		http.Error(w, "nome ausente", http.StatusBadRequest)
		return
	}

	// Segurança: impede travessia de diretório
	if strings.Contains(payload.Name, "..") || strings.Contains(payload.Name, "/") {
		http.Error(w, "nome inválido", http.StatusBadRequest)
		return
	}

	path := filepath.Join(conflictDir, payload.Name)
	err = os.WriteFile(path, []byte(payload.Content), 0644)
	if err != nil {
		http.Error(w, "erro ao salvar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// -------------------------------------------------------------
// API: cria arquivos de exemplo com conflitos
// -------------------------------------------------------------
func apiCreateExamples(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	os.MkdirAll(conflictDir, 0755)

	examples := map[string]string{
		"main.go": `package main

import (
	"fmt"
<<<<<<< HEAD
	"log"
=======
	"os"
>>>>>>> feature/login
)

func main() {
<<<<<<< HEAD
	fmt.Println("Hello, World!")
	log.Println("Iniciando servidor...")
=======
	fmt.Println("Hello, Monaco!")
	if len(os.Args) > 1 {
		fmt.Println("Argumento:", os.Args[1])
	}
>>>>>>> feature/login
}`,
		"config.json": `{
<<<<<<< HEAD
	"port": 8080,
	"host": "localhost",
	"debug": true
=======
	"port": 9090,
	"host": "0.0.0.0",
	"debug": false,
	"timeout": 30
>>>>>>> feature/login
}`,
		"styles.css": `body {
<<<<<<< HEAD
    background-color: #ffffff;
    font-family: Arial, sans-serif;
    color: #333333;
=======
    background-color: #1e1e1e;
    font-family: 'Courier New', monospace;
    color: #d4d4d4;
>>>>>>> feature/login
}

.container {
    max-width: 1200px;
    margin: 0 auto;
<<<<<<< HEAD
    padding: 20px;
=======
    padding: 40px 20px;
    border-radius: 8px;
>>>>>>> feature/login
}`,
	}

	for name, content := range examples {
		path := filepath.Join(conflictDir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			http.Error(w, "erro ao criar exemplos", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Rotas de página
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/editor", editorHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/diff", diffHandler)
	http.HandleFunc("/teste", testeHandler)

	// API
	http.HandleFunc("/api/files", apiListFiles)
	http.HandleFunc("/api/read", apiReadFile)
	http.HandleFunc("/api/save", apiSaveFile)
	http.HandleFunc("/api/examples", apiCreateExamples)

	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

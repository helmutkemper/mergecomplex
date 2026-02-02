package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitmerge/internal/git"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ErrorStr struct {
	Error string
}

type PageData struct {
	Title       string
	CurrentPage string
}

// Diretório onde ficam os arquivos com conflito
const conflictDir = "conflicts"

func renderTemplate(w http.ResponseWriter, page string, data PageData) {
	d, _ := os.ReadDir(".")
	for _, d := range d {
		log.Printf("Templates directory contains file: %s", d.Name())
	}

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

// todo: não deixar global
var globalControl *git.Control

func main() {

	globalControl = new(git.Control)
	globalControl.Init()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Rotas de página
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/editor", editorHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/diff", diffHandler)
	http.HandleFunc("/teste", testeHandler)

	// API
	http.HandleFunc("/api/save", apiSaveFile)
	http.HandleFunc("/api/examples", apiCreateExamples)
	http.HandleFunc("/api/files", filesHandler)
	http.HandleFunc("/api/read", readHandler)

	// Git endpoints
	http.HandleFunc("/git/branchs", gitBranchHandler)
	http.HandleFunc("/git/changes", getChanges)
	http.HandleFunc("/git/diff", getDiff)

	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setJsonHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func setError(w http.ResponseWriter, err error) {
	serverErr := new(ErrorStr)
	serverErr.Error = err.Error()
	data, _ := json.Marshal(serverErr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(data)
}

// getChanges retorna a lista de arquivos modificados entre duas branches
func getChanges(w http.ResponseWriter, r *http.Request) {
	setJsonHeaders(w)

	if !globalControl.IsInitialized() {
		setError(
			w,
			errors.Join(
				fmt.Errorf("no git control found"),
				fmt.Errorf("please, use the endpoint /git/branchs?dir=/absolute/path"),
			),
		)
		return
	}

	yourBranch := r.URL.Query().Get("yourBranch")
	if yourBranch == "" {
		setError(w, fmt.Errorf("yourBranch not provided"))
		return
	}

	baseBranch := r.URL.Query().Get("baseBranch")
	if baseBranch == "" {
		setError(w, fmt.Errorf("baseBranch not provided"))
		return
	}

	// Retorna apenas a lista de arquivos modificados
	list, err := globalControl.GetModifiedFiles(yourBranch, baseBranch)
	if err != nil {
		setError(w, err)
		return
	}

	data, _ := json.Marshal(&list)
	_, _ = w.Write(data)
}

// getDiff retorna o diff de um arquivo específico com marcadores de conflito
func getDiff(w http.ResponseWriter, r *http.Request) {
	if !globalControl.IsInitialized() {
		setError(
			w,
			errors.Join(
				fmt.Errorf("no git control found"),
				fmt.Errorf("please, use the endpoint /git/branchs?dir=/absolute/path"),
			),
		)
		return
	}

	yourBranch := r.URL.Query().Get("yourBranch")
	if yourBranch == "" {
		setError(w, fmt.Errorf("yourBranch not provided"))
		return
	}

	baseBranch := r.URL.Query().Get("baseBranch")
	if baseBranch == "" {
		setError(w, fmt.Errorf("baseBranch not provided"))
		return
	}

	file := r.URL.Query().Get("file")
	if file == "" {
		setError(w, fmt.Errorf("file not provided"))
		return
	}

	// Obtém o diff do arquivo específico
	diffMap, err := globalControl.DiffSpecificFile(yourBranch, baseBranch, file)
	if err != nil {
		setError(w, err)
		return
	}

	// Retorna o conteúdo do diff como texto simples
	diff, exists := diffMap[file]
	if !exists {
		setError(w, fmt.Errorf("file not found in diff"))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(diff))
}

// gitBranchHandler Retorna a lista de todos os branchs do projeto.
//
//	Exemplo: http://localhost:8080/git/branchs?dir=/Users/kemper/go/kemper/gitMerge/testgit
func gitBranchHandler(w http.ResponseWriter, r *http.Request) {
	setJsonHeaders(w)

	dir := r.URL.Query().Get("dir")
	if dir == "" {
		setError(w, fmt.Errorf("no dir provided"))
		return
	}

	// Valida se o diretório existe
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		setError(w, fmt.Errorf("dir does not exist"))
		return
	}

	if err := globalControl.NewRepoLocal(dir); err != nil {
		setError(w, err)
		return
	}

	list, err := globalControl.ListAllBranches()
	if err != nil {
		setError(w, err)
		return
	}

	data, _ := json.Marshal(&list)
	_, _ = w.Write(data)
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	outputDir := r.URL.Query().Get("dir")
	if outputDir == "" {
		outputDir = "./output" // diretório padrão
	}

	var files []string

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignora diretórios
		if info.IsDir() {
			return nil
		}

		// Obtém caminho relativo ao outputDir
		relPath, err := filepath.Rel(outputDir, path)
		if err != nil {
			return err
		}

		// Normaliza para usar "/" em todos os SOs
		files = append(files, filepath.ToSlash(relPath))
		return nil
	})

	if err != nil {
		http.Error(w, "Erro ao listar arquivos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	outputDir := "./output"
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Parâmetro 'name' não informado", http.StatusBadRequest)
		return
	}

	// Garante que o caminho não escape do outputDir
	fullPath := filepath.Join(outputDir, filepath.FromSlash(name))
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(outputDir)) {
		http.Error(w, "Caminho inválido", http.StatusBadRequest)
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "Arquivo não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

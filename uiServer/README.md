# Monaco Editor Site com Go

Um site demonstrativo que integra o Monaco Editor da Microsoft com um servidor web em Go (Golang).

## 🚀 Características

- **Monaco Editor**: Editor de código completo (o mesmo do VS Code)
- **Servidor Go**: Backend leve e performático
- **Menu Vertical**: Navegação intuitiva com ícones Font Awesome
- **Design Responsivo**: Interface adaptável a diferentes tamanhos de tela
- **Open Source**: Apenas tecnologias livres e gratuitas

## 📋 Pré-requisitos

- Go 1.16 ou superior
- Navegador web moderno

## 🔧 Instalação e Execução

1. Clone ou baixe o projeto
2. Navegue até o diretório do projeto:
   ```bash
   cd golang-monaco-site
   ```
3. Execute o servidor:
   ```bash
   go run main.go
   ```
4. Acesse no navegador: `http://localhost:8080`

## 📁 Estrutura do Projeto

```
golang-monaco-site/
├── main.go              # Servidor Go principal
├── templates/           # Templates HTML
│   ├── base.html       # Template base com menu
│   ├── home.html       # Página inicial
│   ├── editor.html     # Página do editor
│   └── about.html      # Página sobre
├── static/             # Arquivos estáticos
│   └── css/
│       └── style.css   # Estilos CSS
└── README.md           # Este arquivo
```

## 🎯 Funcionalidades

### Home
- Visão geral do projeto
- Lista de características
- Tecnologias utilizadas

### Editor
- Monaco Editor totalmente funcional
- Suporte a múltiplas linguagens (JavaScript, Python, Go, HTML, CSS, etc.)
- Troca de temas (Light, Dark, High Contrast)
- Syntax highlighting
- Autocomplete

### Sobre
- Informações detalhadas sobre o projeto
- Documentação do Monaco Editor
- Instruções de uso

## 🛠️ Tecnologias Utilizadas

- **Go (Golang)** - Backend server
- **Monaco Editor** - Editor de código
- **Font Awesome Free** - Ícones
- **HTML5 & CSS3** - Frontend

## 📝 Licença

Este projeto utiliza tecnologias open source:
- Go: Licença BSD
- Monaco Editor: Licença MIT
- Font Awesome Free: Licença SIL OFL 1.1 / MIT

## 🔗 Links Úteis

- [Go Documentation](https://go.dev/doc/)
- [Monaco Editor GitHub](https://github.com/Microsoft/monaco-editor)
- [Font Awesome](https://fontawesome.com/)

## 👨‍💻 Desenvolvimento

Para compilar o projeto:
```bash
go build -o monaco-site main.go
```

Para executar o binário compilado:
```bash
./monaco-site
```

## 🌐 Portas

- Servidor: `8080` (padrão)
- Modifique em `main.go` se necessário

package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

func main() {
	// Configurações - ajuste conforme necessário
	repoPath := "."      // Diretório do repositório (use "." para diretório atual)
	baseBranch := "main" // Branch base para comparação

	// Abre o repositório Git local
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		fmt.Printf("Erro ao abrir repositório: %v\n", err)
		fmt.Println("Certifique-se de que o diretório é um repositório Git válido.")
		return
	}

	// Obtém a branch atual (HEAD)
	head, err := repo.Head()
	if err != nil {
		fmt.Printf("Erro ao obter HEAD: %v\n", err)
		return
	}

	// Extrai o nome da branch atual
	currentBranch := head.Name().Short()

	fmt.Printf("Analisando repositório local: %s\n", repoPath)
	fmt.Printf("Branch atual (detectada): %s\n", currentBranch)
	fmt.Printf("Branch base: %s\n\n", baseBranch)

	// Verifica se estamos na branch base
	if currentBranch == baseBranch {
		fmt.Printf("Você está atualmente na branch base (%s).\n", baseBranch)
		fmt.Println("Mude para a branch que deseja analisar usando 'git checkout <branch>'")
		return
	}

	// Obtém a referência da branch base
	baseRef, err := repo.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", baseBranch)), true)
	if err != nil {
		fmt.Printf("Erro ao obter referência da branch %s: %v\n", baseBranch, err)
		fmt.Println("Verifique se a branch base existe.")
		return
	}

	// Obtém os commits
	currentCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		fmt.Printf("Erro ao obter commit da branch atual: %v\n", err)
		return
	}

	baseCommit, err := repo.CommitObject(baseRef.Hash())
	if err != nil {
		fmt.Printf("Erro ao obter commit da branch %s: %v\n", baseBranch, err)
		return
	}

	// Obtém as árvores de arquivos
	currentTree, err := currentCommit.Tree()
	if err != nil {
		fmt.Printf("Erro ao obter árvore da branch atual: %v\n", err)
		return
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		fmt.Printf("Erro ao obter árvore da branch %s: %v\n", baseBranch, err)
		return
	}

	// Compara as árvores e obtém as diferenças
	changes, err := currentTree.Diff(baseTree)
	if err != nil {
		fmt.Printf("Erro ao comparar árvores: %v\n", err)
		return
	}

	fmt.Printf("\n=== Arquivos alterados na branch '%s' em relação a '%s' ===\n\n", currentBranch, baseBranch)

	if len(changes) == 0 {
		fmt.Println("Nenhuma alteração encontrada.")
		return
	}

	added := 0
	modified := 0
	deleted := 0

	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}

		switch action {
		case merkletrie.Insert:
			fmt.Printf("[ADICIONADO] %s\n", change.To.Name)
			added++
		case merkletrie.Modify:
			fmt.Printf("[MODIFICADO] %s\n", change.To.Name)
			modified++
		case merkletrie.Delete:
			fmt.Printf("[REMOVIDO]   %s\n", change.From.Name)
			deleted++
		}
	}

	fmt.Printf("\n=== Resumo ===\n")
	fmt.Printf("Total de arquivos alterados: %d\n", len(changes))
	fmt.Printf("  - Adicionados: %d\n", added)
	fmt.Printf("  - Modificados: %d\n", modified)
	fmt.Printf("  - Removidos: %d\n", deleted)
}

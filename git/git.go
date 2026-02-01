package git

import (
	"errors"
	"fmt"
	"io"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

// FileChange representa uma mudança em um arquivo
type FileChange struct {
	Path   string // Caminho do arquivo
	Action string // "added", "modified", ou "deleted"
}

type Control struct {
	repository *git.Repository
	progress   io.Writer
}

func (e *Control) Init() {}

func (e *Control) NewRepoRemote(repoURL, localPath string) (err error) {
	//repoURL := "https://github.com/seu-usuario/seu-repositorio.git"
	//branchName := "feature-branch" // Nome da branch que deseja analisar
	//baseBranch := "main"            // Branch base para comparação
	//localPath := "./temp-repo" // Diretório temporário local

	// Clona o repositório
	e.repository, err = git.PlainClone(localPath, false, &git.CloneOptions{
		URL: repoURL,
		//ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
		//SingleBranch:  false, // Clona todas as branches para permitir comparação
		Progress: e.progress,
	})
	if err != nil {
		err = fmt.Errorf("erro ao clonar repositório: %v\n", err)
	}

	return
}

func (e *Control) NewRepoLocal(repoPath string) (err error) {
	// Abre o repositório Git local
	if e.repository, err = git.PlainOpen(repoPath); err != nil {
		err = fmt.Errorf("erro ao abrir repositório: %v\n", err)
		err = errors.Join(fmt.Errorf("certifique-se de que o diretório é um repositório git válido"))
	}

	return
}

// ListBranches retorna uma lista com os nomes de todas as branches do repositório.
// Retorna um slice de strings com os nomes das branches e um erro, se houver.
func (c *Control) ListBranches() ([]string, error) {
	if c.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	// Obtém todas as referências do repositório
	refs, err := c.repository.References()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referências: %w", err)
	}

	var branches []string

	// Itera sobre todas as referências
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Filtra apenas as branches (refs/heads/*)
		if ref.Name().IsBranch() {
			// Obtém o nome curto da branch (sem o prefixo refs/heads/)
			branchName := ref.Name().Short()
			branches = append(branches, branchName)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao iterar sobre as referências: %w", err)
	}

	return branches, nil
}

// ListAllBranches retorna uma lista com os nomes de todas as branches,
// incluindo branches remotas.
// Retorna um slice de strings com os nomes completos das branches e um erro, se houver.
func (c *Control) ListAllBranches() ([]string, error) {
	if c.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	refs, err := c.repository.References()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referências: %w", err)
	}

	var branches []string

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Inclui branches locais e remotas
		if ref.Name().IsBranch() || ref.Name().IsRemote() {
			branchName := ref.Name().Short()
			branches = append(branches, branchName)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao iterar sobre as referências: %w", err)
	}

	return branches, nil
}

// ListLocalBranches retorna apenas as branches locais.
// É um alias para ListBranches() para maior clareza.
func (c *Control) ListLocalBranches() ([]string, error) {
	return c.ListBranches()
}

// ListRemoteBranches retorna uma lista com os nomes de todas as branches remotas.
// Retorna um slice de strings com os nomes das branches remotas e um erro, se houver.
func (c *Control) ListRemoteBranches() ([]string, error) {
	if c.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	refs, err := c.repository.References()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referências: %w", err)
	}

	var branches []string

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Filtra apenas as branches remotas (refs/remotes/*)
		if ref.Name().IsRemote() {
			branchName := ref.Name().Short()
			branches = append(branches, branchName)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao iterar sobre as referências: %w", err)
	}

	return branches, nil
}

// GetModifiedFiles retorna todos os arquivos modificados ou adicionados em uma branch
// comparando com uma branch base.
// Retorna um slice de strings com os caminhos dos arquivos e um erro, se houver.
func (c *Control) GetModifiedFiles(branchName, baseBranch string) ([]string, error) {
	if c.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	// Obtém a referência da branch alvo
	targetRef, err := c.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := c.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", baseBranch)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", baseBranch, err)
	}

	// Obtém os commits
	targetCommit, err := c.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := c.repository.CommitObject(baseRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", baseBranch, err)
	}

	// Obtém as árvores de arquivos
	targetTree, err := targetCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchName, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", baseBranch, err)
	}

	// Compara as árvores e obtém as diferenças
	changes, err := targetTree.Diff(baseTree)
	if err != nil {
		return nil, fmt.Errorf("erro ao comparar árvores: %w", err)
	}

	var modifiedFiles []string

	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}

		// Inclui apenas arquivos modificados ou adicionados
		switch action {
		case merkletrie.Insert: // Arquivo adicionado
			modifiedFiles = append(modifiedFiles, change.To.Name)
		case merkletrie.Modify: // Arquivo modificado
			modifiedFiles = append(modifiedFiles, change.To.Name)
		}
	}

	return modifiedFiles, nil
}

// GetAllChangedFiles retorna todos os arquivos que sofreram qualquer tipo de mudança
// (adicionados, modificados ou removidos) em uma branch comparando com uma branch base.
// Retorna um slice de strings com os caminhos dos arquivos e um erro, se houver.
func (c *Control) GetAllChangedFiles(branchName, baseBranch string) ([]string, error) {
	if c.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	// Obtém a referência da branch alvo
	targetRef, err := c.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := c.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", baseBranch)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", baseBranch, err)
	}

	// Obtém os commits
	targetCommit, err := c.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := c.repository.CommitObject(baseRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", baseBranch, err)
	}

	// Obtém as árvores de arquivos
	targetTree, err := targetCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchName, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", baseBranch, err)
	}

	// Compara as árvores e obtém as diferenças
	changes, err := targetTree.Diff(baseTree)
	if err != nil {
		return nil, fmt.Errorf("erro ao comparar árvores: %w", err)
	}

	var changedFiles []string

	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}

		// Inclui todos os tipos de mudança
		switch action {
		case merkletrie.Insert:
			changedFiles = append(changedFiles, change.To.Name)
		case merkletrie.Modify:
			changedFiles = append(changedFiles, change.To.Name)
		case merkletrie.Delete:
			changedFiles = append(changedFiles, change.From.Name)
		}
	}

	return changedFiles, nil
}

// GetFileChanges retorna informações detalhadas sobre todos os arquivos alterados
// em uma branch comparando com uma branch base.
// Retorna um slice de FileChange com detalhes de cada alteração e um erro, se houver.
func (c *Control) GetFileChanges(branchName, baseBranch string) ([]FileChange, error) {
	if c.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	// Obtém a referência da branch alvo
	targetRef, err := c.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := c.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", baseBranch)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", baseBranch, err)
	}

	// Obtém os commits
	targetCommit, err := c.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := c.repository.CommitObject(baseRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", baseBranch, err)
	}

	// Obtém as árvores de arquivos
	targetTree, err := targetCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchName, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", baseBranch, err)
	}

	// Compara as árvores e obtém as diferenças
	changes, err := targetTree.Diff(baseTree)
	if err != nil {
		return nil, fmt.Errorf("erro ao comparar árvores: %w", err)
	}

	var fileChanges []FileChange

	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}

		var fc FileChange

		switch action {
		case merkletrie.Insert:
			fc = FileChange{
				Path:   change.To.Name,
				Action: "added",
			}
		case merkletrie.Modify:
			fc = FileChange{
				Path:   change.To.Name,
				Action: "modified",
			}
		case merkletrie.Delete:
			fc = FileChange{
				Path:   change.From.Name,
				Action: "deleted",
			}
		}

		fileChanges = append(fileChanges, fc)
	}

	return fileChanges, nil
}

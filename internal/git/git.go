package git

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

func (e *Control) IsInitialized() bool {
	return e.repository != nil
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
		err = fmt.Errorf("erro ao abrir repositório: %v", err)
		err = errors.Join(
			fmt.Errorf(" - "),
			fmt.Errorf("certifique-se de que o diretório é um repositório git válido"),
			fmt.Errorf(" - "),
			fmt.Errorf("diretório: %v", repoPath),
		)
	}

	return
}

// ListBranches retorna uma lista com os nomes de todas as branches do repositório.
// Retorna um slice de strings com os nomes das branches e um erro, se houver.
func (e *Control) ListBranches() ([]string, error) {
	if e.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	// Obtém todas as referências do repositório
	refs, err := e.repository.References()
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
func (e *Control) ListAllBranches() ([]string, error) {
	if e.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	refs, err := e.repository.References()
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
func (e *Control) ListLocalBranches() ([]string, error) {
	return e.ListBranches()
}

// ListRemoteBranches retorna uma lista com os nomes de todas as branches remotas.
// Retorna um slice de strings com os nomes das branches remotas e um erro, se houver.
func (e *Control) ListRemoteBranches() ([]string, error) {
	if e.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	refs, err := e.repository.References()
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
func (e *Control) GetModifiedFiles(branchName, branchBase string) ([]string, error) {
	// Obtém a referência da branch alvo
	targetRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchBase)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchBase, err)
	}

	// Obtém os commits
	targetCommit, err := e.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := e.repository.CommitObject(baseRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchBase, err)
	}

	// Obtém as árvores de arquivos
	targetTree, err := targetCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchName, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchBase, err)
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

// DiffOutputWithBranch compara os arquivos da pasta output com a versão
// dos mesmos arquivos na branch informada.
// Retorna um map onde a chave é o caminho do arquivo e o valor é o diff.
func (e *Control) DiffOutputWithBranch(branchName, outputDir string) (map[string]string, error) {
	// Obtém a referência da branch
	ref, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém o commit e a tree da branch
	commit, err := e.repository.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchName, err)
	}

	diffs := make(map[string]string)

	// Percorre os arquivos da pasta output
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignora diretórios
		if info.IsDir() {
			return nil
		}

		// Obtém o caminho relativo ao outputDir
		relPath, err := filepath.Rel(outputDir, path)
		if err != nil {
			return fmt.Errorf("erro ao obter caminho relativo de %s: %w", path, err)
		}

		// Normaliza o caminho para usar "/" (compatível com git)
		relPath = filepath.ToSlash(relPath)

		//if searchedFile != relPath {
		//	~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
		//}

		// Tenta encontrar o arquivo na branch
		branchFile, err := tree.File(relPath)
		if err != nil {
			// Arquivo não existe na branch, ignora
			return nil
		}

		// Lê o conteúdo do arquivo na branch
		branchContent, err := branchFile.Contents()
		if err != nil {
			return fmt.Errorf("erro ao ler conteúdo de %s na branch: %w", relPath, err)
		}

		// Lê o conteúdo do arquivo local (output)
		localContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("erro ao ler arquivo local %s: %w", path, err)
		}

		// Se os conteúdos são iguais, não há diferença
		if string(localContent) == branchContent {
			return nil
		}

		// Gera o diff entre os dois conteúdos
		diff := generateDiff(branchContent, string(localContent))
		diffs[relPath] = diff

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao percorrer diretório output: %w", err)
	}

	return diffs, nil
}

// generateDiff gera um diff linha por linha entre duas versões de um arquivo.
// branchContent é a versão da branch (antigO), localContent é a versão do output (novo).
// generateDiff gera um arquivo no formato de conflito git,
// compatível com o Diff Editor do frontend.
// branchContent = versão na branch remota (theirs)
// localContent  = versão no output local (ours / HEAD)
//func generateDiff(branchContent, localContent []byte) string {
//	branchLines := strings.Split(string(branchContent), "\n")
//	localLines := strings.Split(string(localContent), "\n")
//
//	var result strings.Builder
//
//	maxLen := len(branchLines)
//	if len(localLines) > maxLen {
//		maxLen = len(localLines)
//	}
//
//	for i := 0; i < maxLen; i++ {
//		branchLine := ""
//		localLine := ""
//
//		if i < len(branchLines) {
//			branchLine = branchLines[i]
//		}
//		if i < len(localLines) {
//			localLine = localLines[i]
//		}
//
//		switch {
//		case i >= len(branchLines):
//			// Linha só existe no local (adicionada no output)
//			result.WriteString(localLine + "\n")
//
//		case i >= len(localLines):
//			// Linha só existe na branch (removida no output)
//			result.WriteString(branchLine + "\n")
//
//		case branchLine == localLine:
//			// Linha igual nas duas versões
//			result.WriteString(localLine + "\n")
//
//		default:
//			// Linha diferente — gera bloco de conflito
//			result.WriteString("<<<<<<< HEAD\n")
//			result.WriteString(localLine + "\n")
//			result.WriteString("=======\n")
//			result.WriteString(branchLine + "\n")
//			result.WriteString(">>>>>>> branch\n")
//		}
//	}
//
//	return result.String()
//}

// GetAllChangedFiles retorna todos os arquivos que sofreram qualquer tipo de mudança
// (adicionados, modificados ou removidos) em uma branch comparando com uma branch base.
// Retorna um slice de strings com os caminhos dos arquivos e um erro, se houver.
func (e *Control) GetAllChangedFiles(branchName, baseBranch string) ([]string, error) {
	if e.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	// Obtém a referência da branch alvo
	targetRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", baseBranch)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", baseBranch, err)
	}

	// Obtém os commits
	targetCommit, err := e.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := e.repository.CommitObject(baseRef.Hash())
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
func (e *Control) GetFileChanges(branchName, baseBranch string) ([]FileChange, error) {
	if e.repository == nil {
		return nil, fmt.Errorf("repositório não inicializado")
	}

	// Obtém a referência da branch alvo
	targetRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", baseBranch)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", baseBranch, err)
	}

	// Obtém os commits
	targetCommit, err := e.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := e.repository.CommitObject(baseRef.Hash())
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

// DownloadModifiedFiles baixa os arquivos modificados entre duas branchs
// e os salva no diretório de destino, preservando a estrutura de pastas.
func (e *Control) DownloadModifiedFiles(branchName, branchBase, destDir string) ([]string, error) {
	_ = os.RemoveAll(destDir)

	// Obtém a referência da branch alvo
	targetRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchBase)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchBase, err)
	}

	// Obtém os commits
	targetCommit, err := e.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := e.repository.CommitObject(baseRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchBase, err)
	}

	// Obtém as árvores
	targetTree, err := targetCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchName, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchBase, err)
	}

	// Compara as árvores
	changes, err := targetTree.Diff(baseTree)
	if err != nil {
		return nil, fmt.Errorf("erro ao comparar árvores: %w", err)
	}

	var downloaded []string

	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}

		// Processa apenas arquivos adicionados ou modificados
		if action != merkletrie.Insert && action != merkletrie.Modify {
			continue
		}

		fileName := change.To.Name

		// Lê o conteúdo do arquivo na tree do commit alvo
		fileInTree, err := targetTree.File(fileName)
		if err != nil {
			return downloaded, fmt.Errorf("erro ao ler arquivo %s: %w", fileName, err)
		}

		content, err := fileInTree.Contents()
		if err != nil {
			return downloaded, fmt.Errorf("erro ao obter conteúdo de %s: %w", fileName, err)
		}

		// Monta o caminho de destino
		destPath := filepath.Join(destDir, fileName)

		// Cria as subpastas se necessário
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return downloaded, fmt.Errorf("erro ao criar diretório %s: %w", filepath.Dir(destPath), err)
		}

		// Salva o arquivo
		if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
			return downloaded, fmt.Errorf("erro ao salvar arquivo %s: %w", destPath, err)
		}

		downloaded = append(downloaded, destPath)
	}

	return downloaded, nil
}

// DiffSpecificFile retorna o diff de um arquivo específico entre duas branches
// com marcadores de conflito git.
func (e *Control) DiffSpecificFile(branchName, branchBase, fileName string) (map[string]string, error) {
	// Obtém a referência da branch alvo
	targetRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchName, err)
	}

	// Obtém a referência da branch base
	baseRef, err := e.repository.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchBase)), true)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter referência da branch %s: %w", branchBase, err)
	}

	// Obtém os commits
	targetCommit, err := e.repository.CommitObject(targetRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchName, err)
	}

	baseCommit, err := e.repository.CommitObject(baseRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("erro ao obter commit da branch %s: %w", branchBase, err)
	}

	// Obtém as árvores
	targetTree, err := targetCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchName, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter árvore da branch %s: %w", branchBase, err)
	}

	// Compara as árvores
	changes, err := targetTree.Diff(baseTree)
	if err != nil {
		return nil, fmt.Errorf("erro ao comparar árvores: %w", err)
	}

	// Procura o arquivo específico nos changes
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			continue
		}

		// Verifica se é o arquivo que estamos procurando
		if change.To.Name != fileName {
			continue
		}

		// Processa apenas arquivos adicionados ou modificados
		if action != merkletrie.Insert && action != merkletrie.Modify {
			return nil, fmt.Errorf("arquivo foi deletado ou não modificado")
		}

		// Lê o conteúdo do arquivo na branch base (branch remota/theirs)
		baseFile, err := baseTree.File(fileName)
		if err != nil {
			// Arquivo não existe na base, considera vazio
			baseFile = nil
		}

		var baseContent string
		if baseFile != nil {
			baseContent, err = baseFile.Contents()
			if err != nil {
				return nil, fmt.Errorf("erro ao obter conteúdo da branch base: %w", err)
			}
		}

		// Lê o conteúdo do arquivo na branch alvo (sua branch/ours)
		targetFile, err := targetTree.File(fileName)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler arquivo %s: %w", fileName, err)
		}

		targetContent, err := targetFile.Contents()
		if err != nil {
			return nil, fmt.Errorf("erro ao obter conteúdo da sua branch: %w", err)
		}

		// Gera o diff com marcadores de conflito
		diff := generateDiff(baseContent, targetContent)

		result := make(map[string]string)
		result[fileName] = diff
		return result, nil
	}

	return nil, fmt.Errorf("arquivo %s não encontrado nas diferenças entre as branches", fileName)
}

// generateDiff gera um arquivo no formato de conflito git.
// branchContent = versão na branch base/remota (theirs)
// localContent  = versão na sua branch (ours / HEAD)
func generateDiff(branchContent, localContent string) string {
	branchLines := strings.Split(branchContent, "\n")
	localLines := strings.Split(localContent, "\n")

	var result strings.Builder

	maxLen := len(branchLines)
	if len(localLines) > maxLen {
		maxLen = len(localLines)
	}

	for i := 0; i < maxLen; i++ {
		branchLine := ""
		localLine := ""

		if i < len(branchLines) {
			branchLine = branchLines[i]
		}
		if i < len(localLines) {
			localLine = localLines[i]
		}

		switch {
		case i >= len(branchLines):
			// Linha só existe no local (adicionada)
			result.WriteString(localLine + "\n")

		case i >= len(localLines):
			// Linha só existe na branch (removida)
			result.WriteString(branchLine + "\n")

		case branchLine == localLine:
			// Linha igual nas duas versões
			result.WriteString(localLine + "\n")

		default:
			// Linha diferente — gera bloco de conflito
			result.WriteString("<<<<<<< HEAD\n")
			result.WriteString(localLine + "\n")
			result.WriteString("=======\n")
			result.WriteString(branchLine + "\n")
			result.WriteString(">>>>>>> branch\n")
		}
	}

	return result.String()
}

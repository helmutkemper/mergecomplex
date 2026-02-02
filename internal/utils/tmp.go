package utils

import (
	"fmt"
	"os"
)

func CreateTempDir(prefix string) (tempDir string, err error) {
	if tempDir, err = os.MkdirTemp("", prefix); err != nil {
		err = fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}

	return
}

func CreateTempDirIn(parentDir, prefix string) (tempDir string, err error) {
	if tempDir, err = os.MkdirTemp(parentDir, prefix); err != nil {
		err = fmt.Errorf("erro ao criar diretório temporário em %s: %w", parentDir, err)
	}

	return
}

func RemoveTempDir(path string) (err error) {
	if err = os.RemoveAll(path); err != nil {
		err = fmt.Errorf("erro ao remover diretório temporário %s: %w", path, err)
	}
	return nil
}

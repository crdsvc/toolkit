package toolkit

import (
	"fmt"
	"os"
)

func (t *Tools) CreateDirIfNotExist(path string) error {
	const mode = 0o755
	if path == "" {
		return fmt.Errorf("empty path provided")
	}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, mode)
		if err != nil {
			return fmt.Errorf("error with mkdirall %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("pther error with stat %w", err)
	}
	return nil

}

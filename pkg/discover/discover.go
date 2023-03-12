package discover

import (
	"context"
	"errors"
	"os"
	"path"
	"strings"
)

var (
	InvalidShuttlePathFile = errors.New("shuttle path did not point ot a shuttle.yaml file")
)

const (
	shuttletaskdir  = "shuttletask"
	shuttlefilename = "shuttle.yaml"
)

type ShuttleTaskDiscovered struct {
	Files     []string
	DirPath   string
	ParentDir string
}

type Discovered struct {
	Local *ShuttleTaskDiscovered
	Plan  *ShuttleTaskDiscovered
}

// path: is a path to the shuttle.yaml file
// It will always look for the shuttletask directory relative to the shuttle.yaml file
//
// 1. Traverse shuttletaskdir
//
// 2. Traverse plan if exists (only 1 layer for now)
//
// 3. Collect file names
//
// 4. Return list of files to move to tmp dir
func Discover(ctx context.Context, shuttlepath string) (*Discovered, error) {
	if !strings.HasSuffix(shuttlepath, shuttlefilename) {
		return nil, InvalidShuttlePathFile
	}
	if _, err := os.Stat(shuttlepath); errors.Is(err, os.ErrNotExist) {
		return nil, InvalidShuttlePathFile
	}

	localdir := path.Dir(shuttlepath)
	localshuttledirentries := make([]string, 0)

	shuttletaskpath := path.Join(localdir, shuttletaskdir)
	if fs, err := os.Stat(shuttletaskpath); err == nil {
		// list all local files
		if fs.IsDir() {
			entries, err := os.ReadDir(shuttletaskpath)
			if err != nil {
				return nil, err
			}

			for _, entry := range entries {
				// skip dirs
				if entry.IsDir() {
					continue
				}

				// skip non go files
				if !strings.HasSuffix(entry.Name(), ".go") {
					continue
				}

				// skip test files
				if strings.HasSuffix(entry.Name(), "test.go") {
					continue
				}

				localshuttledirentries = append(localshuttledirentries, entry.Name())
			}
		}
	}

	discovered := Discovered{
		Local: &ShuttleTaskDiscovered{
			DirPath:   shuttletaskpath,
			Files:     localshuttledirentries,
			ParentDir: localdir,
		},
	}

	//TODO: Add plan as well

	return &discovered, nil
}

package compile

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"text/template"

	"github.com/kjuulh/shuttletask/pkg/discover"
	cp "github.com/otiai10/copy"
	"golang.org/x/mod/sumdb/dirhash"
)

var (
	//go:embed templates/mainFile.tmpl
	mainFileTmpl embed.FS
)

// discovered: Discovered shuttletask projects
//
// 1. Check hash for each dir
//
// 2. Compile for each discovered dir
//
// 2.1. Copy to tmp dir
//
// 2.2. Generate main file
//
// 3. Move binary to .shuttle/shuttletask/binary-<hash>
func Compile(ctx context.Context, discovered *discover.Discovered) error {
	compile(ctx, discovered.Local)

	return nil
}

func compile(ctx context.Context, shuttletask *discover.ShuttleTaskDiscovered) error {
	hash, err := getHash(ctx, shuttletask)
	if err != nil {
		return err
	}

	ok, err := binaryMatches(ctx, hash, shuttletask)
	if err != nil {
		return err
	}

	if ok {
		// The binary is the same so we short circuit
		return nil
	}

	shuttlelocaldir := path.Join(shuttletask.ParentDir, ".shuttle/shuttletask")

	// Generate tmp dir
	if err = generateTmpDir(ctx, shuttlelocaldir); err != nil {
		return err
	}
	// Copy files
	if err = copyFiles(ctx, shuttlelocaldir, shuttletask); err != nil {
		return err
	}

	// Generate AST
	// Generate Main file
	if err = generateMainFile(ctx, shuttlelocaldir, shuttletask); err != nil {
		return err
	}

	// Compile package
	if err = modTidy(ctx, shuttlelocaldir); err != nil {
		return err
	}
	if err = compileBinary(ctx, shuttlelocaldir); err != nil {
		return err
	}
	// Move binary

	return nil
}

func modTidy(ctx context.Context, shuttlelocaldir string) error {
	log.Println("go mod tidy")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = path.Join(shuttlelocaldir, "tmp")

	output, err := cmd.CombinedOutput()
	if err != nil {
		println("%s", string(output))
		return err
	}

	println("%s", string(output))

	return nil
}

func compileBinary(ctx context.Context, shuttlelocaldir string) error {
	log.Println("compiling binary")
	cmd := exec.Command("go", "build")
	cmd.Dir = path.Join(shuttlelocaldir, "tmp")

	output, err := cmd.CombinedOutput()
	if err != nil {
		println("%s", string(output))
		return err
	}

	println("%s", string(output))

	return nil
}

func generateMainFile(ctx context.Context, shuttlelocaldir string, shuttletask *discover.ShuttleTaskDiscovered) error {
	tmpl, err := template.ParseFS(mainFileTmpl, "templates/mainFile.tmpl")
	if err != nil {
		return err
	}

	tmpmainfile := path.Join(shuttlelocaldir, "tmp/main.go")

	file, err := os.Create(tmpmainfile)
	if err != nil {
		return err
	}

	return tmpl.Execute(file, nil)
}

func copyFiles(ctx context.Context, shuttlelocaldir string, shuttletask *discover.ShuttleTaskDiscovered) error {
	tmpdir := path.Join(shuttlelocaldir, "tmp")

	return cp.Copy(shuttletask.DirPath, tmpdir)
}

func generateTmpDir(ctx context.Context, shuttlelocaldir string) error {
	if err := os.MkdirAll(shuttlelocaldir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join(shuttlelocaldir, "binaries"), 0755); err != nil {
		return err
	}

	tmpdir := path.Join(shuttlelocaldir, "tmp")
	if err := os.RemoveAll(tmpdir); err != nil {
		return nil
	}
	if err := os.MkdirAll(tmpdir, 0755); err != nil {
		return err
	}

	return nil
}

func binaryMatches(ctx context.Context, hash string, shuttletask *discover.ShuttleTaskDiscovered) (bool, error) {
	shuttlebindir := path.Join(shuttletask.ParentDir, ".shuttle/shuttletask/binaries")

	if _, err := os.Stat(shuttlebindir); errors.Is(err, os.ErrNotExist) {
		log.Println("DEBUG: package doesn't exist continueing")
		return false, nil
	}

	entries, err := os.ReadDir(shuttlebindir)
	if err != nil {
		return false, err
	}

	if len(entries) == 0 {
		return false, err
	}

	// We only expect a single binary in the folder, so we just take the first entry if it exists
	binary := entries[0]

	if binary.Name() == fmt.Sprintf("shuttletask-%s.go", hash) {
		return true, nil
	} else {
		return false, nil
	}
}

func getHash(ctx context.Context, shuttletask *discover.ShuttleTaskDiscovered) (string, error) {
	entries := make([]string, len(shuttletask.Files))

	for i, task := range shuttletask.Files {
		entries[i] = path.Join(shuttletask.DirPath, task)
	}

	open := func(name string) (io.ReadCloser, error) {
		b, err := os.ReadFile(name)
		if err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(b)), nil
	}

	hash, err := dirhash.Hash1(entries, open)
	if err != nil {
		return "", err
	}

	return hash, nil
}

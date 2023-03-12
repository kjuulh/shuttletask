package compile

import (
	"bytes"
	"context"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/kjuulh/shuttletask/pkg/discover"
	cp "github.com/otiai10/copy"
	"golang.org/x/mod/sumdb/dirhash"
)

var (
	//go:embed templates/mainFile.tmpl
	mainFileTmpl embed.FS
)

const (
	alwaysBuild = true
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
func Compile(ctx context.Context, discovered *discover.Discovered) (string, error) {
	path, err := compile(ctx, discovered.Local)
	if err != nil {
		return "", err
	}

	return path, nil
}

func compile(ctx context.Context, shuttletask *discover.ShuttleTaskDiscovered) (string, error) {
	hash, err := getHash(ctx, shuttletask)
	if err != nil {
		return "", err
	}

	binaryPath, ok, err := binaryMatches(ctx, hash, shuttletask)
	if err != nil {
		return "", err
	}

	if ok && !alwaysBuild { //FIXME: remember to flip to ok
		log.Printf("file already matches continueing\n")
		// The binary is the same so we short circuit
		return binaryPath, nil
	}

	shuttlelocaldir := path.Join(shuttletask.ParentDir, ".shuttle/shuttletask")

	// Generate tmp dir
	if err = generateTmpDir(ctx, shuttlelocaldir); err != nil {
		return "", err
	}
	// Copy files
	if err = copyFiles(ctx, shuttlelocaldir, shuttletask); err != nil {
		return "", err
	}

	// Generate AST
	contents, err := generateAst(ctx, shuttlelocaldir, shuttletask)
	if err != nil {
		return "", err
	}

	// Generate Main file
	if err = generateMainFile(ctx, shuttlelocaldir, shuttletask, contents); err != nil {
		return "", err
	}

	// Compile package
	if err = modTidy(ctx, shuttlelocaldir); err != nil {
		return "", err
	}
	binarypath, err := compileBinary(ctx, shuttlelocaldir)
	if err != nil {
		return "", err
	}
	// Move binary
	finalBinaryPath := path.Join(
		shuttlelocaldir,
		"binaries",
		fmt.Sprintf("shuttletask-%s", hex.EncodeToString([]byte(hash)[:16])),
	)
	os.Rename(binarypath, finalBinaryPath)

	return finalBinaryPath, nil
}

type Function struct {
	Name   string
	Input  []Arg
	Output Output
}

type Arg struct {
	Name string
}

type Output struct {
	Error bool
}

func generateAst(ctx context.Context, shuttlelocaldir string, shuttletask *discover.ShuttleTaskDiscovered) ([]*Function, error) {
	funcs := make([]*Function, 0)

	for _, taskfile := range shuttletask.Files {
		tknSet := token.NewFileSet()
		astfile, err := parser.ParseFile(tknSet, path.Join(shuttlelocaldir, "tmp", taskfile), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		if ast.FileExports(astfile) {
			decls := astfile.Decls
			for _, decl := range decls {
				funcdecl, ok := decl.(*ast.FuncDecl)
				if ok {
					f := Function{}
					f.Name = funcdecl.Name.Name
					param := funcdecl.Type
					paramList := param.Params.List
					for _, param := range paramList {
						for _, name := range param.Names {
							if name != nil && !strings.Contains(fmt.Sprintf("%s", param.Type), "Context") {
								f.Input = append(f.Input, Arg{
									Name: name.Name,
								})
							}
						}
					}
					outputParam := param.Results
					if outputParam != nil {
						if len(outputParam.List) > 1 {
							return nil, errors.New("only error is supported as an output param")
						}
						if len(outputParam.List) == 0 {
							return nil, errors.New("output params are required, only error is supported")
						}

						for _, param := range outputParam.List {
							if fmt.Sprintf("%s", param.Type) != "error" {
								return nil, errors.New("output was not error")
							}
						}

						f.Output = Output{Error: true}
					}

					funcs = append(funcs, &f)
				}
			}
		}
	}

	return funcs, nil
}

func modTidy(ctx context.Context, shuttlelocaldir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = path.Join(shuttlelocaldir, "tmp")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", string(output))
		return err
	}

	return nil
}

func compileBinary(ctx context.Context, shuttlelocaldir string) (string, error) {
	cmd := exec.Command("go", "build")
	cmd.Dir = path.Join(shuttlelocaldir, "tmp")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s\n", string(output))
		return "", err
	}

	return path.Join(shuttlelocaldir, "tmp", "shuttletask"), nil
}

func generateMainFile(
	ctx context.Context,
	shuttlelocaldir string,
	shuttletask *discover.ShuttleTaskDiscovered,
	functions []*Function,
) error {
	tmpmainfile := path.Join(shuttlelocaldir, "tmp/main.go")

	file, err := os.Create(tmpmainfile)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("mainFile.tmpl").Funcs(map[string]any{
		"lower": strings.ToLower,
	}).ParseFS(mainFileTmpl, "templates/mainFile.tmpl"))

	err = tmpl.Execute(file, functions)

	return err
}

func copyFiles(
	ctx context.Context,
	shuttlelocaldir string,
	shuttletask *discover.ShuttleTaskDiscovered,
) error {
	tmpdir := path.Join(shuttlelocaldir, "tmp")

	return cp.Copy(shuttletask.DirPath, tmpdir)
}

func generateTmpDir(ctx context.Context, shuttlelocaldir string) error {
	if err := os.MkdirAll(shuttlelocaldir, 0755); err != nil {
		return err
	}

	binarydir := path.Join(shuttlelocaldir, "binaries")
	if err := os.RemoveAll(binarydir); err != nil {
		return nil
	}
	if err := os.MkdirAll(binarydir, 0755); err != nil {
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

func binaryMatches(
	ctx context.Context,
	hash string,
	shuttletask *discover.ShuttleTaskDiscovered,
) (string, bool, error) {
	shuttlebindir := path.Join(shuttletask.ParentDir, ".shuttle/shuttletask/binaries")

	if _, err := os.Stat(shuttlebindir); errors.Is(err, os.ErrNotExist) {
		log.Println("DEBUG: package doesn't exist continueing")
		return "", false, nil
	}

	entries, err := os.ReadDir(shuttlebindir)
	if err != nil {
		return "", false, err
	}

	if len(entries) == 0 {
		return "", false, err
	}

	log.Printf("%s", entries[0].Name())
	// We only expect a single binary in the folder, so we just take the first entry if it exists
	binary := entries[0]

	if binary.Name() == fmt.Sprintf("shuttletask-%s", hex.EncodeToString([]byte(hash)[:16])) {
		return path.Join(shuttlebindir, binary.Name()), true, nil
	} else {
		return "", false, nil
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

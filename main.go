package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "rewrite-candidates" {
		fmt.Fprintf(os.Stderr, "buggengo [strategy] [repo-directory]\n")
		os.Exit(1)
	}
	repoDir := os.Args[2]
	if _, err := os.Stat(repoDir); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Repo directory does not exist: %q\n", repoDir)
	}

	fset := token.NewFileSet()
	var candidates []map[string]any
	err := filepath.Walk(repoDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Printf("Problem processing repo directory: %q: %v\n", repoDir, err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		sourceFile, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			log.Printf("Problem processing repo directory: %q: %v\n", repoDir, err)
			return nil
		}

		ast.Inspect(sourceFile, func(n ast.Node) bool {
			switch candidateFunc := n.(type) {
			case *ast.FuncDecl:
				lineStart := fset.Position(candidateFunc.Pos()).Line
				lineEnd := fset.Position(candidateFunc.End()).Line

				signatureOnly, err := formatCode(fset, candidateFunc, candidateFunc, func(funcToMod *ast.FuncDecl) {
					funcToMod.Body = nil
				})
				if err != nil {
					return false
				}

				placeholder := placeholderBody()
				withPlaceholder, err := formatCode(fset, candidateFunc, candidateFunc, func(funcToMod *ast.FuncDecl) {
					funcToMod.Body = placeholder
				})
				if err != nil {
					return false
				}

				fullFile, err := formatCode(fset, sourceFile, candidateFunc, func(funcToMod *ast.FuncDecl) {
					funcToMod.Body = placeholder
				})
				if err != nil {
					return false
				}

				candidates = append(candidates, map[string]any{
					"file_path":      path,
					"file_src_code":  fullFile.String(),
					"func_name":      candidateFunc.Name.Name,
					"func_signature": signatureOnly.String(),
					"func_to_write":  withPlaceholder.String(),
					"line_start":     lineStart,
					"line_end":       lineEnd,
				})
			}
			return true
		})
		return nil
	})
	if err != nil {
		log.Printf("Problem processing repo directory: %q: %v\n", repoDir, err)
	}

	b, err := json.Marshal(candidates)
	if err != nil {
		log.Fatalf("Problem marshalling candidates: %v\n", err)
	}
	fmt.Println(string(b))
}

func formatCode(fs *token.FileSet, nodeToOutput any, funcToModify *ast.FuncDecl, modFunc func(*ast.FuncDecl)) (*bytes.Buffer, error) {
	originalBody := funcToModify.Body
	modFunc(funcToModify)
	defer func() {
		funcToModify.Body = originalBody
	}()
	var output bytes.Buffer
	if err := format.Node(&output, fs, nodeToOutput); err != nil {
		return nil, err
	}
	return &output, nil
}

func placeholderBody() *ast.BlockStmt {
	return &ast.BlockStmt{List: []ast.Stmt{&ast.ExprStmt{
		X: &ast.CallExpr{Fun: &ast.Ident{Name: "panic"}, Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"TODO: Implement this function"`}}},
	}}}
}

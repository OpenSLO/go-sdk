package assert

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func AllEnumValuesAreEqual(t *testing.T, filePath, enumName string, validate func(string) error) {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	Require(t, NoError(t, err))

	var enumValues []string

	ast.Inspect(f, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			return true
		}
		for _, spec := range genDecl.Specs {
			vspec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if vspec.Type != nil {
				if ident, ok := vspec.Type.(*ast.Ident); ok && ident.Name == enumName {
					for _, v := range vspec.Values {
						if lit, ok := v.(*ast.BasicLit); ok {
							// Trim quotes from the literal
							val := strings.Trim(lit.Value, "`\"`")
							enumValues = append(enumValues, val)
						}
					}
				}
			}
		}
		return true
	})

	Require(t, NotEmpty(t, enumValues))
	for _, val := range enumValues {
		err = validate(val)
		Require(t, NoError(t, err))
	}
}

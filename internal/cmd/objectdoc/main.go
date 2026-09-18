package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/doc/comment"
	"go/types"
	"io"
	"maps"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/nieomylnieja/govydoc/pkg/govydoc"
	"golang.org/x/tools/go/packages"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/jsonpath"

	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v1alpha"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v2alpha"
)

type generatedObjectDoc struct {
	doc      govydoc.ObjectDoc
	rootType reflect.Type
}

type objectDocGenerator struct {
	name     string
	generate func() (generatedObjectDoc, error)
}

var allDocsGenerators = []objectDocGenerator{
	newObjectDocGenerator(v1alpha.Service{}.GetValidator()),
	newObjectDocGenerator(v1alpha.SLO{}.GetValidator()),
	newObjectDocGenerator(v1.Service{}.GetValidator()),
	newObjectDocGenerator(v1.SLO{}.GetValidator()),
	newObjectDocGenerator(v1.SLI{}.GetValidator()),
	newObjectDocGenerator(v1.AlertCondition{}.GetValidator()),
	newObjectDocGenerator(v1.AlertNotificationTarget{}.GetValidator()),
	newObjectDocGenerator(v1.AlertPolicy{}.GetValidator()),
	newObjectDocGenerator(v1.DataSource{}.GetValidator()),
	newObjectDocGenerator(v2alpha.Service{}.GetValidator()),
	newObjectDocGenerator(v2alpha.SLO{}.GetValidator()),
	newObjectDocGenerator(v2alpha.SLI{}.GetValidator()),
	newObjectDocGenerator(v2alpha.AlertCondition{}.GetValidator()),
	newObjectDocGenerator(v2alpha.AlertNotificationTarget{}.GetValidator()),
	newObjectDocGenerator(v2alpha.AlertPolicy{}.GetValidator()),
	newObjectDocGenerator(v2alpha.DataSource{}.GetValidator()),
}

var (
	apiVersionPath = jsonpath.NewRoot().Name("apiVersion")
	kindPath       = jsonpath.NewRoot().Name("kind")
)

type (
	Versions map[Version]map[Kind]govydoc.ObjectDoc

	Version = string
	Kind    = string
)

func main() {
	versions, err := generateVersions()
	if err != nil {
		panic(err)
	}

	if err = encodeVersions(os.Stdout, versions); err != nil {
		panic(err)
	}
}

func encodeVersions(w io.Writer, versions Versions) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(versions)
}

func newObjectDocGenerator[T any](validator govy.Validator[T]) objectDocGenerator {
	typ := reflect.TypeFor[T]()
	return objectDocGenerator{
		name: typ.String(),
		generate: func() (generatedObjectDoc, error) {
			return generateObjectDoc(validator)
		},
	}
}

func generateObjectDoc[T any](validator govy.Validator[T]) (generatedObjectDoc, error) {
	rootType := reflect.TypeFor[T]()
	doc, err := govydoc.Generate(
		validator,
		govydoc.GenerateGovyOptions(govy.PlanStrictMode()),
	)
	if err != nil {
		return generatedObjectDoc{}, err
	}
	if err = validateRuleDescriptions(doc, rootType); err != nil {
		return generatedObjectDoc{}, err
	}
	return generatedObjectDoc{
		doc:      doc,
		rootType: rootType,
	}, nil
}

func validateRuleDescriptions(doc govydoc.ObjectDoc, rootType reflect.Type) error {
	for _, property := range doc.Properties {
		for ruleIndex, rule := range property.Rules {
			if strings.TrimSpace(rule.Description) == "" {
				return fmt.Errorf(
					"validation rule %d for %s in %s has a blank description",
					ruleIndex+1,
					property.Path,
					rootType,
				)
			}
		}
	}
	return nil
}

func generateVersions() (Versions, error) {
	docs, err := generateAllObjectDocs()
	if err != nil {
		return nil, err
	}
	if err = normalizeGeneratedDocs(docs); err != nil {
		return nil, fmt.Errorf("normalize generated documentation: %w", err)
	}

	slices.SortFunc(docs, func(o1, o2 generatedObjectDoc) int {
		return cmp.Compare(o1.doc.Name, o2.doc.Name)
	})
	return aggregateVersions(docs)
}

func aggregateVersions(docs []generatedObjectDoc) (Versions, error) {
	versions := make(Versions)
	for _, generated := range docs {
		doc := generated.doc
		version, err := discriminatorValue(doc, apiVersionPath)
		if err != nil {
			return nil, err
		}
		kind, err := discriminatorValue(doc, kindPath)
		if err != nil {
			return nil, err
		}
		if versions[version] == nil {
			versions[version] = make(map[Kind]govydoc.ObjectDoc)
		}
		if previous, exists := versions[version][kind]; exists {
			return nil, fmt.Errorf(
				"duplicate version and kind %q %q in document %q: pair is already used by %q",
				version,
				kind,
				doc.Name,
				previous.Name,
			)
		}
		versions[version][kind] = doc
	}
	return versions, nil
}

func discriminatorValue(doc govydoc.ObjectDoc, path jsonpath.Path) (string, error) {
	var value string
	found := false
	for _, property := range doc.Properties {
		if !property.Path.Equal(path) {
			continue
		}
		if found {
			return "", fmt.Errorf(
				"document %q has duplicate discriminator property %s",
				doc.Name,
				path,
			)
		}
		found = true
		if len(property.Values) != 1 {
			return "", fmt.Errorf(
				"document %q discriminator property %s must have exactly one value, but it has %d",
				doc.Name,
				path,
				len(property.Values),
			)
		}
		value = property.Values[0]
	}
	if !found {
		return "", fmt.Errorf("document %q is missing discriminator property %s", doc.Name, path)
	}
	if value == "" {
		return "", fmt.Errorf("document %q discriminator property %s has an empty value", doc.Name, path)
	}
	return value, nil
}

func generateAllObjectDocs() ([]generatedObjectDoc, error) {
	return generateObjectDocs(allDocsGenerators)
}

func generateObjectDocs(generators []objectDocGenerator) ([]generatedObjectDoc, error) {
	docs := make([]generatedObjectDoc, len(generators))
	generatorErrors := make([]error, len(generators))
	var wg sync.WaitGroup
	for i, generator := range generators {
		wg.Go(func() {
			doc, err := generator.generate()
			if err != nil {
				generatorErrors[i] = fmt.Errorf("generate %s documentation: %w", generator.name, err)
				return
			}
			docs[i] = doc
		})
	}
	wg.Wait()
	if err := errors.Join(generatorErrors...); err != nil {
		return nil, err
	}
	return docs, nil
}

const (
	docLinkBaseURL     = "https://pkg.go.dev"
	jsonRawMessageName = "RawMessage"
	jsonRawMessageKind = "[]uint8"
	jsonRawMessagePkg  = "encoding/json"
	jsonValueKind      = "JSON"
)

var deprecatedDocRegex = regexp.MustCompile(`(?m)^Deprecated:\s*(.*)$`)

type fieldOrigin struct {
	owner reflect.Type
	field reflect.StructField
}

type fieldDocKey struct {
	packagePath string
	typeName    string
	fieldName   string
}

type fieldDocResolver struct {
	docs map[fieldDocKey]string
}

func normalizeGeneratedDocs(docs []generatedObjectDoc) error {
	origins := make([]map[string]fieldOrigin, len(docs))
	packagePaths := make(map[string]struct{})
	for i := range docs {
		normalizeRawMessages(&docs[i].doc)

		mappedOrigins, err := mapJSONFieldOrigins(docs[i].rootType)
		if err != nil {
			return fmt.Errorf("map fields for %s: %w", docs[i].rootType, err)
		}
		origins[i] = mappedOrigins
		for _, property := range docs[i].doc.Properties {
			if property.FieldDoc != "" {
				continue
			}
			origin, ok := origins[i][property.Path.String()]
			if ok && origin.owner.PkgPath() != "" {
				packagePaths[origin.owner.PkgPath()] = struct{}{}
			}
		}
	}

	resolver, err := newFieldDocResolver(slices.Sorted(maps.Keys(packagePaths)))
	if err != nil {
		return err
	}
	var recoveryErrors []error
	for i := range docs {
		if err := recoverFieldDocs(&docs[i].doc, origins[i], resolver); err != nil {
			recoveryErrors = append(recoveryErrors, err)
		}
	}
	return errors.Join(recoveryErrors...)
}

func normalizeRawMessages(doc *govydoc.ObjectDoc) {
	wildcardPaths := make(map[string]struct{})
	for i := range doc.Properties {
		property := &doc.Properties[i]
		if property.TypeInfo.Name != jsonRawMessageName ||
			property.TypeInfo.Kind != jsonRawMessageKind ||
			property.TypeInfo.Package != jsonRawMessagePkg {
			continue
		}
		property.TypeInfo.Kind = jsonValueKind
		wildcardPaths[property.Path.IndexWildcard().String()] = struct{}{}
	}
	if len(wildcardPaths) == 0 {
		return
	}

	doc.Properties = slices.DeleteFunc(doc.Properties, func(property govydoc.PropertyDoc) bool {
		_, remove := wildcardPaths[property.Path.String()]
		return remove
	})
	for i := range doc.Properties {
		doc.Properties[i].ChildrenPaths = slices.DeleteFunc(
			doc.Properties[i].ChildrenPaths,
			func(path string) bool {
				_, remove := wildcardPaths[path]
				return remove
			},
		)
	}
}

func mapJSONFieldOrigins(root reflect.Type) (map[string]fieldOrigin, error) {
	origins := make(map[string]fieldOrigin)
	if err := walkJSONFields(root, jsonpath.NewRoot(), origins); err != nil {
		return nil, err
	}
	return origins, nil
}

func walkJSONFields(
	typ reflect.Type,
	path jsonpath.Path,
	origins map[string]fieldOrigin,
) error {
	typ = dereferenceType(typ)
	switch typ.Kind() {
	case reflect.Struct:
		for _, visibleField := range reflect.VisibleFields(typ) {
			if !visibleField.IsExported() {
				continue
			}
			name, _, _ := strings.Cut(visibleField.Tag.Get("json"), ",")
			if name == "" || name == "-" {
				continue
			}
			fieldPath := path.Name(name)
			origin, err := resolveFieldOrigin(typ, visibleField.Index)
			if err != nil {
				return err
			}
			pathString := fieldPath.String()
			if previous, exists := origins[pathString]; exists {
				return fmt.Errorf(
					"JSON path %s resolves to both %s.%s and %s.%s",
					pathString,
					previous.owner,
					previous.field.Name,
					origin.owner,
					origin.field.Name,
				)
			}
			origins[pathString] = origin
			if err = walkJSONFields(visibleField.Type, fieldPath, origins); err != nil {
				return err
			}
		}
	case reflect.Array, reflect.Slice:
		return walkJSONFields(typ.Elem(), path.IndexWildcard(), origins)
	case reflect.Map:
		if err := walkJSONFields(typ.Key(), path.KeyWildcard(), origins); err != nil {
			return err
		}
		return walkJSONFields(typ.Elem(), path.ValueWildcard(), origins)
	default:
	}
	return nil
}

func resolveFieldOrigin(typ reflect.Type, index []int) (fieldOrigin, error) {
	owner := dereferenceType(typ)
	for i, fieldIndex := range index {
		if owner.Kind() != reflect.Struct || fieldIndex >= owner.NumField() {
			return fieldOrigin{}, fmt.Errorf("invalid field index %v for %s", index, typ)
		}
		field := owner.Field(fieldIndex)
		if i == len(index)-1 {
			return fieldOrigin{owner: owner, field: field}, nil
		}
		owner = dereferenceType(field.Type)
	}
	return fieldOrigin{}, fmt.Errorf("empty field index for %s", typ)
}

func dereferenceType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

func newFieldDocResolver(packagePaths []string) (*fieldDocResolver, error) {
	resolver := &fieldDocResolver{
		docs: make(map[fieldDocKey]string),
	}
	if len(packagePaths) == 0 {
		return resolver, nil
	}

	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax,
	}, packagePaths...)
	if err != nil {
		return nil, fmt.Errorf("load packages for field documentation: %w", err)
	}

	var packageErrors []error
	for _, pkg := range loaded {
		for _, pkgErr := range pkg.Errors {
			packageErrors = append(packageErrors, fmt.Errorf("package %s: %w", pkg.PkgPath, pkgErr))
		}
	}
	if len(packageErrors) > 0 {
		return nil, fmt.Errorf("load packages for field documentation: %w", errors.Join(packageErrors...))
	}

	for _, pkg := range loaded {
		resolver.indexPackage(pkg)
	}
	return resolver, nil
}

func (r *fieldDocResolver) indexPackage(pkg *packages.Package) {
	parser := newCommentParser(pkg)
	printer := comment.Printer{
		DocLinkURL: func(link *comment.DocLink) string {
			if link.ImportPath == "" {
				link.ImportPath = pkg.PkgPath
			}
			return link.DefaultURL(docLinkBaseURL)
		},
	}
	for _, file := range pkg.Syntax {
		for _, declaration := range file.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				r.indexTypeFields(pkg.PkgPath, parser, &printer, typeSpec)
			}
		}
	}
}

func (r *fieldDocResolver) indexTypeFields(
	packagePath string,
	parser *comment.Parser,
	printer *comment.Printer,
	typeSpec *ast.TypeSpec,
) {
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return
	}
	for _, field := range structType.Fields.List {
		var doc string
		if field.Doc != nil {
			doc = strings.TrimSpace(string(printer.Markdown(parser.Parse(field.Doc.Text()))))
		}
		for _, name := range astFieldNames(field) {
			r.docs[fieldDocKey{
				packagePath: packagePath,
				typeName:    typeSpec.Name.Name,
				fieldName:   name,
			}] = doc
		}
	}
}

func newCommentParser(current *packages.Package) *comment.Parser {
	return &comment.Parser{
		LookupPackage: func(name string) (string, bool) {
			for path, imported := range current.Imports {
				if imported.Name == name {
					return path, true
				}
			}
			return "", false
		},
		LookupSym: func(recv, name string) bool {
			if recv == "" {
				return current.Types.Scope().Lookup(name) != nil
			}
			object := current.Types.Scope().Lookup(recv)
			if object == nil {
				return false
			}
			member, _, _ := types.LookupFieldOrMethod(object.Type(), true, current.Types, name)
			return member != nil
		},
	}
}

func astFieldNames(field *ast.Field) []string {
	if len(field.Names) > 0 {
		names := make([]string, len(field.Names))
		for i := range field.Names {
			names[i] = field.Names[i].Name
		}
		return names
	}
	if name := embeddedFieldName(field.Type); name != "" {
		return []string{name}
	}
	return nil
}

func embeddedFieldName(expression ast.Expr) string {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression.Name
	case *ast.SelectorExpr:
		return expression.Sel.Name
	case *ast.StarExpr:
		return embeddedFieldName(expression.X)
	case *ast.IndexExpr:
		return embeddedFieldName(expression.X)
	case *ast.IndexListExpr:
		return embeddedFieldName(expression.X)
	case *ast.ParenExpr:
		return embeddedFieldName(expression.X)
	default:
		return ""
	}
}

func recoverFieldDocs(
	doc *govydoc.ObjectDoc,
	origins map[string]fieldOrigin,
	resolver *fieldDocResolver,
) error {
	var recoveryErrors []error
	for i := range doc.Properties {
		property := &doc.Properties[i]
		if property.FieldDoc != "" {
			continue
		}
		path := property.Path.String()
		origin, ok := origins[path]
		if !ok {
			if isRootOrSyntheticWildcardPath(path, origins) {
				continue
			}
			recoveryErrors = append(recoveryErrors, fmt.Errorf(
				"recover field documentation for %s at %s: path has no Go field origin",
				doc.Name,
				property.Path,
			))
			continue
		}
		fieldDoc, indexed := resolver.docs[fieldDocKey{
			packagePath: origin.owner.PkgPath(),
			typeName:    origin.owner.Name(),
			fieldName:   origin.field.Name,
		}]
		if !indexed {
			recoveryErrors = append(recoveryErrors, fmt.Errorf(
				"recover field documentation for %s at %s: %s.%s is missing from the AST field index",
				doc.Name,
				property.Path,
				origin.owner,
				origin.field.Name,
			))
			continue
		}
		if fieldDoc == "" {
			continue
		}
		property.FieldDoc = fieldDoc
		if match := deprecatedDocRegex.FindStringSubmatch(fieldDoc); len(match) > 1 {
			if property.DeprecatedDoc == "" {
				property.DeprecatedDoc = strings.TrimSpace(match[1])
			}
			property.FieldDoc = strings.TrimSpace(deprecatedDocRegex.ReplaceAllString(fieldDoc, ""))
		}
	}
	return errors.Join(recoveryErrors...)
}

func isRootOrSyntheticWildcardPath(path string, origins map[string]fieldOrigin) bool {
	if path == jsonpath.NewRoot().String() {
		return true
	}
	for {
		parent, ok := cutWildcardSuffix(path)
		if !ok {
			return false
		}
		if _, ok = origins[parent]; ok {
			return true
		}
		path = parent
	}
}

func cutWildcardSuffix(path string) (string, bool) {
	for _, suffix := range [...]string{"[*]", ".*", ".*~"} {
		if parent, ok := strings.CutSuffix(path, suffix); ok {
			return parent, true
		}
	}
	return "", false
}

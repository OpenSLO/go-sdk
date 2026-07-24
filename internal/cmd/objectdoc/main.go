package main

import (
	"cmp"
	"encoding/json"
	"os"
	"slices"

	"github.com/nieomylnieja/govydoc/pkg/govydoc"
	"golang.org/x/sync/errgroup"

	"github.com/nobl9/govy/pkg/jsonpath"

	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v1alpha"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v2alpha"
)

var allDocsGeneratorFuncs = []func() (govydoc.ObjectDoc, error){
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1alpha.Service{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1alpha.SLO{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1.Service{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1.SLO{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1.SLI{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1.AlertCondition{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) {
		return govydoc.Generate(v1.AlertNotificationTarget{}.GetValidator())
	},
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1.AlertPolicy{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v1.DataSource{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v2alpha.Service{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v2alpha.SLO{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v2alpha.SLI{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v2alpha.AlertCondition{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) {
		return govydoc.Generate(v2alpha.AlertNotificationTarget{}.GetValidator())
	},
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v2alpha.AlertPolicy{}.GetValidator()) },
	func() (govydoc.ObjectDoc, error) { return govydoc.Generate(v2alpha.DataSource{}.GetValidator()) },
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
	docs := make([]govydoc.ObjectDoc, len(allDocsGeneratorFuncs))
	var group errgroup.Group
	for i, generate := range allDocsGeneratorFuncs {
		group.Go(func() error {
			doc, err := generate()
			if err != nil {
				return err
			}
			docs[i] = doc
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		panic(err)
	}

	slices.SortFunc(docs, func(o1, o2 govydoc.ObjectDoc) int { return cmp.Compare(o1.Name, o2.Name) })

	versions := make(Versions)
	for _, doc := range docs {
		var (
			version Version
			kind    Kind
		)
		for _, prop := range doc.Properties {
			switch {
			case prop.Path.Equal(apiVersionPath):
				version = prop.Values[0]
			case prop.Path.Equal(kindPath):
				kind = prop.Values[0]
			}
		}
		if version == "" || kind == "" {
			panic("missing version or kind in doc: " + doc.Name)
		}
		if versions[version] == nil {
			versions[version] = make(map[Kind]govydoc.ObjectDoc)
		}
		if _, exists := versions[version][kind]; exists {
			panic("duplicate version and kind: " + version + " " + kind)
		}
		versions[version][kind] = doc
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(versions); err != nil {
		panic(err)
	}
}

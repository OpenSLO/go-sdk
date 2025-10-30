package main

import (
	"cmp"
	"encoding/json"
	"os"
	"slices"
	"sync"

	"github.com/OpenSLO/go-sdk/pkg/openslo/v1"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v1alpha"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v2alpha"
	"github.com/nieomylnieja/govydoc/pkg/govydoc"
)

type (
	Versions map[Version]map[Kind]govydoc.ObjectDoc

	Version = string
	Kind    = string
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

func main() {
	wg := sync.WaitGroup{}
	wg.Add(len(allDocsGeneratorFuncs))
	docs := make([]govydoc.ObjectDoc, 0, len(allDocsGeneratorFuncs))
	mu := sync.Mutex{}
	for _, f := range allDocsGeneratorFuncs {
		go func() {
			defer wg.Done()
			doc, err := f()
			if err != nil {
				panic(err)
			}
			mu.Lock()
			docs = append(docs, doc)
			mu.Unlock()
		}()
	}
	wg.Wait()

	slices.SortFunc(docs, func(o1, o2 govydoc.ObjectDoc) int { return cmp.Compare(o1.Name, o2.Name) })

	versions := make(Versions)
	for _, doc := range docs {
		var (
			version Version
			kind    Kind
		)
		for _, prop := range doc.Properties {
			switch prop.Path {
			case "$.apiVersion":
				version = prop.Values[0]
			case "$.kind":
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

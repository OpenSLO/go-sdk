package main

import (
	"bytes"
	"errors"
	"maps"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/nieomylnieja/govydoc/pkg/govydoc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/jsonpath"

	v1 "github.com/OpenSLO/go-sdk/pkg/openslo/v1"
	"github.com/OpenSLO/go-sdk/pkg/openslo/v2alpha"
)

func TestGenerateObjectDocRequiresPredicateDescriptions(t *testing.T) {
	validator := govy.New[v1.Service]().When(func(v1.Service) bool { return true })

	_, err := generateObjectDoc(validator)

	require.ErrorContains(t, err, "predicates without description found at: validator level")
}

func TestGenerateObjectDocRejectsBlankRuleDescriptions(t *testing.T) {
	validator := govy.New(
		govy.For(govy.GetSelf[v1.Service]()).Rules(
			govy.NewRule(func(v1.Service) error { return nil }).WithDescription(" \t"),
		),
	)

	_, err := generateObjectDoc(validator)

	require.ErrorContains(t, err, "validation rule 1 for $ in v1.Service has a blank description")
}

func TestRegisteredValidatorsHaveCompletePlans(t *testing.T) {
	for _, generator := range allDocsGenerators {
		t.Run(generator.name, func(t *testing.T) {
			_, err := generator.generate()
			require.NoError(t, err)
		})
	}
}

func TestGenerateObjectDocsReturnsAllErrorsInGeneratorOrder(t *testing.T) {
	firstErr := errors.New("first failure")
	secondErr := errors.New("second failure")
	secondFinished := make(chan struct{})
	generators := []objectDocGenerator{
		{
			name: "first",
			generate: func() (generatedObjectDoc, error) {
				<-secondFinished
				return generatedObjectDoc{}, firstErr
			},
		},
		{
			name: "second",
			generate: func() (generatedObjectDoc, error) {
				close(secondFinished)
				return generatedObjectDoc{}, secondErr
			},
		},
	}

	_, err := generateObjectDocs(generators)

	require.Error(t, err)
	assert.ErrorIs(t, err, firstErr)
	assert.ErrorIs(t, err, secondErr)
	assert.EqualError(
		t,
		err,
		"generate first documentation: first failure\ngenerate second documentation: second failure",
	)
}

func TestAggregateVersionsRequiresOneDiscriminatorValue(t *testing.T) {
	tests := map[string]struct {
		versionValues []string
		kindValues    []string
		expectedError string
	}{
		"apiVersion without values": {
			kindValues:    []string{"Service"},
			expectedError: `document "Example" discriminator property $.apiVersion must have exactly one value, but it has 0`,
		},
		"kind with multiple values": {
			versionValues: []string{"openslo/v1"},
			kindValues:    []string{"Service", "SLO"},
			expectedError: `document "Example" discriminator property $.kind must have exactly one value, but it has 2`,
		},
		"kind with empty value": {
			versionValues: []string{"openslo/v1"},
			kindValues:    []string{""},
			expectedError: `document "Example" discriminator property $.kind has an empty value`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := aggregateVersions([]generatedObjectDoc{
				newGeneratedObjectDoc("Example", test.versionValues, test.kindValues),
			})

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestAggregateVersionsRejectsDuplicateAndMissingDiscriminatorPaths(t *testing.T) {
	versionProperty := newDiscriminatorProperty(apiVersionPath, "openslo/v1")
	kindProperty := newDiscriminatorProperty(kindPath, "Service")
	tests := map[string]struct {
		properties    []govydoc.PropertyDoc
		expectedError string
	}{
		"duplicate apiVersion": {
			properties:    []govydoc.PropertyDoc{versionProperty, versionProperty, kindProperty},
			expectedError: `document "Example" has duplicate discriminator property $.apiVersion`,
		},
		"duplicate kind": {
			properties:    []govydoc.PropertyDoc{versionProperty, kindProperty, kindProperty},
			expectedError: `document "Example" has duplicate discriminator property $.kind`,
		},
		"missing apiVersion": {
			properties:    []govydoc.PropertyDoc{kindProperty},
			expectedError: `document "Example" is missing discriminator property $.apiVersion`,
		},
		"missing kind": {
			properties:    []govydoc.PropertyDoc{versionProperty},
			expectedError: `document "Example" is missing discriminator property $.kind`,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := aggregateVersions([]generatedObjectDoc{{
				doc: govydoc.ObjectDoc{Name: "Example", Properties: test.properties},
			}})

			require.EqualError(t, err, test.expectedError)
		})
	}
}

func TestAggregateVersionsRejectsDuplicateVersionAndKind(t *testing.T) {
	docs := []generatedObjectDoc{
		newGeneratedObjectDoc("First", []string{"openslo/v1"}, []string{"Service"}),
		newGeneratedObjectDoc("Second", []string{"openslo/v1"}, []string{"Service"}),
	}

	_, err := aggregateVersions(docs)

	require.ErrorContains(t, err, `duplicate version and kind "openslo/v1" "Service"`)
	assert.ErrorContains(t, err, `document "Second"`)
	assert.ErrorContains(t, err, `already used by "First"`)
}

func TestNormalizeRawMessages(t *testing.T) {
	rawMessagePath := jsonpath.Parse("$.spec.connectionDetails")
	rawMessageWildcardPath := rawMessagePath.IndexWildcard()
	byteSlicePath := jsonpath.Parse("$.spec.payload")
	byteSliceWildcardPath := byteSlicePath.IndexWildcard()
	doc := govydoc.ObjectDoc{Properties: []govydoc.PropertyDoc{
		{
			PropertyPlan: govy.PropertyPlan{
				Path: rawMessagePath,
				TypeInfo: govy.TypeInfo{
					Name:    jsonRawMessageName,
					Kind:    jsonRawMessageKind,
					Package: jsonRawMessagePkg,
				},
			},
			ChildrenPaths: []string{rawMessageWildcardPath.String()},
		},
		{
			PropertyPlan: govy.PropertyPlan{
				Path:     rawMessageWildcardPath,
				TypeInfo: govy.TypeInfo{Name: "uint8", Kind: "uint8"},
			},
		},
		{
			PropertyPlan: govy.PropertyPlan{
				Path:     byteSlicePath,
				TypeInfo: govy.TypeInfo{Name: "", Kind: jsonRawMessageKind},
			},
			ChildrenPaths: []string{byteSliceWildcardPath.String()},
		},
		{
			PropertyPlan: govy.PropertyPlan{
				Path:     byteSliceWildcardPath,
				TypeInfo: govy.TypeInfo{Name: "uint8", Kind: "uint8"},
			},
		},
	}}

	normalizeRawMessages(&doc)

	require.Len(t, doc.Properties, 3)
	rawMessage := requireProperty(t, doc, rawMessagePath)
	assert.Equal(t, jsonRawMessageName, rawMessage.TypeInfo.Name)
	assert.Equal(t, jsonValueKind, rawMessage.TypeInfo.Kind)
	assert.Equal(t, jsonRawMessagePkg, rawMessage.TypeInfo.Package)
	assert.Empty(t, rawMessage.ChildrenPaths)
	assert.Nil(t, findProperty(doc, rawMessageWildcardPath))
	assert.NotNil(t, findProperty(doc, byteSliceWildcardPath))
	assert.Equal(t, jsonRawMessageKind, requireProperty(t, doc, byteSlicePath).TypeInfo.Kind)
}

func TestNormalizeGeneratedDocsRecoversPromotedFieldDocs(t *testing.T) {
	typeDocumentation := "existing type documentation"
	docs := []generatedObjectDoc{
		{
			rootType: reflect.TypeFor[v1.AlertPolicy](),
			doc: govydoc.ObjectDoc{
				Name: "v1.AlertPolicy",
				Properties: []govydoc.PropertyDoc{
					newPropertyDoc("$.spec.conditions[*].conditionRef", typeDocumentation),
					newPropertyDoc("$.spec.notificationTargets[*].targetRef", typeDocumentation),
				},
			},
		},
		{
			rootType: reflect.TypeFor[v1.SLO](),
			doc: govydoc.ObjectDoc{
				Name: "v1.SLO",
				Properties: []govydoc.PropertyDoc{
					newPropertyDoc("$.spec.alertPolicies[*].alertPolicyRef", typeDocumentation),
				},
			},
		},
		{
			rootType: reflect.TypeFor[v2alpha.AlertPolicy](),
			doc: govydoc.ObjectDoc{
				Name: "v2alpha.AlertPolicy",
				Properties: []govydoc.PropertyDoc{
					newPropertyDoc("$.spec.conditions[*].conditionRef", typeDocumentation),
					newPropertyDoc("$.spec.notificationTargets[*].targetRef", typeDocumentation),
				},
			},
		},
		{
			rootType: reflect.TypeFor[v2alpha.SLO](),
			doc: govydoc.ObjectDoc{
				Name: "v2alpha.SLO",
				Properties: []govydoc.PropertyDoc{
					newPropertyDoc("$.spec.alertPolicies[*].alertPolicyRef", typeDocumentation),
				},
			},
		},
	}

	require.NoError(t, normalizeGeneratedDocs(docs))

	tests := []struct {
		name     string
		doc      govydoc.ObjectDoc
		path     string
		fieldDoc string
	}{
		{
			name:     "v1 conditionRef",
			doc:      docs[0].doc,
			path:     "$.spec.conditions[*].conditionRef",
			fieldDoc: "ConditionRef names an existing alert condition.",
		},
		{
			name:     "v1 targetRef",
			doc:      docs[0].doc,
			path:     "$.spec.notificationTargets[*].targetRef",
			fieldDoc: "TargetRef names an existing notification target.",
		},
		{
			name:     "v1 alertPolicyRef",
			doc:      docs[1].doc,
			path:     "$.spec.alertPolicies[*].alertPolicyRef",
			fieldDoc: "AlertPolicyRef names an existing alert policy.",
		},
		{
			name:     "v2alpha conditionRef",
			doc:      docs[2].doc,
			path:     "$.spec.conditions[*].conditionRef",
			fieldDoc: "ConditionRef names the alert condition to use.",
		},
		{
			name:     "v2alpha targetRef",
			doc:      docs[2].doc,
			path:     "$.spec.notificationTargets[*].targetRef",
			fieldDoc: "TargetRef names the notification target to use.",
		},
		{
			name:     "v2alpha alertPolicyRef",
			doc:      docs[3].doc,
			path:     "$.spec.alertPolicies[*].alertPolicyRef",
			fieldDoc: "AlertPolicyRef names the alert policy to use.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			property := requireProperty(t, test.doc, jsonpath.Parse(test.path))
			assert.Equal(t, test.fieldDoc, property.FieldDoc)
			assert.Equal(t, typeDocumentation, property.TypeDoc)
		})
	}
}

func TestRecoverFieldDocsReportsFieldsMissingFromASTIndex(t *testing.T) {
	fixtureType := reflect.TypeFor[fieldDocRecoveryFixture]()
	fields := []struct {
		name string
		path string
	}{
		{name: "Documented", path: "$.documented"},
		{name: "Undocumented", path: "$.undocumented"},
		{name: "MissingFirst", path: "$.missingFirst"},
		{name: "MissingSecond", path: "$.missingSecond"},
	}
	doc := govydoc.ObjectDoc{Name: "FixtureDoc"}
	origins := make(map[string]fieldOrigin, len(fields))
	resolver := &fieldDocResolver{docs: make(map[fieldDocKey]string)}
	for _, item := range fields {
		field, ok := fixtureType.FieldByName(item.name)
		require.True(t, ok)
		doc.Properties = append(doc.Properties, newPropertyDoc(item.path, ""))
		origins[item.path] = fieldOrigin{owner: fixtureType, field: field}
		key := fieldDocKey{
			packagePath: fixtureType.PkgPath(),
			typeName:    fixtureType.Name(),
			fieldName:   field.Name,
		}
		switch item.name {
		case "Documented":
			resolver.docs[key] = "Recovered field documentation."
		case "Undocumented":
			resolver.docs[key] = ""
		}
	}

	err := recoverFieldDocs(&doc, origins, resolver)

	require.Error(t, err)
	assert.ErrorContains(t, err, "recover field documentation for FixtureDoc at $.missingFirst")
	assert.ErrorContains(t, err, "fieldDocRecoveryFixture.MissingFirst is missing from the AST field index")
	assert.ErrorContains(t, err, "recover field documentation for FixtureDoc at $.missingSecond")
	assert.ErrorContains(t, err, "fieldDocRecoveryFixture.MissingSecond is missing from the AST field index")
	assert.Equal(t, "Recovered field documentation.", requireProperty(t, doc, jsonpath.Parse("$.documented")).FieldDoc)
	assert.Empty(t, requireProperty(t, doc, jsonpath.Parse("$.undocumented")).FieldDoc)
}

func TestRecoverFieldDocsRejectsOriginlessRealPaths(t *testing.T) {
	fixtureType := reflect.TypeFor[fieldDocRecoveryFixture]()
	field, ok := fixtureType.FieldByName("Documented")
	require.True(t, ok)
	origins := map[string]fieldOrigin{
		"$.labels": {owner: fixtureType, field: field},
	}
	doc := govydoc.ObjectDoc{
		Name: "FixtureDoc",
		Properties: []govydoc.PropertyDoc{
			newPropertyDoc("$", ""),
			newPropertyDoc("$.labels[*]", ""),
			newPropertyDoc("$.labels.*", ""),
			newPropertyDoc("$.labels.*~", ""),
			newPropertyDoc("$.labels.*[*]", ""),
			newPropertyDoc("$.labels.*~[*]", ""),
			newPropertyDoc("$.items[*].promoted", ""),
		},
	}
	resolver := &fieldDocResolver{docs: make(map[fieldDocKey]string)}

	err := recoverFieldDocs(&doc, origins, resolver)

	require.EqualError(
		t,
		err,
		"recover field documentation for FixtureDoc at $.items[*].promoted: path has no Go field origin",
	)
}

func TestGenerateVersionsMatchesCanonicalManifest(t *testing.T) {
	first, err := generateVersions()
	require.NoError(t, err)
	second, err := generateVersions()
	require.NoError(t, err)

	var firstOutput bytes.Buffer
	require.NoError(t, encodeVersions(&firstOutput, first))
	var secondOutput bytes.Buffer
	require.NoError(t, encodeVersions(&secondOutput, second))
	require.True(
		t,
		bytes.Equal(firstOutput.Bytes(), secondOutput.Bytes()),
		"successive generations produced different serialized output",
	)

	expectedKinds := map[Version][]Kind{
		"openslo/v1alpha": {"SLO", "Service"},
		"openslo/v1": {
			"AlertCondition",
			"AlertNotificationTarget",
			"AlertPolicy",
			"DataSource",
			"SLI",
			"SLO",
			"Service",
		},
		"openslo.com/v2alpha": {
			"AlertCondition",
			"AlertNotificationTarget",
			"AlertPolicy",
			"DataSource",
			"SLI",
			"SLO",
			"Service",
		},
	}
	assert.Equal(t, expectedKinds, collectVersionKinds(first))

	connectionDetailsPath := jsonpath.Parse("$.spec.connectionDetails")
	connectionDetailsWildcardPath := connectionDetailsPath.IndexWildcard()
	for _, version := range []Version{"openslo/v1", "openslo.com/v2alpha"} {
		documents, ok := first[version]
		require.True(t, ok, "version %s not found", version)
		dataSource, ok := documents["DataSource"]
		require.True(t, ok, "DataSource not found for version %s", version)
		connectionDetails := requireProperty(t, dataSource, connectionDetailsPath)
		assert.Equal(t, jsonRawMessageName, connectionDetails.TypeInfo.Name)
		assert.Equal(t, jsonValueKind, connectionDetails.TypeInfo.Kind)
		assert.Equal(t, jsonRawMessagePkg, connectionDetails.TypeInfo.Package)
		assert.NotContains(t, connectionDetails.ChildrenPaths, connectionDetailsWildcardPath.String())
		assert.Nil(t, findProperty(dataSource, connectionDetailsWildcardPath))
	}

	checkedIn, err := os.ReadFile("../../../docs/manifest.json")
	require.NoError(t, err)
	require.True(
		t,
		bytes.Equal(checkedIn, firstOutput.Bytes()),
		"docs/manifest.json is stale: checked-in size %d, generated size %d. Run make generate",
		len(checkedIn),
		firstOutput.Len(),
	)
}

type fieldDocRecoveryFixture struct {
	Documented    string `json:"documented"`
	Undocumented  string `json:"undocumented"`
	MissingFirst  string `json:"missingFirst"`
	MissingSecond string `json:"missingSecond"`
}

func newGeneratedObjectDoc(name string, versionValues, kindValues []string) generatedObjectDoc {
	return generatedObjectDoc{doc: govydoc.ObjectDoc{
		Name: name,
		Properties: []govydoc.PropertyDoc{
			newDiscriminatorProperty(apiVersionPath, versionValues...),
			newDiscriminatorProperty(kindPath, kindValues...),
		},
	}}
}

func newDiscriminatorProperty(path jsonpath.Path, values ...string) govydoc.PropertyDoc {
	return govydoc.PropertyDoc{PropertyPlan: govy.PropertyPlan{
		Path:   path,
		Values: values,
	}}
}

func collectVersionKinds(versions Versions) map[Version][]Kind {
	kinds := make(map[Version][]Kind, len(versions))
	for version, documents := range versions {
		kinds[version] = slices.Sorted(maps.Keys(documents))
	}
	return kinds
}

func newPropertyDoc(path, typeDoc string) govydoc.PropertyDoc {
	return govydoc.PropertyDoc{
		PropertyPlan: govy.PropertyPlan{Path: jsonpath.Parse(path)},
		TypeDoc:      typeDoc,
	}
}

func requireProperty(t *testing.T, doc govydoc.ObjectDoc, path jsonpath.Path) *govydoc.PropertyDoc {
	t.Helper()
	property := findProperty(doc, path)
	require.NotNil(t, property, "property %s not found", path)
	return property
}

func findProperty(doc govydoc.ObjectDoc, path jsonpath.Path) *govydoc.PropertyDoc {
	for i := range doc.Properties {
		if doc.Properties[i].Path.Equal(path) {
			return &doc.Properties[i]
		}
	}
	return nil
}

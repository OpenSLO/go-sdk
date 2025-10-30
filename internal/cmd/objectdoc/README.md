## objectdoc

This little utility registers all OpenSLO objects across all versions, and:
extracts their [govy validators](https://pkg.go.dev/github.com/nobl9/govy/pkg/govy#Validator)
and feeds it into [govydoc](https://github.com/nieomylnieja/govydoc)
which in turn produces a complete documentation for each object based on their
[govy validation plan](https://pkg.go.dev/github.com/nobl9/govy/pkg/govy#Plan),
type definitions, code docs and examples.

The output is a structured JSON which is written to stdout.

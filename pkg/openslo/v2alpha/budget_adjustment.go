package v2alpha

import (
	"time"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var (
	_ = Object(BudgetAdjustment{})
	_ = openslo.ObjectValidator[BudgetAdjustment](BudgetAdjustment{})
)

func NewBudgetAdjustment(metadata Metadata, spec BudgetAdjustmentSpec) BudgetAdjustment {
	return BudgetAdjustment{
		APIVersion: APIVersion,
		Kind:       openslo.KindBudgetAdjustment,
		Metadata:   metadata,
		Spec:       spec,
	}
}

type BudgetAdjustment struct {
	APIVersion openslo.Version      `json:"apiVersion"`
	Kind       openslo.Kind         `json:"kind"`
	Metadata   Metadata             `json:"metadata"`
	Spec       BudgetAdjustmentSpec `json:"spec"`
}

func (b BudgetAdjustment) GetVersion() openslo.Version {
	return APIVersion
}

func (b BudgetAdjustment) GetKind() openslo.Kind {
	return openslo.KindBudgetAdjustment
}

func (b BudgetAdjustment) GetName() string {
	return b.Metadata.Name
}

func (b BudgetAdjustment) Validate() error {
	return budgetAdjustmentValidation.Validate(b)
}

func (b BudgetAdjustment) String() string {
	return internal.GetObjectName(b)
}

func (b BudgetAdjustment) GetMetadata() Metadata {
	return b.Metadata
}

func (b BudgetAdjustment) GetValidator() govy.Validator[BudgetAdjustment] {
	return budgetAdjustmentValidation
}

type BudgetAdjustmentSpec struct {
	Description  string             `json:"description"`
	Service      string             `json:"service"`
	IndicatorRef string             `json:"indicatorRef"`
	StartTime    string             `json:"startTime"`
	EndTime      *string            `json:"endTime,omitempty"`
	Duration     *DurationShorthand `json:"duration,omitempty"`
}

var budgetAdjustmentValidation = govy.New(
	validationRulesAPIVersion(func(b BudgetAdjustment) openslo.Version { return b.APIVersion }),
	validationRulesKind(func(b BudgetAdjustment) openslo.Kind { return b.Kind }, openslo.KindBudgetAdjustment),
	validationRulesMetadata(func(b BudgetAdjustment) Metadata { return b.Metadata }),
	govy.For(func(b BudgetAdjustment) BudgetAdjustmentSpec { return b.Spec }).
		WithName("spec").
		Include(govy.New(
			govy.For(func(spec BudgetAdjustmentSpec) string { return spec.Description }).
				WithName("description").
				Required().
				Rules(rules.StringMaxLength(1050)),
			govy.For(func(spec BudgetAdjustmentSpec) string { return spec.Service }).
				WithName("service").
				Required().
				Rules(rules.StringDNSLabel()),
			govy.For(func(spec BudgetAdjustmentSpec) string { return spec.IndicatorRef }).
				WithName("indicatorRef").
				Required().
				Rules(rules.StringDNSLabel()),
			govy.For(func(spec BudgetAdjustmentSpec) string { return spec.StartTime }).
				WithName("startTime").
				Required().
				Rules(rules.StringDateTime(time.RFC3339)),
			govy.ForPointer(func(spec BudgetAdjustmentSpec) *string { return spec.EndTime }).
				WithName("endTime").
				Rules(rules.StringDateTime(time.RFC3339)),
			govy.ForPointer(func(spec BudgetAdjustmentSpec) *DurationShorthand { return spec.Duration }).
				WithName("duration").
				Include(durationShortHandValidation),
			govy.For(govy.GetSelf[BudgetAdjustmentSpec]()).
				WithName("spec").
				Rules(govy.NewRule(func(spec BudgetAdjustmentSpec) error {
					if spec.EndTime == nil && spec.Duration == nil {
						return govy.NewRuleError("one of 'endTime' or 'duration' is required")
					}
					if spec.EndTime != nil && spec.Duration != nil {
						return govy.NewRuleError("only one of 'endTime' or 'duration' can be provided")
					}
					return nil
				})),
		)),
).WithNameFunc(internal.GetObjectName[BudgetAdjustment])

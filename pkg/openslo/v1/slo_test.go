package v1

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nobl9/govy/pkg/govy"
	"github.com/nobl9/govy/pkg/govytest"
	"github.com/nobl9/govy/pkg/rules"

	"github.com/OpenSLO/go-sdk/internal/assert"
	"github.com/OpenSLO/go-sdk/pkg/openslo"
)

var sloValidationMessageRegexp = getValidationMessageRegexp(openslo.KindSLO)

func TestSLO_Validate_Ok(t *testing.T) {
	for _, slo := range []SLO{
		validRatioSLO(),
		validThresholdSLO(),
		validSLOWithSLIRef(),
		validRatioSLOWithInlinedAlertPolicy(),
		validCompositeSLOWithSLIRef(),
		validCompositeSLOWithInlinedSLI(),
	} {
		err := slo.Validate()
		govytest.AssertNoError(t, err)
	}
}

func TestSLO_Validate_VersionAndKind(t *testing.T) {
	slo := validRatioSLO()
	slo.APIVersion = "v0.1"
	slo.Kind = openslo.KindService
	err := slo.Validate()
	assert.Require(t, assert.Error(t, err))
	assert.True(t, sloValidationMessageRegexp.MatchString(err.Error()))
	govytest.AssertError(t, err,
		govytest.ExpectedRuleError{
			PropertyName: "apiVersion",
			Code:         rules.ErrorCodeEqualTo,
		},
		govytest.ExpectedRuleError{
			PropertyName: "kind",
			Code:         rules.ErrorCodeEqualTo,
		},
	)
}

func TestSLO_Validate_Metadata(t *testing.T) {
	runMetadataTests(t, "metadata", func(m Metadata) SLO {
		condition := validRatioSLO()
		condition.Metadata = m
		return condition
	})
}

func TestSLO_Validate_Spec(t *testing.T) {
	t.Run("description ok", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.Description = strings.Repeat("A", 1050)
		err := slo.Validate()
		govytest.AssertNoError(t, err)
	})
	t.Run("description too long", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.Description = strings.Repeat("A", 1051)
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.description",
			Code:         rules.ErrorCodeStringMaxLength,
		})
	})
	t.Run("invalid budgetingMethod", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.BudgetingMethod = "invalid"
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.budgetingMethod",
			Code:         rules.ErrorCodeOneOf,
		})
	})
	for _, method := range validSLOBudgetingMethods {
		t.Run(fmt.Sprintf("budgetingMethod %s", method), func(t *testing.T) {
			slo := validRatioSLO()
			slo.Spec.BudgetingMethod = method
			err := slo.Validate()
			govytest.AssertNoError(t, err)
		})
	}
	t.Run("missing service", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.Service = ""
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.service",
			Code:         rules.ErrorCodeRequired,
		})
	})
	t.Run("missing both indicator definition in spec and objectives", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.Indicator = nil
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec",
			Message: "'indicator' or 'indicatorRef' fields must either be defined on the 'spec' level (standard SLOs)" +
				" or on the 'spec.objectives[*]' level (composite SLOs), but none were provided",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("indicator definition both in spec and objectives", func(t *testing.T) {
		slo := validCompositeSLOWithInlinedSLI()
		slo.Spec.IndicatorRef = ptr("my-sli")
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec",
			Message: "'indicator' or 'indicatorRef' fields must either be defined on the 'spec' level (standard SLOs)" +
				" or on the 'spec.objectives[*]' level (composite SLOs), but not both",
			Code: rules.ErrorCodeMutuallyExclusive,
		})
	})
}

func TestSLO_Validate_Spec_Indicator(t *testing.T) {
	runSLOIndicatorTests(t, "spec", func(indicator *SLOIndicatorInline, ref *string) SLO {
		slo := validRatioSLO()
		if indicator != nil && indicator.Spec.ThresholdMetric != nil {
			slo = validThresholdSLO()
		}
		slo.Spec.Indicator = indicator
		slo.Spec.IndicatorRef = ref
		return slo
	})
}

func TestSLO_Validate_Spec_TimeWindows(t *testing.T) {
	t.Run("missing timeWindow", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.TimeWindow = []SLOTimeWindow{}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.timeWindow",
			Code:         rules.ErrorCodeSliceLength,
		})
	})
	t.Run("too many timeWindows", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.TimeWindow = []SLOTimeWindow{
			slo.Spec.TimeWindow[0],
			slo.Spec.TimeWindow[0],
		}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.timeWindow",
			Code:         rules.ErrorCodeSliceLength,
		})
	})
	t.Run("missing duration", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.TimeWindow[0].Duration = DurationShorthand{}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.timeWindow[0].duration",
			Code:         rules.ErrorCodeRequired,
		})
	})
	t.Run("duration", func(t *testing.T) {
		runDurationShorthandTests(t, "spec.timeWindow[0].duration", func(d DurationShorthand) SLO {
			slo := validRatioSLO()
			slo.Spec.TimeWindow[0].Duration = d
			return slo
		})
	})
	t.Run("calendar set when isRolling is true", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.TimeWindow[0] = SLOTimeWindow{
			Duration:  NewDurationShorthand(1, DurationShorthandUnitWeek),
			IsRolling: true,
			Calendar: &SLOCalendar{
				StartTime: "2022-01-01 12:00:00",
				TimeZone:  "America/New_York",
			},
		}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.timeWindow[0]",
			Message:      "'calendar' cannot be set when 'isRolling' is true",
		})
	})
	t.Run("calendar missing when isRolling is false", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.TimeWindow[0] = SLOTimeWindow{
			Duration:  NewDurationShorthand(1, DurationShorthandUnitWeek),
			IsRolling: false,
			Calendar:  nil,
		}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.timeWindow[0]",
			Message:      "'calendar' must be set when 'isRolling' is false",
		})
	})
}

func TestSLO_Validate_Spec_Objectives(t *testing.T) {
	t.Run("target", func(t *testing.T) {
		for _, tc := range []struct {
			in        float64
			errorCode govy.ErrorCode
		}{
			{0.0, ""},
			{0.9999, ""},
			{1.0, rules.ErrorCodeLessThan},
			{-0.1, rules.ErrorCodeGreaterThanOrEqualTo},
		} {
			slo := validRatioSLO()
			slo.Spec.Objectives[0].Target = ptr(tc.in)
			slo.Spec.Objectives[0].TargetPercent = nil
			err := slo.Validate()
			if tc.errorCode != "" {
				govytest.AssertError(t, err, govytest.ExpectedRuleError{
					PropertyName: "spec.objectives[0].target",
					Code:         tc.errorCode,
				})
			} else {
				govytest.AssertNoError(t, err)
			}
		}
	})
	t.Run("target percent", func(t *testing.T) {
		for _, tc := range []struct {
			in        float64
			errorCode govy.ErrorCode
		}{
			{0.0, ""},
			{0.9999, ""},
			{99.9999, ""},
			{100.0, rules.ErrorCodeLessThan},
			{-0.1, rules.ErrorCodeGreaterThanOrEqualTo},
		} {
			slo := validRatioSLO()
			slo.Spec.Objectives[0].Target = nil
			slo.Spec.Objectives[0].TargetPercent = ptr(tc.in)
			err := slo.Validate()
			if tc.errorCode != "" {
				govytest.AssertError(t, err, govytest.ExpectedRuleError{
					PropertyName: "spec.objectives[0].targetPercent",
					Code:         tc.errorCode,
				})
			} else {
				govytest.AssertNoError(t, err)
			}
		}
	})
	t.Run("both target and targetPercent are missing", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.Objectives[0].Target = nil
		slo.Spec.Objectives[0].TargetPercent = nil
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.objectives[0]",
			Message:      "one of [target, targetPercent] properties must be set, none was provided",
			Code:         rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("both target and targetPercent are set", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.Objectives[0].Target = ptr(0.1)
		slo.Spec.Objectives[0].TargetPercent = ptr(10.0)
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.objectives[0]",
			Message:      "[target, targetPercent] properties are mutually exclusive, provide only one of them",
			Code:         rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("empty operator and value for ratio SLO", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.Objectives[0].Operator = ""
		slo.Spec.Objectives[0].Value = nil
		err := slo.Validate()
		govytest.AssertNoError(t, err)
	})
	t.Run("empty operator and value for threshold SLO with SLI ref", func(t *testing.T) {
		slo := validThresholdSLO()
		slo.Spec.Indicator = nil
		slo.Spec.IndicatorRef = ptr("my-sli")
		slo.Spec.Objectives[0].Operator = ""
		slo.Spec.Objectives[0].Value = nil
		err := slo.Validate()
		govytest.AssertNoError(t, err)
	})
	t.Run("empty operator and value for threshold SLO", func(t *testing.T) {
		slo := validThresholdSLO()
		slo.Spec.Objectives[0].Operator = ""
		slo.Spec.Objectives[0].Value = nil
		err := slo.Validate()
		govytest.AssertError(t, err,
			govytest.ExpectedRuleError{
				PropertyName: "spec.objectives[0].op",
				Code:         rules.ErrorCodeRequired,
			},
			govytest.ExpectedRuleError{
				PropertyName: "spec.objectives[0].value",
				Code:         rules.ErrorCodeRequired,
			},
		)
	})
	t.Run("operator", func(t *testing.T) {
		runOperatorTests(t, "spec.objectives[0].op", func(o Operator) SLO {
			slo := validThresholdSLO()
			slo.Spec.Objectives[0].Operator = o
			return slo
		})
	})
}

func TestSLO_Validate_Spec_CompositeObjectives(t *testing.T) {
	t.Run("indicator", func(t *testing.T) {
		runSLOIndicatorTests(t, "spec.objectives[0]", func(indicator *SLOIndicatorInline, ref *string) SLO {
			slo := validRatioSLO()
			slo.Spec.Indicator = nil
			slo.Spec.IndicatorRef = nil
			slo.Spec.Objectives[0].Indicator = indicator
			slo.Spec.Objectives[0].IndicatorRef = ref
			return slo
		})
	})
	t.Run("compositeWeight", func(t *testing.T) {
		for _, tc := range []struct {
			in        float64
			errorCode govy.ErrorCode
		}{
			{20.0, ""},
			{999999999999.9999, ""},
			{0.0, rules.ErrorCodeGreaterThan},
			{-2.0, rules.ErrorCodeGreaterThan},
		} {
			slo := validCompositeSLOWithSLIRef()
			slo.Spec.Objectives[0].CompositeWeight = ptr(tc.in)
			err := slo.Validate()
			if tc.errorCode != "" {
				govytest.AssertError(t, err, govytest.ExpectedRuleError{
					PropertyName: "spec.objectives[0].compositeWeight",
					Code:         tc.errorCode,
				})
			} else {
				govytest.AssertNoError(t, err)
			}
		}
	})
}

func TestSLO_Validate_Spec_Objectives_TimeSliceTarget(t *testing.T) {
	for _, method := range validSLOBudgetingMethods {
		t.Run(fmt.Sprintf("missing for %s method", method), func(t *testing.T) {
			slo := validRatioSLO()
			slo.Spec.BudgetingMethod = method
			slo.Spec.Objectives[0].TimeSliceTarget = nil
			slo.Spec.Objectives[0].TimeSliceWindow = ptr(NewDurationShorthand(1, "w"))
			err := slo.Validate()
			switch method {
			case SLOBudgetingMethodTimeslices:
				govytest.AssertError(t, err, govytest.ExpectedRuleError{
					PropertyName: "spec.objectives[0].timeSliceTarget",
					Code:         rules.ErrorCodeRequired,
				})
			default:
				govytest.AssertNoError(t, err)
			}
		})
	}
	testCases := []struct {
		in        float64
		errorCode govy.ErrorCode
	}{
		{0.1, ""},
		{1.0, ""},
		{0, rules.ErrorCodeGreaterThan},
		{1.1, rules.ErrorCodeLessThanOrEqualTo},
	}
	for _, tc := range testCases {
		slo := validRatioSLO()
		slo.Spec.Objectives[0].TimeSliceTarget = ptr(tc.in)
		err := slo.Validate()
		if tc.errorCode != "" {
			govytest.AssertError(t, err, govytest.ExpectedRuleError{
				PropertyName: "spec.objectives[0].timeSliceTarget",
				Code:         tc.errorCode,
			})
		} else {
			govytest.AssertNoError(t, err)
		}
	}
}

func TestSLO_Validate_Spec_Objectives_TimeSliceWindow(t *testing.T) {
	for _, method := range validSLOBudgetingMethods {
		t.Run(fmt.Sprintf("missing for %s method", method), func(t *testing.T) {
			slo := validRatioSLO()
			slo.Spec.BudgetingMethod = method
			slo.Spec.Objectives[0].TimeSliceTarget = ptr(0.9)
			slo.Spec.Objectives[0].TimeSliceWindow = nil
			err := slo.Validate()
			switch method {
			case SLOBudgetingMethodTimeslices, SLOBudgetingMethodRatioTimeslices:
				govytest.AssertError(t, err, govytest.ExpectedRuleError{
					PropertyName: "spec.objectives[0].timeSliceWindow",
					Code:         rules.ErrorCodeRequired,
				})
			default:
				govytest.AssertNoError(t, err)
			}
		})
	}
	t.Run("duration", func(t *testing.T) {
		runDurationShorthandTests(t, "spec.objectives[0].timeSliceWindow", func(d DurationShorthand) SLO {
			slo := validRatioSLO()
			slo.Spec.Objectives[0].TimeSliceWindow = &d
			return slo
		})
	})
}

func TestSLO_Validate_Spec_AlertPolicies(t *testing.T) {
	t.Run("no policies", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.AlertPolicies = nil
		err := slo.Validate()
		govytest.AssertNoError(t, err)
	})
	t.Run("both ref and inline are set", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.AlertPolicies[0].SLOAlertPolicyRef = &SLOAlertPolicyRef{}
		slo.Spec.AlertPolicies[0].SLOAlertPolicyInline = &SLOAlertPolicyInline{}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.alertPolicies[0]",
			Code:         rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("ref missing", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.AlertPolicies[0].SLOAlertPolicyRef = &SLOAlertPolicyRef{}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.alertPolicies[0].alertPolicyRef",
			Code:         rules.ErrorCodeRequired,
		})
	})
	t.Run("invalid condition ref", func(t *testing.T) {
		slo := validRatioSLO()
		slo.Spec.AlertPolicies[0].SLOAlertPolicyRef = &SLOAlertPolicyRef{
			AlertPolicyRef: "invalid ref",
		}
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.alertPolicies[0].alertPolicyRef",
			Code:         rules.ErrorCodeStringDNSLabel,
		})
	})
	t.Run("invalid inline kind", func(t *testing.T) {
		slo := validRatioSLOWithInlinedAlertPolicy()
		slo.Spec.AlertPolicies[0].Kind = openslo.KindDataSource
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: "spec.alertPolicies[0].kind",
			Code:         rules.ErrorCodeEqualTo,
		})
	})
	t.Run("metadata", func(t *testing.T) {
		runMetadataTests(t, "spec.alertPolicies[0].metadata", func(m Metadata) SLO {
			slo := validRatioSLOWithInlinedAlertPolicy()
			slo.Spec.AlertPolicies[0].Metadata = m
			return slo
		})
	})
	t.Run("spec", func(t *testing.T) {
		runAlertPolicySpecTests(t, "spec.alertPolicies[0].spec", func(s AlertPolicySpec) SLO {
			slo := validRatioSLOWithInlinedAlertPolicy()
			slo.Spec.AlertPolicies[0].Spec = s
			return slo
		})
	})
}

func runSLOIndicatorTests(t *testing.T, path string, sloGetter func(*SLOIndicatorInline, *string) SLO) {
	t.Helper()

	t.Run("both indicator and indicatorRef are provided", func(t *testing.T) {
		slo := sloGetter(&SLOIndicatorInline{}, new(string))
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: path,
			Message:      "[indicator, indicatorRef] properties are mutually exclusive, provide only one of them",
			Code:         rules.ErrorCodeMutuallyExclusive,
		})
	})
	t.Run("valid indicatorRef", func(t *testing.T) {
		slo := sloGetter(nil, ptr("my-sli"))
		err := slo.Validate()
		govytest.AssertNoError(t, err)
	})
	t.Run("invalid indicatorRef", func(t *testing.T) {
		slo := sloGetter(nil, ptr("my sli"))
		err := slo.Validate()
		govytest.AssertError(t, err, govytest.ExpectedRuleError{
			PropertyName: path + ".indicatorRef",
			Code:         rules.ErrorCodeStringDNSLabel,
		})
	})
	t.Run("indicator.metadata", func(t *testing.T) {
		runMetadataTests(t, path+".indicator.metadata", func(m Metadata) SLO {
			return sloGetter(&SLOIndicatorInline{
				Metadata: m,
				Spec:     validSLI().Spec,
			}, nil)
		})
	})
	t.Run("indicator.spec", func(t *testing.T) {
		runSLISpecTests(t, path+".indicator.spec", func(spec SLISpec) SLO {
			return sloGetter(&SLOIndicatorInline{
				Metadata: validSLI().Metadata,
				Spec:     spec,
			}, nil)
		})
	})
}

func TestSLO_IsComposite(t *testing.T) {
	slo := validRatioSLO()
	assert.False(t, slo.IsComposite())

	slo = validCompositeSLOWithSLIRef()
	assert.True(t, slo.IsComposite())

	t.Run("at least one objective is composite", func(t *testing.T) {
		slo.Spec.Objectives = append(slo.Spec.Objectives, slo.Spec.Objectives[0])
		slo.Spec.Objectives[0].Indicator = nil
		assert.True(t, slo.IsComposite())
	})
}

func validRatioSLO() SLO {
	return NewSLO(
		Metadata{
			Name:        "web-availability",
			DisplayName: "SLO for web availability",
			Labels: Labels{
				"team": {"team-a", "team-b"},
				"env":  {"prod"},
			},
		},
		SLOSpec{
			Description: "X% of search requests are successful",
			Service:     "web",
			Indicator: &SLOIndicatorInline{
				Metadata: Metadata{
					Name: "web-successful-requests-ratio",
				},
				Spec: SLISpec{
					RatioMetric: &SLIRatioMetric{
						Counter: true,
						Good: &SLIMetricSpec{
							MetricSource: SLIMetricSource{
								Type: "Prometheus",
								Spec: map[string]any{
									"query": `sum(http_requests{k8s_cluster="prod",component="web",code=~"2xx|4xx"})`,
								},
							},
						},
						Total: &SLIMetricSpec{
							MetricSource: SLIMetricSource{
								Type: "Prometheus",
								Spec: map[string]any{
									"query": `sum(http_requests{k8s_cluster="prod",component="web"})`,
								},
							},
						},
					},
				},
			},
			TimeWindow: []SLOTimeWindow{
				{
					Duration:  NewDurationShorthand(1, DurationShorthandUnitWeek),
					IsRolling: false,
					Calendar: &SLOCalendar{
						StartTime: "2022-01-01 12:00:00",
						TimeZone:  "America/New_York",
					},
				},
			},
			BudgetingMethod: SLOBudgetingMethodTimeslices,
			Objectives: []SLOObjective{
				{
					DisplayName:     "Good",
					Target:          ptr(0.995),
					TimeSliceTarget: ptr(0.95),
					TimeSliceWindow: ptr(NewDurationShorthand(1, "m")),
				},
			},
			AlertPolicies: []SLOAlertPolicy{
				{SLOAlertPolicyRef: &SLOAlertPolicyRef{AlertPolicyRef: "alert-policy-1"}},
			},
		},
	)
}

func validThresholdSLO() SLO {
	return NewSLO(
		Metadata{
			Name:        "annotator-throughput",
			DisplayName: "SLO for Annotator service throughput",
			Labels: Labels{
				"team": {"team-a", "team-b"},
				"env":  {"prod"},
			},
		},
		SLOSpec{
			Description: "X% of time messages are processed without delay by the processing pipeline (expected value ~100%)",
			Service:     "annotator",
			Indicator: &SLOIndicatorInline{
				Metadata: Metadata{
					Name: "inlined-sli",
				},
				Spec: SLISpec{
					ThresholdMetric: &SLIMetricSpec{
						MetricSource: SLIMetricSource{
							Type: "Prometheus",
							Spec: map[string]any{
								"query": `sum(min_over_time(kafka_consumergroup_lag{k8s_cluster="prod", consumergroup="annotator", topic="annotator-in"}[2m]))`,
							},
						},
					},
				},
			},
			TimeWindow: []SLOTimeWindow{
				{
					Duration:  NewDurationShorthand(1, DurationShorthandUnitWeek),
					IsRolling: false,
					Calendar: &SLOCalendar{
						StartTime: "2022-01-01 12:00:00",
						TimeZone:  "America/New_York",
					},
				},
			},
			BudgetingMethod: SLOBudgetingMethodTimeslices,
			Objectives: []SLOObjective{
				{
					DisplayName:     "Good",
					Operator:        OperatorGTE,
					Value:           ptr(10.0),
					Target:          ptr(0.995),
					TimeSliceTarget: ptr(0.95),
					TimeSliceWindow: ptr(NewDurationShorthand(1, "m")),
				},
			},
			AlertPolicies: []SLOAlertPolicy{
				{SLOAlertPolicyRef: &SLOAlertPolicyRef{AlertPolicyRef: "alert-policy-1"}},
			},
		},
	)
}
func validSLOWithSLIRef() SLO {
	return NewSLO(
		Metadata{
			Name:        "web-availability",
			DisplayName: "SLO for web availability",
			Labels: Labels{
				"team": {"team-a", "team-b"},
				"env":  {"prod"},
			},
		},
		SLOSpec{
			Description:  "X% of search requests are successful",
			Service:      "web",
			IndicatorRef: ptr("my-sli"),
			TimeWindow: []SLOTimeWindow{
				{
					Duration:  NewDurationShorthand(1, DurationShorthandUnitWeek),
					IsRolling: false,
					Calendar: &SLOCalendar{
						StartTime: "2022-01-01 12:00:00",
						TimeZone:  "America/New_York",
					},
				},
			},
			BudgetingMethod: SLOBudgetingMethodTimeslices,
			Objectives: []SLOObjective{
				{
					DisplayName:     "Good",
					Target:          ptr(0.995),
					TimeSliceTarget: ptr(0.95),
					TimeSliceWindow: ptr(NewDurationShorthand(1, "m")),
				},
			},
			AlertPolicies: []SLOAlertPolicy{
				{SLOAlertPolicyRef: &SLOAlertPolicyRef{AlertPolicyRef: "alert-policy-1"}},
			},
		},
	)
}

func validRatioSLOWithInlinedAlertPolicy() SLO {
	slo := validRatioSLO()
	alertPolicy := validAlertPolicy()
	slo.Spec.AlertPolicies[0] = SLOAlertPolicy{
		SLOAlertPolicyInline: &SLOAlertPolicyInline{
			Kind:     alertPolicy.Kind,
			Metadata: alertPolicy.Metadata,
			Spec:     alertPolicy.Spec,
		},
	}
	return slo
}

func validCompositeSLOWithSLIRef() SLO {
	slo := validRatioSLO()
	slo.Spec.Indicator = nil
	slo.Spec.IndicatorRef = nil
	slo.Spec.Objectives[0].Indicator = nil
	slo.Spec.Objectives[0].IndicatorRef = ptr("my-sli")
	slo.Spec.Objectives[0].CompositeWeight = ptr(1.0)
	return slo
}

func validCompositeSLOWithInlinedSLI() SLO {
	slo := validRatioSLO()
	sli := validSLI()
	slo.Spec.Indicator = nil
	slo.Spec.IndicatorRef = nil
	slo.Spec.Objectives[0].IndicatorRef = nil
	slo.Spec.Objectives[0].Indicator = &SLOIndicatorInline{
		Metadata: sli.Metadata,
		Spec:     sli.Spec,
	}
	slo.Spec.Objectives[0].CompositeWeight = ptr(1.0)
	return slo
}

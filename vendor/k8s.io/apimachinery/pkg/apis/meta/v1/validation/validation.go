/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validation

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func ValidateLabelSelector(ps *metav1.LabelSelector, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if ps == nil {
		return allErrs
	}
	allErrs = append(allErrs, ValidateLabels(ps.MatchLabels, fldPath.Child("matchLabels"))...)
	for i, expr := range ps.MatchExpressions {
		allErrs = append(allErrs, ValidateLabelSelectorRequirement(expr, fldPath.Child("matchExpressions").Index(i))...)
	}
	return allErrs
}

func ValidateLabelSelectorRequirement(sr metav1.LabelSelectorRequirement, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	switch sr.Operator {
	case metav1.LabelSelectorOpIn, metav1.LabelSelectorOpNotIn:
		if len(sr.Values) == 0 {
			allErrs = append(allErrs, field.Required(fldPath.Child("values"), "must be specified when `operator` is 'In' or 'NotIn'"))
		}
	case metav1.LabelSelectorOpExists, metav1.LabelSelectorOpDoesNotExist:
		if len(sr.Values) > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("values"), "may not be specified when `operator` is 'Exists' or 'DoesNotExist'"))
		}
	default:
		allErrs = append(allErrs, field.Invalid(fldPath.Child("operator"), sr.Operator, "not a valid selector operator"))
	}
	allErrs = append(allErrs, ValidateLabelName(sr.Key, fldPath.Child("key"))...)
	return allErrs
}

// ValidateLabelName validates that the label name is correctly defined.
func ValidateLabelName(labelName string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for _, msg := range validation.IsQualifiedName(labelName) {
		allErrs = append(allErrs, field.Invalid(fldPath, labelName, msg))
	}
	return allErrs
}

// ValidateLabels validates that a set of labels are correctly defined.
func ValidateLabels(labels map[string]string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	for k, v := range labels {
		allErrs = append(allErrs, ValidateLabelName(k, fldPath)...)
		for _, msg := range validation.IsValidLabelValue(v) {
			allErrs = append(allErrs, field.Invalid(fldPath, v, msg))
		}
	}
	return allErrs
}

func ValidateDeleteOptions(options *metav1.DeleteOptions) field.ErrorList {
	allErrs := field.ErrorList{}
	if options.OrphanDependents != nil && options.PropagationPolicy != nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("propagationPolicy"), options.PropagationPolicy, "orphanDependents and deletionPropagation cannot be both set"))
	}
	if options.PropagationPolicy != nil &&
		*options.PropagationPolicy != metav1.DeletePropagationForeground &&
		*options.PropagationPolicy != metav1.DeletePropagationBackground &&
		*options.PropagationPolicy != metav1.DeletePropagationOrphan {
		allErrs = append(allErrs, field.NotSupported(field.NewPath("propagationPolicy"), options.PropagationPolicy, []string{string(metav1.DeletePropagationForeground), string(metav1.DeletePropagationBackground), string(metav1.DeletePropagationOrphan), "nil"}))
	}
	allErrs = append(allErrs, validateDryRun(field.NewPath("dryRun"), options.DryRun)...)
	return allErrs
}

func ValidateCreateOptions(options *metav1.CreateOptions) field.ErrorList {
	return validateDryRun(field.NewPath("dryRun"), options.DryRun)
}

func ValidateUpdateOptions(options *metav1.UpdateOptions) field.ErrorList {
	return validateDryRun(field.NewPath("dryRun"), options.DryRun)
}

func ValidatePatchOptions(options *metav1.PatchOptions, patchType types.PatchType) field.ErrorList {
	allErrs := field.ErrorList{}
	if patchType != types.ApplyPatchType && options.Force != nil {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("force"), "may not be specified for non-apply patch"))
	}
	allErrs = append(allErrs, validateDryRun(field.NewPath("dryRun"), options.DryRun)...)
	return allErrs
}

var allowedDryRunValues = sets.NewString(metav1.DryRunAll)

func validateDryRun(fldPath *field.Path, dryRun []string) field.ErrorList {
	allErrs := field.ErrorList{}
	if !allowedDryRunValues.HasAll(dryRun...) {
		allErrs = append(allErrs, field.NotSupported(fldPath, dryRun, allowedDryRunValues.List()))
	}
	return allErrs
}

const UninitializedStatusUpdateErrorMsg string = `must not update status when the object is uninitialized`

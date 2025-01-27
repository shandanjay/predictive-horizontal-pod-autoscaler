/*
Copyright 2021 The Predictive Horizontal Pod Autoscaler Authors.

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

package fake

import (
	"github.com/jthomperoo/custom-pod-autoscaler/v2/evaluate"
	"github.com/jthomperoo/horizontal-pod-autoscaler/metric"
)

// Evaluater (fake) provides a way to insert functionality into a Evaluater
type Evaluater struct {
	GetEvaluationReactor func(gatheredMetrics []*metric.Metric) (*evaluate.Evaluation, error)
}

// GetEvaluation calls the fake Evaluater function
func (f *Evaluater) GetEvaluation(gatheredMetrics []*metric.Metric) (*evaluate.Evaluation, error) {
	return f.GetEvaluationReactor(gatheredMetrics)
}

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

package linear_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/config"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/fake"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/prediction/linear"
	"github.com/jthomperoo/predictive-horizontal-pod-autoscaler/internal/stored"
)

func TestPredict_GetPrediction(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    int32
		expectedErr error
		predicter   *linear.Predict
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Fail no Linear configuration",
			0,
			errors.New("No Linear configuration provided for model"),
			&linear.Predict{},
			&config.Model{},
			[]*stored.Evaluation{},
		},
		{
			"Fail no evaluations",
			0,
			errors.New("No evaluations provided for Linear regression model"),
			&linear.Predict{},
			&config.Model{
				Type: linear.Type,
				Linear: &config.Linear{
					StoredValues: 5,
					LookAhead:    0,
				},
			},
			[]*stored.Evaluation{},
		},
		{
			"Success, only one evaluation, return without the prediction",
			32,
			nil,
			&linear.Predict{},
			&config.Model{
				Type: linear.Type,
				Linear: &config.Linear{
					StoredValues: 5,
					LookAhead:    0,
				},
			},
			[]*stored.Evaluation{
				{
					ID: 0,
					Evaluation: stored.DBEvaluation{
						TargetReplicas: 32,
					},
				},
			},
		},
		{
			"Fail execution of algorithm fails",
			0,
			errors.New("algorithm fail"),
			&linear.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "", errors.New("algorithm fail")
					},
				},
			},
			&config.Model{
				Type: linear.Type,
				Linear: &config.Linear{
					StoredValues: 5,
					LookAhead:    0,
				},
			},
			[]*stored.Evaluation{
				{
					ID: 0,
				},
				{
					ID: 1,
				},
			},
		},
		{
			"Fail algorithm returns non-integer castable value",
			0,
			errors.New(`strconv.Atoi: parsing "invalid": invalid syntax`),
			&linear.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "invalid", nil
					},
				},
			},
			&config.Model{
				Type: linear.Type,
				Linear: &config.Linear{
					StoredValues: 5,
					LookAhead:    0,
				},
			},
			[]*stored.Evaluation{
				{
					ID: 0,
				},
				{
					ID: 1,
				},
			},
		},
		{
			"Success",
			3,
			nil,
			&linear.Predict{
				Runner: &fake.Run{
					RunAlgorithmWithValueReactor: func(algorithmPath, value string, timeout int) (string, error) {
						return "3", nil
					},
				},
			},
			&config.Model{
				Type: linear.Type,
				Linear: &config.Linear{
					StoredValues: 5,
					LookAhead:    0,
				},
			},
			[]*stored.Evaluation{
				{
					ID: 0,
				},
				{
					ID: 1,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result, err := test.predicter.GetPrediction(test.model, test.evaluations)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("result mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestModelPredict_GetIDsToRemove(t *testing.T) {
	equateErrorMessage := cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})

	var tests = []struct {
		description string
		expected    []int
		expectedErr error
		model       *config.Model
		evaluations []*stored.Evaluation
	}{
		{
			"Fail no Linear configuration",
			nil,
			errors.New("No Linear configuration provided for model"),
			&config.Model{},
			[]*stored.Evaluation{},
		},
		{
			"3 IDs too many, mark 3 for removal",
			[]int{5, 3, 8},
			nil,
			&config.Model{
				Linear: &config.Linear{
					StoredValues: 3,
				},
			},
			[]*stored.Evaluation{
				{
					ID:      1,
					Created: time.Time{}.Add(time.Duration(4) * time.Second),
				},
				{
					ID:      2,
					Created: time.Time{}.Add(time.Duration(5) * time.Second),
				},
				// START OLDEST
				{
					ID:      5,
					Created: time.Time{}.Add(time.Duration(1) * time.Second),
				},
				{
					ID:      3,
					Created: time.Time{}.Add(time.Duration(2) * time.Second),
				},
				{
					ID:      8,
					Created: time.Time{}.Add(time.Duration(3) * time.Second),
				},
				// END OLDEST
				{
					ID:      4,
					Created: time.Time{}.Add(time.Duration(6) * time.Second),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &linear.Predict{}
			result, err := predicter.GetIDsToRemove(test.model, test.evaluations)
			if !cmp.Equal(&err, &test.expectedErr, equateErrorMessage) {
				t.Errorf("error mismatch (-want +got):\n%s", cmp.Diff(test.expectedErr, err, equateErrorMessage))
				return
			}
			if !cmp.Equal(test.expected, result) {
				t.Errorf("remove IDs mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestPredict_GetType(t *testing.T) {
	var tests = []struct {
		description string
		expected    string
	}{
		{
			"Successful get type",
			"Linear",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			predicter := &linear.Predict{}
			result := predicter.GetType()
			if !cmp.Equal(test.expected, result) {
				t.Errorf("type mismatch (-want +got):\n%s", cmp.Diff(test.expected, result))
			}
		})
	}
}

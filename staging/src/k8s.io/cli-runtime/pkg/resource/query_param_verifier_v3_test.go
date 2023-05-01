/*
Copyright 2023 The Kubernetes Authors.

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

package resource

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/openapi/cached"
	"k8s.io/client-go/openapi/openapitest"
	"k8s.io/client-go/openapi3"
)

func TestV3SupportsQueryParamBatchV1(t *testing.T) {
	tests := map[string]struct {
		crds             []schema.GroupKind      // CRDFinder returns these CRD's
		gvk              schema.GroupVersionKind // GVK whose OpenAPI V3 spec is checked
		queryParam       VerifiableQueryParam    // Usually "fieldValidation"
		expectedSupports bool
	}{
		"Field validation query param is supported for batch/v1/Job": {
			crds: []schema.GroupKind{},
			gvk: schema.GroupVersionKind{
				Group:   "batch",
				Version: "v1",
				Kind:    "Job",
			},
			queryParam:       QueryParamFieldValidation,
			expectedSupports: true,
		},
		"Field validation query param supported for core/v1/Namespace": {
			crds: []schema.GroupKind{},
			gvk: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Namespace",
			},
			queryParam:       QueryParamFieldValidation,
			expectedSupports: true,
		},
		"Field validation unsupported for unknown GVK": {
			crds: []schema.GroupKind{},
			gvk: schema.GroupVersionKind{
				Group:   "bad",
				Version: "v1",
				Kind:    "Uknown",
			},
			queryParam:       QueryParamFieldValidation,
			expectedSupports: false,
		},
		"Unknown query param unsupported (for all GVK's)": {
			crds: []schema.GroupKind{},
			gvk: schema.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    "Deployment",
			},
			queryParam:       "UnknownQueryParam",
			expectedSupports: false,
		},
		"Field validation query param supported for found CRD": {
			crds: []schema.GroupKind{
				{
					Group: "example.com",
					Kind:  "ExampleCRD",
				},
			},
			// GVK matches above CRD GroupKind
			gvk: schema.GroupVersionKind{
				Group:   "example.com",
				Version: "v1",
				Kind:    "ExampleCRD",
			},
			queryParam:       QueryParamFieldValidation,
			expectedSupports: true,
		},
		"Field validation query param unsupported for missing CRD": {
			crds: []schema.GroupKind{
				{
					Group: "different.com",
					Kind:  "DifferentCRD",
				},
			},
			// GVK does NOT match above CRD GroupKind
			gvk: schema.GroupVersionKind{
				Group:   "example.com",
				Version: "v1",
				Kind:    "ExampleCRD",
			},
			queryParam:       QueryParamFieldValidation,
			expectedSupports: false,
		},
		"List GVK is specifically unsupported": {
			crds: []schema.GroupKind{},
			gvk: schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "List",
			},
			queryParam:       QueryParamFieldValidation,
			expectedSupports: false,
		},
	}

	root := openapi3.NewRoot(cached.NewClient(openapitest.NewEmbeddedFileClient()))
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			verifier := &queryParamVerifierV3{
				finder: NewCRDFinder(func() ([]schema.GroupKind, error) {
					return tc.crds, nil
				}),
				root:       root,
				queryParam: tc.queryParam,
			}
			err := verifier.HasSupport(tc.gvk)
			if tc.expectedSupports && err != nil {
				t.Errorf("Expected supports, but returned err for GVK (%s)", tc.gvk)
			} else if !tc.expectedSupports && err == nil {
				t.Errorf("Expected not supports, but returned no err for GVK (%s)", tc.gvk)
			}
		})
	}
}

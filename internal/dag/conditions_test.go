// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dag

import (
	"testing"

	contour_api_v1 "github.com/projectcontour/contour/apis/projectcontour/v1"
	"github.com/stretchr/testify/assert"
)

func TestPathMatchCondition(t *testing.T) {
	tests := map[string]struct {
		matchconditions []contour_api_v1.MatchCondition
		want            MatchCondition
	}{
		"empty condition list": {
			matchconditions: nil,
			want:            &PrefixMatchCondition{Prefix: "/"},
		},
		"single slash": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/",
			}},
			want: &PrefixMatchCondition{Prefix: "/"},
		},
		"two slashes": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/",
			}, {
				Prefix: "/",
			}},
			want: &PrefixMatchCondition{Prefix: "/"},
		},
		"mixed matchconditions": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/a/",
			}, {
				Prefix: "/b",
			}},
			want: &PrefixMatchCondition{Prefix: "/a/b"},
		},
		"trailing slash": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/a/",
			}},
			want: &PrefixMatchCondition{Prefix: "/a/"},
		},
		"trailing slash on second prefix condition": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/a",
			},
				{
					Prefix: "/b/",
				}},
			want: &PrefixMatchCondition{Prefix: "/a/b/"},
		},
		"nothing but slashes": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Prefix: "///",
				},
				{
					Prefix: "/",
				}},
			want: &PrefixMatchCondition{Prefix: "/"},
		},
		"header condition": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: new(contour_api_v1.HeaderMatchCondition),
			}},
			want: &PrefixMatchCondition{Prefix: "/"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := mergePathMatchConditions(tc.matchconditions)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestHeaderMatchConditions(t *testing.T) {
	tests := map[string]struct {
		matchconditions []contour_api_v1.MatchCondition
		want            []HeaderMatchCondition
	}{
		"empty condition list": {
			matchconditions: nil,
			want:            nil,
		},
		"prefix": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/",
			}},
			want: nil,
		},
		"header condition empty": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: new(contour_api_v1.HeaderMatchCondition),
			}},
			want: nil,
		},
		"header present": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:    "x-request-id",
					Present: true,
				},
			}},
			want: []HeaderMatchCondition{{
				Name:      "x-request-id",
				MatchType: "present",
			}},
		},
		"header name but missing condition": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name: "x-request-id",
				},
			}},
			// this should be filtered out beforehand, but in case it leaks
			// through the behavior is to ignore the header contains entry.
			want: nil,
		},
		"header contains": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-request-id",
					Contains: "abcdef",
				},
			}},
			want: []HeaderMatchCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
			}},
		},
		"header not contains": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:        "x-request-id",
					NotContains: "abcdef",
				},
			}},
			want: []HeaderMatchCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
				Invert:    true,
			}},
		},
		"header exact": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:  "x-request-id",
					Exact: "abcdef",
				},
			}},
			want: []HeaderMatchCondition{{
				Name:      "x-request-id",
				MatchType: "exact",
				Value:     "abcdef",
			}},
		},
		"header not exact": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-request-id",
					NotExact: "abcdef",
				},
			}},
			want: []HeaderMatchCondition{{
				Name:      "x-request-id",
				MatchType: "exact",
				Value:     "abcdef",
				Invert:    true,
			}},
		},
		"two header contains": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-request-id",
					Contains: "abcdef",
				},
			}, {
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-request-id",
					Contains: "cedfg",
				},
			}},
			want: []HeaderMatchCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
			}, {
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "cedfg",
			}},
		},
		"two header contains different case": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-request-id",
					Contains: "abcdef",
				},
			}, {
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "X-Request-Id",
					Contains: "abcdef",
				},
			}},
			want: []HeaderMatchCondition{{
				Name:      "x-request-id",
				MatchType: "contains",
				Value:     "abcdef",
			}, {
				Name:      "X-Request-Id",
				MatchType: "contains",
				Value:     "abcdef",
			}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := mergeHeaderMatchConditions(tc.matchconditions)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestPrefixMatchConditionsValid(t *testing.T) {
	tests := map[string]struct {
		matchconditions []contour_api_v1.MatchCondition
		want            bool
	}{
		"empty condition list": {
			matchconditions: nil,
			want:            true,
		},
		"valid path condition only": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/api",
			}},
			want: true,
		},
		"valid path condition with headers": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/api",
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-header",
					Contains: "abc",
				},
			}},
			want: true,
		},
		"two prefix matchconditions": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/api",
			}, {
				Prefix: "/v1",
			}},
			want: false,
		},
		"two prefix matchconditions with headers": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "/api",
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-header",
					Contains: "abc",
				},
			}, {
				Prefix: "/v1",
			}},
			want: false,
		},
		"invalid prefix condition": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "api",
			}},
			want: false,
		},
		"invalid prefix condition with headers": {
			matchconditions: []contour_api_v1.MatchCondition{{
				Prefix: "api",
				Header: &contour_api_v1.HeaderMatchCondition{
					Name:     "x-header",
					Contains: "abc",
				},
			}},
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := pathMatchConditionsValid(tc.matchconditions)
			assert.Equal(t, tc.want, err == nil)
		})
	}
}

func TestValidateHeaderMatchConditions(t *testing.T) {
	tests := map[string]struct {
		matchconditions []contour_api_v1.MatchCondition
		wantErr         bool
	}{
		"empty condition list": {
			matchconditions: nil,
			wantErr:         false,
		},
		"prefix only": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Prefix: "/blog",
				},
			},
			wantErr: false,
		},
		"valid matchconditions": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				},
			},
			wantErr: false,
		},
		"prefix matchconditions + valid headers": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Prefix: "/blog",
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:        "x-header",
						NotContains: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:        "another-header",
						NotContains: "123",
					},
				},
			},
			wantErr: false,
		},
		"multiple 'exact' matchconditions for the same header are invalid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:  "x-header",
						Exact: "123",
					},
				},
			},
			wantErr: true,
		},
		"multiple 'exact' matchconditions for different headers are valid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:  "x-different-header",
						Exact: "123",
					},
				},
			},
			wantErr: false,
		},
		"'exact' and 'notexact' matchconditions for the same header with the same value are invalid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:     "x-header",
						NotExact: "abc",
					},
				},
			},
			wantErr: true,
		},
		"'exact' and 'notexact' matchconditions for the same header with different values are valid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:     "x-header",
						NotExact: "def",
					},
				},
			},
			wantErr: false,
		},
		"'exact' and 'notexact' matchconditions for different headers with the same value are valid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:  "x-header",
						Exact: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:     "x-another-header",
						NotExact: "abc",
					},
				},
			},
			wantErr: false,
		},
		"'contains' and 'notcontains' matchconditions for the same header with the same value are invalid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:        "x-header",
						NotContains: "abc",
					},
				},
			},
			wantErr: true,
		},
		"'contains' and 'notcontains' matchconditions for the same header with different values are valid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:        "x-header",
						NotContains: "def",
					},
				},
			},
			wantErr: false,
		},
		"'contains' and 'notcontains' matchconditions for different headers with the same value are valid": {
			matchconditions: []contour_api_v1.MatchCondition{
				{
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:     "x-header",
						Contains: "abc",
					},
				}, {
					Header: &contour_api_v1.HeaderMatchCondition{
						Name:        "x-another-header",
						NotContains: "abc",
					},
				},
			},
			wantErr: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotErr := headerMatchConditionsValid(tc.matchconditions)

			if !tc.wantErr {
				assert.NoError(t, gotErr)
			}

			if tc.wantErr {
				assert.Error(t, gotErr)
			}
		})
	}
}

package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getAgenciesGSI(t *testing.T) {
	type testCase struct {
		platformAdmin bool
		sort          map[string]string
	}

	testCases := map[string]testCase{
		"platform admin sorting": {
			platformAdmin: true,
			sort: map[string]string{
				pagesSortCreatedAsc:   "type-created-index",
				pagesSortCreatedDesc:  "type-created-index",
				pagesSortModifiedAsc:  "type-modified-index",
				pagesSortModifiedDesc: "type-modified-index",
				pagesSortNameAsc:      "type-name-index",
				pagesSortNameDesc:     "type-name-index",
			},
		},
		"member sorting": {
			platformAdmin: false,
			sort: map[string]string{
				pagesSortCreatedAsc:   "idpid-agency_created-index",
				pagesSortCreatedDesc:  "idpid-agency_created-index",
				pagesSortModifiedAsc:  "idpid-agency_modified-index",
				pagesSortModifiedDesc: "idpid-agency_modified-index",
				pagesSortNameAsc:      "idpid-name-index",
				pagesSortNameDesc:     "idpid-name-index",
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			for sort, expectedGsi := range tc.sort {
				t.Run(sort, func(t *testing.T) {
					assert.Equal(t, expectedGsi, getAgenciesGSI(tc.platformAdmin, sort))
				})
			}
		})
	}
}

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
				agenciesSortCreatedAsc:   "type-created-index",
				agenciesSortCreatedDesc:  "type-created-index",
				agenciesSortModifiedAsc:  "type-modified-index",
				agenciesSortModifiedDesc: "type-modified-index",
				agenciesSortNameAsc:      "type-name-index",
				agenciesSortNameDesc:     "type-name-index",
			},
		},
		"member sorting": {
			platformAdmin: false,
			sort: map[string]string{
				agenciesSortCreatedAsc:   "idpid-agency_created-index",
				agenciesSortCreatedDesc:  "idpid-agency_created-index",
				agenciesSortModifiedAsc:  "idpid-agency_modified-index",
				agenciesSortModifiedDesc: "idpid-agency_modified-index",
				agenciesSortNameAsc:      "idpid-name-index",
				agenciesSortNameDesc:     "idpid-name-index",
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

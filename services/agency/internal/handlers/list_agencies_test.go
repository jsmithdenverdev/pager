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
				agenciesSortModifiedAsc:  "type-created-index",
				agenciesSortModifiedDesc: "type-created-index",
				agenciesSortNameAsc:      "type-created-index",
				agenciesSortNameDesc:     "type-created-index",
			},
		},
		"member sorting": {
			platformAdmin: false,
			sort: map[string]string{
				agenciesSortCreatedAsc:   "idpid-created-index",
				agenciesSortCreatedDesc:  "idpid-created-index",
				agenciesSortModifiedAsc:  "idpid-created-index",
				agenciesSortModifiedDesc: "idpid-created-index",
				agenciesSortNameAsc:      "idpid-created-index",
				agenciesSortNameDesc:     "idpid-created-index",
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			for direction, expectedGsi := range tc.sort {
				t.Run(direction, func(t *testing.T) {
					assert.Equal(t, expectedGsi, getAgenciesGSI(tc.platformAdmin, listAgenciesRequest{
						Sort: direction,
					}))
				})
			}
		})
	}
}

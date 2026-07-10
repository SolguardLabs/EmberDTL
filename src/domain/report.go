package domain

import (
	"sort"

	"emberdtl/src/amount"
)

type AmountEntry struct {
	ID     string        `json:"id"`
	Amount amount.Amount `json:"amount"`
}

type AssetReport struct {
	ID       string            `json:"id"`
	Symbol   string            `json:"symbol"`
	Decimals int               `json:"decimals"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type AccountReport struct {
	ID            string        `json:"id"`
	Role          AccountRole   `json:"role"`
	Status        AccountStatus `json:"status"`
	Free          []AmountEntry `json:"free"`
	Held          []AmountEntry `json:"held"`
	SettledIn     []AmountEntry `json:"settledIn"`
	Withdrawn     []AmountEntry `json:"withdrawn"`
	FeesPaid      []AmountEntry `json:"feesPaid"`
	Contributed   []AmountEntry `json:"contributed"`
	ClaimsPaid    []AmountEntry `json:"claimsPaid"`
	DefaultsTaken []AmountEntry `json:"defaultsTaken"`
}

type PoolReport struct {
	ID                  string        `json:"id"`
	Asset               string        `json:"asset"`
	Status              PoolStatus    `json:"status"`
	Balance             amount.Amount `json:"balance"`
	Contributions       amount.Amount `json:"contributions"`
	FeeContributions    amount.Amount `json:"feeContributions"`
	ClaimsPaid          amount.Amount `json:"claimsPaid"`
	Recovered           amount.Amount `json:"recovered"`
	ReservedForDefaults amount.Amount `json:"reservedForDefaults"`
	PendingDefaults     amount.Amount `json:"pendingDefaults"`
	PendingClaims       amount.Amount `json:"pendingClaims"`
}

type ReserveReport struct {
	ID                     string        `json:"id"`
	Asset                  string        `json:"asset"`
	Owner                  string        `json:"owner"`
	Status                 ReserveStatus `json:"status"`
	Balance                amount.Amount `json:"balance"`
	Held                   amount.Amount `json:"held"`
	GrossDeposits          amount.Amount `json:"grossDeposits"`
	GrossWithdrawals       amount.Amount `json:"grossWithdrawals"`
	SettlementVolume       amount.Amount `json:"settlementVolume"`
	AccruedFees            amount.Amount `json:"accruedFees"`
	InsuranceContributions amount.Amount `json:"insuranceContributions"`
}

type FacilityReport struct {
	ID          string         `json:"id"`
	ReserveID   string         `json:"reserveId"`
	Asset       string         `json:"asset"`
	Borrower    string         `json:"borrower"`
	Beneficiary string         `json:"beneficiary"`
	Operator    string         `json:"operator"`
	Status      FacilityStatus `json:"status"`
	Principal   amount.Amount  `json:"principal"`
	Outstanding amount.Amount  `json:"outstanding"`
	Repaid      amount.Amount  `json:"repaid"`
	FeesPaid    amount.Amount  `json:"feesPaid"`
	OpenedEpoch int            `json:"openedEpoch"`
	DueEpoch    int            `json:"dueEpoch"`
	DefaultID   string         `json:"defaultId,omitempty"`
}

type DefaultReport struct {
	ID               string        `json:"id"`
	FacilityID       string        `json:"facilityId"`
	Asset            string        `json:"asset"`
	Reporter         string        `json:"reporter"`
	Borrower         string        `json:"borrower"`
	Beneficiary      string        `json:"beneficiary"`
	Status           DefaultStatus `json:"status"`
	Amount           amount.Amount `json:"amount"`
	ExpectedRecovery amount.Amount `json:"expectedRecovery"`
	CoverageCeiling  amount.Amount `json:"coverageCeiling"`
	PendingExposure  amount.Amount `json:"pendingExposure"`
	ClaimedCoverage  amount.Amount `json:"claimedCoverage"`
	Recovered        amount.Amount `json:"recovered"`
	ReportedEpoch    int           `json:"reportedEpoch"`
	AcceptedEpoch    int           `json:"acceptedEpoch,omitempty"`
	ResolvedEpoch    int           `json:"resolvedEpoch,omitempty"`
	Claims           []string      `json:"claims"`
}

type ClaimReport struct {
	ID                     string        `json:"id"`
	DefaultID              string        `json:"defaultId"`
	FacilityID             string        `json:"facilityId"`
	Asset                  string        `json:"asset"`
	Claimant               string        `json:"claimant"`
	PayoutAccount          string        `json:"payoutAccount"`
	Status                 ClaimStatus   `json:"status"`
	Amount                 amount.Amount `json:"amount"`
	Coverage               amount.Amount `json:"coverage"`
	CapacityAtRegistration amount.Amount `json:"capacityAtRegistration"`
	RegisteredEpoch        int           `json:"registeredEpoch"`
	ExecutedEpoch          int           `json:"executedEpoch,omitempty"`
}

type ReconciliationRow struct {
	Asset             string        `json:"asset"`
	ReserveBalance    amount.Amount `json:"reserveBalance"`
	FacilityExposure  amount.Amount `json:"facilityExposure"`
	PoolBalance       amount.Amount `json:"poolBalance"`
	PendingDefaults   amount.Amount `json:"pendingDefaults"`
	PendingClaims     amount.Amount `json:"pendingClaims"`
	NetCoverageBuffer amount.Amount `json:"netCoverageBuffer"`
}

type EventReport = Event

type SystemReport struct {
	Name           string                   `json:"name"`
	Epoch          int                      `json:"epoch"`
	Assets         []AssetReport            `json:"assets"`
	Accounts       []AccountReport          `json:"accounts"`
	Pools          []PoolReport             `json:"pools"`
	Reserves       []ReserveReport          `json:"reserves"`
	Facilities     []FacilityReport         `json:"facilities"`
	Defaults       []DefaultReport          `json:"defaults"`
	Claims         []ClaimReport            `json:"claims"`
	Reconciliation []ReconciliationRow      `json:"reconciliation"`
	Metrics        map[string]amount.Amount `json:"metrics"`
	Events         []EventReport            `json:"events,omitempty"`
	Validations    []string                 `json:"validations,omitempty"`
}

func BuildReport(state *State, includeEvents bool, validations []string) SystemReport {
	report := SystemReport{
		Name:           state.Name,
		Epoch:          state.Epoch,
		Assets:         assetReports(state),
		Accounts:       accountReports(state),
		Pools:          poolReports(state),
		Reserves:       reserveReports(state),
		Facilities:     facilityReports(state),
		Defaults:       defaultReports(state),
		Claims:         claimReports(state),
		Reconciliation: reconciliationRows(state),
		Metrics:        copyMetrics(state.Metrics),
		Validations:    append([]string(nil), validations...),
	}
	if includeEvents {
		report.Events = append([]EventReport(nil), state.Events...)
	}
	return report
}

func assetReports(state *State) []AssetReport {
	result := []AssetReport{}
	for _, asset := range state.SortedAssets() {
		result = append(result, AssetReport{
			ID:       asset.ID,
			Symbol:   asset.Symbol,
			Decimals: asset.Decimals,
			Metadata: copyStringMap(asset.Metadata),
		})
	}
	return result
}

func accountReports(state *State) []AccountReport {
	result := []AccountReport{}
	for _, account := range state.SortedAccounts() {
		report := AccountReport{
			ID:            account.ID,
			Role:          account.Role,
			Status:        account.Status,
			Free:          []AmountEntry{},
			Held:          []AmountEntry{},
			SettledIn:     []AmountEntry{},
			Withdrawn:     []AmountEntry{},
			FeesPaid:      []AmountEntry{},
			Contributed:   []AmountEntry{},
			ClaimsPaid:    []AmountEntry{},
			DefaultsTaken: []AmountEntry{},
		}
		keys := make([]string, 0, len(account.Balances))
		for key := range account.Balances {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			balance := account.Balances[key]
			report.Free = appendAmount(report.Free, key, balance.Free)
			report.Held = appendAmount(report.Held, key, balance.Held)
			report.SettledIn = appendAmount(report.SettledIn, key, balance.SettledIn)
			report.Withdrawn = appendAmount(report.Withdrawn, key, balance.Withdrawn)
			report.FeesPaid = appendAmount(report.FeesPaid, key, balance.FeesPaid)
			report.Contributed = appendAmount(report.Contributed, key, balance.Contributed)
			report.ClaimsPaid = appendAmount(report.ClaimsPaid, key, balance.ClaimsPaid)
			report.DefaultsTaken = appendAmount(report.DefaultsTaken, key, balance.DefaultsTaken)
		}
		result = append(result, report)
	}
	return result
}

func poolReports(state *State) []PoolReport {
	result := []PoolReport{}
	for _, pool := range state.SortedPools() {
		result = append(result, PoolReport{
			ID:                  pool.ID,
			Asset:               pool.Asset,
			Status:              pool.Status,
			Balance:             pool.Balance,
			Contributions:       pool.Contributions,
			FeeContributions:    pool.FeeContributions,
			ClaimsPaid:          pool.ClaimsPaid,
			Recovered:           pool.Recovered,
			ReservedForDefaults: pool.ReservedForDefaults,
			PendingDefaults:     state.PendingDefaultExposure(pool.Asset),
			PendingClaims:       state.PendingClaimCoverage(pool.Asset),
		})
	}
	return result
}

func reserveReports(state *State) []ReserveReport {
	result := []ReserveReport{}
	for _, reserve := range state.SortedReserves() {
		result = append(result, ReserveReport{
			ID:                     reserve.ID,
			Asset:                  reserve.Asset,
			Owner:                  reserve.Owner,
			Status:                 reserve.Status,
			Balance:                reserve.Balance,
			Held:                   reserve.Held,
			GrossDeposits:          reserve.GrossDeposits,
			GrossWithdrawals:       reserve.GrossWithdrawals,
			SettlementVolume:       reserve.SettlementVolume,
			AccruedFees:            reserve.AccruedFees,
			InsuranceContributions: reserve.InsuranceContributions,
		})
	}
	return result
}

func facilityReports(state *State) []FacilityReport {
	result := []FacilityReport{}
	for _, facility := range state.SortedFacilities() {
		result = append(result, FacilityReport{
			ID:          facility.ID,
			ReserveID:   facility.ReserveID,
			Asset:       facility.Asset,
			Borrower:    facility.Borrower,
			Beneficiary: facility.Beneficiary,
			Operator:    facility.Operator,
			Status:      facility.Status,
			Principal:   facility.Principal,
			Outstanding: facility.Outstanding,
			Repaid:      facility.Repaid,
			FeesPaid:    facility.FeesPaid,
			OpenedEpoch: facility.OpenedEpoch,
			DueEpoch:    facility.DueEpoch,
			DefaultID:   facility.DefaultID,
		})
	}
	return result
}

func defaultReports(state *State) []DefaultReport {
	result := []DefaultReport{}
	for _, def := range state.SortedDefaults() {
		result = append(result, DefaultReport{
			ID:               def.ID,
			FacilityID:       def.FacilityID,
			Asset:            def.Asset,
			Reporter:         def.Reporter,
			Borrower:         def.Borrower,
			Beneficiary:      def.Beneficiary,
			Status:           def.Status,
			Amount:           def.Amount,
			ExpectedRecovery: def.ExpectedRecovery,
			CoverageCeiling:  def.CoverageCeiling,
			PendingExposure:  def.PendingExposure,
			ClaimedCoverage:  def.ClaimedCoverage,
			Recovered:        def.Recovered,
			ReportedEpoch:    def.ReportedEpoch,
			AcceptedEpoch:    def.AcceptedEpoch,
			ResolvedEpoch:    def.ResolvedEpoch,
			Claims:           append([]string(nil), def.Claims...),
		})
	}
	return result
}

func claimReports(state *State) []ClaimReport {
	result := []ClaimReport{}
	for _, claim := range state.SortedClaims() {
		result = append(result, ClaimReport{
			ID:                     claim.ID,
			DefaultID:              claim.DefaultID,
			FacilityID:             claim.FacilityID,
			Asset:                  claim.Asset,
			Claimant:               claim.Claimant,
			PayoutAccount:          claim.PayoutAccount,
			Status:                 claim.Status,
			Amount:                 claim.Amount,
			Coverage:               claim.Coverage,
			CapacityAtRegistration: claim.CapacityAtRegistration,
			RegisteredEpoch:        claim.RegisteredEpoch,
			ExecutedEpoch:          claim.ExecutedEpoch,
		})
	}
	return result
}

func reconciliationRows(state *State) []ReconciliationRow {
	rows := []ReconciliationRow{}
	for _, asset := range state.SortedAssets() {
		var poolBalance amount.Amount
		if pool, ok := state.Pools[asset.ID]; ok {
			poolBalance = pool.Balance
		}
		pendingDefaults := state.PendingDefaultExposure(asset.ID)
		pendingClaims := state.PendingClaimCoverage(asset.ID)
		buffer := amount.Amount(int64(poolBalance) - int64(pendingDefaults) - int64(pendingClaims))
		rows = append(rows, ReconciliationRow{
			Asset:             asset.ID,
			ReserveBalance:    state.ReserveBalance(asset.ID),
			FacilityExposure:  state.FacilityExposure(asset.ID),
			PoolBalance:       poolBalance,
			PendingDefaults:   pendingDefaults,
			PendingClaims:     pendingClaims,
			NetCoverageBuffer: buffer,
		})
	}
	return rows
}

func appendAmount(entries []AmountEntry, id string, value amount.Amount) []AmountEntry {
	if value == 0 {
		return entries
	}
	return append(entries, AmountEntry{ID: id, Amount: value})
}

func copyMetrics(input map[string]amount.Amount) map[string]amount.Amount {
	output := map[string]amount.Amount{}
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		output[key] = input[key]
	}
	return output
}

func copyStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := map[string]string{}
	for key, value := range input {
		output[key] = value
	}
	return output
}

package risk

import (
	"fmt"

	"emberdtl/src/amount"
	"emberdtl/src/domain"
	"emberdtl/src/policy"
)

type ClaimEvaluation struct {
	Asset                  string        `json:"asset"`
	DefaultID              string        `json:"defaultId"`
	ClaimAmount            amount.Amount `json:"claimAmount"`
	Coverage               amount.Amount `json:"coverage"`
	PoolBalance            amount.Amount `json:"poolBalance"`
	PendingDefaultExposure amount.Amount `json:"pendingDefaultExposure"`
	AvailableCoverage      amount.Amount `json:"availableCoverage"`
	CeilingRemaining       amount.Amount `json:"ceilingRemaining"`
}

type FacilityEvaluation struct {
	FacilityID       string        `json:"facilityId"`
	ReserveID        string        `json:"reserveId"`
	Asset            string        `json:"asset"`
	Principal        amount.Amount `json:"principal"`
	ReserveBalance   amount.Amount `json:"reserveBalance"`
	ReserveAfterOpen amount.Amount `json:"reserveAfterOpen"`
}

func EvaluateFacility(reserve *domain.ReserveBucket, principal amount.Amount) (FacilityEvaluation, error) {
	if principal > reserve.Balance {
		return FacilityEvaluation{}, fmt.Errorf("reserve cannot fund facility principal")
	}
	return FacilityEvaluation{
		ReserveID:        reserve.ID,
		Asset:            reserve.Asset,
		Principal:        principal,
		ReserveBalance:   reserve.Balance,
		ReserveAfterOpen: reserve.Balance.MustSub(principal),
	}, nil
}

func EvaluateDefault(rules policy.Policy, facility *domain.Facility, value amount.Amount, epoch int) (amount.Amount, amount.Amount, error) {
	if !rules.CanReportDefault(epoch, facility.DueEpoch) {
		return 0, 0, fmt.Errorf("facility is still inside the default grace window")
	}
	if value == 0 {
		value = facility.Outstanding
	}
	if value > facility.Outstanding {
		return 0, 0, fmt.Errorf("default amount exceeds outstanding")
	}
	ceiling, err := rules.CoverageCeiling(value)
	if err != nil {
		return 0, 0, err
	}
	pending, err := rules.PendingDefaultExposure(value)
	if err != nil {
		return 0, 0, err
	}
	return ceiling, pending, nil
}

func EvaluateClaim(state *domain.State, rules policy.Policy, pool *domain.InsurancePool, def *domain.DefaultCase, claimAmount amount.Amount) (ClaimEvaluation, error) {
	if def.Status == domain.DefaultResolved {
		return ClaimEvaluation{}, fmt.Errorf("default is already resolved")
	}
	if err := rules.CheckClaimAmount(claimAmount, def.Amount); err != nil {
		return ClaimEvaluation{}, err
	}
	remainingCeiling := def.CoverageCeiling.SubFloor(def.ClaimedCoverage)
	coverage, err := rules.ClaimCoverage(claimAmount, remainingCeiling)
	if err != nil {
		return ClaimEvaluation{}, err
	}
	if coverage == 0 {
		return ClaimEvaluation{}, fmt.Errorf("claim has no remaining coverage")
	}
	pendingDefaults := state.PendingDefaultExposure(pool.Asset)
	available := pool.Balance.SubFloor(pendingDefaults)
	if coverage > available {
		return ClaimEvaluation{}, fmt.Errorf("claim coverage exceeds insurance capacity")
	}
	return ClaimEvaluation{
		Asset:                  pool.Asset,
		DefaultID:              def.ID,
		ClaimAmount:            claimAmount,
		Coverage:               coverage,
		PoolBalance:            pool.Balance,
		PendingDefaultExposure: pendingDefaults,
		AvailableCoverage:      available,
		CeilingRemaining:       remainingCeiling,
	}, nil
}

func ExposureRatio(pool *domain.InsurancePool, exposure amount.Amount) int64 {
	if pool.Balance <= 0 {
		return 0
	}
	return int64(exposure) * 10_000 / int64(pool.Balance)
}

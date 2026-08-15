package solvency

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"emberdtl/src/amount"
)

const basisPoints int64 = 10_000

type Band string

const (
	BandNominal  Band = "nominal"
	BandWatch    Band = "watch"
	BandGuarded  Band = "guarded"
	BandCritical Band = "critical"
)

type AssetPosition struct {
	Asset              string
	PoolBalance        amount.Amount
	ReserveBalance     amount.Amount
	FacilityExposure   amount.Amount
	PendingDefaults    amount.Amount
	PendingClaims      amount.Amount
	ExpectedRecoveries amount.Amount
	LargestFacility    amount.Amount
}

type StressPolicy struct {
	LiquidityHaircutBps   int64
	DefaultShockBps       int64
	ClaimShockBps         int64
	RecoveryHaircutBps    int64
	ConcentrationAddonBps int64
	MinimumCoverageBps    int64
}

type Assessment struct {
	Asset                string
	AvailableLiquidity   amount.Amount
	StressedDefaults     amount.Amount
	StressedClaims       amount.Amount
	RecognizedRecoveries amount.Amount
	ConcentrationAddon   amount.Amount
	RequiredLiquidity    amount.Amount
	NetLiquidityBuffer   int64
	CoverageBps          int64
	FacilityReserveBps   int64
	ConcentrationBps     int64
	Band                 Band
	Signals              []string
}

func DefaultStressPolicy() StressPolicy {
	return StressPolicy{
		LiquidityHaircutBps:   1_000,
		DefaultShockBps:       2_500,
		ClaimShockBps:         1_500,
		RecoveryHaircutBps:    6_000,
		ConcentrationAddonBps: 1_250,
		MinimumCoverageBps:    11_500,
	}
}

func (p StressPolicy) Validate() error {
	checks := []struct {
		name    string
		value   int64
		maximum int64
	}{
		{"liquidity haircut", p.LiquidityHaircutBps, basisPoints},
		{"default shock", p.DefaultShockBps, 50_000},
		{"claim shock", p.ClaimShockBps, 50_000},
		{"recovery haircut", p.RecoveryHaircutBps, basisPoints},
		{"concentration addon", p.ConcentrationAddonBps, basisPoints},
		{"minimum coverage", p.MinimumCoverageBps, 50_000},
	}
	for _, check := range checks {
		if check.value < 0 || check.value > check.maximum {
			return fmt.Errorf("%s bps out of range", check.name)
		}
	}
	return nil
}

func (p AssetPosition) Validate() error {
	if strings.TrimSpace(p.Asset) == "" {
		return fmt.Errorf("asset is required")
	}
	values := []amount.Amount{
		p.PoolBalance,
		p.ReserveBalance,
		p.FacilityExposure,
		p.PendingDefaults,
		p.PendingClaims,
		p.ExpectedRecoveries,
		p.LargestFacility,
	}
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
	}
	if p.LargestFacility > p.FacilityExposure {
		return fmt.Errorf("largest facility exceeds total facility exposure")
	}
	return nil
}

func Assess(position AssetPosition, rules StressPolicy) (Assessment, error) {
	if err := position.Validate(); err != nil {
		return Assessment{}, err
	}
	if err := rules.Validate(); err != nil {
		return Assessment{}, err
	}

	available, err := position.PoolBalance.MulBps(basisPoints - rules.LiquidityHaircutBps)
	if err != nil {
		return Assessment{}, err
	}
	stressedDefaults, err := scaleCeil(position.PendingDefaults, basisPoints+rules.DefaultShockBps)
	if err != nil {
		return Assessment{}, err
	}
	stressedClaims, err := scaleCeil(position.PendingClaims, basisPoints+rules.ClaimShockBps)
	if err != nil {
		return Assessment{}, err
	}
	recoveries, err := position.ExpectedRecoveries.MulBps(basisPoints - rules.RecoveryHaircutBps)
	if err != nil {
		return Assessment{}, err
	}
	addon, err := scaleCeil(position.LargestFacility, rules.ConcentrationAddonBps)
	if err != nil {
		return Assessment{}, err
	}
	grossRequired, err := amount.Sum(stressedDefaults, stressedClaims, addon)
	if err != nil {
		return Assessment{}, err
	}
	required := grossRequired.SubFloor(recoveries)
	coverage, err := ratioBps(available, required)
	if err != nil {
		return Assessment{}, err
	}
	reserveCoverage, err := ratioBps(position.ReserveBalance, position.FacilityExposure)
	if err != nil {
		return Assessment{}, err
	}
	concentration, err := ratioBps(position.LargestFacility, position.FacilityExposure)
	if err != nil {
		return Assessment{}, err
	}

	assessment := Assessment{
		Asset:                position.Asset,
		AvailableLiquidity:   available,
		StressedDefaults:     stressedDefaults,
		StressedClaims:       stressedClaims,
		RecognizedRecoveries: recoveries,
		ConcentrationAddon:   addon,
		RequiredLiquidity:    required,
		NetLiquidityBuffer:   int64(available) - int64(required),
		CoverageBps:          coverage,
		FacilityReserveBps:   reserveCoverage,
		ConcentrationBps:     concentration,
	}
	assessment.Band = classify(assessment, rules.MinimumCoverageBps)
	assessment.Signals = signals(assessment)
	return assessment, nil
}

func AssessPortfolio(positions []AssetPosition, rules StressPolicy) ([]Assessment, error) {
	seen := map[string]struct{}{}
	result := make([]Assessment, 0, len(positions))
	for _, position := range positions {
		asset := strings.ToLower(strings.TrimSpace(position.Asset))
		if _, exists := seen[asset]; exists {
			return nil, fmt.Errorf("duplicate asset %s", position.Asset)
		}
		seen[asset] = struct{}{}
		assessment, err := Assess(position, rules)
		if err != nil {
			return nil, fmt.Errorf("asset %s: %w", position.Asset, err)
		}
		result = append(result, assessment)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Asset < result[j].Asset })
	return result, nil
}

func scaleCeil(value amount.Amount, bps int64) (amount.Amount, error) {
	product, err := value.CheckedMul(bps)
	if err != nil {
		return 0, err
	}
	if product == 0 {
		return 0, nil
	}
	quotient := int64(product) / basisPoints
	if int64(product)%basisPoints != 0 {
		quotient++
	}
	return amount.Amount(quotient), nil
}

func ratioBps(numerator, denominator amount.Amount) (int64, error) {
	if denominator == 0 {
		if numerator == 0 {
			return basisPoints, nil
		}
		return math.MaxInt32, nil
	}
	product, err := numerator.CheckedMul(basisPoints)
	if err != nil {
		return 0, err
	}
	return int64(product) / int64(denominator), nil
}

func classify(assessment Assessment, minimum int64) Band {
	if assessment.RequiredLiquidity == 0 || assessment.CoverageBps >= minimum+2_000 {
		return BandNominal
	}
	if assessment.CoverageBps >= minimum {
		return BandWatch
	}
	if assessment.CoverageBps >= 9_000 {
		return BandGuarded
	}
	return BandCritical
}

func signals(assessment Assessment) []string {
	result := []string{}
	if assessment.NetLiquidityBuffer < 0 {
		result = append(result, "negative_stressed_buffer")
	}
	if assessment.ConcentrationBps > 5_000 {
		result = append(result, "facility_concentration")
	}
	if assessment.FacilityReserveBps < 10_000 {
		result = append(result, "reserve_below_facility_exposure")
	}
	if assessment.StressedClaims > assessment.AvailableLiquidity {
		result = append(result, "claims_above_available_liquidity")
	}
	return result
}

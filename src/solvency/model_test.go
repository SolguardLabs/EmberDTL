package solvency

import (
	"math"
	"reflect"
	"testing"

	"emberdtl/src/amount"
)

func baselinePosition() AssetPosition {
	return AssetPosition{
		Asset:              "usd",
		PoolBalance:        amount.Must(10_000),
		ReserveBalance:     amount.Must(25_000),
		FacilityExposure:   amount.Must(20_000),
		PendingDefaults:    amount.Must(2_000),
		PendingClaims:      amount.Must(1_000),
		ExpectedRecoveries: amount.Must(500),
		LargestFacility:    amount.Must(5_000),
	}
}

func TestAssessCalculatesStressedWaterfall(t *testing.T) {
	got, err := Assess(baselinePosition(), DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.AvailableLiquidity != 9_000 || got.StressedDefaults != 2_500 || got.StressedClaims != 1_150 {
		t.Fatalf("unexpected stress legs: %+v", got)
	}
	if got.RecognizedRecoveries != 200 || got.ConcentrationAddon != 625 || got.RequiredLiquidity != 4_075 {
		t.Fatalf("unexpected requirement: %+v", got)
	}
	if got.NetLiquidityBuffer != 4_925 || got.CoverageBps != 22_085 || got.Band != BandNominal {
		t.Fatalf("unexpected coverage: %+v", got)
	}
}

func TestAssessFlagsCriticalLiquidity(t *testing.T) {
	position := baselinePosition()
	position.PoolBalance = 2_000
	position.PendingClaims = 4_000
	got, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.Band != BandCritical || got.NetLiquidityBuffer >= 0 {
		t.Fatalf("expected critical negative buffer: %+v", got)
	}
	if !contains(got.Signals, "negative_stressed_buffer") || !contains(got.Signals, "claims_above_available_liquidity") {
		t.Fatalf("missing liquidity signals: %v", got.Signals)
	}
}

func TestAssessRecognizesRecoveryHaircut(t *testing.T) {
	rules := DefaultStressPolicy()
	rules.RecoveryHaircutBps = 2_500
	position := baselinePosition()
	position.ExpectedRecoveries = 2_000
	got, err := Assess(position, rules)
	if err != nil {
		t.Fatal(err)
	}
	if got.RecognizedRecoveries != 1_500 {
		t.Fatalf("got %d", got.RecognizedRecoveries)
	}
}

func TestAssessUsesConservativeCeiling(t *testing.T) {
	position := baselinePosition()
	position.PendingDefaults = 1
	position.PendingClaims = 1
	position.ExpectedRecoveries = 0
	position.LargestFacility = 1
	got, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.StressedDefaults != 2 || got.StressedClaims != 2 || got.ConcentrationAddon != 1 {
		t.Fatalf("expected conservative ceilings: %+v", got)
	}
}

func TestAssessReportsConcentration(t *testing.T) {
	position := baselinePosition()
	position.LargestFacility = 15_000
	got, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.ConcentrationBps != 7_500 || !contains(got.Signals, "facility_concentration") {
		t.Fatalf("unexpected concentration: %+v", got)
	}
}

func TestAssessZeroRequirementIsNominal(t *testing.T) {
	position := AssetPosition{Asset: "eur", PoolBalance: 100}
	got, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.Band != BandNominal || got.CoverageBps != math.MaxInt32 {
		t.Fatalf("unexpected empty book assessment: %+v", got)
	}
}

func TestPositionRejectsLargestFacilityAboveExposure(t *testing.T) {
	position := baselinePosition()
	position.LargestFacility = position.FacilityExposure + 1
	if _, err := Assess(position, DefaultStressPolicy()); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPolicyRejectsInvalidHaircut(t *testing.T) {
	rules := DefaultStressPolicy()
	rules.LiquidityHaircutBps = 10_001
	if _, err := Assess(baselinePosition(), rules); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestAssessClassifiesWatchBand(t *testing.T) {
	position := baselinePosition()
	position.PoolBalance = 5_300
	got, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.Band != BandWatch {
		t.Fatalf("expected watch, got %s at %d bps", got.Band, got.CoverageBps)
	}
}

func TestAssessClassifiesGuardedBand(t *testing.T) {
	position := baselinePosition()
	position.PoolBalance = 4_300
	got, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got.Band != BandGuarded {
		t.Fatalf("expected guarded, got %s at %d bps", got.Band, got.CoverageBps)
	}
}

func TestAssessPortfolioSortsAssets(t *testing.T) {
	eur := baselinePosition()
	eur.Asset = "eur"
	usd := baselinePosition()
	got, err := AssessPortfolio([]AssetPosition{usd, eur}, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Asset != "eur" || got[1].Asset != "usd" {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestAssessPortfolioRejectsDuplicateAsset(t *testing.T) {
	one := baselinePosition()
	two := baselinePosition()
	two.Asset = " USD "
	if _, err := AssessPortfolio([]AssetPosition{one, two}, DefaultStressPolicy()); err == nil {
		t.Fatal("expected duplicate asset error")
	}
}

func TestSignalsAreDeterministic(t *testing.T) {
	position := baselinePosition()
	position.PoolBalance = 1_000
	position.ReserveBalance = 2_000
	position.LargestFacility = 15_000
	one, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	two, err := Assess(position, DefaultStressPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(one.Signals, two.Signals) {
		t.Fatalf("signals changed: %v != %v", one.Signals, two.Signals)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

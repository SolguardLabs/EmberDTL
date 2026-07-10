package policy

import (
	"fmt"
	"strings"

	"emberdtl/src/amount"
)

type Policy struct {
	Name                       string        `json:"name"`
	MinimumAccountFunding      amount.Amount `json:"minimumAccountFunding"`
	MinimumReserveDeposit      amount.Amount `json:"minimumReserveDeposit"`
	MinimumContribution        amount.Amount `json:"minimumContribution"`
	MinimumFacilityPrincipal   amount.Amount `json:"minimumFacilityPrincipal"`
	MaximumFacilityPrincipal   amount.Amount `json:"maximumFacilityPrincipal"`
	SettlementFeeBps           int64         `json:"settlementFeeBps"`
	MinimumSettlementFee       amount.Amount `json:"minimumSettlementFee"`
	MaximumSettlementFee       amount.Amount `json:"maximumSettlementFee"`
	InsuranceFeeShareBps       int64         `json:"insuranceFeeShareBps"`
	MaxDefaultCoverageBps      int64         `json:"maxDefaultCoverageBps"`
	PendingDefaultReserveBps   int64         `json:"pendingDefaultReserveBps"`
	MaxClaimAmountBps          int64         `json:"maxClaimAmountBps"`
	MinimumClaimAmount         amount.Amount `json:"minimumClaimAmount"`
	DefaultGraceEpochs         int           `json:"defaultGraceEpochs"`
	ClaimExecutionDelayEpochs  int           `json:"claimExecutionDelayEpochs"`
	AllowEarlyDefaultReport    bool          `json:"allowEarlyDefaultReport"`
	RequireOperatorForFacility bool          `json:"requireOperatorForFacility"`
	RequireReserveSolvency     bool          `json:"requireReserveSolvency"`
	AllowManualRecovery        bool          `json:"allowManualRecovery"`
}

func Default() Policy {
	return Policy{
		Name:                       "ember-mainnet-policy",
		MinimumAccountFunding:      amount.Must(1),
		MinimumReserveDeposit:      amount.Must(100),
		MinimumContribution:        amount.Must(10),
		MinimumFacilityPrincipal:   amount.Must(100),
		MaximumFacilityPrincipal:   amount.Must(10_000_000),
		SettlementFeeBps:           120,
		MinimumSettlementFee:       amount.Must(1),
		MaximumSettlementFee:       amount.Must(50_000),
		InsuranceFeeShareBps:       5_000,
		MaxDefaultCoverageBps:      8_000,
		PendingDefaultReserveBps:   5_000,
		MaxClaimAmountBps:          10_000,
		MinimumClaimAmount:         amount.Must(1),
		DefaultGraceEpochs:         1,
		ClaimExecutionDelayEpochs:  0,
		AllowEarlyDefaultReport:    false,
		RequireOperatorForFacility: true,
		RequireReserveSolvency:     true,
		AllowManualRecovery:        true,
	}
}

func (p Policy) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("policy name is required")
	}
	checks := []struct {
		name  string
		value amount.Amount
	}{
		{"minimum account funding", p.MinimumAccountFunding},
		{"minimum reserve deposit", p.MinimumReserveDeposit},
		{"minimum contribution", p.MinimumContribution},
		{"minimum facility principal", p.MinimumFacilityPrincipal},
		{"maximum facility principal", p.MaximumFacilityPrincipal},
		{"minimum settlement fee", p.MinimumSettlementFee},
		{"maximum settlement fee", p.MaximumSettlementFee},
		{"minimum claim amount", p.MinimumClaimAmount},
	}
	for _, check := range checks {
		if err := check.value.Validate(); err != nil {
			return fmt.Errorf("%s: %w", check.name, err)
		}
	}
	if p.MinimumFacilityPrincipal > p.MaximumFacilityPrincipal {
		return fmt.Errorf("minimum facility principal exceeds maximum")
	}
	if p.MinimumSettlementFee > p.MaximumSettlementFee {
		return fmt.Errorf("minimum settlement fee exceeds maximum")
	}
	if err := checkBps("settlement fee", p.SettlementFeeBps, 0, 10_000); err != nil {
		return err
	}
	if err := checkBps("insurance fee share", p.InsuranceFeeShareBps, 0, 10_000); err != nil {
		return err
	}
	if err := checkBps("max default coverage", p.MaxDefaultCoverageBps, 0, 10_000); err != nil {
		return err
	}
	if err := checkBps("pending default reserve", p.PendingDefaultReserveBps, 0, 10_000); err != nil {
		return err
	}
	if err := checkBps("max claim amount", p.MaxClaimAmountBps, 0, 10_000); err != nil {
		return err
	}
	if p.DefaultGraceEpochs < 0 || p.ClaimExecutionDelayEpochs < 0 {
		return fmt.Errorf("epoch delays cannot be negative")
	}
	return nil
}

func checkBps(name string, value, minimum, maximum int64) error {
	if value < minimum || value > maximum {
		return fmt.Errorf("%s bps out of range", name)
	}
	return nil
}

func (p Policy) CheckFunding(value amount.Amount) error {
	if value < p.MinimumAccountFunding {
		return fmt.Errorf("funding below minimum")
	}
	return value.Validate()
}

func (p Policy) CheckReserveDeposit(value amount.Amount) error {
	if value < p.MinimumReserveDeposit {
		return fmt.Errorf("reserve deposit below minimum")
	}
	return value.Validate()
}

func (p Policy) CheckContribution(value amount.Amount) error {
	if value < p.MinimumContribution {
		return fmt.Errorf("contribution below minimum")
	}
	return value.Validate()
}

func (p Policy) CheckFacilityPrincipal(value amount.Amount) error {
	if value < p.MinimumFacilityPrincipal {
		return fmt.Errorf("facility principal below minimum")
	}
	if value > p.MaximumFacilityPrincipal {
		return fmt.Errorf("facility principal above maximum")
	}
	return value.Validate()
}

func (p Policy) CheckClaimAmount(value amount.Amount, defaultAmount amount.Amount) error {
	if value < p.MinimumClaimAmount {
		return fmt.Errorf("claim amount below minimum")
	}
	limit, err := defaultAmount.MulBps(p.MaxClaimAmountBps)
	if err != nil {
		return err
	}
	if value > limit {
		return fmt.Errorf("claim amount exceeds default policy")
	}
	return value.Validate()
}

func (p Policy) SettlementFee(value amount.Amount) (amount.Amount, error) {
	fee, err := value.MulBpsCeil(p.SettlementFeeBps)
	if err != nil {
		return 0, err
	}
	fee = amount.Clamp(fee, p.MinimumSettlementFee, p.MaximumSettlementFee)
	if value == 0 {
		return 0, nil
	}
	return fee, nil
}

func (p Policy) InsuranceFeeShare(fee amount.Amount) (amount.Amount, error) {
	return fee.MulBps(p.InsuranceFeeShareBps)
}

func (p Policy) OperatorFeeShare(fee amount.Amount) (amount.Amount, error) {
	insurance, err := p.InsuranceFeeShare(fee)
	if err != nil {
		return 0, err
	}
	return fee.CheckedSub(insurance)
}

func (p Policy) CoverageCeiling(defaultAmount amount.Amount) (amount.Amount, error) {
	return defaultAmount.MulBps(p.MaxDefaultCoverageBps)
}

func (p Policy) PendingDefaultExposure(defaultAmount amount.Amount) (amount.Amount, error) {
	return defaultAmount.MulBps(p.PendingDefaultReserveBps)
}

func (p Policy) ClaimCoverage(claimAmount amount.Amount, ceiling amount.Amount) (amount.Amount, error) {
	coverage, err := claimAmount.MulBps(p.MaxDefaultCoverageBps)
	if err != nil {
		return 0, err
	}
	if coverage > ceiling {
		return ceiling, nil
	}
	return coverage, nil
}

func (p Policy) CanReportDefault(epoch, dueEpoch int) bool {
	if p.AllowEarlyDefaultReport {
		return true
	}
	return epoch >= dueEpoch+p.DefaultGraceEpochs
}

func (p Policy) CanExecuteClaim(currentEpoch, registeredEpoch int) bool {
	return currentEpoch >= registeredEpoch+p.ClaimExecutionDelayEpochs
}

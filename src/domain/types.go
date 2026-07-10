package domain

import (
	"fmt"
	"strings"

	"emberdtl/src/amount"
)

type Asset struct {
	ID       string            `json:"id"`
	Symbol   string            `json:"symbol"`
	Decimals int               `json:"decimals"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func NewAsset(id, symbol string, decimals int) *Asset {
	if strings.TrimSpace(symbol) == "" {
		symbol = strings.ToUpper(id)
	}
	return &Asset{
		ID:       id,
		Symbol:   symbol,
		Decimals: decimals,
		Metadata: map[string]string{},
	}
}

type Balance struct {
	Asset         string        `json:"asset"`
	Free          amount.Amount `json:"free"`
	Held          amount.Amount `json:"held"`
	SettledIn     amount.Amount `json:"settledIn"`
	Withdrawn     amount.Amount `json:"withdrawn"`
	FeesPaid      amount.Amount `json:"feesPaid"`
	Contributed   amount.Amount `json:"contributed"`
	ClaimsPaid    amount.Amount `json:"claimsPaid"`
	DefaultsTaken amount.Amount `json:"defaultsTaken"`
}

func NewBalance(asset string) *Balance {
	return &Balance{Asset: asset}
}

func (b *Balance) Available() amount.Amount {
	return b.Free
}

func (b *Balance) Credit(value amount.Amount) error {
	next, err := b.Free.CheckedAdd(value)
	if err != nil {
		return err
	}
	b.Free = next
	return nil
}

func (b *Balance) Debit(value amount.Amount) error {
	next, err := b.Free.CheckedSub(value)
	if err != nil {
		return err
	}
	b.Free = next
	return nil
}

func (b *Balance) Hold(value amount.Amount) error {
	nextFree, err := b.Free.CheckedSub(value)
	if err != nil {
		return err
	}
	nextHeld, err := b.Held.CheckedAdd(value)
	if err != nil {
		return err
	}
	b.Free = nextFree
	b.Held = nextHeld
	return nil
}

func (b *Balance) Release(value amount.Amount) error {
	nextHeld, err := b.Held.CheckedSub(value)
	if err != nil {
		return err
	}
	nextFree, err := b.Free.CheckedAdd(value)
	if err != nil {
		return err
	}
	b.Held = nextHeld
	b.Free = nextFree
	return nil
}

func (b *Balance) BurnHeld(value amount.Amount) error {
	nextHeld, err := b.Held.CheckedSub(value)
	if err != nil {
		return err
	}
	b.Held = nextHeld
	return nil
}

type Account struct {
	ID          string              `json:"id"`
	Role        AccountRole         `json:"role"`
	Status      AccountStatus       `json:"status"`
	DisplayName string              `json:"displayName,omitempty"`
	Balances    map[string]*Balance `json:"balances"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
}

func NewAccount(id string, role AccountRole) *Account {
	if role == "" {
		role = RoleUser
	}
	return &Account{
		ID:       id,
		Role:     role,
		Status:   AccountActive,
		Balances: map[string]*Balance{},
		Metadata: map[string]string{},
	}
}

func (a *Account) Balance(asset string) *Balance {
	balance, ok := a.Balances[asset]
	if !ok {
		balance = NewBalance(asset)
		a.Balances[asset] = balance
	}
	return balance
}

func (a *Account) Available(asset string) amount.Amount {
	return a.Balance(asset).Available()
}

func (a *Account) Credit(asset string, value amount.Amount) error {
	return a.Balance(asset).Credit(value)
}

func (a *Account) Debit(asset string, value amount.Amount) error {
	return a.Balance(asset).Debit(value)
}

func (a *Account) Hold(asset string, value amount.Amount) error {
	return a.Balance(asset).Hold(value)
}

func (a *Account) Release(asset string, value amount.Amount) error {
	return a.Balance(asset).Release(value)
}

func (a *Account) BurnHeld(asset string, value amount.Amount) error {
	return a.Balance(asset).BurnHeld(value)
}

type InsurancePool struct {
	ID                  string            `json:"id"`
	Asset               string            `json:"asset"`
	Status              PoolStatus        `json:"status"`
	Balance             amount.Amount     `json:"balance"`
	Contributions       amount.Amount     `json:"contributions"`
	FeeContributions    amount.Amount     `json:"feeContributions"`
	ClaimsPaid          amount.Amount     `json:"claimsPaid"`
	Recovered           amount.Amount     `json:"recovered"`
	ReservedForDefaults amount.Amount     `json:"reservedForDefaults"`
	Metadata            map[string]string `json:"metadata,omitempty"`
}

func NewInsurancePool(asset string) *InsurancePool {
	return &InsurancePool{
		ID:       "pool-" + asset,
		Asset:    asset,
		Status:   PoolActive,
		Metadata: map[string]string{},
	}
}

func (p *InsurancePool) CreditContribution(value amount.Amount) error {
	nextBalance, err := p.Balance.CheckedAdd(value)
	if err != nil {
		return err
	}
	nextContribution, err := p.Contributions.CheckedAdd(value)
	if err != nil {
		return err
	}
	p.Balance = nextBalance
	p.Contributions = nextContribution
	return nil
}

func (p *InsurancePool) CreditFee(value amount.Amount) error {
	nextBalance, err := p.Balance.CheckedAdd(value)
	if err != nil {
		return err
	}
	nextFee, err := p.FeeContributions.CheckedAdd(value)
	if err != nil {
		return err
	}
	p.Balance = nextBalance
	p.FeeContributions = nextFee
	return nil
}

func (p *InsurancePool) DebitClaim(value amount.Amount) error {
	p.Balance = amount.Amount(int64(p.Balance) - int64(value))
	nextPaid, err := p.ClaimsPaid.CheckedAdd(value)
	if err != nil {
		return err
	}
	p.ClaimsPaid = nextPaid
	return nil
}

func (p *InsurancePool) CreditRecovery(value amount.Amount) error {
	nextBalance, err := p.Balance.CheckedAdd(value)
	if err != nil {
		return err
	}
	nextRecovered, err := p.Recovered.CheckedAdd(value)
	if err != nil {
		return err
	}
	p.Balance = nextBalance
	p.Recovered = nextRecovered
	return nil
}

type ReserveBucket struct {
	ID                     string            `json:"id"`
	Asset                  string            `json:"asset"`
	Owner                  string            `json:"owner"`
	Status                 ReserveStatus     `json:"status"`
	Balance                amount.Amount     `json:"balance"`
	Held                   amount.Amount     `json:"held"`
	GrossDeposits          amount.Amount     `json:"grossDeposits"`
	GrossWithdrawals       amount.Amount     `json:"grossWithdrawals"`
	SettlementVolume       amount.Amount     `json:"settlementVolume"`
	AccruedFees            amount.Amount     `json:"accruedFees"`
	InsuranceContributions amount.Amount     `json:"insuranceContributions"`
	Metadata               map[string]string `json:"metadata,omitempty"`
}

func NewReserveBucket(id, owner, asset string) *ReserveBucket {
	return &ReserveBucket{
		ID:       id,
		Asset:    asset,
		Owner:    owner,
		Status:   ReserveOpen,
		Metadata: map[string]string{},
	}
}

func (r *ReserveBucket) Deposit(value amount.Amount) error {
	nextBalance, err := r.Balance.CheckedAdd(value)
	if err != nil {
		return err
	}
	nextGross, err := r.GrossDeposits.CheckedAdd(value)
	if err != nil {
		return err
	}
	r.Balance = nextBalance
	r.GrossDeposits = nextGross
	return nil
}

func (r *ReserveBucket) Withdraw(value amount.Amount) error {
	nextBalance, err := r.Balance.CheckedSub(value)
	if err != nil {
		return err
	}
	nextGross, err := r.GrossWithdrawals.CheckedAdd(value)
	if err != nil {
		return err
	}
	r.Balance = nextBalance
	r.GrossWithdrawals = nextGross
	return nil
}

func (r *ReserveBucket) Disburse(value amount.Amount) error {
	nextBalance, err := r.Balance.CheckedSub(value)
	if err != nil {
		return err
	}
	nextHeld, err := r.Held.CheckedAdd(value)
	if err != nil {
		return err
	}
	r.Balance = nextBalance
	r.Held = nextHeld
	return nil
}

func (r *ReserveBucket) ReleaseHeld(value amount.Amount) error {
	nextHeld, err := r.Held.CheckedSub(value)
	if err != nil {
		return err
	}
	nextBalance, err := r.Balance.CheckedAdd(value)
	if err != nil {
		return err
	}
	r.Held = nextHeld
	r.Balance = nextBalance
	return nil
}

func (r *ReserveBucket) BurnHeld(value amount.Amount) error {
	nextHeld, err := r.Held.CheckedSub(value)
	if err != nil {
		return err
	}
	r.Held = nextHeld
	return nil
}

type Facility struct {
	ID          string            `json:"id"`
	ReserveID   string            `json:"reserveId"`
	Asset       string            `json:"asset"`
	Borrower    string            `json:"borrower"`
	Beneficiary string            `json:"beneficiary"`
	Operator    string            `json:"operator"`
	Status      FacilityStatus    `json:"status"`
	Principal   amount.Amount     `json:"principal"`
	Outstanding amount.Amount     `json:"outstanding"`
	Repaid      amount.Amount     `json:"repaid"`
	FeesPaid    amount.Amount     `json:"feesPaid"`
	OpenedEpoch int               `json:"openedEpoch"`
	DueEpoch    int               `json:"dueEpoch"`
	ClosedEpoch int               `json:"closedEpoch,omitempty"`
	DefaultID   string            `json:"defaultId,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func NewFacility(id, reserveID, borrower, beneficiary, operator, asset string, principal amount.Amount, epoch, dueEpoch int) *Facility {
	if dueEpoch <= epoch {
		dueEpoch = epoch + 1
	}
	return &Facility{
		ID:          id,
		ReserveID:   reserveID,
		Asset:       asset,
		Borrower:    borrower,
		Beneficiary: beneficiary,
		Operator:    operator,
		Status:      FacilityPerforming,
		Principal:   principal,
		Outstanding: principal,
		OpenedEpoch: epoch,
		DueEpoch:    dueEpoch,
		Metadata:    map[string]string{},
	}
}

func (f *Facility) ApplyRepayment(value, fee amount.Amount) error {
	nextOutstanding, err := f.Outstanding.CheckedSub(value)
	if err != nil {
		return err
	}
	nextRepaid, err := f.Repaid.CheckedAdd(value)
	if err != nil {
		return err
	}
	nextFees, err := f.FeesPaid.CheckedAdd(fee)
	if err != nil {
		return err
	}
	f.Outstanding = nextOutstanding
	f.Repaid = nextRepaid
	f.FeesPaid = nextFees
	if f.Outstanding == 0 {
		f.Status = FacilityClosed
	}
	return nil
}

type DefaultCase struct {
	ID               string            `json:"id"`
	FacilityID       string            `json:"facilityId"`
	Asset            string            `json:"asset"`
	Reporter         string            `json:"reporter"`
	Borrower         string            `json:"borrower"`
	Beneficiary      string            `json:"beneficiary"`
	Status           DefaultStatus     `json:"status"`
	Amount           amount.Amount     `json:"amount"`
	ExpectedRecovery amount.Amount     `json:"expectedRecovery"`
	CoverageCeiling  amount.Amount     `json:"coverageCeiling"`
	PendingExposure  amount.Amount     `json:"pendingExposure"`
	ClaimedCoverage  amount.Amount     `json:"claimedCoverage"`
	Recovered        amount.Amount     `json:"recovered"`
	ReportedEpoch    int               `json:"reportedEpoch"`
	AcceptedEpoch    int               `json:"acceptedEpoch,omitempty"`
	ResolvedEpoch    int               `json:"resolvedEpoch,omitempty"`
	Claims           []string          `json:"claims"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

func NewDefaultCase(id string, facility *Facility, reporter string, value, ceiling, pending amount.Amount, epoch int) *DefaultCase {
	return &DefaultCase{
		ID:              id,
		FacilityID:      facility.ID,
		Asset:           facility.Asset,
		Reporter:        reporter,
		Borrower:        facility.Borrower,
		Beneficiary:     facility.Beneficiary,
		Status:          DefaultReported,
		Amount:          value,
		CoverageCeiling: ceiling,
		PendingExposure: pending,
		ReportedEpoch:   epoch,
		Claims:          []string{},
		Metadata:        map[string]string{},
	}
}

func (d *DefaultCase) AddClaim(claimID string, coverage amount.Amount) error {
	d.Claims = append(d.Claims, claimID)
	next, err := d.ClaimedCoverage.CheckedAdd(coverage)
	if err != nil {
		return err
	}
	d.ClaimedCoverage = next
	if d.Status == DefaultReported {
		d.Status = DefaultAccepted
		d.AcceptedEpoch = d.ReportedEpoch
	}
	return nil
}

func (d *DefaultCase) ApplyCoverage(value amount.Amount) error {
	if value > d.PendingExposure {
		d.PendingExposure = 0
	} else {
		d.PendingExposure = d.PendingExposure.MustSub(value)
	}
	return nil
}

func (d *DefaultCase) Resolve(epoch int, recovery amount.Amount) error {
	d.Status = DefaultResolved
	d.ResolvedEpoch = epoch
	d.Recovered = recovery
	d.PendingExposure = 0
	return nil
}

type Claim struct {
	ID                     string            `json:"id"`
	DefaultID              string            `json:"defaultId"`
	FacilityID             string            `json:"facilityId"`
	Asset                  string            `json:"asset"`
	Claimant               string            `json:"claimant"`
	PayoutAccount          string            `json:"payoutAccount"`
	Status                 ClaimStatus       `json:"status"`
	Amount                 amount.Amount     `json:"amount"`
	Coverage               amount.Amount     `json:"coverage"`
	CapacityAtRegistration amount.Amount     `json:"capacityAtRegistration"`
	RegisteredEpoch        int               `json:"registeredEpoch"`
	ExecutedEpoch          int               `json:"executedEpoch,omitempty"`
	Metadata               map[string]string `json:"metadata,omitempty"`
}

func NewClaim(id string, def *DefaultCase, claimant, payout string, value, coverage, capacity amount.Amount, epoch int) *Claim {
	return &Claim{
		ID:                     id,
		DefaultID:              def.ID,
		FacilityID:             def.FacilityID,
		Asset:                  def.Asset,
		Claimant:               claimant,
		PayoutAccount:          payout,
		Status:                 ClaimApproved,
		Amount:                 value,
		Coverage:               coverage,
		CapacityAtRegistration: capacity,
		RegisteredEpoch:        epoch,
		Metadata:               map[string]string{},
	}
}

func (c *Claim) Execute(epoch int) error {
	if c.Status != ClaimRegistered && c.Status != ClaimApproved {
		return fmt.Errorf("claim %s cannot execute from %s", c.ID, c.Status)
	}
	c.Status = ClaimExecuted
	c.ExecutedEpoch = epoch
	return nil
}

type Event struct {
	ID         string            `json:"id"`
	Epoch      int               `json:"epoch"`
	Type       EventType         `json:"type"`
	AccountID  string            `json:"accountId,omitempty"`
	ReserveID  string            `json:"reserveId,omitempty"`
	FacilityID string            `json:"facilityId,omitempty"`
	DefaultID  string            `json:"defaultId,omitempty"`
	ClaimID    string            `json:"claimId,omitempty"`
	Asset      string            `json:"asset,omitempty"`
	Amount     amount.Amount     `json:"amount,omitempty"`
	Message    string            `json:"message"`
	Fields     map[string]string `json:"fields,omitempty"`
}

func NewEvent(id string, epoch int, kind EventType) Event {
	return Event{
		ID:     id,
		Epoch:  epoch,
		Type:   kind,
		Fields: map[string]string{},
	}
}

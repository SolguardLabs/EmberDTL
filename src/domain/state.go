package domain

import (
	"fmt"
	"sort"

	"emberdtl/src/amount"
)

type State struct {
	Name       string
	Epoch      int
	Sequence   int
	Assets     map[string]*Asset
	Accounts   map[string]*Account
	Pools      map[string]*InsurancePool
	Reserves   map[string]*ReserveBucket
	Facilities map[string]*Facility
	Defaults   map[string]*DefaultCase
	Claims     map[string]*Claim
	Events     []Event
	Metrics    map[string]amount.Amount
}

func NewState(name string) *State {
	if name == "" {
		name = "EmberDTL"
	}
	return &State{
		Name:       name,
		Epoch:      1,
		Assets:     map[string]*Asset{},
		Accounts:   map[string]*Account{},
		Pools:      map[string]*InsurancePool{},
		Reserves:   map[string]*ReserveBucket{},
		Facilities: map[string]*Facility{},
		Defaults:   map[string]*DefaultCase{},
		Claims:     map[string]*Claim{},
		Events:     []Event{},
		Metrics:    map[string]amount.Amount{},
	}
}

func (s *State) AddAsset(asset *Asset) error {
	if _, exists := s.Assets[asset.ID]; exists {
		return fmt.Errorf("asset %s already exists", asset.ID)
	}
	s.Assets[asset.ID] = asset
	if _, exists := s.Pools[asset.ID]; !exists {
		s.Pools[asset.ID] = NewInsurancePool(asset.ID)
	}
	return nil
}

func (s *State) RequireAsset(id string) (*Asset, error) {
	asset, ok := s.Assets[id]
	if !ok {
		return nil, fmt.Errorf("asset %s not found", id)
	}
	return asset, nil
}

func (s *State) AddAccount(account *Account) error {
	if _, exists := s.Accounts[account.ID]; exists {
		return fmt.Errorf("account %s already exists", account.ID)
	}
	s.Accounts[account.ID] = account
	return nil
}

func (s *State) RequireAccount(id string) (*Account, error) {
	account, ok := s.Accounts[id]
	if !ok {
		return nil, fmt.Errorf("account %s not found", id)
	}
	return account, nil
}

func (s *State) RequirePool(asset string) (*InsurancePool, error) {
	pool, ok := s.Pools[asset]
	if !ok {
		return nil, fmt.Errorf("pool for asset %s not found", asset)
	}
	return pool, nil
}

func (s *State) AddReserve(reserve *ReserveBucket) error {
	if _, exists := s.Reserves[reserve.ID]; exists {
		return fmt.Errorf("reserve %s already exists", reserve.ID)
	}
	if _, err := s.RequireAsset(reserve.Asset); err != nil {
		return err
	}
	if _, err := s.RequireAccount(reserve.Owner); err != nil {
		return err
	}
	s.Reserves[reserve.ID] = reserve
	return nil
}

func (s *State) RequireReserve(id string) (*ReserveBucket, error) {
	reserve, ok := s.Reserves[id]
	if !ok {
		return nil, fmt.Errorf("reserve %s not found", id)
	}
	return reserve, nil
}

func (s *State) AddFacility(facility *Facility) error {
	if _, exists := s.Facilities[facility.ID]; exists {
		return fmt.Errorf("facility %s already exists", facility.ID)
	}
	if _, err := s.RequireReserve(facility.ReserveID); err != nil {
		return err
	}
	if _, err := s.RequireAccount(facility.Borrower); err != nil {
		return err
	}
	if _, err := s.RequireAccount(facility.Beneficiary); err != nil {
		return err
	}
	if facility.Operator != "" {
		if _, err := s.RequireAccount(facility.Operator); err != nil {
			return err
		}
	}
	s.Facilities[facility.ID] = facility
	return nil
}

func (s *State) RequireFacility(id string) (*Facility, error) {
	facility, ok := s.Facilities[id]
	if !ok {
		return nil, fmt.Errorf("facility %s not found", id)
	}
	return facility, nil
}

func (s *State) AddDefault(def *DefaultCase) error {
	if _, exists := s.Defaults[def.ID]; exists {
		return fmt.Errorf("default %s already exists", def.ID)
	}
	if _, err := s.RequireFacility(def.FacilityID); err != nil {
		return err
	}
	s.Defaults[def.ID] = def
	return nil
}

func (s *State) RequireDefault(id string) (*DefaultCase, error) {
	def, ok := s.Defaults[id]
	if !ok {
		return nil, fmt.Errorf("default %s not found", id)
	}
	return def, nil
}

func (s *State) AddClaim(claim *Claim) error {
	if _, exists := s.Claims[claim.ID]; exists {
		return fmt.Errorf("claim %s already exists", claim.ID)
	}
	if _, err := s.RequireDefault(claim.DefaultID); err != nil {
		return err
	}
	if _, err := s.RequireAccount(claim.Claimant); err != nil {
		return err
	}
	if _, err := s.RequireAccount(claim.PayoutAccount); err != nil {
		return err
	}
	s.Claims[claim.ID] = claim
	return nil
}

func (s *State) RequireClaim(id string) (*Claim, error) {
	claim, ok := s.Claims[id]
	if !ok {
		return nil, fmt.Errorf("claim %s not found", id)
	}
	return claim, nil
}

func (s *State) NextEvent(kind EventType) Event {
	s.Sequence++
	return NewEvent(fmt.Sprintf("evt-%06d", s.Sequence), s.Epoch, kind)
}

func (s *State) AppendEvent(event Event) {
	s.Events = append(s.Events, event)
}

func (s *State) AddMetric(name string, value amount.Amount) {
	current := s.Metrics[name]
	s.Metrics[name] = current.MustAdd(value)
}

func (s *State) SetMetric(name string, value amount.Amount) {
	s.Metrics[name] = value
}

func (s *State) SortedAssets() []*Asset {
	keys := make([]string, 0, len(s.Assets))
	for key := range s.Assets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]*Asset, 0, len(keys))
	for _, key := range keys {
		result = append(result, s.Assets[key])
	}
	return result
}

func (s *State) SortedAccounts() []*Account {
	keys := make([]string, 0, len(s.Accounts))
	for key := range s.Accounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]*Account, 0, len(keys))
	for _, key := range keys {
		result = append(result, s.Accounts[key])
	}
	return result
}

func (s *State) SortedPools() []*InsurancePool {
	keys := make([]string, 0, len(s.Pools))
	for key := range s.Pools {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]*InsurancePool, 0, len(keys))
	for _, key := range keys {
		result = append(result, s.Pools[key])
	}
	return result
}

func (s *State) SortedReserves() []*ReserveBucket {
	keys := make([]string, 0, len(s.Reserves))
	for key := range s.Reserves {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]*ReserveBucket, 0, len(keys))
	for _, key := range keys {
		result = append(result, s.Reserves[key])
	}
	return result
}

func (s *State) SortedFacilities() []*Facility {
	keys := make([]string, 0, len(s.Facilities))
	for key := range s.Facilities {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]*Facility, 0, len(keys))
	for _, key := range keys {
		result = append(result, s.Facilities[key])
	}
	return result
}

func (s *State) SortedDefaults() []*DefaultCase {
	keys := make([]string, 0, len(s.Defaults))
	for key := range s.Defaults {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]*DefaultCase, 0, len(keys))
	for _, key := range keys {
		result = append(result, s.Defaults[key])
	}
	return result
}

func (s *State) SortedClaims() []*Claim {
	keys := make([]string, 0, len(s.Claims))
	for key := range s.Claims {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]*Claim, 0, len(keys))
	for _, key := range keys {
		result = append(result, s.Claims[key])
	}
	return result
}

func (s *State) PendingDefaultExposure(asset string) amount.Amount {
	var total amount.Amount
	for _, def := range s.Defaults {
		if def.Asset != asset || def.Status == DefaultResolved {
			continue
		}
		total = total.MustAdd(def.PendingExposure)
	}
	return total
}

func (s *State) PendingClaimCoverage(asset string) amount.Amount {
	var total amount.Amount
	for _, claim := range s.Claims {
		if claim.Asset != asset || claim.Status == ClaimExecuted || claim.Status == ClaimRejected {
			continue
		}
		total = total.MustAdd(claim.Coverage)
	}
	return total
}

func (s *State) FacilityExposure(asset string) amount.Amount {
	var total amount.Amount
	for _, facility := range s.Facilities {
		if facility.Asset != asset || facility.Status == FacilityClosed {
			continue
		}
		total = total.MustAdd(facility.Outstanding)
	}
	return total
}

func (s *State) ReserveBalance(asset string) amount.Amount {
	var total amount.Amount
	for _, reserve := range s.Reserves {
		if reserve.Asset != asset {
			continue
		}
		total = total.MustAdd(reserve.Balance)
	}
	return total
}

func (s *State) Validate() []string {
	messages := []string{}
	for _, asset := range s.SortedAssets() {
		pool, ok := s.Pools[asset.ID]
		if !ok {
			messages = append(messages, fmt.Sprintf("missing pool for %s", asset.ID))
			continue
		}
		if pool.Balance < 0 {
			messages = append(messages, fmt.Sprintf("pool %s balance below zero", pool.ID))
		}
	}
	for _, reserve := range s.SortedReserves() {
		if reserve.Balance < 0 {
			messages = append(messages, fmt.Sprintf("reserve %s balance below zero", reserve.ID))
		}
	}
	for _, facility := range s.SortedFacilities() {
		if facility.Outstanding > facility.Principal {
			messages = append(messages, fmt.Sprintf("facility %s outstanding exceeds principal", facility.ID))
		}
	}
	if len(messages) == 0 {
		messages = append(messages, "ok")
	}
	return messages
}

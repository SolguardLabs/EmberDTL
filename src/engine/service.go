package engine

import (
	"fmt"
	"strings"

	"emberdtl/src/domain"
	"emberdtl/src/ids"
	"emberdtl/src/insurance"
	"emberdtl/src/ledger"
	"emberdtl/src/policy"
	"emberdtl/src/risk"
)

type Service struct {
	state       *domain.State
	policy      policy.Policy
	journal     *ledger.Journal
	book        *insurance.Book
	validations []string
}

func New(name string, rules policy.Policy) (*Service, error) {
	if strings.TrimSpace(name) == "" {
		name = "EmberDTL"
	}
	if err := rules.Validate(); err != nil {
		return nil, err
	}
	state := domain.NewState(name)
	journal := ledger.NewJournal()
	return &Service{
		state:       state,
		policy:      rules,
		journal:     journal,
		book:        insurance.NewBook(state, journal),
		validations: []string{},
	}, nil
}

func MustNew(name string, rules policy.Policy) *Service {
	service, err := New(name, rules)
	if err != nil {
		panic(err)
	}
	return service
}

func (s *Service) State() *domain.State {
	return s.state
}

func (s *Service) Journal() *ledger.Journal {
	return s.journal
}

func (s *Service) Policy() policy.Policy {
	return s.policy
}

func (s *Service) RegisterAsset(req AssetRequest) error {
	id, err := ids.NewAssetID(req.ID)
	if err != nil {
		return invalid(err.Error())
	}
	symbol := strings.TrimSpace(req.Symbol)
	asset := domain.NewAsset(id.String(), symbol, req.Decimals)
	if req.Decimals < 0 || req.Decimals > 18 {
		return invalid("asset decimals out of range")
	}
	for key, value := range req.Metadata {
		asset.Metadata[key] = value
	}
	if err := s.state.AddAsset(asset); err != nil {
		return stateError(err.Error())
	}
	event := s.state.NextEvent(domain.EventAssetRegistered)
	event.Asset = asset.ID
	event.Message = "asset registered with insurance pool"
	s.state.AppendEvent(event)
	return nil
}

func roleFromString(raw string) domain.AccountRole {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "treasury":
		return domain.RoleTreasury
	case "borrower":
		return domain.RoleBorrower
	case "merchant":
		return domain.RoleMerchant
	case "operator":
		return domain.RoleOperator
	case "underwriter":
		return domain.RoleUnderwriter
	default:
		return domain.RoleUser
	}
}

func (s *Service) RegisterAccount(req AccountRequest) error {
	id, err := ids.NewAccountID(req.ID)
	if err != nil {
		return invalid(err.Error())
	}
	account := domain.NewAccount(id.String(), roleFromString(req.Role))
	account.DisplayName = strings.TrimSpace(req.DisplayName)
	for key, value := range req.Metadata {
		account.Metadata[key] = value
	}
	if err := s.state.AddAccount(account); err != nil {
		return stateError(err.Error())
	}
	event := s.state.NextEvent(domain.EventAccountRegistered)
	event.AccountID = account.ID
	event.Message = "account registered"
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) FundAccount(req FundAccountRequest) error {
	accountID, err := ids.NewAccountID(req.AccountID)
	if err != nil {
		return invalid(err.Error())
	}
	assetID, err := ids.NewAssetID(req.Asset)
	if err != nil {
		return invalid(err.Error())
	}
	if err := s.policy.CheckFunding(req.Amount); err != nil {
		return policyError(err.Error())
	}
	if _, err := s.state.RequireAsset(assetID.String()); err != nil {
		return notFound(err.Error())
	}
	account, err := s.state.RequireAccount(accountID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if err := account.Credit(assetID.String(), req.Amount); err != nil {
		return stateError(err.Error())
	}
	s.journal.Append(ledger.Credit(s.state.Epoch, account.ID, assetID.String(), req.Amount, "external funding"))
	s.state.AddMetric("funding", req.Amount)
	event := s.state.NextEvent(domain.EventFunding)
	event.AccountID = account.ID
	event.Asset = assetID.String()
	event.Amount = req.Amount
	event.Message = "account funding posted"
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) OpenReserve(req ReserveRequest) error {
	reserveID, err := ids.NewReserveID(req.ID)
	if err != nil {
		return invalid(err.Error())
	}
	ownerID, err := ids.NewAccountID(req.Owner)
	if err != nil {
		return invalid(err.Error())
	}
	assetID, err := ids.NewAssetID(req.Asset)
	if err != nil {
		return invalid(err.Error())
	}
	reserve := domain.NewReserveBucket(reserveID.String(), ownerID.String(), assetID.String())
	for key, value := range req.Metadata {
		reserve.Metadata[key] = value
	}
	if err := s.state.AddReserve(reserve); err != nil {
		return stateError(err.Error())
	}
	event := s.state.NextEvent(domain.EventReserveOpened)
	event.ReserveID = reserve.ID
	event.AccountID = reserve.Owner
	event.Asset = reserve.Asset
	event.Message = "reserve bucket opened"
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) DepositReserve(req ReserveDepositRequest) error {
	reserveID, err := ids.NewReserveID(req.ReserveID)
	if err != nil {
		return invalid(err.Error())
	}
	accountID, err := ids.NewAccountID(req.AccountID)
	if err != nil {
		return invalid(err.Error())
	}
	assetID, err := ids.NewAssetID(req.Asset)
	if err != nil {
		return invalid(err.Error())
	}
	if err := s.policy.CheckReserveDeposit(req.Amount); err != nil {
		return policyError(err.Error())
	}
	reserve, err := s.state.RequireReserve(reserveID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if reserve.Asset != assetID.String() {
		return invalid("reserve asset mismatch")
	}
	account, err := s.state.RequireAccount(accountID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if err := account.Debit(assetID.String(), req.Amount); err != nil {
		return solvencyError("account cannot fund reserve deposit")
	}
	if err := reserve.Deposit(req.Amount); err != nil {
		return stateError(err.Error())
	}
	s.journal.Append(ledger.Reserve(s.state.Epoch, reserve.ID, account.ID, reserve.Asset, req.Amount, "reserve deposit"))
	s.state.AddMetric("reserveDeposits", req.Amount)
	event := s.state.NextEvent(domain.EventReserveDeposit)
	event.ReserveID = reserve.ID
	event.AccountID = account.ID
	event.Asset = reserve.Asset
	event.Amount = req.Amount
	event.Message = "reserve deposit accepted"
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) Contribute(req ContributionRequest) error {
	accountID, err := ids.NewAccountID(req.AccountID)
	if err != nil {
		return invalid(err.Error())
	}
	assetID, err := ids.NewAssetID(req.Asset)
	if err != nil {
		return invalid(err.Error())
	}
	if err := s.policy.CheckContribution(req.Amount); err != nil {
		return policyError(err.Error())
	}
	account, err := s.state.RequireAccount(accountID.String())
	if err != nil {
		return notFound(err.Error())
	}
	pool, err := s.state.RequirePool(assetID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if err := s.book.Contribute(s.state.Epoch, account, pool, req.Amount); err != nil {
		return solvencyError(err.Error())
	}
	s.state.AddMetric("insuranceContributions", req.Amount)
	event := s.state.NextEvent(domain.EventContribution)
	event.AccountID = account.ID
	event.Asset = pool.Asset
	event.Amount = req.Amount
	event.Message = "insurance contribution accepted"
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) OpenFacility(req FacilityRequest) error {
	facilityID, err := ids.NewFacilityID(req.ID)
	if err != nil {
		return invalid(err.Error())
	}
	reserveID, err := ids.NewReserveID(req.ReserveID)
	if err != nil {
		return invalid(err.Error())
	}
	borrowerID, err := ids.NewAccountID(req.Borrower)
	if err != nil {
		return invalid(err.Error())
	}
	beneficiaryID, err := ids.NewAccountID(req.Beneficiary)
	if err != nil {
		return invalid(err.Error())
	}
	operator := strings.TrimSpace(req.Operator)
	if s.policy.RequireOperatorForFacility && operator == "" {
		return policyError("facility operator required")
	}
	if operator != "" {
		operatorID, err := ids.NewAccountID(operator)
		if err != nil {
			return invalid(err.Error())
		}
		operator = operatorID.String()
	}
	assetID, err := ids.NewAssetID(req.Asset)
	if err != nil {
		return invalid(err.Error())
	}
	if err := s.policy.CheckFacilityPrincipal(req.Principal); err != nil {
		return policyError(err.Error())
	}
	reserve, err := s.state.RequireReserve(reserveID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if reserve.Asset != assetID.String() {
		return invalid("facility asset does not match reserve")
	}
	evaluation, err := risk.EvaluateFacility(reserve, req.Principal)
	if err != nil {
		return solvencyError(err.Error())
	}
	evaluation.FacilityID = facilityID.String()
	beneficiary, err := s.state.RequireAccount(beneficiaryID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if _, err := s.state.RequireAccount(borrowerID.String()); err != nil {
		return notFound(err.Error())
	}
	if operator != "" {
		if _, err := s.state.RequireAccount(operator); err != nil {
			return notFound(err.Error())
		}
	}
	if err := reserve.Disburse(req.Principal); err != nil {
		return solvencyError(err.Error())
	}
	if err := beneficiary.Credit(assetID.String(), req.Principal); err != nil {
		return stateError(err.Error())
	}
	beneficiary.Balance(assetID.String()).SettledIn = beneficiary.Balance(assetID.String()).SettledIn.MustAdd(req.Principal)
	facility := domain.NewFacility(facilityID.String(), reserve.ID, borrowerID.String(), beneficiary.ID, operator, assetID.String(), req.Principal, s.state.Epoch, req.DueEpoch)
	for key, value := range req.Metadata {
		facility.Metadata[key] = value
	}
	if err := s.state.AddFacility(facility); err != nil {
		return stateError(err.Error())
	}
	s.journal.Append(ledger.Release(s.state.Epoch, reserve.ID, beneficiary.ID, assetID.String(), req.Principal, "facility principal disbursed"))
	s.state.AddMetric("facilityPrincipal", req.Principal)
	event := s.state.NextEvent(domain.EventFacilityOpened)
	event.ReserveID = reserve.ID
	event.FacilityID = facility.ID
	event.AccountID = beneficiary.ID
	event.Asset = assetID.String()
	event.Amount = req.Principal
	event.Message = "facility opened from reserve"
	event.Fields["reserveAfterOpen"] = evaluation.ReserveAfterOpen.String()
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) Repay(req RepaymentRequest) error {
	facilityID, err := ids.NewFacilityID(req.FacilityID)
	if err != nil {
		return invalid(err.Error())
	}
	facility, err := s.state.RequireFacility(facilityID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if facility.Status == domain.FacilityClosed {
		return stateError("facility already closed")
	}
	if req.Amount == 0 {
		req.Amount = facility.Outstanding
	}
	if req.Amount > facility.Outstanding {
		return stateError("repayment exceeds outstanding")
	}
	borrowerID := facility.Borrower
	if strings.TrimSpace(req.Borrower) != "" {
		parsed, err := ids.NewAccountID(req.Borrower)
		if err != nil {
			return invalid(err.Error())
		}
		borrowerID = parsed.String()
	}
	if borrowerID != facility.Borrower {
		return invalid("repayment borrower does not match facility")
	}
	borrower, err := s.state.RequireAccount(borrowerID)
	if err != nil {
		return notFound(err.Error())
	}
	reserve, err := s.state.RequireReserve(facility.ReserveID)
	if err != nil {
		return notFound(err.Error())
	}
	pool, err := s.state.RequirePool(facility.Asset)
	if err != nil {
		return notFound(err.Error())
	}
	fee, err := s.policy.SettlementFee(req.Amount)
	if err != nil {
		return policyError(err.Error())
	}
	totalDebit := req.Amount.MustAdd(fee)
	if borrower.Available(facility.Asset) < totalDebit {
		return solvencyError("borrower balance cannot cover repayment and fee")
	}
	insuranceFee, err := s.policy.InsuranceFeeShare(fee)
	if err != nil {
		return policyError(err.Error())
	}
	operatorFee, err := s.policy.OperatorFeeShare(fee)
	if err != nil {
		return policyError(err.Error())
	}
	if err := borrower.Debit(facility.Asset, totalDebit); err != nil {
		return solvencyError(err.Error())
	}
	balance := borrower.Balance(facility.Asset)
	balance.FeesPaid = balance.FeesPaid.MustAdd(fee)
	if err := reserve.ReleaseHeld(req.Amount); err != nil {
		return stateError(err.Error())
	}
	reserve.SettlementVolume = reserve.SettlementVolume.MustAdd(req.Amount)
	reserve.AccruedFees = reserve.AccruedFees.MustAdd(fee)
	if err := facility.ApplyRepayment(req.Amount, fee); err != nil {
		return stateError(err.Error())
	}
	if err := s.book.AddFeeShare(s.state.Epoch, reserve, pool, insuranceFee); err != nil {
		return stateError(err.Error())
	}
	if operatorFee > 0 && facility.Operator != "" {
		operator, err := s.state.RequireAccount(facility.Operator)
		if err != nil {
			return notFound(err.Error())
		}
		if err := operator.Credit(facility.Asset, operatorFee); err != nil {
			return stateError(err.Error())
		}
		s.journal.Append(ledger.Fee(s.state.Epoch, operator.ID, facility.ID, facility.Asset, operatorFee, "operator settlement fee"))
	}
	s.journal.Append(ledger.Debit(s.state.Epoch, borrower.ID, facility.Asset, totalDebit, "repayment debit"))
	s.journal.Append(ledger.Reserve(s.state.Epoch, reserve.ID, borrower.ID, facility.Asset, req.Amount, "repayment returned to reserve"))
	s.state.AddMetric("repayments", req.Amount)
	s.state.AddMetric("settlementFees", fee)
	event := s.state.NextEvent(domain.EventRepayment)
	event.FacilityID = facility.ID
	event.ReserveID = reserve.ID
	event.AccountID = borrower.ID
	event.Asset = facility.Asset
	event.Amount = req.Amount
	event.Message = "facility repayment settled"
	event.Fields["fee"] = fee.String()
	event.Fields["insuranceFee"] = insuranceFee.String()
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) ReportDefault(req DefaultRequest) error {
	defaultID, err := ids.NewDefaultID(req.ID)
	if err != nil {
		return invalid(err.Error())
	}
	facilityID, err := ids.NewFacilityID(req.FacilityID)
	if err != nil {
		return invalid(err.Error())
	}
	reporterID, err := ids.NewAccountID(req.Reporter)
	if err != nil {
		return invalid(err.Error())
	}
	facility, err := s.state.RequireFacility(facilityID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if _, err := s.state.RequireAccount(reporterID.String()); err != nil {
		return notFound(err.Error())
	}
	ceiling, pending, err := risk.EvaluateDefault(s.policy, facility, req.Amount, s.state.Epoch)
	if err != nil {
		return riskError(err.Error())
	}
	if req.Amount == 0 {
		req.Amount = facility.Outstanding
	}
	def := domain.NewDefaultCase(defaultID.String(), facility, reporterID.String(), req.Amount, ceiling, pending, s.state.Epoch)
	def.ExpectedRecovery = req.ExpectedRecovery
	for key, value := range req.Metadata {
		def.Metadata[key] = value
	}
	if err := s.state.AddDefault(def); err != nil {
		return stateError(err.Error())
	}
	facility.Status = domain.FacilityDefaulted
	facility.DefaultID = def.ID
	s.state.AddMetric("reportedDefaults", req.Amount)
	event := s.state.NextEvent(domain.EventDefaultReported)
	event.DefaultID = def.ID
	event.FacilityID = facility.ID
	event.AccountID = reporterID.String()
	event.Asset = facility.Asset
	event.Amount = req.Amount
	event.Message = "facility default reported"
	event.Fields["coverageCeiling"] = ceiling.String()
	event.Fields["pendingExposure"] = pending.String()
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) RegisterClaim(req ClaimRequest) error {
	claimID, err := ids.NewClaimID(req.ID)
	if err != nil {
		return invalid(err.Error())
	}
	defaultID, err := ids.NewDefaultID(req.DefaultID)
	if err != nil {
		return invalid(err.Error())
	}
	claimantID, err := ids.NewAccountID(req.Claimant)
	if err != nil {
		return invalid(err.Error())
	}
	payoutID := claimantID
	if strings.TrimSpace(req.PayoutAccount) != "" {
		payoutID, err = ids.NewAccountID(req.PayoutAccount)
		if err != nil {
			return invalid(err.Error())
		}
	}
	def, err := s.state.RequireDefault(defaultID.String())
	if err != nil {
		return notFound(err.Error())
	}
	pool, err := s.state.RequirePool(def.Asset)
	if err != nil {
		return notFound(err.Error())
	}
	if _, err := s.state.RequireAccount(claimantID.String()); err != nil {
		return notFound(err.Error())
	}
	if _, err := s.state.RequireAccount(payoutID.String()); err != nil {
		return notFound(err.Error())
	}
	if req.Amount == 0 {
		req.Amount = def.Amount
	}
	evaluation, err := risk.EvaluateClaim(s.state, s.policy, pool, def, req.Amount)
	if err != nil {
		return riskError(err.Error())
	}
	claim := domain.NewClaim(claimID.String(), def, claimantID.String(), payoutID.String(), req.Amount, evaluation.Coverage, evaluation.AvailableCoverage, s.state.Epoch)
	for key, value := range req.Metadata {
		claim.Metadata[key] = value
	}
	if err := s.state.AddClaim(claim); err != nil {
		return stateError(err.Error())
	}
	if err := def.AddClaim(claim.ID, claim.Coverage); err != nil {
		return stateError(err.Error())
	}
	s.state.AddMetric("registeredClaims", claim.Coverage)
	event := s.state.NextEvent(domain.EventClaimRegistered)
	event.DefaultID = def.ID
	event.ClaimID = claim.ID
	event.FacilityID = def.FacilityID
	event.AccountID = claimantID.String()
	event.Asset = def.Asset
	event.Amount = claim.Coverage
	event.Message = "insurance claim registered"
	event.Fields["availableCoverage"] = evaluation.AvailableCoverage.String()
	event.Fields["poolBalance"] = evaluation.PoolBalance.String()
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) ExecuteClaim(req ExecuteClaimRequest) error {
	claimID, err := ids.NewClaimID(req.ClaimID)
	if err != nil {
		return invalid(err.Error())
	}
	claim, err := s.state.RequireClaim(claimID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if !s.policy.CanExecuteClaim(s.state.Epoch, claim.RegisteredEpoch) {
		return policyError("claim execution delay has not elapsed")
	}
	def, err := s.state.RequireDefault(claim.DefaultID)
	if err != nil {
		return notFound(err.Error())
	}
	pool, err := s.state.RequirePool(claim.Asset)
	if err != nil {
		return notFound(err.Error())
	}
	payout, err := s.state.RequireAccount(claim.PayoutAccount)
	if err != nil {
		return notFound(err.Error())
	}
	if err := claim.Execute(s.state.Epoch); err != nil {
		return stateError(err.Error())
	}
	if err := s.book.PayClaim(s.state.Epoch, pool, claim, payout); err != nil {
		return stateError(err.Error())
	}
	if err := def.ApplyCoverage(claim.Coverage); err != nil {
		return stateError(err.Error())
	}
	if facility, err := s.state.RequireFacility(claim.FacilityID); err == nil {
		if claim.Coverage > facility.Outstanding {
			facility.Outstanding = 0
			facility.Status = domain.FacilityClosed
		} else {
			facility.Outstanding = facility.Outstanding.MustSub(claim.Coverage)
		}
	}
	s.state.AddMetric("executedClaims", claim.Coverage)
	event := s.state.NextEvent(domain.EventClaimExecuted)
	event.DefaultID = def.ID
	event.ClaimID = claim.ID
	event.FacilityID = claim.FacilityID
	event.AccountID = payout.ID
	event.Asset = claim.Asset
	event.Amount = claim.Coverage
	event.Message = "insurance claim executed"
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) ResolveDefault(req ResolveDefaultRequest) error {
	defaultID, err := ids.NewDefaultID(req.DefaultID)
	if err != nil {
		return invalid(err.Error())
	}
	def, err := s.state.RequireDefault(defaultID.String())
	if err != nil {
		return notFound(err.Error())
	}
	pool, err := s.state.RequirePool(def.Asset)
	if err != nil {
		return notFound(err.Error())
	}
	var recoveryFrom *domain.Account
	if req.RecoveryAmount > 0 {
		if !s.policy.AllowManualRecovery {
			return policyError("manual recovery disabled")
		}
		recoveryID, err := ids.NewAccountID(req.RecoveryFrom)
		if err != nil {
			return invalid(err.Error())
		}
		recoveryFrom, err = s.state.RequireAccount(recoveryID.String())
		if err != nil {
			return notFound(err.Error())
		}
		if err := s.book.Recover(s.state.Epoch, pool, recoveryFrom, def, req.RecoveryAmount); err != nil {
			return solvencyError(err.Error())
		}
	} else if err := def.Resolve(s.state.Epoch, 0); err != nil {
		return stateError(err.Error())
	}
	if facility, err := s.state.RequireFacility(def.FacilityID); err == nil && facility.Outstanding == 0 {
		facility.Status = domain.FacilityClosed
		facility.ClosedEpoch = s.state.Epoch
	}
	s.state.AddMetric("resolvedDefaults", def.Amount)
	event := s.state.NextEvent(domain.EventDefaultResolved)
	event.DefaultID = def.ID
	event.FacilityID = def.FacilityID
	event.Asset = def.Asset
	event.Amount = req.RecoveryAmount
	event.Message = "default resolution recorded"
	if recoveryFrom != nil {
		event.AccountID = recoveryFrom.ID
	}
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) Withdraw(req WithdrawalRequest) error {
	accountID, err := ids.NewAccountID(req.AccountID)
	if err != nil {
		return invalid(err.Error())
	}
	assetID, err := ids.NewAssetID(req.Asset)
	if err != nil {
		return invalid(err.Error())
	}
	account, err := s.state.RequireAccount(accountID.String())
	if err != nil {
		return notFound(err.Error())
	}
	if req.Amount <= 0 {
		return invalid("withdrawal amount required")
	}
	if err := account.Debit(assetID.String(), req.Amount); err != nil {
		return solvencyError("account balance cannot cover withdrawal")
	}
	balance := account.Balance(assetID.String())
	balance.Withdrawn = balance.Withdrawn.MustAdd(req.Amount)
	s.journal.Append(ledger.Debit(s.state.Epoch, account.ID, assetID.String(), req.Amount, "account withdrawal"))
	s.state.AddMetric("withdrawals", req.Amount)
	event := s.state.NextEvent(domain.EventWithdrawal)
	event.AccountID = account.ID
	event.Asset = assetID.String()
	event.Amount = req.Amount
	event.Message = "account withdrawal processed"
	s.state.AppendEvent(event)
	return nil
}

func (s *Service) Advance(req AdvanceRequest) error {
	if req.To > 0 {
		if req.To < s.state.Epoch {
			return stateError("cannot move epoch backwards")
		}
		s.state.Epoch = req.To
		return nil
	}
	if req.Epochs <= 0 {
		req.Epochs = 1
	}
	s.state.Epoch += req.Epochs
	return nil
}

func (s *Service) Validate() []string {
	messages := s.state.Validate()
	s.validations = messages
	return messages
}

func (s *Service) Report(includeEvents bool) domain.SystemReport {
	validations := s.Validate()
	return domain.BuildReport(s.state, includeEvents, validations)
}

func (s *Service) Summary() string {
	report := s.Report(false)
	return fmt.Sprintf("%s epoch=%d accounts=%d reserves=%d facilities=%d claims=%d", report.Name, report.Epoch, len(report.Accounts), len(report.Reserves), len(report.Facilities), len(report.Claims))
}

package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"emberdtl/src/amount"
	"emberdtl/src/domain"
	"emberdtl/src/engine"
	"emberdtl/src/policy"
)

type PolicyConfig struct {
	Name                       *string        `json:"name"`
	MinimumAccountFunding      *amount.Amount `json:"minimumAccountFunding"`
	MinimumReserveDeposit      *amount.Amount `json:"minimumReserveDeposit"`
	MinimumContribution        *amount.Amount `json:"minimumContribution"`
	MinimumFacilityPrincipal   *amount.Amount `json:"minimumFacilityPrincipal"`
	MaximumFacilityPrincipal   *amount.Amount `json:"maximumFacilityPrincipal"`
	SettlementFeeBps           *int64         `json:"settlementFeeBps"`
	MinimumSettlementFee       *amount.Amount `json:"minimumSettlementFee"`
	MaximumSettlementFee       *amount.Amount `json:"maximumSettlementFee"`
	InsuranceFeeShareBps       *int64         `json:"insuranceFeeShareBps"`
	MaxDefaultCoverageBps      *int64         `json:"maxDefaultCoverageBps"`
	PendingDefaultReserveBps   *int64         `json:"pendingDefaultReserveBps"`
	MaxClaimAmountBps          *int64         `json:"maxClaimAmountBps"`
	MinimumClaimAmount         *amount.Amount `json:"minimumClaimAmount"`
	DefaultGraceEpochs         *int           `json:"defaultGraceEpochs"`
	ClaimExecutionDelayEpochs  *int           `json:"claimExecutionDelayEpochs"`
	AllowEarlyDefaultReport    *bool          `json:"allowEarlyDefaultReport"`
	RequireOperatorForFacility *bool          `json:"requireOperatorForFacility"`
	RequireReserveSolvency     *bool          `json:"requireReserveSolvency"`
	AllowManualRecovery        *bool          `json:"allowManualRecovery"`
}

func (c PolicyConfig) Apply(base policy.Policy) policy.Policy {
	if c.Name != nil {
		base.Name = *c.Name
	}
	if c.MinimumAccountFunding != nil {
		base.MinimumAccountFunding = *c.MinimumAccountFunding
	}
	if c.MinimumReserveDeposit != nil {
		base.MinimumReserveDeposit = *c.MinimumReserveDeposit
	}
	if c.MinimumContribution != nil {
		base.MinimumContribution = *c.MinimumContribution
	}
	if c.MinimumFacilityPrincipal != nil {
		base.MinimumFacilityPrincipal = *c.MinimumFacilityPrincipal
	}
	if c.MaximumFacilityPrincipal != nil {
		base.MaximumFacilityPrincipal = *c.MaximumFacilityPrincipal
	}
	if c.SettlementFeeBps != nil {
		base.SettlementFeeBps = *c.SettlementFeeBps
	}
	if c.MinimumSettlementFee != nil {
		base.MinimumSettlementFee = *c.MinimumSettlementFee
	}
	if c.MaximumSettlementFee != nil {
		base.MaximumSettlementFee = *c.MaximumSettlementFee
	}
	if c.InsuranceFeeShareBps != nil {
		base.InsuranceFeeShareBps = *c.InsuranceFeeShareBps
	}
	if c.MaxDefaultCoverageBps != nil {
		base.MaxDefaultCoverageBps = *c.MaxDefaultCoverageBps
	}
	if c.PendingDefaultReserveBps != nil {
		base.PendingDefaultReserveBps = *c.PendingDefaultReserveBps
	}
	if c.MaxClaimAmountBps != nil {
		base.MaxClaimAmountBps = *c.MaxClaimAmountBps
	}
	if c.MinimumClaimAmount != nil {
		base.MinimumClaimAmount = *c.MinimumClaimAmount
	}
	if c.DefaultGraceEpochs != nil {
		base.DefaultGraceEpochs = *c.DefaultGraceEpochs
	}
	if c.ClaimExecutionDelayEpochs != nil {
		base.ClaimExecutionDelayEpochs = *c.ClaimExecutionDelayEpochs
	}
	if c.AllowEarlyDefaultReport != nil {
		base.AllowEarlyDefaultReport = *c.AllowEarlyDefaultReport
	}
	if c.RequireOperatorForFacility != nil {
		base.RequireOperatorForFacility = *c.RequireOperatorForFacility
	}
	if c.RequireReserveSolvency != nil {
		base.RequireReserveSolvency = *c.RequireReserveSolvency
	}
	if c.AllowManualRecovery != nil {
		base.AllowManualRecovery = *c.AllowManualRecovery
	}
	return base
}

type Step struct {
	Action string `json:"action"`

	ID               string            `json:"id"`
	AccountID        string            `json:"accountId"`
	ReserveID        string            `json:"reserveId"`
	FacilityID       string            `json:"facilityId"`
	DefaultID        string            `json:"defaultId"`
	ClaimID          string            `json:"claimId"`
	Role             string            `json:"role"`
	DisplayName      string            `json:"displayName"`
	Owner            string            `json:"owner"`
	Asset            string            `json:"asset"`
	Symbol           string            `json:"symbol"`
	Source           string            `json:"source"`
	Borrower         string            `json:"borrower"`
	Beneficiary      string            `json:"beneficiary"`
	Operator         string            `json:"operator"`
	Reporter         string            `json:"reporter"`
	Claimant         string            `json:"claimant"`
	PayoutAccount    string            `json:"payoutAccount"`
	RecoveryFrom     string            `json:"recoveryFrom"`
	Decimals         int               `json:"decimals"`
	DueEpoch         int               `json:"dueEpoch"`
	Epochs           int               `json:"epochs"`
	To               int               `json:"to"`
	Amount           amount.Amount     `json:"amount"`
	Principal        amount.Amount     `json:"principal"`
	ExpectedRecovery amount.Amount     `json:"expectedRecovery"`
	RecoveryAmount   amount.Amount     `json:"recoveryAmount"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

type Scenario struct {
	Name     string                  `json:"name"`
	Policy   PolicyConfig            `json:"policy"`
	Assets   []engine.AssetRequest   `json:"assets"`
	Accounts []engine.AccountRequest `json:"accounts"`
	Steps    []Step                  `json:"steps"`
}

type Result struct {
	Service *engine.Service
	Report  domain.SystemReport
}

func LoadFile(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var scenario Scenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		return nil, err
	}
	if strings.TrimSpace(scenario.Name) == "" {
		scenario.Name = "EmberDTL"
	}
	return &scenario, nil
}

func RunFile(path string, includeEvents bool) (*Result, error) {
	loaded, err := LoadFile(path)
	if err != nil {
		return nil, err
	}
	return loaded.Run(includeEvents)
}

func ValidateFile(path string) ([]string, error) {
	result, err := RunFile(path, false)
	if err != nil {
		return nil, err
	}
	return result.Service.Validate(), nil
}

func (s Scenario) Run(includeEvents bool) (*Result, error) {
	rules := s.Policy.Apply(policy.Default())
	service, err := engine.New(s.Name, rules)
	if err != nil {
		return nil, err
	}
	for _, asset := range s.Assets {
		if err := service.RegisterAsset(asset); err != nil {
			return nil, fmt.Errorf("asset %s: %w", asset.ID, err)
		}
	}
	for _, account := range s.Accounts {
		if err := service.RegisterAccount(account); err != nil {
			return nil, fmt.Errorf("account %s: %w", account.ID, err)
		}
	}
	for index, step := range s.Steps {
		if err := executeStep(service, step); err != nil {
			return nil, fmt.Errorf("step %d %s: %w", index+1, step.Action, err)
		}
	}
	report := service.Report(includeEvents)
	return &Result{Service: service, Report: report}, nil
}

func executeStep(service *engine.Service, step Step) error {
	switch normalizeAction(step.Action) {
	case "account", "registeraccount", "createaccount":
		return service.RegisterAccount(engine.AccountRequest{
			ID:          firstNonEmpty(step.ID, step.AccountID),
			Role:        step.Role,
			DisplayName: step.DisplayName,
			Metadata:    step.Metadata,
		})
	case "asset", "registerasset":
		return service.RegisterAsset(engine.AssetRequest{
			ID:       step.ID,
			Symbol:   step.Symbol,
			Decimals: step.Decimals,
			Metadata: step.Metadata,
		})
	case "fund", "fundaccount", "mint":
		return service.FundAccount(engine.FundAccountRequest{
			AccountID: step.AccountID,
			Asset:     step.Asset,
			Amount:    step.Amount,
			Source:    step.Source,
		})
	case "reserve", "openreserve":
		return service.OpenReserve(engine.ReserveRequest{
			ID:       firstNonEmpty(step.ID, step.ReserveID),
			Owner:    step.Owner,
			Asset:    step.Asset,
			Metadata: step.Metadata,
		})
	case "deposit", "depositreserve", "reservedeposit":
		return service.DepositReserve(engine.ReserveDepositRequest{
			ReserveID: firstNonEmpty(step.ReserveID, step.ID),
			AccountID: step.AccountID,
			Asset:     step.Asset,
			Amount:    step.Amount,
		})
	case "contribute", "poolcontribution":
		return service.Contribute(engine.ContributionRequest{
			AccountID: step.AccountID,
			Asset:     step.Asset,
			Amount:    step.Amount,
		})
	case "facility", "openfacility":
		return service.OpenFacility(engine.FacilityRequest{
			ID:          firstNonEmpty(step.ID, step.FacilityID),
			ReserveID:   step.ReserveID,
			Borrower:    step.Borrower,
			Beneficiary: step.Beneficiary,
			Operator:    step.Operator,
			Asset:       step.Asset,
			Principal:   step.Principal,
			DueEpoch:    step.DueEpoch,
			Metadata:    step.Metadata,
		})
	case "repay", "repayment", "settle":
		return service.Repay(engine.RepaymentRequest{
			FacilityID: firstNonEmpty(step.FacilityID, step.ID),
			Borrower:   step.Borrower,
			Amount:     step.Amount,
		})
	case "default", "reportdefault":
		return service.ReportDefault(engine.DefaultRequest{
			ID:               firstNonEmpty(step.ID, step.DefaultID),
			FacilityID:       step.FacilityID,
			Reporter:         step.Reporter,
			Amount:           step.Amount,
			ExpectedRecovery: step.ExpectedRecovery,
			Metadata:         step.Metadata,
		})
	case "claim", "registerclaim":
		return service.RegisterClaim(engine.ClaimRequest{
			ID:            firstNonEmpty(step.ID, step.ClaimID),
			DefaultID:     step.DefaultID,
			Claimant:      step.Claimant,
			PayoutAccount: step.PayoutAccount,
			Amount:        step.Amount,
			Metadata:      step.Metadata,
		})
	case "executeclaim", "payclaim":
		return service.ExecuteClaim(engine.ExecuteClaimRequest{
			ClaimID: firstNonEmpty(step.ClaimID, step.ID),
		})
	case "resolve", "resolvedefault":
		return service.ResolveDefault(engine.ResolveDefaultRequest{
			DefaultID:      firstNonEmpty(step.DefaultID, step.ID),
			RecoveryFrom:   step.RecoveryFrom,
			RecoveryAmount: step.RecoveryAmount,
		})
	case "withdraw", "withdrawal":
		return service.Withdraw(engine.WithdrawalRequest{
			AccountID: step.AccountID,
			Asset:     step.Asset,
			Amount:    step.Amount,
		})
	case "advance", "advanceepoch":
		return service.Advance(engine.AdvanceRequest{Epochs: step.Epochs, To: step.To})
	case "validate":
		service.Validate()
		return nil
	default:
		return fmt.Errorf("unknown action %q", step.Action)
	}
}

func EncodeReport(report domain.SystemReport, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(report, "", "  ")
	}
	return json.Marshal(report)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func normalizeAction(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	action = strings.ReplaceAll(action, "_", "")
	action = strings.ReplaceAll(action, "-", "")
	action = strings.ReplaceAll(action, " ", "")
	return action
}

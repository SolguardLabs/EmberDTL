package engine

import "emberdtl/src/amount"

type AssetRequest struct {
	ID       string            `json:"id"`
	Symbol   string            `json:"symbol"`
	Decimals int               `json:"decimals"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type AccountRequest struct {
	ID          string            `json:"id"`
	Role        string            `json:"role"`
	DisplayName string            `json:"displayName,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type FundAccountRequest struct {
	AccountID string        `json:"accountId"`
	Asset     string        `json:"asset"`
	Amount    amount.Amount `json:"amount"`
	Source    string        `json:"source,omitempty"`
}

type ReserveRequest struct {
	ID       string            `json:"id"`
	Owner    string            `json:"owner"`
	Asset    string            `json:"asset"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type ReserveDepositRequest struct {
	ReserveID string        `json:"reserveId"`
	AccountID string        `json:"accountId"`
	Asset     string        `json:"asset"`
	Amount    amount.Amount `json:"amount"`
}

type ContributionRequest struct {
	AccountID string        `json:"accountId"`
	Asset     string        `json:"asset"`
	Amount    amount.Amount `json:"amount"`
}

type FacilityRequest struct {
	ID          string            `json:"id"`
	ReserveID   string            `json:"reserveId"`
	Borrower    string            `json:"borrower"`
	Beneficiary string            `json:"beneficiary"`
	Operator    string            `json:"operator"`
	Asset       string            `json:"asset"`
	Principal   amount.Amount     `json:"principal"`
	DueEpoch    int               `json:"dueEpoch"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type RepaymentRequest struct {
	FacilityID string        `json:"facilityId"`
	Borrower   string        `json:"borrower,omitempty"`
	Amount     amount.Amount `json:"amount"`
}

type DefaultRequest struct {
	ID               string            `json:"id"`
	FacilityID       string            `json:"facilityId"`
	Reporter         string            `json:"reporter"`
	Amount           amount.Amount     `json:"amount"`
	ExpectedRecovery amount.Amount     `json:"expectedRecovery"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

type ClaimRequest struct {
	ID            string            `json:"id"`
	DefaultID     string            `json:"defaultId"`
	Claimant      string            `json:"claimant"`
	PayoutAccount string            `json:"payoutAccount"`
	Amount        amount.Amount     `json:"amount"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type ExecuteClaimRequest struct {
	ClaimID string `json:"claimId"`
}

type ResolveDefaultRequest struct {
	DefaultID      string        `json:"defaultId"`
	RecoveryFrom   string        `json:"recoveryFrom"`
	RecoveryAmount amount.Amount `json:"recoveryAmount"`
}

type WithdrawalRequest struct {
	AccountID string        `json:"accountId"`
	Asset     string        `json:"asset"`
	Amount    amount.Amount `json:"amount"`
}

type AdvanceRequest struct {
	Epochs int `json:"epochs"`
	To     int `json:"to"`
}

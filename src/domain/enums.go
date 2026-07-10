package domain

type AccountRole string

const (
	RoleTreasury    AccountRole = "treasury"
	RoleBorrower    AccountRole = "borrower"
	RoleMerchant    AccountRole = "merchant"
	RoleOperator    AccountRole = "operator"
	RoleUnderwriter AccountRole = "underwriter"
	RoleUser        AccountRole = "user"
)

type AccountStatus string

const (
	AccountActive AccountStatus = "active"
	AccountFrozen AccountStatus = "frozen"
)

type ReserveStatus string

const (
	ReserveOpen   ReserveStatus = "open"
	ReserveClosed ReserveStatus = "closed"
)

type FacilityStatus string

const (
	FacilityOpen       FacilityStatus = "open"
	FacilityPerforming FacilityStatus = "performing"
	FacilityDefaulted  FacilityStatus = "defaulted"
	FacilityClosed     FacilityStatus = "closed"
)

type DefaultStatus string

const (
	DefaultReported DefaultStatus = "reported"
	DefaultAccepted DefaultStatus = "accepted"
	DefaultResolved DefaultStatus = "resolved"
)

type ClaimStatus string

const (
	ClaimRegistered ClaimStatus = "registered"
	ClaimApproved   ClaimStatus = "approved"
	ClaimExecuted   ClaimStatus = "executed"
	ClaimRejected   ClaimStatus = "rejected"
)

type PoolStatus string

const (
	PoolActive PoolStatus = "active"
	PoolPaused PoolStatus = "paused"
)

type EventType string

const (
	EventAssetRegistered   EventType = "asset_registered"
	EventAccountRegistered EventType = "account_registered"
	EventReserveOpened     EventType = "reserve_opened"
	EventFunding           EventType = "funding"
	EventReserveDeposit    EventType = "reserve_deposit"
	EventFacilityOpened    EventType = "facility_opened"
	EventRepayment         EventType = "repayment"
	EventContribution      EventType = "contribution"
	EventDefaultReported   EventType = "default_reported"
	EventClaimRegistered   EventType = "claim_registered"
	EventClaimExecuted     EventType = "claim_executed"
	EventDefaultResolved   EventType = "default_resolved"
	EventWithdrawal        EventType = "withdrawal"
	EventValidation        EventType = "validation"
)

package insurance

import (
	"fmt"

	"emberdtl/src/amount"
	"emberdtl/src/domain"
	"emberdtl/src/ledger"
)

type Book struct {
	state   *domain.State
	journal *ledger.Journal
}

func NewBook(state *domain.State, journal *ledger.Journal) *Book {
	return &Book{state: state, journal: journal}
}

func (b *Book) Contribute(epoch int, account *domain.Account, pool *domain.InsurancePool, value amount.Amount) error {
	if err := account.Debit(pool.Asset, value); err != nil {
		return err
	}
	if err := pool.CreditContribution(value); err != nil {
		return err
	}
	balance := account.Balance(pool.Asset)
	balance.Contributed = balance.Contributed.MustAdd(value)
	b.journal.Append(ledger.Contribution(epoch, account.ID, pool.Asset, value, "insurance pool contribution"))
	return nil
}

func (b *Book) AddFeeShare(epoch int, reserve *domain.ReserveBucket, pool *domain.InsurancePool, value amount.Amount) error {
	if value == 0 {
		return nil
	}
	if err := pool.CreditFee(value); err != nil {
		return err
	}
	reserve.InsuranceContributions = reserve.InsuranceContributions.MustAdd(value)
	b.journal.Append(ledger.Contribution(epoch, reserve.Owner, pool.Asset, value, "fee share credited to insurance pool"))
	return nil
}

func (b *Book) PayClaim(epoch int, pool *domain.InsurancePool, claim *domain.Claim, payout *domain.Account) error {
	if err := pool.DebitClaim(claim.Coverage); err != nil {
		return err
	}
	if err := payout.Credit(claim.Asset, claim.Coverage); err != nil {
		return err
	}
	balance := payout.Balance(claim.Asset)
	balance.ClaimsPaid = balance.ClaimsPaid.MustAdd(claim.Coverage)
	b.journal.Append(ledger.Claim(epoch, payout.ID, claim.DefaultID, claim.ID, claim.Asset, claim.Coverage, "insurance claim payout"))
	return nil
}

func (b *Book) Recover(epoch int, pool *domain.InsurancePool, account *domain.Account, def *domain.DefaultCase, value amount.Amount) error {
	if value == 0 {
		return nil
	}
	if err := account.Debit(def.Asset, value); err != nil {
		return err
	}
	if err := pool.CreditRecovery(value); err != nil {
		return err
	}
	if err := def.Resolve(epoch, value); err != nil {
		return err
	}
	b.journal.Append(ledger.Recovery(epoch, account.ID, def.ID, def.Asset, value, "default recovery credited to pool"))
	return nil
}

func (b *Book) AssertAsset(pool *domain.InsurancePool, asset string) error {
	if pool.Asset != asset {
		return fmt.Errorf("pool asset %s does not match %s", pool.Asset, asset)
	}
	return nil
}

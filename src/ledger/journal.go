package ledger

import (
	"fmt"
	"sort"
	"time"

	"emberdtl/src/amount"
)

type EntryKind string

const (
	EntryCredit       EntryKind = "credit"
	EntryDebit        EntryKind = "debit"
	EntryReserve      EntryKind = "reserve"
	EntryRelease      EntryKind = "release"
	EntryContribution EntryKind = "contribution"
	EntryFee          EntryKind = "fee"
	EntryClaim        EntryKind = "claim"
	EntryRecovery     EntryKind = "recovery"
)

type Entry struct {
	Sequence int           `json:"sequence"`
	Epoch    int           `json:"epoch"`
	Kind     EntryKind     `json:"kind"`
	Asset    string        `json:"asset"`
	Account  string        `json:"account,omitempty"`
	Reserve  string        `json:"reserve,omitempty"`
	Facility string        `json:"facility,omitempty"`
	Default  string        `json:"default,omitempty"`
	Claim    string        `json:"claim,omitempty"`
	Amount   amount.Amount `json:"amount"`
	Memo     string        `json:"memo"`
	Time     time.Time     `json:"time"`
}

type Journal struct {
	entries []Entry
}

func NewJournal() *Journal {
	return &Journal{entries: []Entry{}}
}

func (j *Journal) Append(entry Entry) Entry {
	entry.Sequence = len(j.entries) + 1
	if entry.Time.IsZero() {
		entry.Time = time.Unix(0, int64(entry.Sequence)).UTC()
	}
	j.entries = append(j.entries, entry)
	return entry
}

func (j *Journal) AppendAll(entries []Entry) {
	for _, entry := range entries {
		j.Append(entry)
	}
}

func (j *Journal) Entries() []Entry {
	return append([]Entry(nil), j.entries...)
}

func (j *Journal) ByAsset(asset string) []Entry {
	result := []Entry{}
	for _, entry := range j.entries {
		if entry.Asset == asset {
			result = append(result, entry)
		}
	}
	return result
}

func (j *Journal) TotalsByKind(asset string) map[EntryKind]amount.Amount {
	totals := map[EntryKind]amount.Amount{}
	for _, entry := range j.entries {
		if asset != "" && entry.Asset != asset {
			continue
		}
		totals[entry.Kind] = totals[entry.Kind].MustAdd(entry.Amount)
	}
	return totals
}

func (j *Journal) ReconcileAssets() []AssetReconciliation {
	assets := map[string]struct{}{}
	for _, entry := range j.entries {
		assets[entry.Asset] = struct{}{}
	}
	keys := make([]string, 0, len(assets))
	for key := range assets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := []AssetReconciliation{}
	for _, asset := range keys {
		totals := j.TotalsByKind(asset)
		rows = append(rows, AssetReconciliation{
			Asset:         asset,
			Credits:       totals[EntryCredit],
			Debits:        totals[EntryDebit],
			Reserves:      totals[EntryReserve],
			Releases:      totals[EntryRelease],
			Contributions: totals[EntryContribution],
			Fees:          totals[EntryFee],
			Claims:        totals[EntryClaim],
			Recoveries:    totals[EntryRecovery],
		})
	}
	return rows
}

type AssetReconciliation struct {
	Asset         string        `json:"asset"`
	Credits       amount.Amount `json:"credits"`
	Debits        amount.Amount `json:"debits"`
	Reserves      amount.Amount `json:"reserves"`
	Releases      amount.Amount `json:"releases"`
	Contributions amount.Amount `json:"contributions"`
	Fees          amount.Amount `json:"fees"`
	Claims        amount.Amount `json:"claims"`
	Recoveries    amount.Amount `json:"recoveries"`
}

func Credit(epoch int, account, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryCredit, Account: account, Asset: asset, Amount: value, Memo: memo}
}

func Debit(epoch int, account, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryDebit, Account: account, Asset: asset, Amount: value, Memo: memo}
}

func Reserve(epoch int, reserve, account, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryReserve, Reserve: reserve, Account: account, Asset: asset, Amount: value, Memo: memo}
}

func Release(epoch int, reserve, account, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryRelease, Reserve: reserve, Account: account, Asset: asset, Amount: value, Memo: memo}
}

func Contribution(epoch int, account, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryContribution, Account: account, Asset: asset, Amount: value, Memo: memo}
}

func Fee(epoch int, account, facility, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryFee, Account: account, Facility: facility, Asset: asset, Amount: value, Memo: memo}
}

func Claim(epoch int, account, def, claim, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryClaim, Account: account, Default: def, Claim: claim, Asset: asset, Amount: value, Memo: memo}
}

func Recovery(epoch int, account, def, asset string, value amount.Amount, memo string) Entry {
	return Entry{Epoch: epoch, Kind: EntryRecovery, Account: account, Default: def, Asset: asset, Amount: value, Memo: memo}
}

func Describe(entry Entry) string {
	target := entry.Account
	if target == "" {
		target = entry.Reserve
	}
	if target == "" {
		target = entry.Facility
	}
	return fmt.Sprintf("%06d %s %s %s %s", entry.Sequence, entry.Kind, target, entry.Amount, entry.Asset)
}

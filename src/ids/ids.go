package ids

import (
	"fmt"
	"strings"
	"unicode"
)

type Kind string

const (
	KindAccount  Kind = "account"
	KindAsset    Kind = "asset"
	KindReserve  Kind = "reserve"
	KindFacility Kind = "facility"
	KindDefault  Kind = "default"
	KindClaim    Kind = "claim"
	KindEvent    Kind = "event"
)

type Identifier struct {
	kind Kind
	raw  string
}

func New(kind Kind, raw string) (Identifier, error) {
	normalized := normalize(raw)
	if normalized == "" {
		return Identifier{}, fmt.Errorf("%s id is required", kind)
	}
	if len(normalized) < 2 || len(normalized) > 80 {
		return Identifier{}, fmt.Errorf("%s id length out of range", kind)
	}
	if !isAllowedStart(rune(normalized[0])) {
		return Identifier{}, fmt.Errorf("%s id must start with a letter or number", kind)
	}
	for _, char := range normalized {
		if !isAllowed(char) {
			return Identifier{}, fmt.Errorf("%s id contains unsupported character %q", kind, char)
		}
	}
	return Identifier{kind: kind, raw: normalized}, nil
}

func NewAccountID(raw string) (Identifier, error) {
	return New(KindAccount, raw)
}

func NewAssetID(raw string) (Identifier, error) {
	return New(KindAsset, raw)
}

func NewReserveID(raw string) (Identifier, error) {
	return New(KindReserve, raw)
}

func NewFacilityID(raw string) (Identifier, error) {
	return New(KindFacility, raw)
}

func NewDefaultID(raw string) (Identifier, error) {
	return New(KindDefault, raw)
}

func NewClaimID(raw string) (Identifier, error) {
	return New(KindClaim, raw)
}

func NewEventID(raw string) (Identifier, error) {
	return New(KindEvent, raw)
}

func (id Identifier) String() string {
	return id.raw
}

func (id Identifier) Kind() Kind {
	return id.kind
}

func (id Identifier) Equal(other Identifier) bool {
	return id.kind == other.kind && id.raw == other.raw
}

func normalize(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "_", "-")
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return strings.Trim(value, "-")
}

func isAllowedStart(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char)
}

func isAllowed(char rune) bool {
	return unicode.IsLetter(char) || unicode.IsDigit(char) || char == '-' || char == '.'
}

func RequireKind(id Identifier, kind Kind) error {
	if id.Kind() != kind {
		return fmt.Errorf("expected %s id, got %s", kind, id.Kind())
	}
	return nil
}

package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MsgTypeCreateSideChainValidator = "side_create_validator"
	MsgTypeEditSideChainValidator   = "side_edit_validator"
	MsgTypeSideChainDelegate        = "side_delegate"
	MsgTypeSideChainBeginRedelegate = "side_begin_redelegate"
	MsgTypeSideChainUndelegate      = "side_undelegate"

	MaxSideChainIdLength = 20
)

type MsgCreateSideChainValidator struct {
	Description   Description    `json:"description"`
	Commission    CommissionMsg  `json:"commission"`
	DelegatorAddr sdk.AccAddress `json:"delegator_address"`
	ValidatorAddr sdk.ValAddress `json:"validator_address"`
	Delegation    sdk.Coin       `json:"delegation"`
	SideChainId   string         `json:"side_chain_id"`
	SideConsAddr  []byte         `json:"side_cons_addr"`
	SideFeeAddr   []byte         `json:"side_fee_addr"`
}

func NewMsgCreateSideChainValidator(valAddr sdk.ValAddress, delegation sdk.Coin,
	description Description, commission CommissionMsg, sideChainId string, sideConsAddr []byte, sideFeeAddr []byte) MsgCreateSideChainValidator {
	return NewMsgCreateSideChainValidatorOnBehalfOf(sdk.AccAddress(valAddr), valAddr, delegation, description, commission, sideChainId, sideConsAddr, sideFeeAddr)
}

func NewMsgCreateSideChainValidatorOnBehalfOf(delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, delegation sdk.Coin,
	description Description, commission CommissionMsg, sideChainId string, sideConsAddr []byte, sideFeeAddr []byte) MsgCreateSideChainValidator {
	return MsgCreateSideChainValidator{
		Description:   description,
		Commission:    commission,
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: valAddr,
		Delegation:    delegation,
		SideChainId:   sideChainId,
		SideConsAddr:  sideConsAddr,
		SideFeeAddr:   sideFeeAddr,
	}
}

//nolint
func (msg MsgCreateSideChainValidator) Route() string { return MsgRoute }
func (msg MsgCreateSideChainValidator) Type() string  { return MsgTypeCreateSideChainValidator }

// Return address(es) that must sign over msg.GetSignBytes()
func (msg MsgCreateSideChainValidator) GetSigners() []sdk.AccAddress {
	// delegator is first signer so delegator pays fees
	addrs := []sdk.AccAddress{msg.DelegatorAddr}

	if !bytes.Equal(msg.DelegatorAddr.Bytes(), msg.ValidatorAddr.Bytes()) {
		// if validator addr is not same as delegator addr, validator must sign
		// msg as well
		addrs = append(addrs, sdk.AccAddress(msg.ValidatorAddr))
	}
	return addrs
}

// get the bytes for the message signer to sign on
func (msg MsgCreateSideChainValidator) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgCreateSideChainValidator) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}

func (msg MsgCreateSideChainValidator) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}
	if msg.Delegation.Amount < 1e8 {
		return ErrBadDelegationAmount(DefaultCodespace, "self delegation must not be less than 1e8")
	}
	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}
	commission := NewCommission(msg.Commission.Rate, msg.Commission.MaxRate, msg.Commission.MaxChangeRate)
	if err := commission.Validate(); err != nil {
		return err
	}

	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > MaxSideChainIdLength {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "side chain id must be included and max length is 20 bytes")
	}

	if len(msg.SideConsAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected SideConsAddr length is %d, actual length is %d", sdk.AddrLen, len(msg.SideConsAddr)))
	}

	if len(msg.SideFeeAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected SideFeeAddr length is %d, actual length is %d", sdk.AddrLen, len(msg.SideFeeAddr)))
	}

	return nil
}

//______________________________________________________________________
type MsgEditSideChainValidator struct {
	Description   Description    `json:"description"`
	ValidatorAddr sdk.ValAddress `json:"address"`

	// We pass a reference to the new commission rate as it's not mandatory to
	// update. If not updated, the deserialized rate will be zero with no way to
	// distinguish if an update was intended.
	//
	// REF: #2373
	CommissionRate *sdk.Dec `json:"commission_rate"`

	SideChainId string `json:"side_chain_id"`
	// for SideConsAddr and SideFeeAddr, we do not update the values if they are not provided.
	SideConsAddr []byte `json:"side_cons_addr"`
	SideFeeAddr  []byte `json:"side_fee_addr"`
}

func NewMsgEditSideChainValidator(sideChainId string, validatorAddr sdk.ValAddress, description Description, commissionRate *sdk.Dec, sideConsAddr, sideFeeAddr []byte) MsgEditSideChainValidator {
	return MsgEditSideChainValidator{
		Description:    description,
		ValidatorAddr:  validatorAddr,
		CommissionRate: commissionRate,
		SideChainId:    sideChainId,
		SideConsAddr:   sideConsAddr,
		SideFeeAddr:    sideFeeAddr,
	}
}

//nolint
func (msg MsgEditSideChainValidator) Route() string { return MsgRoute }
func (msg MsgEditSideChainValidator) Type() string  { return MsgTypeEditSideChainValidator }
func (msg MsgEditSideChainValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.AccAddress(msg.ValidatorAddr)}
}

// get the bytes for the message signer to sign on
func (msg MsgEditSideChainValidator) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgEditSideChainValidator) ValidateBasic() sdk.Error {
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}

	if msg.Description == (Description{}) {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "description must be included")
	}
	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}

	if msg.CommissionRate != nil {
		if msg.CommissionRate.GT(sdk.OneDec()) || msg.CommissionRate.LT(sdk.ZeroDec()) {
			return sdk.NewError(DefaultCodespace, CodeInvalidInput, "commission rate must be between 0 and 1 (inclusive)")
		}
	}

	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > MaxSideChainIdLength {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "side chain id must be included and max length is 20 bytes")
	}

	// if SideConsAddr is empty, we do not update it.
	if len(msg.SideConsAddr) != 0 {
		if len(msg.SideConsAddr) != sdk.AddrLen {
			return sdk.ErrInvalidAddress(fmt.Sprintf("Expected SideConsAddr length is %d, actual length is %d", sdk.AddrLen, len(msg.SideConsAddr)))
		}
	}

	// if SideFeeAddr is empty, we do not update it.
	if len(msg.SideFeeAddr) != 0 {
		if len(msg.SideFeeAddr) != sdk.AddrLen {
			return sdk.ErrInvalidAddress(fmt.Sprintf("Expected SideFeeAddr length is %d, actual length is %d", sdk.AddrLen, len(msg.SideFeeAddr)))
		}
	}

	return nil
}

func (msg MsgEditSideChainValidator) GetInvolvedAddresses() []sdk.AccAddress {
	return msg.GetSigners()
}

//______________________________________________________________________
type MsgSideChainDelegate struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
	Delegation    sdk.Coin       `json:"delegation"`

	SideChainId string `json:"side_chain_id"`
}

func NewMsgSideChainDelegate(sideChainId string, delAddr sdk.AccAddress, valAddr sdk.ValAddress, delegation sdk.Coin) MsgSideChainDelegate {
	return MsgSideChainDelegate{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Delegation:    delegation,
		SideChainId:   sideChainId,
	}
}

//nolint
func (msg MsgSideChainDelegate) Route() string { return MsgRoute }
func (msg MsgSideChainDelegate) Type() string  { return MsgTypeSideChainDelegate }
func (msg MsgSideChainDelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgSideChainDelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// quick validity check
func (msg MsgSideChainDelegate) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}
	// we need this lower limit to prevent too many delegation records.
	if msg.Delegation.Amount < 1e8 {
		return ErrBadDelegationAmount(DefaultCodespace, "delegation must not be less than 1e8")
	}
	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > MaxSideChainIdLength {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "side chain id must be included and max length is 20 bytes")
	}
	return nil
}

func (msg MsgSideChainDelegate) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr, sdk.AccAddress(msg.ValidatorAddr)}
}

//______________________________________________________________________
type MsgSideChainBeginRedelegate struct {
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`
	ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"`
	ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"`
	Amount           sdk.Coin        `json:"amount"`
	SideChainId      string         `json:"side_chain_id"`
}

func NewMsgSideChainBeginRedelegate(sideChainId string, delegatorAddr sdk.AccAddress, valSrcAddr sdk.ValAddress, valDstAddr sdk.ValAddress, amount sdk.Coin) MsgSideChainBeginRedelegate {
	return MsgSideChainBeginRedelegate{
		DelegatorAddr:    delegatorAddr,
		ValidatorSrcAddr: valSrcAddr,
		ValidatorDstAddr: valDstAddr,
		Amount:           amount,
		SideChainId:      sideChainId,
	}
}

//nolint
func (msg MsgSideChainBeginRedelegate) Route() string { return MsgRoute }
func (msg MsgSideChainBeginRedelegate) Type() string  { return MsgTypeSideChainBeginRedelegate }
func (msg MsgSideChainBeginRedelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

// get the bytes for the message signer to sign on
func (msg MsgSideChainBeginRedelegate) GetSignBytes() []byte {
	b := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(b)
}

func (msg MsgSideChainBeginRedelegate) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr, sdk.AccAddress(msg.ValidatorSrcAddr), sdk.AccAddress(msg.DelegatorAddr)}
}

func (msg MsgSideChainBeginRedelegate) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorSrcAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected ValidatorSrcAddr length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorSrcAddr)))
	}
	if len(msg.ValidatorDstAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected ValidatorDstAddr length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorDstAddr)))
	}
	if bytes.Equal(msg.ValidatorSrcAddr, msg.ValidatorDstAddr) {
		return ErrSelfRedelegation(DefaultCodespace)
	}
	if msg.Amount.Amount < 1e8 {
		return ErrBadDelegationAmount(DefaultCodespace, "redelegation amount must not be less than 1e8 ")
	}
	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > MaxSideChainIdLength {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "side chain id must be included and max length is 20 bytes")
	}
	return nil
}

//______________________________________________________________________
type MsgSideChainUndelegate struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.ValAddress `json:"validator_addr"`
	Amount        sdk.Coin        `json:"amount"`
	SideChainId   string         `json:"side_chain_id"`
}

func NewMsgSideChainUndelegate(sideChainId string, delegatorAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) MsgSideChainUndelegate {
	return MsgSideChainUndelegate{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: valAddr,
		Amount:        amount,
		SideChainId:   sideChainId,
	}
}

//nolint
func (msg MsgSideChainUndelegate) Route() string { return MsgRoute }
func (msg MsgSideChainUndelegate) Type() string  { return MsgTypeSideChainUndelegate }
func (msg MsgSideChainUndelegate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr}
}

func (msg MsgSideChainUndelegate) GetSignBytes() []byte {
	bz := MsgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgSideChainUndelegate) ValidateBasic() sdk.Error {
	if len(msg.DelegatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected delegator address length is %d, actual length is %d", sdk.AddrLen, len(msg.DelegatorAddr)))
	}
	if len(msg.ValidatorAddr) != sdk.AddrLen {
		return sdk.ErrInvalidAddress(fmt.Sprintf("Expected validator address length is %d, actual length is %d", sdk.AddrLen, len(msg.ValidatorAddr)))
	}
	if msg.Amount.Amount <= 0 {
		return ErrBadDelegationAmount(DefaultCodespace, "undelegation amount must be positive")
	}
	if len(msg.SideChainId) == 0 || len(msg.SideChainId) > MaxSideChainIdLength {
		return sdk.NewError(DefaultCodespace, CodeInvalidInput, "side chain id must be included and max length is 20 bytes")
	}
	return nil
}

func (msg MsgSideChainUndelegate) GetInvolvedAddresses() []sdk.AccAddress {
	return []sdk.AccAddress{msg.DelegatorAddr, sdk.AccAddress(msg.ValidatorAddr)}
}

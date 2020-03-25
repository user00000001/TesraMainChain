// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/TesraSupernet/TesraMainChain/common"
	"github.com/TesraSupernet/TesraMainChain/common/hexutil"
)

func (r Receipt) MarshalJSON() ([]byte, error) {
	type Receipt struct {
		PostState         hexutil.Bytes  `json:"root"`
		Status            hexutil.Uint   `json:"status"`
		CumulativeGasUsed *hexutil.Big   `json:"cumulativeGasUsed" gencodec:"required"`
		Bloom             Bloom          `json:"logsBloom"         gencodec:"required"`
		Logs              []*Log         `json:"logs"              gencodec:"required"`
		TxHash            common.Hash    `json:"transactionHash" gencodec:"required"`
		ContractAddress   common.Address `json:"contractAddress"`
		GasUsed           *hexutil.Big   `json:"gasUsed" gencodec:"required"`
	}
	var enc Receipt
	enc.PostState = r.PostState
	enc.Status = hexutil.Uint(r.Status)
	enc.CumulativeGasUsed = (*hexutil.Big)(r.CumulativeGasUsed)
	enc.Bloom = r.Bloom
	enc.Logs = r.Logs
	enc.TxHash = r.TxHash
	enc.ContractAddress = r.ContractAddress
	enc.GasUsed = (*hexutil.Big)(r.GasUsed)
	return json.Marshal(&enc)
}

func (r *Receipt) UnmarshalJSON(input []byte) error {
	type Receipt struct {
		PostState         hexutil.Bytes   `json:"root"`
		Status            *hexutil.Uint   `json:"status"`
		CumulativeGasUsed *hexutil.Big    `json:"cumulativeGasUsed" gencodec:"required"`
		Bloom             *Bloom          `json:"logsBloom"         gencodec:"required"`
		Logs              []*Log          `json:"logs"              gencodec:"required"`
		TxHash            *common.Hash    `json:"transactionHash" gencodec:"required"`
		ContractAddress   *common.Address `json:"contractAddress"`
		GasUsed           *hexutil.Big    `json:"gasUsed" gencodec:"required"`
	}
	var dec Receipt
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.PostState != nil {
		r.PostState = dec.PostState
	}
	if dec.Status != nil {
		r.Status = uint(*dec.Status)
	}
	if dec.CumulativeGasUsed == nil {
		return errors.New("missing required field 'cumulativeGasUsed' for Receipt")
	}
	r.CumulativeGasUsed = (*big.Int)(dec.CumulativeGasUsed)
	if dec.Bloom == nil {
		return errors.New("missing required field 'logsBloom' for Receipt")
	}
	r.Bloom = *dec.Bloom
	if dec.Logs == nil {
		return errors.New("missing required field 'logs' for Receipt")
	}
	r.Logs = dec.Logs
	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionHash' for Receipt")
	}
	r.TxHash = *dec.TxHash
	if dec.ContractAddress != nil {
		r.ContractAddress = *dec.ContractAddress
	}
	if dec.GasUsed == nil {
		return errors.New("missing required field 'gasUsed' for Receipt")
	}
	r.GasUsed = (*big.Int)(dec.GasUsed)
	return nil
}

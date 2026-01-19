package hivego

import (
	"encoding/hex"
	"fmt"

	"github.com/decred/dcrd/dcrec/secp256k1/v2"
)

type HiveTransaction struct {
	RefBlockNum    uint16           `json:"ref_block_num"`
	RefBlockPrefix uint32           `json:"ref_block_prefix"`
	Expiration     string           `json:"expiration"`
	Operations     []HiveOperation  `json:"-"`
	OperationsJs   [][2]interface{} `json:"operations"`
	Extensions     []string         `json:"extensions"`
	Signatures     []string         `json:"signatures"`
	ChainID        string           `json:"-"`
}

func (t *HiveTransaction) GenerateTrxId() (string, error) {
	tB, err := SerializeTx(*t)
	if err != nil {
		return "", err
	}
	digest := HashTx(tB)

	return hex.EncodeToString(digest)[0:40], nil
}

func (t *HiveTransaction) Sign(keyPair KeyPair) (string, error) {
	message, err := SerializeTx(*t)

	if err != nil {
		return "", err
	}

	// Use custom chain ID if provided, otherwise use default
	if t.ChainID != "" {
		digest := HashTxForSig(message, t.ChainID)
		sig, err := secp256k1.SignCompact(keyPair.PrivateKey, digest, true)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(sig), nil
	} else {
		digest := HashTxForSig(message)
		sig, err := secp256k1.SignCompact(keyPair.PrivateKey, digest, true)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(sig), nil
	}
}

func (t *HiveTransaction) AddSig(sig string) {
	t.Signatures = append(t.Signatures, sig)
}

func (t *HiveTransaction) prepareJson() {
	var opsContainer [][2]interface{}
	for _, op := range t.Operations {
		var opContainer [2]interface{}
		opContainer[0] = op.OpName()
		opContainer[1] = op
		opsContainer = append(opsContainer, opContainer)
	}
	if t.Extensions == nil {
		t.Extensions = []string{}
	}
	t.OperationsJs = opsContainer
}

func (h *HiveRpcNode) Broadcast(ops []HiveOperation, wif *string) (string, error) {
	signingData, err := h.GetSigningData()
	if err != nil {
		return "", err
	}
	tx := HiveTransaction{
		RefBlockNum:    signingData.refBlockNum,
		RefBlockPrefix: signingData.refBlockPrefix,
		Expiration:     signingData.expiration,
		Operations:     ops,
	}

	// Set chain ID from HiveRpcNode if available
	if h.ChainID != "" {
		tx.ChainID = h.ChainID
	}

	message, err := SerializeTx(tx)

	if err != nil {
		return "", err
	}

	// Use custom chain ID for signing if provided
	var digest []byte
	if tx.ChainID != "" {
		digest = HashTxForSig(message, tx.ChainID)
	} else {
		digest = HashTxForSig(message)
	}

	txId, err := tx.GenerateTrxId()
	if err != nil {
		return "", err
	}
	sig, err := SignDigest(digest, wif)
	if err != nil {
		return "", err
	}

	tx.Signatures = append(tx.Signatures, hex.EncodeToString(sig))

	tx.prepareJson()

	var params []interface{}
	params = append(params, tx)
	if !h.NoBroadcast {
		q := hrpcQuery{"condenser_api.broadcast_transaction", params}
		res, err := h.rpcExec(q)
		if err != nil {
			return string(res), err
		}
	}

	return txId, nil
}

func (h *HiveRpcNode) BroadcastRaw(tx HiveTransaction) (string, error) {
	if len(tx.Signatures) == 0 {
		return "", fmt.Errorf("transaction is not signed")
	}

	tx.prepareJson()
	var params []interface{}
	params = append(params, tx)
	if !h.NoBroadcast {
		q := hrpcQuery{"condenser_api.broadcast_transaction", params}
		res, err := h.rpcExec(q)
		if err != nil {
			return string(res), err
		}
	}
	txId, err := tx.GenerateTrxId()
	if err != nil {
		return "", err
	}
	return txId, nil
}

package contract

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/D-Cous/go-web3"
	"github.com/D-Cous/go-web3/abi"
	"github.com/D-Cous/go-web3/jsonrpc"
	"github.com/D-Cous/go-web3/wallet"
)

// Contract is an Ethereum contract
type Contract struct {
	addr     web3.Address
	from     *web3.Address
	abi      *abi.ABI
	provider *jsonrpc.Client
}

// DeployContract deploys a contract
func DeployContract(provider *jsonrpc.Client, from web3.Address, abi *abi.ABI, bin []byte, args ...interface{}) *Txn {
	return &Txn{
		From:     from,
		Provider: provider,
		Method:   abi.Constructor,
		Args:     args,
		Bin:      bin,
	}
}

// NewContract creates a new contract instance
func NewContract(addr web3.Address, abi *abi.ABI, provider *jsonrpc.Client) *Contract {
	return &Contract{
		addr:     addr,
		abi:      abi,
		provider: provider,
	}
}

// ABI returns the abi of the contract
func (c *Contract) ABI() *abi.ABI {
	return c.abi
}

// Addr returns the address of the contract
func (c *Contract) Addr() web3.Address {
	return c.addr
}

// SetFrom sets the origin of the calls
func (c *Contract) SetFrom(addr web3.Address) {
	c.from = &addr
}

// EstimateGas estimates the gas for a contract call
func (c *Contract) EstimateGas(method string, args ...interface{}) (uint64, error) {
	return c.Txn(method, args).EstimateGas()
}

// Call calls a method in the contract
func (c *Contract) Call(method string, block web3.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	m, ok := c.abi.Methods[method]
	if !ok {
		return nil, fmt.Errorf("method %s not found", method)
	}

	// Encode input
	data, err := abi.Encode(args, m.Inputs)
	if err != nil {
		return nil, err
	}
	data = append(m.ID(), data...)

	// Call function
	msg := &web3.CallMsg{
		To:   &c.addr,
		Data: data,
	}
	if c.from != nil {
		msg.From = *c.from
	}

	rawStr, err := c.provider.Eth().Call(msg, block)
	if err != nil {
		return nil, err
	}

	// Decode output
	raw, err := hex.DecodeString(rawStr[2:])
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	respInterface, err := abi.Decode(m.Outputs, raw)
	if err != nil {
		return nil, err
	}

	resp := respInterface.(map[string]interface{})
	return resp, nil
}

// Txn creates a new transaction object
func (c *Contract) Txn(method string, args ...interface{}) *Txn {
	m, ok := c.abi.Methods[method]
	if !ok {
		// TODO, return error
		fmt.Println("Method not found")
		panic(fmt.Errorf("method %s not found", method))
	}

	return &Txn{
		Addr:     &c.addr,
		Provider: c.provider,
		Method:   m,
		Args:     args,
	}
}

// Txn is a transaction object
type Txn struct {
	From     web3.Address
	Addr     *web3.Address
	Provider *jsonrpc.Client
	Method   *abi.Method
	Args     []interface{}
	Data     []byte
	Bin      []byte
	GasLimit uint64
	GasPrice uint64
	Value    *big.Int
	Hash     web3.Hash
	Receipt  *web3.Receipt
}

func (t *Txn) isContractDeployment() bool {
	return t.Bin != nil
}

// AddArgs is used to set the arguments of the transaction
func (t *Txn) AddArgs(args ...interface{}) *Txn {
	t.Args = args
	return t
}

// SetValue sets the value for the txn
func (t *Txn) SetValue(v *big.Int) *Txn {
	t.Value = new(big.Int).Set(v)
	return t
}

// EstimateGas estimates the gas for the call
func (t *Txn) EstimateGas() (uint64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	return t.estimateGas()
}

func (t *Txn) estimateGas() (uint64, error) {
	if t.isContractDeployment() {
		return t.Provider.Eth().EstimateGasContract(t.Data)
	}

	msg := &web3.CallMsg{
		From:  t.From,
		To:    t.Addr,
		Data:  t.Data,
		Value: t.Value,
	}
	return t.Provider.Eth().EstimateGas(msg)
}

// SignSendAndWait is a blocking query that combines
// both SignAndSend and Wait functions
func (t *Txn) SignSendAndWait(key *wallet.Key, chainID uint64) error {
	if err := t.SignAndSend(key, chainID); err != nil {
		return err
	}
	if err := t.Wait(); err != nil {
		return err
	}
	return nil
}

func (t *Txn) ConvertToWeb3Transaction() (*web3.Transaction, error) {

	err := t.Validate()
	if err != nil {
		fmt.Println("Error when validating the transaction")
		return nil, err
	}

	// estimate gas price
	if t.GasPrice == 0 {
		t.GasPrice, err = t.Provider.Eth().GasPrice()
		if err != nil {
			fmt.Println("Error when getting the current gas price")
			return nil, err
		}
	}
	// estimate gas limit
	if t.GasLimit == 0 {
		t.GasLimit, err = t.estimateGas()
		if err != nil {
			fmt.Println("Error when estimating gas")
			return nil, err
		}
	}

	//get the nonce
	nonce, err := t.Provider.Eth().GetNonce(t.From, web3.Latest)
	if err != nil {
		fmt.Printf("Error retrieving nonce for %s\n", t.From)
		return nil, err
	}

	// send transaction
	txn := &web3.Transaction{
		From:     t.From,
		Input:    t.Data,
		GasPrice: t.GasPrice,
		Gas:      t.GasLimit,
		Value:    t.Value,
		Nonce:    nonce,
	}
	if t.Addr != nil {
		txn.To = t.Addr
	}

	return txn, nil
}

// Signs and sends the transaction to the network
func (t *Txn) SignAndSend(key *wallet.Key, chainID uint64) error {
	err := t.Validate()
	if err != nil {
		fmt.Println("Error when validating the transaction")
		return err
	}

	// estimate gas price
	if t.GasPrice == 0 {
		t.GasPrice, err = t.Provider.Eth().GasPrice()
		if err != nil {
			fmt.Println("Error when getting the current gas price")
			return err
		}
	}
	// estimate gas limit
	if t.GasLimit == 0 {
		t.GasLimit, err = t.estimateGas()
		if err != nil {
			fmt.Println("Error when estimating gas")
			return err
		}
	}

	//get the nonce
	nonce, err := t.Provider.Eth().GetNonce(t.From, web3.Latest)
	if err != nil {
		fmt.Printf("Error retrieving nonce for %s\n", t.From)
		return err
	}

	// send transaction
	txn := &web3.Transaction{
		From:     t.From,
		Input:    t.Data,
		GasPrice: t.GasPrice,
		Gas:      t.GasLimit,
		Value:    t.Value,
		Nonce:    nonce,
	}
	if t.Addr != nil {
		txn.To = t.Addr
	}

	// Create the signer object and sign the transaction
	signer := wallet.NewEIP155Signer(chainID)
	signedTxn, err := signer.SignTx(txn, key)
	if err != nil {
		return err
	}
	//send the transaction and return the transacation hash
	t.Hash, err = t.Provider.Eth().SendRawTransaction(signedTxn.MarshalRLP())

	if err != nil {
		return err
	}
	return nil
}

// Validate validates the arguments of the transaction
func (t *Txn) Validate() error {
	if t.Data != nil {
		// Already validated
		return nil
	}
	if t.isContractDeployment() {
		t.Data = append(t.Data, t.Bin...)
	}
	if t.Method != nil {
		data, err := abi.Encode(t.Args, t.Method.Inputs)
		if err != nil {
			return fmt.Errorf("failed to encode arguments: %v", err)
		}
		if !t.isContractDeployment() {
			t.Data = append(t.Method.ID(), data...)
		} else {
			t.Data = append(t.Data, data...)
		}
	}
	return nil
}

// SetGasPrice sets the gas price of the transaction
func (t *Txn) SetGasPrice(gasPrice uint64) *Txn {
	t.GasPrice = gasPrice
	return t
}

// SetGasLimit sets the gas limit of the transaction
func (t *Txn) SetGasLimit(gasLimit uint64) *Txn {
	t.GasLimit = gasLimit
	return t
}

// Wait waits till the transaction is mined
func (t *Txn) Wait() error {
	if (t.Hash == web3.Hash{}) {
		panic("transaction not executed")
	}

	var err error
	for {
		t.Receipt, err = t.Provider.Eth().GetTransactionReceipt(t.Hash)
		if err != nil {
			if err.Error() != "not found" {
				return err
			}
		}
		if t.Receipt != nil {
			break
		}
	}
	return nil
}

// Receipt returns the receipt of the transaction after wait
func (t *Txn) GetReceipt() *web3.Receipt {
	return t.Receipt
}

// Event is a solidity event
type Event struct {
	event *abi.Event
}

// Encode encodes an event
func (e *Event) Encode() web3.Hash {
	return e.event.ID()
}

// ParseLog parses a log
func (e *Event) ParseLog(log *web3.Log) (map[string]interface{}, error) {
	return abi.ParseLog(e.event.Inputs, log)
}

// Event returns a specific event
func (c *Contract) Event(name string) (*Event, bool) {
	event, ok := c.abi.Events[name]
	if !ok {
		return nil, false
	}
	return &Event{event}, true
}

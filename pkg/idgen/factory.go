package idgen

type IDFactory struct {
	wallet_id_gen      IDGenerator
	transaction_id_gen IDGenerator
	session_id_gen     IDGenerator
}

func NewIDFactory(idgen IDGenerator) *IDFactory {
	return &IDFactory{wallet_id_gen: idgen.Clone(), transaction_id_gen: idgen.Clone(), session_id_gen: idgen.Clone()}
}

func (f *IDFactory) GenWalletID() (string, error) {
	id, err := f.wallet_id_gen.GenerateString()
	if err != nil {
		return "", err
	}
	return "W" + id, nil
}

func (f *IDFactory) GenTransactionID() (string, error) {
	id, err := f.transaction_id_gen.GenerateString()
	if err != nil {
		return "", err
	}
	return "T" + id, nil
}

func (f *IDFactory) GenSessionID() (string, error) {
	id, err := f.session_id_gen.GenerateString()
	if err != nil {
		return "", err
	}
	return "S" + id, nil
}

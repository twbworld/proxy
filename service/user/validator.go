package user

type Validator struct{}

// 检验LoginPost参数
func (v *Validator) ValidatorSubscribe() error {
	return nil
}

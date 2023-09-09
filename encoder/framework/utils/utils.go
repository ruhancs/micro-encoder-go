package utils

import "encoding/json"
func IsJson(s string) error {
	var js struct{}

	//converte s em json e preenche em js
	if err := json.Unmarshal([]byte(s), &js); err != nil {
		return err
	}
	return nil
}
package writer

import (
	"encoding/json"
	"os"
)

func MarshallAndSave(fileName string, val any) error {
	json, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, json, 0644)
	return err
}

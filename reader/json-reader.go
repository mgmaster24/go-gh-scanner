package reader

import (
	"encoding/json"
	"io"
	"os"
)

// ReadJSNOData is a generic method that will read JSON data into the
// provided generic type instance
func ReadJSONData[T any](fileName string) (T, error) {
	var val T
	configFile, err := os.Open(fileName)
	if err != nil {
		return val, err
	}

	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		return val, err
	}

	err = json.Unmarshal(bytes, &val)
	if err != nil {
		return val, err
	}

	return val, nil
}

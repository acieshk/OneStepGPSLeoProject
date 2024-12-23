package common

import (
	"encoding/json"
	"io/ioutil"
)

func ReadDevicesFromJSON(filename string) ([]map[string]interface{}, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var response struct {
		ResultList []map[string]interface{} `json:"result_list"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return response.ResultList, nil
}

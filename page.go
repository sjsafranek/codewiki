package main

import (
	"encoding/json"
)

type Page struct {
	Title           string
	Data            string
	CurrentVersion  int
	SelectedVersion int
}

func (self *Page) Unmarshal() (string, error) {
	if "" == self.Data {
		return "", nil
	}
	wrapper := make(map[string]string)
	err := json.Unmarshal([]byte(self.Data), &wrapper)
	return wrapper["content"], err
}

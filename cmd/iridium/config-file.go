package main

import (
	"os"
	"path"
	"encoding/json"
	"io/ioutil"
)

type GameConfig struct {
	Title string
	Subtitle string
	Author string
	InitialPassage string `json:"init"`
}

func readConfig(srcdir string) (GameConfig, error) {
	jsonFile, err := os.Open(path.Join(srcdir, SRC_JSON))
	if err != nil {
		return GameConfig{}, err
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config GameConfig
	json.Unmarshal(byteValue, &config)
	//fmt.Println("json = ", config)
	return config, nil
}

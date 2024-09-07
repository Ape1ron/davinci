package mongo

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/go/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

type FindOneAction struct {
	Filter bson.M
	Opts   []*options.FindOneOptions
}

func parseFindOneActions(cmd string) *FindOneAction {
	filter, optsArray, err := parseFilterAndOpts(cmd)
	if err != nil {
		panic(err)
	}
	var findOneOptions []*options.FindOneOptions
	for _, opt := range optsArray {
		option, err := parseFindOneOptions(opt)
		if err != nil {
			panic(err)
		}
		findOneOptions = append(findOneOptions, option)
	}
	return &FindOneAction{
		Filter: filter,
		Opts:   findOneOptions,
	}
}

func parseFindOneOptions(option string) (*options.FindOneOptions, error) {
	optionName, optionDesc, err := parseOpNameAndDesc(option)
	if err != nil {
		return nil, err
	}
	switch optionName {
	case Sort:
		if optionDesc == "" {
			return options.FindOne(), nil
		}
		var sortDesc bson.M
		if err := json.Unmarshal([]byte(optionDesc), &sortDesc); err != nil {
			return nil, err
		}
		return options.FindOne().SetSort(sortDesc), nil
	}
	return nil, fmt.Errorf("can not find the FindOneOption: %s, available options: [%s]", optionName,
		strings.Join([]string{Sort}, ","))
}

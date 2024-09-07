package mongo

import (
	"davinci/common"
	"encoding/json"
	"fmt"
	"github.com/siddontang/go/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"strings"
)

type FindAction struct {
	Filter bson.M
	Opts   []*options.FindOptions
}

type SizeAction struct {
	*FindAction
}

func parseFindActions(cmd string) interface{} {
	filter, optsArray, err := parseFilterAndOpts(cmd)
	if err != nil {
		panic(err)
	}
	var findOptions []*options.FindOptions
	var size = false
	for i := 0; i < len(optsArray); i++ {
		opt := optsArray[i]
		option, err := parseFindOptions(opt)
		if err != nil {
			panic(err)
		}
		if option.Comment != nil && *option.Comment == Size {
			if i != len(optsArray)-1 {
				return fmt.Errorf("size() function can only be used at the end")
			}
			size = true
			break
		}
		findOptions = append(findOptions, option)
	}
	if size {
		return &SizeAction{
			FindAction: &FindAction{
				Filter: filter,
				Opts:   findOptions,
			},
		}
	}
	return &FindAction{
		Filter: filter,
		Opts:   findOptions,
	}
}

func parseFilterAndOpts(cmd string) (bson.M, []string, error) {
	splitOptsIndex, err := findCloseBrackets(cmd)
	if err != nil {
		return nil, nil, err
	}
	filterDesc := cmd[1:splitOptsIndex]
	var optsArray []string

	if splitOptsIndex < len(cmd)-2 {
		optsString := cmd[splitOptsIndex+1:]
		if optsString[0] != '.' {
			return nil, nil, fmt.Errorf("invalid character: " + optsString)
		}
		optsArray = common.SplitCmd(optsString[1:], '.')
	}

	var filter = bson.M{}
	if filterDesc != "" {
		err := json.Unmarshal([]byte(filterDesc), &filter)
		if err != nil {
			return nil, nil, err
		}
	}

	return filter, optsArray, nil
}

func parseFindOptions(option string) (*options.FindOptions, error) {
	optionName, optionDesc, err := parseOpNameAndDesc(option)
	if err != nil {
		return nil, err
	}
	switch optionName {
	case Limit:
		if limit, err := strconv.ParseInt(optionDesc, 10, 64); err != nil {
			return nil, err
		} else {
			return options.Find().SetLimit(limit), nil
		}
	case Sort:
		if optionDesc == "" {
			return options.Find(), nil
		}
		var sortDesc bson.M
		if err := json.Unmarshal([]byte(optionDesc), &sortDesc); err != nil {
			return nil, err
		}
		return options.Find().SetSort(sortDesc), nil
	case Size:
		return options.Find().SetComment("size"), nil
	}
	return nil, fmt.Errorf("can not find the FindOption: %s, available options: [%s]", optionName,
		strings.Join([]string{Sort, Limit, Size}, ","))
}

func parseOpNameAndDesc(option string) (string, string, error) {
	startFuncIndex := strings.Index(option, "(")
	endFuncIndex := strings.LastIndex(option, ")")
	if endFuncIndex < startFuncIndex || startFuncIndex < 0 || endFuncIndex < 0 || endFuncIndex != len(option)-1 {
		return "", "", fmt.Errorf("invalid character: " + option)
	}
	optionName := option[:startFuncIndex]
	optionDesc := option[startFuncIndex+1 : endFuncIndex]
	return optionName, optionDesc, nil
}

/*
 * Copyright 2022 New Relic Corporation. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"errors"
	"io/ioutil"
	"regexp"

	"github.com/newrelic/infra-integrations-sdk/v3/log"

	"gopkg.in/yaml.v3"

	"github.com/newrelic/infra-integrations-sdk/v3/data/inventory"
)

var (
	errNoInventoryData = errors.New("config path not correctly set, cannot fetch inventory data")
)

func getInventory() (map[string]interface{}, error) {
	rawYamlFile, err := ioutil.ReadFile(args.ConfigPath)
	if err != nil {
		return nil, err
	}

	i := make(inventory.Item)
	err = yaml.Unmarshal(rawYamlFile, &i)
	if err != nil {
		return nil, err
	}

	if len(i) == 0 {
		return nil, errNoInventoryData
	}

	return i, nil
}

func populateInventory(i *inventory.Inventory, rawInventory inventory.Item) error {
	for k, v := range rawInventory {
		switch value := v.(type) {
		case map[interface{}]interface{}:
			for subk, subv := range value {
				switch subVal := subv.(type) {
				case []interface{}:
					//TODO: Do not include lists for now
				default:
					setValue(i, k, subk.(string), subVal)
				}
			}
		case []interface{}:
			//TODO: Do not include lists for now
		default:
			setValue(i, k, "value", value)
		}
	}
	return nil
}

func setValue(i *inventory.Inventory, key string, field string, value interface{}) {
	re, _ := regexp.Compile("(?i)password")

	if re.MatchString(key) || re.MatchString(field) {
		value = "(omitted value)"
	}
	err := i.SetItem(key, field, value)
	if err != nil {
		log.Error("setting item: %v", err)
	}
}

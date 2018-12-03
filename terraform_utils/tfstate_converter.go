package terraform_utils

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform/flatmap"
	"github.com/hashicorp/terraform/terraform"
)

type TfstateConverter struct{}

func (c TfstateConverter) Convert(pathToTfstate string, metadata map[string]ResourceMetaData) ([]TerraformResource, error) {
	resources := []TerraformResource{}
	// read full tfstate file
	data, err := ioutil.ReadFile(pathToTfstate)
	if err != nil {
		return resources, err
	}
	tfState := terraform.State{}
	// parse json to tfstate struct from terraform code
	err = json.Unmarshal(data, &tfState)
	if err != nil {
		return resources, err
	}
	for _, module := range tfState.Modules {
		for key, resource := range module.Resources {
			item := map[string]interface{}{}
			allAttributes := []string{}
			for key := range resource.Primary.Attributes {
				allAttributes = append(allAttributes, key)
			}
			// delete empty array
			for _, key := range allAttributes {
				if strings.HasSuffix(key, ".#") && resource.Primary.Attributes[key] == "0" {
					delete(resource.Primary.Attributes, key)
				}
			}
			// delete ignored keys by regex patter
			for keyAttribute := range resource.Primary.Attributes {
				for patter := range metadata[resource.Primary.ID].IgnoreKeys {
					match, err := regexp.MatchString(patter, keyAttribute)
					if match && err == nil {
						delete(resource.Primary.Attributes, keyAttribute)
					}
				}
			}
			// delete empty keys with empty value, but not from AllowEmptyValue list
			for keyAttribute, value := range resource.Primary.Attributes {
				if value != "" {
					continue
				}
				allowEmptyValue := false
				for patter := range metadata[resource.Primary.ID].AllowEmptyValue {
					match, err := regexp.MatchString(patter, keyAttribute)
					if match && err == nil {
						allowEmptyValue = true
					}
				}
				if !allowEmptyValue {
					delete(resource.Primary.Attributes, keyAttribute)
				}
			}
			// parse Attributes to go string with flatmap package
			for key := range resource.Primary.Attributes {
				blockName := strings.Split(key, ".")[0]

				if _, exist := item[blockName]; exist {
					continue
				}

				item[blockName] = flatmap.Expand(resource.Primary.Attributes, blockName)
			}
			// add Additional Fields to resource
			for key, value := range metadata[resource.Primary.ID].AdditionalFields {
				item[key] = value
			}
			resources = append(resources, TerraformResource{
				ResourceType: strings.Split(key, ".")[0],
				ResourceName: strings.Split(key, ".")[1],
				Item:         item,
				ID:           resource.Primary.ID,
				Provider:     metadata[resource.Primary.ID].Provider,
			})
		}
	}
	return resources, nil
}

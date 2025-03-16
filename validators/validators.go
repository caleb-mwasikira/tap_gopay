package validators

import (
	"fmt"
	"log"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func validateStruct(obj interface{}) map[string]string {
	obj_value := reflect.ValueOf(obj)
	obj_type := reflect.TypeOf(obj)
	errs := map[string]string{}

	num_fields := obj_value.NumField()
	for i := 0; i < num_fields; i++ {
		field := obj_type.Field(i)
		value := obj_value.Field(i)

		validateTag := field.Tag.Get("validate")
		jsonTag := field.Tag.Get("json")

		if validateTag == "" {
			// no custom validateTag, skip validation
			continue
		}

		fieldName := field.Name
		if jsonTag != "" {
			fieldName = jsonTag
		}

		// split validateTag values
		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			switch {
			case rule == "required":
				if value.Kind() == reflect.String && strings.Trim(value.String(), " ") == "" {
					errs[fieldName] = fmt.Sprintf("%v is required", fieldName)
				}

			case strings.HasPrefix(rule, "min="):
				min_value, err := strconv.Atoi(strings.TrimPrefix(rule, "min="))
				if err != nil {
					log.Fatalf("invalid min= value on struct tag for %v\n", fieldName)
				}

				if value.Kind() == reflect.Int && value.Int() < int64(min_value) {
					errs[fieldName] = fmt.Sprintf("%v must be at least %v", fieldName, min_value)
				}

				if value.Kind() == reflect.Float64 && value.Float() < float64(min_value) {
					errs[fieldName] = fmt.Sprintf("%v must be at least %v", fieldName, min_value)
				}

				if value.Kind() == reflect.String && len(strings.Trim(value.String(), " ")) < min_value {
					errs[fieldName] = fmt.Sprintf("%v must be at least %v characters long", fieldName, min_value)
				}

			case strings.HasPrefix(rule, "max="):
				max_value, err := strconv.Atoi(strings.TrimPrefix(rule, "max="))
				if err != nil {
					log.Fatalf("invalid max= value on struct tag for %v\n", fieldName)
				}

				if value.Kind() == reflect.Int && value.Int() > int64(max_value) {
					errs[fieldName] = fmt.Sprintf("%v must be at most %v", fieldName, max_value)
				}

				if value.Kind() == reflect.Float64 && value.Float() > float64(max_value) {
					errs[fieldName] = fmt.Sprintf("%v must be at most %v", fieldName, max_value)
				}

				if value.Kind() == reflect.String && len(strings.Trim(value.String(), " ")) > max_value {
					errs[fieldName] = fmt.Sprintf("%v must be at most %v characters long", fieldName, max_value)
				}

			case rule == "email":
				if value.Kind() != reflect.String {
					errs[fieldName] = fmt.Sprintf("%v must be a string", fieldName)
					continue
				}

				err := validateEmail(value.String())
				if err != nil {
					errs[fieldName] = err.Error()
				}

			case rule == "password":
				if value.Kind() != reflect.String {
					errs[fieldName] = fmt.Sprintf("%v must be a string", fieldName)
					continue
				}

				err := validatePassword(value.String(), false)
				if err != nil {
					errs[fieldName] = err.Error()
				}

			case rule == "account_type":
				if value.Kind() != reflect.String {
					errs[fieldName] = fmt.Sprintf("%v must be a string", fieldName)
					continue
				}

				validAccountTypes := []string{"user", "agent", "admin"}
				if !slices.Contains(validAccountTypes, value.String()) {
					errs[fieldName] = "Invalid account type. Valid account types include: ['user','agent','admin']"
				}
			}
		}
	}

	return errs
}

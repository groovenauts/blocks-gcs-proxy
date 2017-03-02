package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type (
	InvalidExpression struct{
		msg string
	}
)

func (e *InvalidExpression) Error() string {
	return e.msg
}

type (
	Variable struct {
		data map[string]interface{}
		// quoteString boolean
		separator string
	}
)

const (
	DefaultExpandedArraySeparator = "[[GCSPROXY:SEP]]"
)

func (v *Variable) expand(str string) (string, error) {
	if v.separator == "" {
		v.separator = DefaultExpandedArraySeparator
	}
	re0 := regexp.MustCompile(`\%\{\s*([\w.]+)\s*\}`)
	re1 := regexp.MustCompile(`\A\%\{\s*`)
	re2 := regexp.MustCompile(`\s*\}\z`)
	errors := []error{}
	res := re0.ReplaceAllStringFunc(str, func(raw string) string {
		expr := re1.ReplaceAllString(re2.ReplaceAllString(raw, ""), "")
		value, err := v.dive(expr)
		if err != nil {
			errors = append(errors, err)
			return ""
		}
		if value == nil {
			errors = append(errors, &InvalidExpression{"No value found for " + expr})
		}
		switch value.(type) {
		case string:
			return value.(string)
		case []interface{}:
			return v.flatten(value)
		case map[string]interface{}:
			return v.flatten(value)
		case map[string]string:
			return v.flatten(value)
		default:
			return fmt.Sprintf("%v", value)
		}
	})
	if len(errors) > 0 {
		return "", &CompositeError{errors}
	}
	return res, nil
}

func (v *Variable) dive(expr string) (interface{}, error) {
	var_names := strings.Split(expr, expr_separator)
	// fmt.Printf("var_names: %v\n", var_names)
	res, err := v.inject(var_names, v.data, func(tmp interface{}, name string) (interface{}, error) {
		res, err := v.dig(tmp, name, expr)
		// fmt.Printf("res: %v err: %v\n", res, err)
		if err != nil {
			return nil, err
		}
		return res, nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *Variable) dig(tmp interface{}, name, expr string) (interface{}, error) {
	result, err := v.digIn(tmp, name, expr)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("name: %v, result: %v\n", name, result)
	switch result.(type) {
	case string:
		re := regexp.MustCompile(`\A\{.*\}\z|\A\[.*\]\z`)
		matched := re.MatchString(result.(string))
		if matched {
			var obj interface{}
			src := []byte(result.(string))
			err := json.Unmarshal(src, &obj)
			if err != nil {
				return result.(string), nil
			}
			return obj, nil
		} else {
			return result, nil
		}
	default:
		return result, nil
	}
}

func (v *Variable) digIn(tmp interface{}, name, expr string) (interface{}, error) {
	switch tmp.(type) {
	case []string, []interface{}:
		return v.getFromArray(tmp, name)
	case map[string]interface{}, map[string]string:
		return v.getFromMap(tmp, name)
	default:
		return nil, fmt.Errorf("Unsupported Object type: [%T]%v", tmp, tmp)
	}
}

func (v *Variable) getFromArray(array interface{}, name string) (interface{}, error) {
	idx, err := v.parseIndex(name)
	if err != nil {
		return nil, &InvalidExpression{ "Invalid index for array: " + name }
	}
	switch array.(type) {
	case []string:
		l := len(array.([]string))
		if idx >= l {
			return nil, &InvalidExpression{ fmt.Sprintf("Invalid index %v for array who has %v items", idx, l) }
		}
		return array.([]string)[idx], nil
	case []interface{}:
		l := len(array.([]interface{}))
		if idx >= l {
			return nil, &InvalidExpression{ fmt.Sprintf("Invalid index %v for array who has %v items", idx, l) }
		}
		return array.([]interface{})[idx], nil
	default:
		return nil, fmt.Errorf("Unsupported object given as an array: [%T]%v", array, array)
	}
}

func (v *Variable) getFromMap(obj interface{}, name string) (interface{}, error) {
	switch obj.(type) {
	case map[string]interface{}:
		m := obj.(map[string]interface{})
		v, ok := m[name]
		if !ok {
			return nil, &InvalidExpression{ fmt.Sprintf("Invalid key %v for map %v", name, m) }
		}
		return v, nil
	case map[string]string:
		m := obj.(map[string]string)
		v, ok := m[name]
		if !ok {
			return nil, &InvalidExpression{ fmt.Sprintf("Invalid key %v for map %v", name, m) }
		}
		return v, nil
	default:
		return nil, fmt.Errorf("Unsupported object given as a map: [%T]%v", obj, obj)
	}
}


func (v *Variable) parseIndex(str string) (int, error) {
	re := regexp.MustCompile(`\A\d+\z`)
	matched := re.MatchString(str)
	if !matched {
		return 0, fmt.Errorf("Invalid Reference: %v", str)
	}
	idx, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("Error to Atoi(%v): %v\n", str, err)
		return 0, err
	}
	return idx, nil
}

func (v *Variable) inject(var_names []string, tmp interface{}, f func(interface{}, string) (interface{}, error)) (interface{}, error) {
	name := var_names[0]
	rest := var_names[1:]
	res, err := f(tmp, name)
	if err != nil {
		return nil, err
	}
	if len(rest) == 0 {
		return res, nil
	} else {
		return v.inject(rest, res, f)
	}
}

const (
	expr_separator = "."
)

func (v *Variable) flatten(obj interface{}) string {
	switch obj.(type) {
	case string:
		return obj.(string)
	case []string:
		return strings.Join(obj.([]string), v.separator)
	case []interface{}:
		res := []string{}
		for _, i := range obj.([]interface{}) {
			res = append(res, v.flatten(i))
		}
		return strings.Join(res, v.separator)
	case map[string]interface{}:
		res := []string{}
		for _, i := range obj.(map[string]interface{}) {
			res = append(res, v.flatten(i))
		}
		return strings.Join(res, v.separator)
	case map[string]string:
		res := []string{}
		for _, i := range obj.(map[string]string) {
			res = append(res, v.flatten(i))
		}
		return strings.Join(res, v.separator)
	default:
		return fmt.Sprintf("%v", obj)
	}
}

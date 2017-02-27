package gcsproxy

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type (
	Variable struct {
		data map[string]interface{}
		// quoteString boolean
	}
)

func (v *Variable)expand(str string) (string, error) {
	re0 := regexp.MustCompile(`\%\{\s*([\w.]+)\s*\}`)
	re1 := regexp.MustCompile(`\A\%\{\s*`)
	re2 := regexp.MustCompile(`\s*\}\z`)
	res := re0.ReplaceAllStringFunc("A/%{  foo  }/B/%{bar}/D", func(raw string) string{
		expr := re1.ReplaceAllString(re2.ReplaceAllString(raw, ""), "")
		value, err := v.dive(expr)
		if err != nil {
			// return err
			value = ""
		}
		switch value.(type) {
		case string: return value.(string)
		case []interface{}:
			return v.flatten(value)
		case map[string]interface{}:
			return v.flatten(value)
		default:
			return fmt.Sprintf("%v", value)
		}
	})
	return res, nil
}

func (v *Variable)dive(expr string) (interface{}, error) {
	var_names := strings.Split(expr, expr_separator)
	res, err := v.inject(var_names, v.data, func(tmp interface{}, name string) (interface{}, error) {
		res, err := v.dig(tmp, name, expr)
		if err != nil { return nil, err }
		return res, nil
	})
	if err != nil { return nil, err }
	return res, nil
}

func (v *Variable)dig(tmp interface{}, name, expr string) (interface{}, error) {
	matched, err := regexp.MatchString(`\A\d+\z`, name)
	if err != nil {
		return nil, err
	}
	if matched {
		idx, err := strconv.Atoi(name)
		if err != nil { return nil, err }
		switch tmp.(type) {
		case []string:
			return tmp.([]string)[idx], nil
		case []interface{}:
			return tmp.([]interface{})[idx], nil
		case map[string]interface{}:
			return tmp.(map[string]interface{})[name], nil
		}
	} else {
		switch tmp.(type) {
		case map[string]interface{}:
			return tmp.(map[string]interface{})[name], nil
		}
	}
	return nil, fmt.Errorf("Invalid Reference")
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
	variable_separator = " "
)

func (v *Variable)flatten(obj interface{}) string {
	switch obj.(type) {
	case string:
		return obj.(string)
	case []string:
		return strings.Join(obj.([]string), variable_separator)
	case []interface{}:
		res := []string{}
		for _, i := range obj.([]interface{}) {
			res = append(res, v.flatten(i))
		}
		return strings.Join(res, variable_separator)
	case map[string]interface{}:
		res := []string{}
		for _, i := range obj.(map[string]interface{}) {
			res = append(res, v.flatten(i))
		}
		return strings.Join(res, variable_separator)
	default:
		return fmt.Sprintf("%v", obj)
	}
}

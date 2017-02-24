package gcsproxy

import (
	"fmt"
	"regexp"
	"strconv"
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
		value, err := dig_variables(expr)
		if err != nil {
			return err
		}
		switch value.(type) {
		case string: return value
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

func (v *Variable)dig_variables(expr string) interface{} {
	var_names := strings.Split(expr, expr_separator)
	return v.inject(var_names, v.data, func(tmp interface{}, name string) interface{}{
		return v.dig_variable(tmp, name, expr)
	})
}

func (v *Variable)dig_variable(tmp interface{}, name, expr string) (interface{}, error) {
	if regexp.MatchString(`\A\d+\z`, name) {
		idx := strconv.Atoi(name)
		switch tmp.(type) {
		case []string:
			return tmp.([]string)[idx]
		case []interface{}:
			return tmp.([]interface{})[idx]
		case map[string]interface{}:
			return tmp.(map[string]interface{})[name]
		}
	} else {
		switch tmp.(type) {
		case map[string]interface{}:
			return tmp.(map[string]interface{})[name]
		}
	}
	retur nil, fmt.Errorf("Invalid Reference")
}


func (v *Variable) inject(var_names []string, tmp interface{}, f func(interface{}, name string) interface{}) interface{} {
	name := var_names[0]
	rest := var_names[1:]
	res := f(tmp, name)
	if len(rest) == 0 {
		return res
	} else {
		return inject(rest, res, f)
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
		res := []string
		for _, i := range obj.([]interface{}) {
			res = append(res, v.flatten(i))
		}
		return strings.Join(res, variable_separator)
	case map[string]interface{}:
		res := []string
		for _, i := range obj.(map[string]interface{}) {
			res = append(res, v.flatten(i))
		}
		return strings.Join(res, variable_separator)
	default:
		return fmt.Sprintf("%v", obj)
	}
}

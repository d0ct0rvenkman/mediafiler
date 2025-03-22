package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/multierr"
)

type PathIgnorePattern struct {
	Type    string
	Pattern string
}

func (p PathIgnorePattern) IsValid() (bool, error) {
	var err error
	var merr error
	var valid = true

	err = nil
	merr = nil
	switch p.Type {
	case `string`:
		err = nil
	case `regex`:
		_, err = regexp.Compile(p.Pattern)
		if err != nil {
			err = errors.New("value for Pattern is not a valid regex")
			valid = false
		}
	default:
		err = fmt.Errorf("invalid Type '%s'", p.Type)
	}

	if err != nil {
		merr = multierr.Append(merr, err)
		valid = false
	}

	// the thing we're searching for shouldn't be empty
	if len(p.Pattern) == 0 {
		merr = multierr.Append(merr, errors.New("the Pattern cannot be empty"))
		valid = false
	}

	return valid, merr

}

type PathIgnoreFilter []PathIgnorePattern

func (p *PathIgnoreFilter) AddPattern(newPattern PathIgnorePattern) error {
	valid, err := newPattern.IsValid()
	if !valid {
		return fmt.Errorf("new pattern is not valid. reason: %s", err)
	} else {
		*p = append(*p, newPattern)
		return nil
	}
}

func (p *PathIgnoreFilter) IsPathFiltered(path string) (bool, error) {
	for _, v := range *p {
		switch v.Type {
		case "string":
			if strings.Contains(path, v.Pattern) {
				return true, nil
			}
		case "regex":
			rx := regexp.MustCompile(v.Pattern)
			if rx.MatchString(path) {
				return true, nil
			}
		default:
			return false, errors.New("non-sensical Type found while executing IsPathFiltered()")
		}
	}

	return false, nil
}

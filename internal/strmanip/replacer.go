package strmanip

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	multierr "github.com/hashicorp/go-multierror"
)

type ReplacerRule struct {
	Type        string
	Find        string
	ReplaceWith string
}

func (rr ReplacerRule) IsValid() (bool, error) {
	var err error
	var merr error
	var valid = true

	err = nil
	merr = nil

	switch rr.Type {
	case "regex":
		_, err = regexp.Compile(rr.Find)
		if err != nil {
			err = errors.New("value Find is not a valid regex")
			valid = false
		}
	case "string":
		err = nil
	default:
		err = fmt.Errorf("invalid ReplacerRule type '%s'. valid types are 'regex' or 'string'", rr.Type)
	}

	if err != nil {
		merr = multierr.Append(merr, err)
		valid = false
	}

	// the thing we're searching for shouldn't be empty
	if len(rr.Find) == 0 {
		merr = multierr.Append(merr, errors.New("the Find term cannot be empty"))
		valid = false
	}

	return valid, merr

}

type Replacer []ReplacerRule

func (r *Replacer) AddRule(Rule ReplacerRule) error {
	valid, err := Rule.IsValid()
	if !valid {
		return fmt.Errorf("new replacer rule is not valid. reason: %s", err)
	} else {
		*r = append(*r, Rule)
		return nil
	}
}

func (r Replacer) Replace(input string) (string, error) {
	output := input

	// fmt.Printf("before modification: '%s'\n", input)

	for _, v := range r {
		switch v.Type {
		case "string":
			output = strings.ReplaceAll(output, v.Find, v.ReplaceWith)
		case "regex":
			rx := regexp.MustCompile(v.Find)
			output = rx.ReplaceAllString(output, v.ReplaceWith)
		default:
			// this *should* get caught by AddRule/IsValid, but a sufficiently motivated individual could get around these
			return "", fmt.Errorf("rule type '%s' is non-sensical", v.Type)
		}

		// fmt.Printf("after modification #%d: '%s'\n", k, output)
	}

	return output, nil
}

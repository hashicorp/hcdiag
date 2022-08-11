package redact

import (
	"encoding/json"
)

type RedactedString struct {
	inputString string
	redactions  []*Redact

	redactedString string
}

func NewRedactedString(input string, redactions []*Redact) RedactedString {
	return RedactedString{
		inputString: input,
		redactions:  redactions,
	}
}

func NewRedactedStringSlice(inSlice []string, r []*Redact) []RedactedString {
	var outSlice []RedactedString
	for _, in := range inSlice {
		outSlice = append(outSlice, NewRedactedString(in, r))
	}
	return outSlice
}

func (r *RedactedString) String() string {
	red, err := String(r.inputString, r.redactions)
	if err != nil {
		// TODO (nwchandler): I'm not sure if this is the best way to handle this... Seems like we don't want to blow the whole process up, but we should err on the side of not returning anything...
		return ""
	}
	return red
}

func (r *RedactedString) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

package redact

func GetDefaultAgentRedactions() ([]*Redact, error) {
	var defaultAgentRedactions = make([]*Redact, 0)

	redactions := []struct {
		name    string
		matcher string
		replace string
	}{
		{
			name:    "regex",
			matcher: "/myRegex/",
			replace: "agentdefault-redaction",
		},
		{
			name:    "redacts once",
			matcher: "myRegex",
			replace: "agentdefault-redaction",
		},
		{
			name:    "redacts something",
			matcher: "something",
			replace: "agentdefault-redaction",
		},
	}
	for _, redaction := range redactions {
		redact, err := New(redaction.matcher, redaction.name, redaction.replace)
		if err != nil {
			// If there's an issue, return an empty slice so that we can just ignore these redactions
			return make([]*Redact, 0), err
		}
		defaultAgentRedactions = append(defaultAgentRedactions, redact)
	}
	return defaultAgentRedactions, nil
}

package redact

func GetDefaultAgentRedactions() ([]*Redact, error) {
	var defaultAgentRedactions = make([]*Redact, 0)

	redactions := []struct {
		name    string
		matcher string
		replace string
	}{
		{
			name:    "empty input",
			matcher: "/myRegex/",
			replace: "agentdefault-test-replace",
		},
		{
			name:    "redacts once",
			matcher: "myRegex",
			replace: "agentdefault-test-replace",
		},
		{
			name:    "redacts many",
			matcher: "test",
		},
	}
	for _, redaction := range redactions {
		redact, err := New(redaction.matcher, "", redaction.replace)
		if err != nil {
			// If there's an issue, return an empty slice so that we can just ignore agent redactions
			return make([]*Redact, 0), err
		}
		defaultAgentRedactions = append(defaultAgentRedactions, redact)
	}
	return defaultAgentRedactions, nil
}

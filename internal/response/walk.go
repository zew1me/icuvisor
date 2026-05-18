package response

import "fmt"

type jsonWalkContainer int

const (
	walkRoot jsonWalkContainer = iota
	walkMapValue
	walkSliceValue
)

type jsonWalkDecision struct {
	Drop    bool
	Stop    bool
	Missing []string
}

type jsonWalkVisitor func(path string, value any, container jsonWalkContainer) jsonWalkDecision

func walkJSON(value any, path string, container jsonWalkContainer, visitor jsonWalkVisitor) (any, []string, bool) {
	decision := visitor(path, value, container)
	if decision.Drop {
		return nil, decision.Missing, true
	}
	if decision.Stop {
		return value, decision.Missing, false
	}
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		missing := append([]string(nil), decision.Missing...)
		for key, item := range typed {
			itemPath := joinPath(path, key)
			walked, nestedMissing, dropped := walkJSON(item, itemPath, walkMapValue, visitor)
			missing = append(missing, nestedMissing...)
			if dropped {
				continue
			}
			out[key] = walked
		}
		return out, missing, false
	case []any:
		out := make([]any, 0, len(typed))
		missing := append([]string(nil), decision.Missing...)
		for i, item := range typed {
			walked, nestedMissing, dropped := walkJSON(item, indexPath(path, i), walkSliceValue, visitor)
			missing = append(missing, nestedMissing...)
			if dropped {
				out = append(out, nil)
				continue
			}
			out = append(out, walked)
		}
		return out, missing, false
	default:
		return value, decision.Missing, false
	}
}

func joinPath(base string, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

func indexPath(base string, index int) string {
	if base == "" {
		return fmt.Sprintf("[%d]", index)
	}
	return fmt.Sprintf("%s[%d]", base, index)
}

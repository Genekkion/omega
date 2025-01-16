package structs

type Language int

const (
	LanguageNone Language = iota
	LanguageGo   Language = iota
)

var (
	LanguageMap = map[Language]string{
		LanguageNone: "None",
		LanguageGo:   "Go",
	}

	LanguageMapInverted = func() map[string]Language {
		m := map[string]Language{}
		for k, v := range LanguageMap {
			m[v] = k
		}
		return m
	}()
)

// Returns empty string if not found
func (l Language) ToString() string {
	str, exists := LanguageMap[l]
	if !exists {
		return ""
	}
	return str
}

// Returns -1 if not found
func ParseLanguageString(str string) Language {
	l, exists := LanguageMapInverted[str]
	if !exists {
		return -1
	}
	return l
}

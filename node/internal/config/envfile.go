package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EnvFile struct {
	lines []envLine
	index map[string]int
}

type envLine struct {
	key   string
	value string
	raw   string
	kind  envLineKind
}

type envLineKind int

const (
	envLineRaw envLineKind = iota
	envLineKeyValue
)

func ParseEnvFile(data []byte) (*EnvFile, error) {
	result := &EnvFile{index: map[string]int{}}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if text == "" {
		return result, nil
	}
	parts := strings.Split(text, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	for _, raw := range parts {
		line := envLine{raw: raw, kind: envLineRaw}
		trimmed := strings.TrimSpace(raw)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			key, value, ok := strings.Cut(raw, "=")
			if ok {
				line.kind = envLineKeyValue
				line.key = strings.TrimSpace(key)
				line.value = unquoteEnvValue(strings.TrimSpace(value))
				result.index[line.key] = len(result.lines)
			}
		}
		result.lines = append(result.lines, line)
	}
	return result, nil
}

func LoadEnvFile(path string) (*EnvFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ParseEnvFile(nil)
	}
	if err != nil {
		return nil, err
	}
	return ParseEnvFile(data)
}

func SaveEnvFileAtomic(path string, file *EnvFile) error {
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, file.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (f *EnvFile) Get(key string) string {
	if f == nil || f.index == nil {
		return ""
	}
	if idx, ok := f.index[key]; ok {
		return f.lines[idx].value
	}
	return ""
}

func (f *EnvFile) Set(key, value string) {
	if f.index == nil {
		f.index = map[string]int{}
	}
	if idx, ok := f.index[key]; ok {
		f.lines[idx] = envLine{kind: envLineKeyValue, key: key, value: value}
		return
	}
	f.index[key] = len(f.lines)
	f.lines = append(f.lines, envLine{kind: envLineKeyValue, key: key, value: value})
}

func (f *EnvFile) Values() map[string]string {
	values := map[string]string{}
	if f == nil {
		return values
	}
	for _, line := range f.lines {
		if line.kind == envLineKeyValue {
			values[line.key] = line.value
		}
	}
	return values
}

func (f *EnvFile) Bytes() []byte {
	var out bytes.Buffer
	for _, line := range f.lines {
		switch line.kind {
		case envLineKeyValue:
			_, _ = fmt.Fprintf(&out, "%s=%s\n", line.key, quoteEnvValue(line.value))
		default:
			if line.raw != "" {
				out.WriteString(line.raw)
			}
			out.WriteByte('\n')
		}
	}
	return out.Bytes()
}

func quoteEnvValue(value string) string {
	value = normalizeEnvValue(value)
	if value == "" {
		return ""
	}
	if strings.ContainsAny(value, " \t#\"") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

func normalizeEnvValue(value string) string {
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func unquoteEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		return strings.ReplaceAll(value[1:len(value)-1], `\"`, `"`)
	}
	return value
}

func sortedEnvKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

package basetype

import "fmt"

func (m Metric) NoModelKey() string {
	return fmt.Sprintf("%s$%s$%s", m.Name, m.Unit, m.Query)
}
func (m Metric) WithModelKey(modelType string) string {
	return fmt.Sprintf("%s$%s", m.NoModelKey(), modelType)
}

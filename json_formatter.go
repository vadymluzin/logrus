package logrus

import (
	"encoding/json"
	"fmt"
)

type fieldKey string
type FieldMap map[fieldKey]string

const (
	FieldKeyMsg    = "msg"
	FieldKeyLevel  = "level"
	FieldKeyTime   = "time"
	FieldKeyMethod = "method"
)

func (f FieldMap) resolve(key fieldKey) string {
	if k, ok := f[key]; ok {
		return k
	}

	return string(key)
}

type JSONFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	TimestampFormat string

	// FieldMap allows users to customize the names of keys for various fields.
	// As an example:
	// formatter := &JSONFormatter{
	//   	FieldMap: FieldMap{
	// 		 FieldKeyTime:   "@timestamp",
	// 		 FieldKeyLevel:  "@level",
	// 		 FieldKeyMsg:    "@message",
	// 		 FieldKeyMethod: "@caller",
	//    },
	// }
	FieldMap FieldMap
}

func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	data := make(Fields, len(entry.Data)+4)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}
	prefixFieldClashes(data)

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = DefaultTimestampFormat
	}

	data[f.FieldMap.resolve(FieldKeyTime)] = entry.Time.Format(timestampFormat)
	data[f.FieldMap.resolve(FieldKeyMsg)] = entry.Message
	data[f.FieldMap.resolve(FieldKeyLevel)] = entry.Level.String()
	if ReportMethod() {
		data[f.FieldMap.resolve(FieldKeyMethod)] = entry.Method
	}
	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

package utils

import (
	"maps"

	"github.com/sirupsen/logrus"
)

const logIDKey = "log_id"

type LogID int

const (
	UnknownLogID LogID = iota
	DryRunLogID
	PreExecLogID
	PostExecLogID
	DisablingPowerEventsLogID
)

type LogrusCustomFields struct {
	fields map[string]interface{}
}

func NewLogrusEmptyFields() *LogrusCustomFields {
	return &LogrusCustomFields{}
}

func NewLogrusCustomFields(fields map[string]interface{}) *LogrusCustomFields {
	return &LogrusCustomFields{fields: fields}
}

func (l *LogrusCustomFields) WithLogID(id LogID) logrus.Fields {
	newFields := logrus.Fields{}
	maps.Copy(newFields, l.fields)
	newFields[logIDKey] = id
	return newFields
}

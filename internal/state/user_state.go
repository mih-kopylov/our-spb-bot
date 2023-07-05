package state

import (
	"github.com/samber/lo"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type FormField string

type UserState struct {
	logger             *zap.Logger
	UserId             int64          `firestore:"userId"`
	FullName           string         `firestore:"fullName"`
	MessageHandlerName string         `firestore:"messageHandlerName"`
	LastAccessAt       time.Time      `firestore:"lastAccessAt"`
	Form               map[string]any `firestore:"form"`

	Accounts          []Account `firestore:"accounts"`
	Categories        string    `firestore:"categories"`
	SentMessagesCount int       `firestore:"sentMessagesCount"`
}

func (s *UserState) ClearForm() {
	s.Form = nil
}

func (s *UserState) SetFormField(key FormField, value any) {
	if s.Form == nil {
		s.Form = map[string]any{}
	}

	s.Form[string(key)] = value
}

func (s *UserState) GetStringFormField(key FormField) string {
	if s.Form == nil {
		return ""
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return ""
	}

	stringValue, ok := value.(string)
	if !ok {
		return ""
	}

	return stringValue
}

func (s *UserState) GetIntFormField(key FormField) int {
	if s.Form == nil {
		return 0
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return 0
	}

	intValue, ok := value.(int)
	if !ok {
		return 0
	}

	return intValue
}

func (s *UserState) GetStringSlice(key FormField) []string {
	if s.Form == nil {
		return nil
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return nil
	}

	typeOfValue := reflect.TypeOf(value)
	if typeOfValue.Kind() == reflect.Slice && typeOfValue.Elem().Kind() == reflect.String {
		valueStringSlice, ok := value.([]string)
		if !ok {
			s.logger.Warn("failed to convert value to string array",
				zap.Int64("userId", s.UserId),
				zap.Any("key", key))
			return nil
		}

		return valueStringSlice
	}

	valueArray, ok := value.([]any)
	if !ok {
		s.logger.Warn("failed to convert value to array",
			zap.Int64("userId", s.UserId),
			zap.Any("key", key))
		return nil
	}

	var sliceStrings []string
	for _, item := range valueArray {
		itemString, ok := item.(string)
		if !ok {
			s.logger.Warn("failed to convert array value item to string",
				zap.Int64("userId", s.UserId),
				zap.Any("item", item),
				zap.Any("key", key))
		} else {
			sliceStrings = append(sliceStrings, itemString)
		}
	}

	return sliceStrings
}

func (s *UserState) AddValueToStringSlice(key FormField, value string) {
	if s.Form == nil {
		s.Form = map[string]any{}
	}

	currentValue := s.GetStringSlice(key)
	currentValue = append(currentValue, value)
	s.Form[string(key)] = currentValue
}

func (s *UserState) RemoveValueFromStringSlice(key FormField, value string) {
	if s.Form == nil {
		s.Form = map[string]any{}
	}

	currentValue := s.GetStringSlice(key)
	index := lo.IndexOf(currentValue, value)
	if index >= 0 {
		currentValue = append(currentValue[0:index], currentValue[index+1:]...)
		s.Form[string(key)] = currentValue
	}
}

func (s *UserState) GetStringMap(key FormField) map[string]string {
	if s.Form == nil {
		return map[string]string{}
	}

	value, exists := s.Form[string(key)]
	if !exists {
		return map[string]string{}
	}

	mapValue := value.(map[string]any)

	result := map[string]string{}
	for k, v := range mapValue {
		result[k] = v.(string)
	}
	return result
}

func (s *UserState) PutValueToMap(key FormField, valueKey string, value string) {
	currentValue := s.GetStringMap(key)
	currentValue[valueKey] = value
	s.Form[string(key)] = currentValue
}

package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Validator validates structs based on tags.
type Validator struct {
	cache sync.Map
}

// New creates a new Validator.
func New() *Validator {
	return &Validator{}
}

type structMetadata struct {
	fields []fieldMetadata
}

type fieldMetadata struct {
	name       string
	fieldName  string
	rules      []string
	isEmbedded bool
}

// Struct validates a struct's fields based on tags.
func (v *Validator) Struct(s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("only structs are supported")
	}
	return v.validateStruct(val, "")
}

func (v *Validator) validateStruct(val reflect.Value, prefix string) error {
	t := val.Type()
	meta := v.getMetadata(t)

	for _, fieldMeta := range meta.fields {
		var fVal reflect.Value
		if fieldMeta.isEmbedded {
			fVal = val.FieldByName(fieldMeta.fieldName)
			if err := v.validateStruct(fVal, prefix); err != nil {
				return err
			}
			continue
		}

		fVal = val.FieldByName(fieldMeta.fieldName)
		for _, rule := range fieldMeta.rules {
			if err := validateField(fVal, rule, fieldMeta.name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *Validator) getMetadata(t reflect.Type) *structMetadata {
	if val, ok := v.cache.Load(t); ok {
		return val.(*structMetadata)
	}
	meta := v.parseStruct(t)
	v.cache.Store(t, meta)
	return meta
}

func (v *Validator) parseStruct(t reflect.Type) *structMetadata {
	meta := &structMetadata{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			// We mark it as embedded so we can recursively validate it
			meta.fields = append(meta.fields, fieldMetadata{
				fieldName:  f.Name,
				isEmbedded: true,
			})
		} else {
			tag := f.Tag.Get("validate")
			if tag == "" {
				continue
			}
			rules := strings.Split(tag, ",")
			meta.fields = append(meta.fields, fieldMetadata{
				name:      f.Name,
				fieldName: f.Name,
				rules:     rules,
			})
		}
	}
	return meta
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func validateField(val reflect.Value, rule string, fieldName string) error {
	valStr := fmt.Sprintf("%v", val.Interface())

	switch {
	case rule == "required":
		if valStr == "" || (val.Kind() == reflect.String && strings.TrimSpace(valStr) == "") {
			return fmt.Errorf("field %s is required", fieldName)
		}
	case rule == "email":
		if !emailRegex.MatchString(strings.ToLower(valStr)) {
			return fmt.Errorf("field %s must be a valid email", fieldName)
		}
	case rule == "numeric":
		if _, err := strconv.Atoi(valStr); err != nil {
			return fmt.Errorf("field %s must be numeric", fieldName)
		}
	}
	return nil
}

package data

type Fields map[string]interface{}

type FieldsConverter interface {
	ToFields() Fields
}

func WithField(key string, value interface{}) Fields {
	i := make(Fields, 3)
	return i.WithField(key, value)
}

func WithFields(fields FieldsConverter) Fields {
	return fields.ToFields()
}

func (f Fields) WithField(key string, value interface{}) Fields {
	n := f.copy()
	n[key] = value
	return n
}

func (f Fields) WithFields(data Fields) Fields {
	n := f.copy()
	for k, v := range data {
		n[k] = v
	}
	return n
}

func (f Fields) copy() Fields {
	data := Fields{}
	for k, v := range f {
		data[k] = v
	}
	return data
}

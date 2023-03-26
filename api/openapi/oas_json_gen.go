// Code generated by ogen, DO NOT EDIT.

package openapi

import (
	"math/bits"
	"strconv"

	"github.com/go-faster/errors"
	"github.com/go-faster/jx"

	"github.com/ogen-go/ogen/json"
	"github.com/ogen-go/ogen/validate"
)

// Encode implements json.Marshaler.
func (s *AddPageReq) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *AddPageReq) encodeFields(e *jx.Encoder) {
	{

		e.FieldStart("url")
		e.Str(s.URL)
	}
	{
		if s.Description.Set {
			e.FieldStart("description")
			s.Description.Encode(e)
		}
	}
	{
		if s.Formats != nil {
			e.FieldStart("formats")
			e.ArrStart()
			for _, elem := range s.Formats {
				elem.Encode(e)
			}
			e.ArrEnd()
		}
	}
}

var jsonFieldsNameOfAddPageReq = [3]string{
	0: "url",
	1: "description",
	2: "formats",
}

// Decode decodes AddPageReq from json.
func (s *AddPageReq) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode AddPageReq to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "url":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := d.Str()
				s.URL = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"url\"")
			}
		case "description":
			if err := func() error {
				s.Description.Reset()
				if err := s.Description.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"description\"")
			}
		case "formats":
			if err := func() error {
				s.Formats = make([]Format, 0)
				if err := d.Arr(func(d *jx.Decoder) error {
					var elem Format
					if err := elem.Decode(d); err != nil {
						return err
					}
					s.Formats = append(s.Formats, elem)
					return nil
				}); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"formats\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode AddPageReq")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000001,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfAddPageReq) {
					name = jsonFieldsNameOfAddPageReq[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *AddPageReq) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *AddPageReq) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *Error) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *Error) encodeFields(e *jx.Encoder) {
	{

		e.FieldStart("message")
		e.Str(s.Message)
	}
	{
		if s.Localized.Set {
			e.FieldStart("localized")
			s.Localized.Encode(e)
		}
	}
}

var jsonFieldsNameOfError = [2]string{
	0: "message",
	1: "localized",
}

// Decode decodes Error from json.
func (s *Error) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Error to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "message":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := d.Str()
				s.Message = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"message\"")
			}
		case "localized":
			if err := func() error {
				s.Localized.Reset()
				if err := s.Localized.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"localized\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode Error")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000001,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfError) {
					name = jsonFieldsNameOfError[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *Error) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Error) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes Format as json.
func (s Format) Encode(e *jx.Encoder) {
	e.Str(string(s))
}

// Decode decodes Format from json.
func (s *Format) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Format to nil")
	}
	v, err := d.StrBytes()
	if err != nil {
		return err
	}
	// Try to use constant string.
	switch Format(v) {
	case FormatAll:
		*s = FormatAll
	case FormatPdf:
		*s = FormatPdf
	case FormatSinglePage:
		*s = FormatSinglePage
	case FormatHeaders:
		*s = FormatHeaders
	default:
		*s = Format(v)
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s Format) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Format) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes AddPageReq as json.
func (o OptAddPageReq) Encode(e *jx.Encoder) {
	if !o.Set {
		return
	}
	o.Value.Encode(e)
}

// Decode decodes AddPageReq from json.
func (o *OptAddPageReq) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptAddPageReq to nil")
	}
	o.Set = true
	if err := o.Value.Decode(d); err != nil {
		return err
	}
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s OptAddPageReq) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OptAddPageReq) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes string as json.
func (o OptString) Encode(e *jx.Encoder) {
	if !o.Set {
		return
	}
	e.Str(string(o.Value))
}

// Decode decodes string from json.
func (o *OptString) Decode(d *jx.Decoder) error {
	if o == nil {
		return errors.New("invalid: unable to decode OptString to nil")
	}
	o.Set = true
	v, err := d.Str()
	if err != nil {
		return err
	}
	o.Value = string(v)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s OptString) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *OptString) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *Page) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *Page) encodeFields(e *jx.Encoder) {
	{

		e.FieldStart("id")
		json.EncodeUUID(e, s.ID)
	}
	{

		e.FieldStart("url")
		e.Str(s.URL)
	}
	{

		e.FieldStart("created")
		json.EncodeDateTime(e, s.Created)
	}
	{

		e.FieldStart("formats")
		e.ArrStart()
		for _, elem := range s.Formats {
			elem.Encode(e)
		}
		e.ArrEnd()
	}
	{

		e.FieldStart("status")
		s.Status.Encode(e)
	}
}

var jsonFieldsNameOfPage = [5]string{
	0: "id",
	1: "url",
	2: "created",
	3: "formats",
	4: "status",
}

// Decode decodes Page from json.
func (s *Page) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Page to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "id":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := json.DecodeUUID(d)
				s.ID = v
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"id\"")
			}
		case "url":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.URL = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"url\"")
			}
		case "created":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := json.DecodeDateTime(d)
				s.Created = v
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"created\"")
			}
		case "formats":
			requiredBitSet[0] |= 1 << 3
			if err := func() error {
				s.Formats = make([]Format, 0)
				if err := d.Arr(func(d *jx.Decoder) error {
					var elem Format
					if err := elem.Decode(d); err != nil {
						return err
					}
					s.Formats = append(s.Formats, elem)
					return nil
				}); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"formats\"")
			}
		case "status":
			requiredBitSet[0] |= 1 << 4
			if err := func() error {
				if err := s.Status.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"status\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode Page")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00011111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfPage) {
					name = jsonFieldsNameOfPage[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *Page) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Page) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *PageWithResults) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *PageWithResults) encodeFields(e *jx.Encoder) {
	{

		e.FieldStart("id")
		json.EncodeUUID(e, s.ID)
	}
	{

		e.FieldStart("url")
		e.Str(s.URL)
	}
	{

		e.FieldStart("created")
		json.EncodeDateTime(e, s.Created)
	}
	{

		e.FieldStart("formats")
		e.ArrStart()
		for _, elem := range s.Formats {
			elem.Encode(e)
		}
		e.ArrEnd()
	}
	{

		e.FieldStart("status")
		s.Status.Encode(e)
	}
	{

		e.FieldStart("results")
		e.ArrStart()
		for _, elem := range s.Results {
			elem.Encode(e)
		}
		e.ArrEnd()
	}
}

var jsonFieldsNameOfPageWithResults = [6]string{
	0: "id",
	1: "url",
	2: "created",
	3: "formats",
	4: "status",
	5: "results",
}

// Decode decodes PageWithResults from json.
func (s *PageWithResults) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode PageWithResults to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "id":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := json.DecodeUUID(d)
				s.ID = v
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"id\"")
			}
		case "url":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.URL = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"url\"")
			}
		case "created":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := json.DecodeDateTime(d)
				s.Created = v
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"created\"")
			}
		case "formats":
			requiredBitSet[0] |= 1 << 3
			if err := func() error {
				s.Formats = make([]Format, 0)
				if err := d.Arr(func(d *jx.Decoder) error {
					var elem Format
					if err := elem.Decode(d); err != nil {
						return err
					}
					s.Formats = append(s.Formats, elem)
					return nil
				}); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"formats\"")
			}
		case "status":
			requiredBitSet[0] |= 1 << 4
			if err := func() error {
				if err := s.Status.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"status\"")
			}
		case "results":
			requiredBitSet[0] |= 1 << 5
			if err := func() error {
				s.Results = make([]Result, 0)
				if err := d.Arr(func(d *jx.Decoder) error {
					var elem Result
					if err := elem.Decode(d); err != nil {
						return err
					}
					s.Results = append(s.Results, elem)
					return nil
				}); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"results\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode PageWithResults")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00111111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfPageWithResults) {
					name = jsonFieldsNameOfPageWithResults[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *PageWithResults) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *PageWithResults) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes Pages as json.
func (s Pages) Encode(e *jx.Encoder) {
	unwrapped := []Page(s)

	e.ArrStart()
	for _, elem := range unwrapped {
		elem.Encode(e)
	}
	e.ArrEnd()
}

// Decode decodes Pages from json.
func (s *Pages) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Pages to nil")
	}
	var unwrapped []Page
	if err := func() error {
		unwrapped = make([]Page, 0)
		if err := d.Arr(func(d *jx.Decoder) error {
			var elem Page
			if err := elem.Decode(d); err != nil {
				return err
			}
			unwrapped = append(unwrapped, elem)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return errors.Wrap(err, "alias")
	}
	*s = Pages(unwrapped)
	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s Pages) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Pages) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *Result) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *Result) encodeFields(e *jx.Encoder) {
	{

		e.FieldStart("format")
		s.Format.Encode(e)
	}
	{
		if s.Error.Set {
			e.FieldStart("error")
			s.Error.Encode(e)
		}
	}
	{

		e.FieldStart("files")
		e.ArrStart()
		for _, elem := range s.Files {
			elem.Encode(e)
		}
		e.ArrEnd()
	}
}

var jsonFieldsNameOfResult = [3]string{
	0: "format",
	1: "error",
	2: "files",
}

// Decode decodes Result from json.
func (s *Result) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Result to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "format":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				if err := s.Format.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"format\"")
			}
		case "error":
			if err := func() error {
				s.Error.Reset()
				if err := s.Error.Decode(d); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"error\"")
			}
		case "files":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				s.Files = make([]ResultFilesItem, 0)
				if err := d.Arr(func(d *jx.Decoder) error {
					var elem ResultFilesItem
					if err := elem.Decode(d); err != nil {
						return err
					}
					s.Files = append(s.Files, elem)
					return nil
				}); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"files\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode Result")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00000101,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfResult) {
					name = jsonFieldsNameOfResult[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *Result) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Result) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode implements json.Marshaler.
func (s *ResultFilesItem) Encode(e *jx.Encoder) {
	e.ObjStart()
	s.encodeFields(e)
	e.ObjEnd()
}

// encodeFields encodes fields.
func (s *ResultFilesItem) encodeFields(e *jx.Encoder) {
	{

		e.FieldStart("id")
		json.EncodeUUID(e, s.ID)
	}
	{

		e.FieldStart("name")
		e.Str(s.Name)
	}
	{

		e.FieldStart("mimetype")
		e.Str(s.Mimetype)
	}
	{

		e.FieldStart("size")
		e.Int64(s.Size)
	}
}

var jsonFieldsNameOfResultFilesItem = [4]string{
	0: "id",
	1: "name",
	2: "mimetype",
	3: "size",
}

// Decode decodes ResultFilesItem from json.
func (s *ResultFilesItem) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode ResultFilesItem to nil")
	}
	var requiredBitSet [1]uint8

	if err := d.ObjBytes(func(d *jx.Decoder, k []byte) error {
		switch string(k) {
		case "id":
			requiredBitSet[0] |= 1 << 0
			if err := func() error {
				v, err := json.DecodeUUID(d)
				s.ID = v
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"id\"")
			}
		case "name":
			requiredBitSet[0] |= 1 << 1
			if err := func() error {
				v, err := d.Str()
				s.Name = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"name\"")
			}
		case "mimetype":
			requiredBitSet[0] |= 1 << 2
			if err := func() error {
				v, err := d.Str()
				s.Mimetype = string(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"mimetype\"")
			}
		case "size":
			requiredBitSet[0] |= 1 << 3
			if err := func() error {
				v, err := d.Int64()
				s.Size = int64(v)
				if err != nil {
					return err
				}
				return nil
			}(); err != nil {
				return errors.Wrap(err, "decode field \"size\"")
			}
		default:
			return d.Skip()
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "decode ResultFilesItem")
	}
	// Validate required fields.
	var failures []validate.FieldError
	for i, mask := range [1]uint8{
		0b00001111,
	} {
		if result := (requiredBitSet[i] & mask) ^ mask; result != 0 {
			// Mask only required fields and check equality to mask using XOR.
			//
			// If XOR result is not zero, result is not equal to expected, so some fields are missed.
			// Bits of fields which would be set are actually bits of missed fields.
			missed := bits.OnesCount8(result)
			for bitN := 0; bitN < missed; bitN++ {
				bitIdx := bits.TrailingZeros8(result)
				fieldIdx := i*8 + bitIdx
				var name string
				if fieldIdx < len(jsonFieldsNameOfResultFilesItem) {
					name = jsonFieldsNameOfResultFilesItem[fieldIdx]
				} else {
					name = strconv.Itoa(fieldIdx)
				}
				failures = append(failures, validate.FieldError{
					Name:  name,
					Error: validate.ErrFieldRequired,
				})
				// Reset bit.
				result &^= 1 << bitIdx
			}
		}
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s *ResultFilesItem) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *ResultFilesItem) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

// Encode encodes Status as json.
func (s Status) Encode(e *jx.Encoder) {
	e.Str(string(s))
}

// Decode decodes Status from json.
func (s *Status) Decode(d *jx.Decoder) error {
	if s == nil {
		return errors.New("invalid: unable to decode Status to nil")
	}
	v, err := d.StrBytes()
	if err != nil {
		return err
	}
	// Try to use constant string.
	switch Status(v) {
	case StatusNew:
		*s = StatusNew
	case StatusProcessing:
		*s = StatusProcessing
	case StatusDone:
		*s = StatusDone
	case StatusFailed:
		*s = StatusFailed
	case StatusWithErrors:
		*s = StatusWithErrors
	default:
		*s = Status(v)
	}

	return nil
}

// MarshalJSON implements stdjson.Marshaler.
func (s Status) MarshalJSON() ([]byte, error) {
	e := jx.Encoder{}
	s.Encode(&e)
	return e.Bytes(), nil
}

// UnmarshalJSON implements stdjson.Unmarshaler.
func (s *Status) UnmarshalJSON(data []byte) error {
	d := jx.DecodeBytes(data)
	return s.Decode(d)
}

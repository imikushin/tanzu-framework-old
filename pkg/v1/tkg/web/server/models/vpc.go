// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// Vpc vpc
// swagger:model vpc
type Vpc struct {

	// cidr
	Cidr string `json:"cidr,omitempty"`

	// id
	ID string `json:"id,omitempty"`
}

// Validate validates this vpc
func (m *Vpc) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Vpc) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Vpc) UnmarshalBinary(b []byte) error {
	var res Vpc
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

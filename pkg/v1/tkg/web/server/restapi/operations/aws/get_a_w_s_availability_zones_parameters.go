// Code generated by go-swagger; DO NOT EDIT.

package aws

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
)

// NewGetAWSAvailabilityZonesParams creates a new GetAWSAvailabilityZonesParams object
// no default values defined in spec.
func NewGetAWSAvailabilityZonesParams() GetAWSAvailabilityZonesParams {

	return GetAWSAvailabilityZonesParams{}
}

// GetAWSAvailabilityZonesParams contains all the bound params for the get a w s availability zones operation
// typically these are obtained from a http.Request
//
// swagger:parameters getAWSAvailabilityZones
type GetAWSAvailabilityZonesParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetAWSAvailabilityZonesParams() beforehand.
func (o *GetAWSAvailabilityZonesParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

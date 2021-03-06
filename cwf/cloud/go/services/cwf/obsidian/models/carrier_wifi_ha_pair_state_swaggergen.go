// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// CarrierWifiHaPairState Full state object of an ha pair
// swagger:model carrier_wifi_ha_pair_state
type CarrierWifiHaPairState struct {

	// gateway1 health
	Gateway1Health *CarrierWifiGatewayHealthStatus `json:"gateway1_health,omitempty"`

	// gateway2 health
	Gateway2Health *CarrierWifiGatewayHealthStatus `json:"gateway2_health,omitempty"`

	// ha pair status
	HaPairStatus *CarrierWifiHaPairStatus `json:"ha_pair_status,omitempty"`
}

// Validate validates this carrier wifi ha pair state
func (m *CarrierWifiHaPairState) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateGateway1Health(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateGateway2Health(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateHaPairStatus(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CarrierWifiHaPairState) validateGateway1Health(formats strfmt.Registry) error {

	if swag.IsZero(m.Gateway1Health) { // not required
		return nil
	}

	if m.Gateway1Health != nil {
		if err := m.Gateway1Health.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("gateway1_health")
			}
			return err
		}
	}

	return nil
}

func (m *CarrierWifiHaPairState) validateGateway2Health(formats strfmt.Registry) error {

	if swag.IsZero(m.Gateway2Health) { // not required
		return nil
	}

	if m.Gateway2Health != nil {
		if err := m.Gateway2Health.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("gateway2_health")
			}
			return err
		}
	}

	return nil
}

func (m *CarrierWifiHaPairState) validateHaPairStatus(formats strfmt.Registry) error {

	if swag.IsZero(m.HaPairStatus) { // not required
		return nil
	}

	if m.HaPairStatus != nil {
		if err := m.HaPairStatus.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("ha_pair_status")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *CarrierWifiHaPairState) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *CarrierWifiHaPairState) UnmarshalBinary(b []byte) error {
	var res CarrierWifiHaPairState
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// Code generated by mockery v2.10.4. DO NOT EDIT.

// Regenerate this file using `make store-mocks`.

package mocks

import (
	model "github.com/mattermost/mattermost-server/v6/model"
	mock "github.com/stretchr/testify/mock"
)

// UserTermsOfServiceStore is an autogenerated mock type for the UserTermsOfServiceStore type
type UserTermsOfServiceStore struct {
	mock.Mock
}

// Delete provides a mock function with given fields: userID, termsOfServiceId
func (_m *UserTermsOfServiceStore) Delete(userID string, termsOfServiceId string) error {
	ret := _m.Called(userID, termsOfServiceId)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(userID, termsOfServiceId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByUser provides a mock function with given fields: userID
func (_m *UserTermsOfServiceStore) GetByUser(userID string) (*model.UserTermsOfService, error) {
	ret := _m.Called(userID)

	var r0 *model.UserTermsOfService
	if rf, ok := ret.Get(0).(func(string) *model.UserTermsOfService); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserTermsOfService)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: userTermsOfService
func (_m *UserTermsOfServiceStore) Save(userTermsOfService *model.UserTermsOfService) (*model.UserTermsOfService, error) {
	ret := _m.Called(userTermsOfService)

	var r0 *model.UserTermsOfService
	if rf, ok := ret.Get(0).(func(*model.UserTermsOfService) *model.UserTermsOfService); ok {
		r0 = rf(userTermsOfService)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UserTermsOfService)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.UserTermsOfService) error); ok {
		r1 = rf(userTermsOfService)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
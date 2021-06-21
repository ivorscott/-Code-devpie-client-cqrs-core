// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	squirrel "github.com/Masterminds/squirrel"
	mock "github.com/stretchr/testify/mock"
)

// Squirrler is an autogenerated mock type for the Squirrler type
type Squirrler struct {
	mock.Mock
}

// Delete provides a mock function with given fields: from
func (_m *Squirrler) Delete(from string) squirrel.DeleteBuilder {
	ret := _m.Called(from)

	var r0 squirrel.DeleteBuilder
	if rf, ok := ret.Get(0).(func(string) squirrel.DeleteBuilder); ok {
		r0 = rf(from)
	} else {
		r0 = ret.Get(0).(squirrel.DeleteBuilder)
	}

	return r0
}

// Insert provides a mock function with given fields: into
func (_m *Squirrler) Insert(into string) squirrel.InsertBuilder {
	ret := _m.Called(into)

	var r0 squirrel.InsertBuilder
	if rf, ok := ret.Get(0).(func(string) squirrel.InsertBuilder); ok {
		r0 = rf(into)
	} else {
		r0 = ret.Get(0).(squirrel.InsertBuilder)
	}

	return r0
}

// PlaceholderFormat provides a mock function with given fields: f
func (_m *Squirrler) PlaceholderFormat(f squirrel.PlaceholderFormat) squirrel.StatementBuilderType {
	ret := _m.Called(f)

	var r0 squirrel.StatementBuilderType
	if rf, ok := ret.Get(0).(func(squirrel.PlaceholderFormat) squirrel.StatementBuilderType); ok {
		r0 = rf(f)
	} else {
		r0 = ret.Get(0).(squirrel.StatementBuilderType)
	}

	return r0
}

// Replace provides a mock function with given fields: into
func (_m *Squirrler) Replace(into string) squirrel.InsertBuilder {
	ret := _m.Called(into)

	var r0 squirrel.InsertBuilder
	if rf, ok := ret.Get(0).(func(string) squirrel.InsertBuilder); ok {
		r0 = rf(into)
	} else {
		r0 = ret.Get(0).(squirrel.InsertBuilder)
	}

	return r0
}

// RunWith provides a mock function with given fields: runner
func (_m *Squirrler) RunWith(runner squirrel.BaseRunner) squirrel.StatementBuilderType {
	ret := _m.Called(runner)

	var r0 squirrel.StatementBuilderType
	if rf, ok := ret.Get(0).(func(squirrel.BaseRunner) squirrel.StatementBuilderType); ok {
		r0 = rf(runner)
	} else {
		r0 = ret.Get(0).(squirrel.StatementBuilderType)
	}

	return r0
}

// Select provides a mock function with given fields: columns
func (_m *Squirrler) Select(columns ...string) squirrel.SelectBuilder {
	_va := make([]interface{}, len(columns))
	for _i := range columns {
		_va[_i] = columns[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 squirrel.SelectBuilder
	if rf, ok := ret.Get(0).(func(...string) squirrel.SelectBuilder); ok {
		r0 = rf(columns...)
	} else {
		r0 = ret.Get(0).(squirrel.SelectBuilder)
	}

	return r0
}

// Update provides a mock function with given fields: table
func (_m *Squirrler) Update(table string) squirrel.UpdateBuilder {
	ret := _m.Called(table)

	var r0 squirrel.UpdateBuilder
	if rf, ok := ret.Get(0).(func(string) squirrel.UpdateBuilder); ok {
		r0 = rf(table)
	} else {
		r0 = ret.Get(0).(squirrel.UpdateBuilder)
	}

	return r0
}

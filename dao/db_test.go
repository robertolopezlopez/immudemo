package dao

import (
	"context"
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

type (
	sqlDBMock struct {
		mock.Mock
	}
)

func (m *sqlDBMock) createTable(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *sqlDBMock) createIndex(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestDBase_CreateTables_OK(t *testing.T) {
	m := sqlDBMock{}
	m.On("createTable", mock.Anything).Return(nil)
	m.On("createIndex", mock.Anything).Return(nil)
	dbase := DBase{dbWrapper: &m}

	err := dbase.CreateTables(context.Background())

	require.Nil(t, err)

	m.AssertExpectations(t)
}

func TestDBase_CreateTables_createTable_NOK(t *testing.T) {
	m := sqlDBMock{}
	m.On("createTable", mock.Anything).Return(fmt.Errorf("create table error"))
	dbase := DBase{dbWrapper: &m}

	err := dbase.CreateTables(context.Background())

	require.NotNil(t, err)
	assert.Equal(t, err.Error(), "create table error")

	m.AssertExpectations(t)
}

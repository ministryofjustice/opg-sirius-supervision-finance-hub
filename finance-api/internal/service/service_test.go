package service

import (
	"github.com/opg-sirius-finance-hub/finance-api/internal/testhelpers"
	"os"
	"testing"
)

var testDB *testhelpers.TestDatabase

func TestMain(m *testing.M) {
	testDB = testhelpers.InitDb()
	defer testDB.TearDown()
	os.Exit(m.Run())
}

package exasol_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/exasol/exasol-driver-go"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite
	ctx             context.Context
	exasolContainer testcontainers.Container
	port            int
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.ctx = getContext()
	suite.exasolContainer = runExasolContainer(suite.ctx)
	suite.port = getExasolPort(suite.exasolContainer, suite.ctx)
}

func (suite *IntegrationTestSuite) TestConnect() {

	db, _ := sql.Open("exasol", fmt.Sprintf("exa:localhost:%d;user=sys;password=exasol;encryption=0", suite.port))
	rows, _ := db.Query("SELECT 2 FROM DUAL")
	columns, _ := rows.Columns()
	suite.Equal("2", columns[0])
}

func (suite *IntegrationTestSuite) TestConnectWithWrongPort() {

	exasol, _ := sql.Open("exasol", "exa:localhost:1234;user=sys;password=exasol")
	err := exasol.Ping()
	suite.Error(err)
	suite.Contains(err.Error(), "connect: connection refuse")
}

func (suite *IntegrationTestSuite) TestConnectWithWrongUsername() {

	exasol, _ := sql.Open("exasol", exasol.NewConfig("wronguser", "exasol").Insecure(true).Port(suite.port).String())
	suite.EqualError(exasol.Ping(), "[08004] Connection exception - authentication failed.")
}

func (suite *IntegrationTestSuite) TestConnectWithWrongPassword() {

	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "wrongpassword").Insecure(true).Port(suite.port).String())
	suite.EqualError(exasol.Ping(), "[08004] Connection exception - authentication failed.")
}

func (suite *IntegrationTestSuite) TestExecAndQuery() {

	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Port(suite.port).String())
	schemaName := "TEST_SCHEMA_1"
	_, _ = exasol.Exec("CREATE SCHEMA " + schemaName)
	_, _ = exasol.Exec("CREATE TABLE " + schemaName + ".TEST_TABLE(x INT)")
	_, _ = exasol.Exec("INSERT INTO " + schemaName + ".TEST_TABLE VALUES (15)")
	rows, _ := exasol.Query("SELECT x FROM " + schemaName + ".TEST_TABLE")
	suite.assertSingleValueResult(rows, "15")
}

func (suite *IntegrationTestSuite) TestExecuteWithError() {

	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Port(suite.port).String())
	_, err := exasol.Exec("CREATE SCHEMAA TEST_SCHEMA")
	suite.Error(err)
	suite.True(strings.Contains(err.Error(), "syntax error"))
}

func (suite *IntegrationTestSuite) TestQueryWithError() {

	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Port(suite.port).String())
	schemaName := "TEST_SCHEMA_2"
	_, _ = exasol.Exec("CREATE SCHEMA " + schemaName)
	_, err := exasol.Query("SELECT x FROM " + schemaName + ".TEST_TABLE")
	suite.Error(err)
	suite.True(strings.Contains(err.Error(), "object TEST_SCHEMA_2.TEST_TABLE not found"))
}

func (suite *IntegrationTestSuite) TestPreparedStatement() {

	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Port(suite.port).String())
	schemaName := "TEST_SCHEMA_3"
	_, _ = exasol.Exec("CREATE SCHEMA " + schemaName)
	_, _ = exasol.Exec("CREATE TABLE " + schemaName + ".TEST_TABLE(x INT)")
	preparedStatement, _ := exasol.Prepare("INSERT INTO " + schemaName + ".TEST_TABLE VALUES (?)")
	_, _ = preparedStatement.Exec(15)
	preparedStatement, _ = exasol.Prepare("SELECT x FROM " + schemaName + ".TEST_TABLE WHERE x = ?")
	rows, _ := preparedStatement.Query(15)
	suite.assertSingleValueResult(rows, "15")
}

func (suite *IntegrationTestSuite) TestBeginAndCommit() {
	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Autocommit(false).Port(suite.port).String())
	schemaName := "TEST_SCHEMA_4"
	transaction, _ := exasol.Begin()
	_, _ = transaction.Exec("CREATE SCHEMA " + schemaName)
	_, _ = transaction.Exec("CREATE TABLE " + schemaName + ".TEST_TABLE(x INT)")
	_, _ = transaction.Exec("INSERT INTO " + schemaName + ".TEST_TABLE VALUES (15)")
	_ = transaction.Commit()
	rows, _ := exasol.Query("SELECT x FROM " + schemaName + ".TEST_TABLE")
	suite.assertSingleValueResult(rows, "15")
}

func (suite *IntegrationTestSuite) TestBeginAndRollback() {
	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Autocommit(false).Port(suite.port).String())
	schemaName := "TEST_SCHEMA_5"
	transaction, _ := exasol.Begin()
	_, _ = transaction.Exec("CREATE SCHEMA " + schemaName)
	_, _ = transaction.Exec("CREATE TABLE " + schemaName + ".TEST_TABLE(x INT)")
	_, _ = transaction.Exec("INSERT INTO " + schemaName + ".TEST_TABLE VALUES (15)")
	_ = transaction.Rollback()
	_, err := exasol.Query("SELECT x FROM " + schemaName + ".TEST_TABLE")
	suite.Error(err)
	suite.True(strings.Contains(err.Error(), "object "+schemaName+".TEST_TABLE not found"))
}

func (suite *IntegrationTestSuite) TestPingWithContext() {
	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Port(suite.port).String())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	suite.NoError(exasol.PingContext(ctx))
	cancel()
}

func (suite *IntegrationTestSuite) TestExecuteAndQueryWithContext() {
	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Port(suite.port).String())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	schemaName := "TEST_SCHEMA_6"
	_, _ = exasol.ExecContext(ctx, "CREATE SCHEMA "+schemaName)
	_, _ = exasol.ExecContext(ctx, "CREATE TABLE "+schemaName+".TEST_TABLE(x INT)")
	_, _ = exasol.ExecContext(ctx, "INSERT INTO "+schemaName+".TEST_TABLE VALUES (15)")
	rows, _ := exasol.QueryContext(ctx, "SELECT x FROM "+schemaName+".TEST_TABLE")
	cancel()
	suite.assertSingleValueResult(rows, "15")
}

func (suite *IntegrationTestSuite) TestBeginWithCancelledContext() {
	exasol, _ := sql.Open("exasol", exasol.NewConfig("sys", "exasol").Insecure(true).Port(suite.port).Autocommit(false).String())
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	schemaName := "TEST_SCHEMA_7"
	transaction, _ := exasol.BeginTx(ctx, nil)
	_, _ = transaction.ExecContext(ctx, "CREATE SCHEMA "+schemaName)
	cancel()
	_, err := transaction.ExecContext(ctx, "CREATE TABLE "+schemaName+".TEST_TABLE(x INT)")
	suite.EqualError(err, "context canceled")
}

func (suite *IntegrationTestSuite) assertSingleValueResult(rows *sql.Rows, expected string) {
	rows.Next()
	var testValue string
	_ = rows.Scan(&testValue)
	suite.Equal(expected, testValue)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	err := suite.exasolContainer.Terminate(suite.ctx)
	onError(err)
}

func getContext() context.Context {
	return context.Background()
}

func runExasolContainer(ctx context.Context) testcontainers.Container {
	request := testcontainers.ContainerRequest{
		Image:        "exasol/docker-db:7.0.7",
		ExposedPorts: []string{"8563", "2580"},
		WaitingFor:   wait.ForLog("All stages finished"),
		Privileged:   true,
	}
	exasolContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	onError(err)
	return exasolContainer
}

func getExasolPort(exasolContainer testcontainers.Container, ctx context.Context) int {
	port, err := exasolContainer.MappedPort(ctx, "8563")
	onError(err)
	return port.Int()
}

func onError(err error) {
	if err != nil {
		log.Printf("Error %s", err)
		panic(err)
	}
}
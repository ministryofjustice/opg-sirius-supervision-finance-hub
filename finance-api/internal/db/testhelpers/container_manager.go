package testhelpers

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"os"
)

const (
	dbname       = "finance"
	user         = "user"
	password     = "password"
	stackID      = "test_stack"
	snapshotFile = "../../db_snapshot.sql"
)

// ContainerManager manages the lifecycle of Docker containers for testing
type ContainerManager struct {
	dbContainer *testcontainers.DockerContainer
}

// NewContainerManager creates a new ContainerManager instance
// TODO: Change this to use the existing testhelpers
func NewContainerManager(ctx context.Context) *ContainerManager {
	os.Setenv("TESTCONTAINERS_LOG_LEVEL", "DEBUG")
	logger := log.New(os.Stdout, "testcontainers: ", log.LstdFlags)
	testcontainers.Logger = logger

	identifier := tc.StackIdentifier(stackID)
	compose, err := tc.NewDockerComposeWith(tc.WithStackFiles("../../docker-compose.yml"), identifier, tc.WithLogger(logger))
	if err != nil {
		logger.Fatal(err)
	}

	err = compose.Up(ctx, tc.Wait(true), tc.RunServices("sirius-db", "finance-hub-api"))
	if err != nil {
		logger.Fatal(err)
	}

	err = compose.WaitForService("sirius-db", wait.ForLog("database system is ready to accept connections").WithOccurrence(2).AsRegexp()).Up(ctx, tc.RunServices("finance-migration"))
	if err != nil {
		logger.Fatal(err)
	}

	err = compose.WaitForService("finance-migration", wait.ForLog("goose: successfully migrated database to version: \\d{14}").AsRegexp()).Down(ctx)
	if err != nil {
		logger.Fatal(err)
	}

	db, err := compose.ServiceContainer(ctx, "sirius-db")
	if err != nil {
		logger.Fatal(err)
	}
	err = createSnapshot(ctx, db)
	if err != nil {
		logger.Fatal(err)
	}

	return &ContainerManager{dbContainer: db}
}

func createSnapshot(ctx context.Context, db *testcontainers.DockerContainer) error {
	_, _, err := db.Exec(ctx, []string{"pg_dump", "-U", user, "-d", dbname, "-F", "c", "-b", "-v", "-f", snapshotFile})
	if err != nil {
		return fmt.Errorf("failed to create database snapshot: %w", err)
	}
	return nil
}

func (cm *ContainerManager) Restore(ctx context.Context) error {
	_, _, err := cm.dbContainer.Exec(ctx, []string{"pg_restore", "-U", user, "-d", dbname, "-c", "-v", snapshotFile})
	if err != nil {
		return fmt.Errorf("failed to restore database snapshot: %w", err)
	}
	return nil
}

func (cm *ContainerManager) TearDown(ctx context.Context) {
	identifier := tc.StackIdentifier(stackID)
	compose, err := tc.NewDockerComposeWith(tc.WithStackFiles("../../docker-compose.yml"), identifier)
	if err != nil {
		panic(err)
	}

	err = compose.Down(ctx, tc.RemoveOrphans(true), tc.RemoveImagesLocal)
	if err != nil {
		panic(err)
	}
}

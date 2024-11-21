//go:build seed && !release

package seed

import (
	"context"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/service"
)

func seedData(ctx context.Context, svc *service.Service) error {
	// Implement your seeding logic here
	// Example: Create some initial data using the service functions
	//err := svc.CreateClient(ctx, "Test Client")
	//if err != nil {
	//	return err
	//}

	// Add more seeding logic as needed
	return nil
}

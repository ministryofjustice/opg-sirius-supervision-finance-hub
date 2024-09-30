package service

import (
	"context"
	"fmt"
)

func (s *Service) ProcessFinanceAdminUpload(ctx context.Context, bucketName string, key string) error {
	fmt.Println(bucketName)
	fmt.Println(key)
	return nil
}

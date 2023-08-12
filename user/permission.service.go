package user

import "context"

type IPermissionService interface {
	InitiateInitialPermissions(ctx context.Context) error
}

type PermissionService struct {
	repository IPermissionRepository
}

func NewPermissionService(repository IPermissionRepository) IPermissionService {
	return &PermissionService{
		repository: repository,
	}
}

func (s *PermissionService) InitiateInitialPermissions(ctx context.Context) error {
	permissions := generateInitialPermissions()
	err := s.repository.AddAll(ctx, permissions)
	if err != nil {
		return err
	}
	return nil
}

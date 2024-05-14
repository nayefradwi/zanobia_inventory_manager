package user

import "context"

type IPermissionService interface {
	InitiateInitialPermissions(ctx context.Context) error
	CreatePermission(ctx context.Context, permission Permission) error
	FindPermissionByHandle(ctx context.Context, handle string) (Permission, error)
	GetAllPermissions(ctx context.Context) ([]Permission, error)
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
	s.repository.InitiateAll(ctx, permissions)
	return nil
}

func (s *PermissionService) CreatePermission(ctx context.Context, permission Permission) error {
	validationErr := ValidatePermission(permission)
	if validationErr != nil {
		return validationErr
	}
	permission.Handle = generateHandle(permission.Name)
	err := s.repository.CreatePermssion(ctx, permission)
	if err != nil {
		return err
	}
	return nil
}

func (s *PermissionService) FindPermissionByHandle(ctx context.Context, handle string) (Permission, error) {
	return s.repository.FindByHandle(ctx, handle)
}

func (s *PermissionService) GetAllPermissions(ctx context.Context) ([]Permission, error) {
	return s.repository.GetAllPermissions(ctx)
}

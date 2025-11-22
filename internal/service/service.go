package service

import (
	"context"

	"github.com/vibe-gaming/backend/internal/config"
	"github.com/vibe-gaming/backend/internal/domain"
	"github.com/vibe-gaming/backend/internal/esia"
	"github.com/vibe-gaming/backend/internal/repository"
	"github.com/vibe-gaming/backend/pkg/auth"
	"github.com/vibe-gaming/backend/pkg/hash"
	"github.com/vibe-gaming/backend/pkg/otp"

	"github.com/google/uuid"
)

type Services struct {
	Users         Users
	Benefits      Benefits
	Cities        Cities
	Favorites     Favorites
	Organizations Organizations
}

type Deps struct {
	Config                 *config.Config
	Hasher                 hash.PasswordHasher
	TokenManager           auth.TokenManager
	OtpGenerator           otp.Generator
	OrganizationRepository repository.OrganizationRepository
	Repos                  *repository.Repositories
	EsiaClient             *esia.Client
	GigachatClient         interface {
		EnhanceSearchQuery(ctx context.Context, query string) ([]string, error)
	}
}

func NewServices(deps Deps) *Services {
	return &Services{
		Users: newUserService(deps.Repos.Users,
			deps.Repos.RefreshSession,
			deps.Repos.Cities,
			deps.Repos.UserDocument,
			deps.Hasher,
			deps.TokenManager,
			deps.OtpGenerator,
			deps.EsiaClient,
			deps.Config.Auth,
			deps.Config,
		),
		Benefits:      newBenefitService(deps.Repos.Benefits, deps.Repos.Favorite, deps.Repos.Users, deps.Repos.Organization, deps.GigachatClient),
		Cities:        newCityService(deps.Repos.Cities),
		Favorites:     newFavoriteService(deps.Repos.Favorite),
		Organizations: newOrganizationService(deps.Repos.Organization),
	}
}

type Users interface {
	Auth(ctx context.Context, code string, userAgent string, userIP string) (*Tokens, error)
	createSession(ctx context.Context, userID *uuid.UUID, userAgent *string, userIP *string) (*Tokens, error)
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUserInfo(ctx context.Context, userID uuid.UUID, cityID uuid.UUID, groups domain.GroupTypeList) error
	UpdateUserGroups(ctx context.Context, userID uuid.UUID, groups domain.UserGroupList) error
	CreateDocument(ctx context.Context, document *domain.UserDocument) error
	GetDocumentsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.UserDocument, error)
	GeneratePensionerCertificatePDF(ctx context.Context, userID uuid.UUID) ([]byte, error)
	GenerateUserCertificatePDF(ctx context.Context, userID uuid.UUID, groupType domain.GroupType) ([]byte, error)
	Count(ctx context.Context) (int64, error)
	GetUserGroupsStats(ctx context.Context) (map[string]int64, error)
}

type Cities interface {
	GetAll(ctx context.Context) ([]domain.City, error)
	Count(ctx context.Context) (int64, error)
}

type Benefits interface {
	Create(ctx context.Context, benefit *domain.Benefit) error
	Update(ctx context.Context, benefit *domain.Benefit) error
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context, page, limit int, filters *repository.BenefitFilters) ([]*domain.Benefit, int64, error)
	GetByID(ctx context.Context, id string, userID *uuid.UUID) (*domain.Benefit, error)
	GetByIDWithoutIncrement(ctx context.Context, id string, userID *uuid.UUID) (*domain.Benefit, error)
	IsFavorite(ctx context.Context, userID uuid.UUID, benefitID uuid.UUID) (bool, error)
	MarkAsFavorite(ctx context.Context, userID uuid.UUID, benefitID uuid.UUID) error
	GetFilterStats(ctx context.Context, filters *repository.BenefitFilters) (*repository.FilterStats, error)
	GetUserBenefitsStats(ctx context.Context, userID uuid.UUID) (*repository.UserBenefitsStats, error)
	GeneratePDF(ctx context.Context, benefit *domain.Benefit) ([]byte, error)
	GenerateBenefitsListPDF(ctx context.Context, benefits []*domain.Benefit, total int64, page int, limit int) ([]byte, error)
	Count(ctx context.Context, filters *repository.BenefitFilters) (int64, error)
	GetBenefitTypesStats(ctx context.Context) (map[string]int64, error)
}

type Favorites interface {
	GetTotalCount(ctx context.Context) (int64, error)
}

type Organizations interface {
	Create(ctx context.Context, organization *domain.Organization) error
	Update(ctx context.Context, organization *domain.Organization) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*domain.Organization, error)
	GetAll(ctx context.Context) ([]domain.Organization, error)
	GetAllByCityID(ctx context.Context, cityID string) ([]domain.Organization, error)
}

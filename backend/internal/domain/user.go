package domain

import (
	domainerrors "backend/internal/domain/errors"
	"backend/pkg/errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	OauthProviderGitHub OauthProvider = "github"
	OauthProviderGoogle OauthProvider = "google"
)

type (
	UserAggregate struct {
		BaseUser
		LocalUser      *LocalUser
		ThirdPartyUser *ThirdPartyUser
	}

	LocalUser struct {
		Password string `json:"password"`
	}

	BaseUser struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Email       string    `json:"email"`
		WorkspaceID uuid.UUID `json:"workspace_id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	ThirdPartyUser struct {
		OauthProvider OauthProvider `json:"oauth_provider"`
		OauthID       string        `json:"oauth_id"`
	}
	UserFactory   struct{}
	OauthProvider string
)

func NewLocalUser(password string) (LocalUser, *errors.Error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return LocalUser{}, err
	}
	return LocalUser{Password: hashedPassword}, nil
}

func NewBaseUser(name, email string, workspaceID uuid.UUID) BaseUser {
	return BaseUser{
		ID:          uuid.New(),
		Name:        name,
		Email:       email,
		WorkspaceID: workspaceID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func NewThirdPartyUser(oauthProvider, oauthID string) (*ThirdPartyUser, *errors.Error) {
	return &ThirdPartyUser{
		OauthProvider: OauthProvider(oauthProvider),
		OauthID:       oauthID,
	}, nil
}

func (u *LocalUser) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (f *UserFactory) Create(oauthProvider *OauthProvider, oauthId *uuid.UUID, name, email string, password *string, workspaceID uuid.UUID) (UserAggregate, *errors.Error) {
	baseUser := NewBaseUser(name, email, workspaceID)
	if oauthProvider != nil && oauthId != nil {
		thirdPartyUser, err := NewThirdPartyUser(string(*oauthProvider), oauthId.String())
		if err != nil {
			return UserAggregate{}, err
		}
		return UserAggregate{
			BaseUser:       baseUser,
			ThirdPartyUser: thirdPartyUser,
		}, nil
	}
	if password != nil {
		hashedPassword, err := hashPassword(*password)
		if err != nil {
			return UserAggregate{}, err
		}
		localUser, err := NewLocalUser(hashedPassword)
		if err != nil {
			return UserAggregate{}, err
		}
		return UserAggregate{
			BaseUser:  baseUser,
			LocalUser: &localUser,
		}, nil
	}

	return UserAggregate{}, domainerrors.InvalidInput("authentication method", "either password or oauth credentials must be provided")

}

func hashPassword(password string) (string, *errors.Error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "failed to hash password").
			WithHTTPStatus(500).
			WithSeverity(errors.SeverityCritical)
	}

	return string(hashed), nil
}

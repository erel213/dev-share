package domain

import (
	domainerrors "backend/internal/domain/errors"
	"backend/pkg/errors"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

const (
	OauthProviderGitHub OauthProvider = "github"
	OauthProviderGoogle OauthProvider = "google"

	// Argon2id parameters (OWASP recommendations)
	argon2Time      = 2
	argon2Memory    = 19 * 1024 // 19 MB
	argon2Threads   = 1
	argon2KeyLength = 32
	argon2SaltLen   = 16
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
	valid, err := verifyArgon2idHash(password, u.Password)
	if err != nil {
		return false
	}
	return valid
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
		localUser, err := NewLocalUser(*password)
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
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", errors.Wrap(err, "failed to generate salt").
			WithHTTPStatus(500).
			WithSeverity(errors.SeverityCritical)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLength)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	encodedPassword := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, encodedSalt, encodedHash)

	return encodedPassword, nil
}

func verifyArgon2idHash(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	var version int
	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	keyLength := uint32(len(decodedHash))
	comparisonHash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLength)

	if len(comparisonHash) != len(decodedHash) {
		return false, nil
	}

	for i := range comparisonHash {
		if comparisonHash[i] != decodedHash[i] {
			return false, nil
		}
	}

	return true, nil
}

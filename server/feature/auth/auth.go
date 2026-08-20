package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/plex"
	"github.com/sbondCo/Watcharr/token"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// We use a separate struct for registration to avoid confusion
// and possible accidents where we allow a user to pass in a
// property from the main User struct that shouldn't be allowed.
type UserRegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UseAdminTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type JellyfinAuth struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

type JellyfinAuthResponse struct {
	User struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	} `json:"User"`
	AccessToken string `json:"AccessToken"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type UserPasswordUpdateRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type AvailableAuthProvidersResponse struct {
	AvailableAuthProviders []string `json:"available"`
	SignupEnabled          bool     `json:"signupEnabled"`
	IsInSetup              bool     `json:"isInSetup"`
	UseEmby                bool     `json:"useEmby"`
	HeaderAuthAutoLogin    bool     `json:"headerAuthAutoLogin"`
}

type PlexProvider interface {
	FetchPlexAccountFromToken(token string) (plex.PlexUser, error)
	GetPlexHomeServerAuthToken(plexAuth string, userClientId string) (string, error)
}

type Service struct {
	db           *gorm.DB
	cfg          *config.ServerConfig
	plexProvider PlexProvider
}

func NewService(db *gorm.DB, cfg *config.ServerConfig, plexProvider PlexProvider) *Service {
	return &Service{
		db,
		cfg,
		plexProvider,
	}
}

func (s *Service) Register(ur *UserRegisterRequest, initialPerm int) (AuthResponse, error) {
	if !s.cfg.SIGNUP_ENABLED {
		slog.Warn("Register: Register called, but signing up is disabled.")
		return AuthResponse{}, errors.New("registering is disabled")
	}
	var user entity.User = entity.User{Username: ur.Username, Password: ur.Password}
	slog.Info("Register: A user is registering", "username", user.Username)
	hash, err := s.hashPassword(user.Password, entity.GetPassArgonParams())
	if err != nil {
		slog.Error("Register: Hashing password failed!", "error", err)
		return AuthResponse{}, errors.New("registering failed")
	}

	// Update user obj to replace the plaintext pass with hash
	user.Password = hash

	// Update user permissions if an initial perm is passed in (1 is default)
	if initialPerm != 0 && initialPerm != entity.PERM_NONE {
		slog.Info("Register: User being registered has been given extra initial permissions", "initial_perm", initialPerm)
		user.Permissions = initialPerm
	}

	user.Country = &s.cfg.DEFAULT_COUNTRY

	res := s.db.Create(&user)
	if res.Error != nil {
		// If error is because unique contraint failed.. user already exists
		if res.Error == gorm.ErrDuplicatedKey {
			slog.Error("Registration failed", "error", res.Error.Error(), "error_pretty", "User already exists")
			return AuthResponse{}, errors.New("User already exists")
		}
		slog.Error("Registration failed", "error", err, "error_pretty", "Watcharr does not know why this failed, assume database operation failed")
		return AuthResponse{}, errors.New("unknown error")
	}

	// Gorm fills our user obj with the ID from db after insert,
	// just ensure it actually has.
	if user.ID == 0 {
		slog.Error("user.ID not filled out after registration", "userId", user.ID)
		return AuthResponse{}, errors.New("failed to get user id, try login")
	}

	token, err := s.signJWT(user.ID, user.Username, user.Type)
	if err != nil {
		slog.Error("Registration: Failed to sign new jwt", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func (s *Service) RegisterFirstUser(urr *UserRegisterRequest) (AuthResponse, error) {
	// Ensure no users exist
	var userCount int64
	uresp := s.db.Model(&entity.User{}).Count(&userCount)
	if uresp.Error != nil {
		slog.Error("registerFirstUser: User count query failed!", "error", uresp.Error)
		return AuthResponse{}, errors.New("failed to query db for a count of users")
	}
	if userCount != 0 {
		slog.Warn("registerFirstUser: registered users already exist.")
		return AuthResponse{}, errors.New("first user already registered")
	}
	slog.Info("Registering first user.")
	return s.Register(urr, entity.PERM_ADMIN)
}

func (s *Service) Login(userL *entity.User) (AuthResponse, error) {
	slog.Debug("A User Is Logging In", "username", userL.Username)

	type WatcharrUser struct {
		ID       uint
		Username string
		Password string
	}

	dbUser := new(WatcharrUser)
	res := s.db.
		Model(&entity.User{}).
		Where("username = ? AND (type IS NULL OR type = 0)", userL.Username).
		Take(&dbUser)
	if res.Error != nil {
		slog.Error("Failed to select user from database for login", "error", res.Error)
		return AuthResponse{}, errors.New("User does not exist")
	}

	match, err := s.compareHash(userL.Password, dbUser.Password)
	if err != nil {
		slog.Error("Failed to compare pass to hash for login", "error", err)
		return AuthResponse{}, errors.New("failed to login")
	}
	if !match {
		slog.Error("User failed to provide correct password for login", "hash_matched", match)
		return AuthResponse{}, errors.New("incorrect details")
	}

	token, err := s.signJWT(dbUser.ID, dbUser.Username, entity.WATCHARR_USER)
	if err != nil {
		slog.Error("Failed to sign new jwt!", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func (s *Service) LoginJellyfin(userL *entity.User) (AuthResponse, error) {
	if s.cfg.JELLYFIN_HOST == "" {
		slog.Error("Request made to login via Jellyfin, but JELLYFIN_HOST has not been configured.")
		return AuthResponse{}, errors.New("jellyfin login not enabled")
	}

	base, err := url.Parse(s.cfg.JELLYFIN_HOST + "/Users/AuthenticateByName")
	if err != nil {
		slog.Error("Failed to parse AuthenticateByName api endpoint url", "error", err.Error())
		return AuthResponse{}, errors.New("failed to parse api uri")
	}

	// Marshall struct as json
	usrJSON, err := json.Marshal(JellyfinAuth{Username: userL.Username, Pw: userL.Password})
	if err != nil {
		slog.Error("Error marshalling JellyfinAuth JSON", "error", err.Error())
		return AuthResponse{}, errors.New("failed to marshal json")
	}
	// Run auth request
	client := &http.Client{}
	req, err := http.NewRequest("POST", base.String(), bytes.NewBuffer(usrJSON))
	if err != nil {
		slog.Error("Creating request to jellyfin for auth failed", "error", err)
		return AuthResponse{}, errors.New("request failed")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Emby-Authorization", "MediaBrowser Client=\"Watcharr\", Device=\"HTTP\", DeviceId=\"WatcharrFor"+userL.Username+"\", Version=\"10.8.0\"")
	res, err := client.Do(req)
	if err != nil {
		slog.Error("making request to jellyfin for auth failed", "error", err)
		return AuthResponse{}, errors.New("request failed")
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		slog.Error("Error reading jellyfin auth response", "error", err.Error())
		return AuthResponse{}, err
	}
	if res.StatusCode != 200 {
		slog.Error("Jellyfin auth non 200 status code", "status_code", res.StatusCode, "error", string(body))
		return AuthResponse{}, errors.New("incorrect details")
	}
	// Process auth response
	resp := new(JellyfinAuthResponse)
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return AuthResponse{}, errors.New("failed to process response")
	}
	if resp.User.ID == "" {
		return AuthResponse{}, errors.New("jellyfin returned empty user id")
	}

	dbUser := new(entity.User)

	userID, err := s.GetUserIDFromServiceClientId(
		entity.UserServiceNameJellyfin,
		resp.User.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If the userID cant be found because the service doesn't exist,
			// then the user must not exist either, so create one.
			err := s.createJellyfinUser(dbUser, *resp)
			if err != nil {
				slog.Error("LoginJellyfin: Creating user failed.", "error", err)
				return AuthResponse{},
					errors.New("failed to create new user from jellyfin")
			}
		} else {
			return AuthResponse{}, errors.New("couldn't find a user_id")
		}
	} else {
		dbRes := s.db.
			Where("id = ? AND type = ?",
				userID,
				entity.JELLYFIN_USER).
			Preload("UserServices").
			Take(dbUser)
		if dbRes.Error != nil {
			return AuthResponse{}, errors.New("error locating user in db")
		} else if resp.AccessToken != "" {
			// We got the user.. update their access token in db
			slog.Debug("LoginJellyfin: Updating user with new access token")
			for i, v := range dbUser.UserServices {
				if v.Name == entity.UserServiceNameJellyfin {
					slog.Info("LoginJellyfin: Found jellyfin user service.. attemping to update")
					dbUser.UserServices[i].AuthToken = resp.AccessToken
					break
				}
			}
			s.db.Save(&dbUser.UserServices)
		}
	}

	token, err := s.signJWT(dbUser.ID, dbUser.Username, entity.JELLYFIN_USER)
	if err != nil {
		slog.Error("LoginJellyfin: Failed to sign new jwt!", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

// Create Watcharr user from jellyfin auth response.
// Only to be ran after verifying the user doesn't exist.
func (s *Service) createJellyfinUser(
	dbUser *entity.User,
	resp JellyfinAuthResponse,
) error {
	if dbUser == nil {
		slog.Error("createJellyfinUser: dbUser is nil")
		return errors.New("dbUser is nil")
	}
	dbUser.UserServices = append(dbUser.UserServices, entity.UserServices{
		Name:      entity.UserServiceNameJellyfin,
		ClientID:  resp.User.ID,
		AuthToken: resp.AccessToken,
	})
	dbUser.Username = resp.User.Name
	dbUser.Type = entity.JELLYFIN_USER
	dbUser.Country = &s.cfg.DEFAULT_COUNTRY

	err := s.db.Create(&dbUser).Error
	if err != nil {
		slog.Error("createJellyfinUser: Create Failed!", "error", err)
		return err
	}
	return nil
}

// Login via Plex.
func (s *Service) LoginPlex(lr *plex.PlexLoginRequest) (AuthResponse, error) {
	if s.cfg.PLEX_HOST == "" || s.cfg.PLEX_MACHINE_ID == "" {
		slog.Error("Request made to login via Plex, but Plex authentication is disabled")
		return AuthResponse{}, errors.New("plex login not enabled")
	}
	slog.Debug("A Plex User Is Logging In")
	account, err := s.plexProvider.FetchPlexAccountFromToken(lr.AuthToken)
	if err != nil {
		slog.Error("loginPlex: Could not fetch Plex account", "error", err)
		return AuthResponse{}, errors.New("could not fetch plex acount")
	}
	if account.Username == "" || account.Id == 0 {
		slog.Error("loginPlex: Username or id missing from account response:", "username", account.Username, "id", account.Id)
		return AuthResponse{}, errors.New("data is missing from the plex account response")
	}
	// Get users auth token against our home plex server.
	// If no auth token, assume they don't have access to our plex server.
	homeAuthToken, err := s.plexProvider.GetPlexHomeServerAuthToken(lr.AuthToken, lr.ClientIdentifier)
	if err != nil || homeAuthToken == "" {
		slog.Error("loginPlex: Failed to get home server auth token for user! If not because the request failed, then ensure the user has access to our home servers library.", "error", err)
		return AuthResponse{}, errors.New("failed to verify plex access")
	}
	dbUser := new(entity.User)
	userIdQ := s.db.Select("user_id").Where("name = ? AND client_id = ?", "plex", account.Id).Table("user_services")
	dbRes := s.db.Where("type = ?", entity.PLEX_USER).Where("id = (?)", userIdQ).Preload("UserServices").Take(&dbUser)
	if dbRes.Error != nil {
		if errors.Is(dbRes.Error, gorm.ErrRecordNotFound) {
			slog.Debug("loginPlex: New plex user attempted login.. creating Watcharr account now.")
			dbUser.Username = account.Username
			dbUser.Type = entity.PLEX_USER
			dbUser.UserServices = append(dbUser.UserServices, entity.UserServices{
				Name:       "plex",
				ClientID:   strconv.FormatUint(account.Id, 10),
				AuthToken:  lr.AuthToken,
				AuthToken2: homeAuthToken,
			})
			dbUser.Country = &s.cfg.DEFAULT_COUNTRY
			dbRes = s.db.Create(&dbUser)
			if dbRes.Error != nil {
				slog.Error("loginPlex: Failed to create new user in db from plex response", "error", dbRes.Error)
				return AuthResponse{}, errors.New("failed to create new user from plex")
			}
		} else {
			slog.Error("loginPlex: Failed to select user from database for login", "error", dbRes.Error)
			return AuthResponse{}, errors.New("failed to locate user")
		}
	} else {
		// If user exists.. update their access tokens in db
		for i, v := range dbUser.UserServices {
			if v.Name == "plex" {
				slog.Info("loginPlex: Found plex user service.. attemping to update")
				dbUser.UserServices[i].AuthToken = lr.AuthToken
				if homeAuthToken != "" {
					dbUser.UserServices[i].AuthToken2 = homeAuthToken
				}
				break
			}
		}
		s.db.Save(&dbUser.UserServices)
	}
	token, err := s.signJWT(dbUser.ID, dbUser.Username, entity.PLEX_USER)
	if err != nil {
		slog.Error("loginPlex: Failed to sign new jwt!", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

// TODO the logic that gets and validated a token should be moved to Token service.
func (s *Service) UseAdminToken(req *UseAdminTokenRequest, userId uint) error {
	var dbToken entity.Token
	resp := s.db.Where("value = ?", req.Token).Take(&dbToken)
	if resp.Error != nil {
		slog.Info("useAdminToken failed", "error", "token not found in db")
		return errors.New("invalid token")
	}
	if dbToken.Type != entity.TOKENTYPE_ADMIN {
		slog.Info("useAdminToken failed", "error", "token is of wrong type", "type_wanted", entity.TOKENTYPE_ADMIN, "type_actual", dbToken.Type)
		return errors.New("invalid token")
	}
	dur := time.Since(dbToken.CreatedAt)
	if dur > token.TokenMaxAge {
		slog.Info("useAdminToken failed", "error", "token in db has expired")
		return errors.New("invalid token")
	}
	if dbToken.UserID != userId {
		slog.Info("useAdminToken failed", "error", "token in db is not for this user")
		return errors.New("invalid token")
	}
	// Token is valid and for current user.. give user admin.
	// Incase removing the token after used fails, this is in a transaction so user wont be admin.
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Give user admin
		if err := tx.Model(&entity.User{}).Where("id = ?", userId).Update("permissions", entity.PERM_ADMIN).Error; err != nil {
			return err
		}
		// Delete used token
		if err := tx.Where("value = ?", req.Token).Delete(&entity.Token{}).Error; err != nil {
			return err
		}
		// commit transaction if no errors
		return nil
	})
	if err != nil {
		slog.Info("useAdminToken failed", "error", err, "error_pretty", "using token transaction failed")
		return errors.New("failed to use token")
	}
	return nil
}

// Get user_id from UserServices by service name & client_id.
func (s *Service) GetUserIDFromServiceClientId(
	name entity.UserServiceName,
	cid string,
) (uint, error) {
	if name == "" || cid == "" {
		slog.Error("GetUserIDFromServiceClientId: name or cid empty",
			"name", name,
			"cid", cid)
		return 0, errors.New("name or cid is empty")
	}
	var userId uint = 0
	err := s.db.
		Model(&entity.UserServices{}).
		Select("user_id").
		Where("name = ? AND client_id = ?", name, cid).
		Scan(&userId).
		Error
	if err != nil {
		slog.Error("GetUserIDFromServiceClientId: Lookup failed.", "error", err)
		// This importantly should always return `err` from gorm as dependants
		// check for type of error.
		return 0, err
	}
	if userId == 0 {
		slog.Error("GetUserIDFromServiceClientId: No user id.")
		return 0, errors.New("didn't get a valid user id")
	}
	return userId, nil
}

func (s *Service) signJWT(
	userID uint,
	userName string,
	userType entity.UserType,
) (string, error) {
	slog.Debug("signJWT: Signing a new token.",
		"user_id", userID, "user_name", userName, "user_type", userType)

	if userID == 0 {
		return "", errors.New("no userID provided")
	} else if userName == "" {
		return "", errors.New("no userName provided")
	}

	// Create new jwt with claim data
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.TokenClaims{
		UserID:   userID,
		Username: userName,
		Type:     userType,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer:   "watcharr",
		},
	})

	// Sign and get the complete encoded token as a string using the secret
	return jwt.SignedString([]byte(s.cfg.JWT_SECRET))
}

func (s *Service) hashPassword(password string, p *entity.ArgonParams) (encodedHash string, err error) {
	salt, err := s.generateRandomBytes(p.SaltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format hash in standard way.
	encodedHash = fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.Memory,
		p.Iterations,
		p.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func (s *Service) generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *Service) compareHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := s.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func (s *Service) decodeHash(encodedHash string) (p *entity.ArgonParams, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("the encoded hash is not in the correct format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errors.New("incompatible version of argon2")
	}

	p = &entity.ArgonParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func (s *Service) UserChangePassword(p UserPasswordUpdateRequest, userId uint) error {
	slog.Debug("UserChangePassword: Running.", "user_id", userId)

	type User struct {
		ID       uint
		Password string
		Type     entity.UserType
	}

	user := new(User)
	res := s.db.
		Model(&entity.User{}).
		Where("id = ?", userId).
		Take(&user)
	if res.Error != nil {
		slog.Error("UserChangePassword: Failed to retrieve user from database!",
			"user_id", userId, "error", res.Error)
		return errors.New("failed to retrieve user")
	}
	if user.Type != entity.WATCHARR_USER {
		slog.Error("UserChangePassword: Only Watcharr users can change their password!",
			"user_id", userId, "user_type", user.Type)
		return errors.New("incorrect user type")
	}
	slog.Debug("UserChangePassword: User found.", "user_id", userId)

	match, err := s.compareHash(p.OldPassword, user.Password)
	if err != nil {
		slog.Error("UserChangePassword: Failed to compare passwords!",
			"user_id", userId, "error", err)
		return errors.New("failed to compare passwords")
	}
	if !match {
		slog.Error("UserChangePassword: Passwords do not match.",
			"user_id", userId, "error", err)
		return errors.New("password incorrect")
	}
	slog.Debug("UserChangePassword: Password matched. Hashing new pass.",
		"user_id", userId)

	hash, err := s.hashPassword(p.NewPassword, entity.GetPassArgonParams())
	if err != nil {
		slog.Error("UserChangePassword: Failed to hash new password!",
			"user_id", userId, "error", err)
		return errors.New("failed to hash new password")
	}
	slog.Debug("UserChangePassword: New password hashed.", "user_id", userId)
	err = s.db.
		Model(&entity.User{}).
		Where("id = ?", userId).
		Update("password", hash).
		Error
	if err != nil {
		slog.Error("UserChangePassword: Update password query failed!",
			"user_id", userId, "error", err)
		return errors.New("failed to update password")
	}
	slog.Debug("UserChangePassword: Password updated.", "user_id", userId)

	return nil
}

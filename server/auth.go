package main

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

type UserType uint8

var (
	// Assume watcharr user if none of these...
	JELLYFIN_USER UserType = 1
	PLEX_USER     UserType = 2
)

// User Perms
// iota auto increments for us so when adding new
// perms, add to bottom as to not change other perm
// values.
const (
	PERM_NONE int = 1 << iota
	PERM_ADMIN
	PERM_REQUEST_CONTENT
)

// uniqueIndex applied between Username and UserType, so same usernames can exist, but only with different types.
// This is incase different users with same name from different services try to signup.
type User struct {
	GormModel
	Username string `gorm:"uniqueIndex:usr_name_to_type;not null" json:"username" binding:"required"`
	Password string `gorm:"not null" json:"password" binding:"required"`
	AvatarID uint   `json:"-"`
	Avatar   Image  `json:"avatar"`
	Bio      string `json:"bio"`
	// The type of user/which auth service they originate from.
	// Empty if from Watcharr, or the name of the service (eg. jellyfin)
	Type UserType `gorm:"uniqueIndex:usr_name_to_type;not null;default:0" json:"type"`
	// ID of user from the third party service, this will be used purely for lookup of user at signin.
	ThirdPartyID string `json:"-"`
	// Auth token from third party (jellyfin)
	ThirdPartyAuth string `json:"-"`
	Watched        []Watched
	Permissions    int `gorm:"default:1" json:"-"`
	// All user settings cols, in another struct for reusability
	UserSettings
}

func (u *User) GetSafe() PublicUser {
	return PublicUser{
		ID:       u.ID,
		Username: u.Username,
		Avatar:   u.Avatar,
		Bio:      u.Bio,
	}
}

// This struct uses pointer to the values, so in update user settings,
// we can tell which setting is being updated (if not nil..).
type UserSettings struct {
	// Is profile private
	Private *bool `gorm:"default:false" json:"private"`
	// Are watched list content thoughts public (profile must also be public is false)
	PrivateThoughts *bool `gorm:"default:false" json:"privateThoughts"`
	// If ui 'spoilers' should be shown
	HideSpoilers *bool `gorm:"default:false" json:"hideSpoilers"`
	// If user wants previously watched items to show in 'Finished' filter,
	// even if the watched item state has since been changed.
	// Also if user wants to show in watched stats.
	IncludePreviouslyWatched *bool `gorm:"default:false" json:"includePreviouslyWatched"`
}

// We use a separate struct for registration to avoid confusion
// and possible accidents where we allow a user to pass in a
// property from the main User struct that shouldn't be allowed.
type UserRegisterRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	Type           UserType
	ThirdPartyAuth string
}

type PlexUserRequest struct {
	AuthToken string `json:"authtoken" binding:"required"`
}

type PlexUser struct {
	Id       uint32 `json:"id" binding:"required"`
	Uuid     string `json:"uuid" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type PlexUserResponse struct {
	Plexuser PlexUser `json:"user" binding:"required"`
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

type AvailableAuthProvidersResponse struct {
	AvailableAuthProviders []string `json:"available"`
	SignupEnabled          bool     `json:"signupEnabled"`
	IsInSetup              bool     `json:"isInSetup"`
	PlexOauthId            string   `json:"plexOauthId"`
}

type ArgonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func GetPassArgonParams() *ArgonParams {
	return &ArgonParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
}

type TokenClaims struct {
	UserID   uint     `json:"userId"`
	Username string   `json:"username"`
	Type     UserType `json:"type"`
	jwt.RegisteredClaims
}

type UserPasswordUpdateRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// Auth middleware
// If db is passed, extra user info from the database will be fetched.
func AuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Debug("AuthRequired middleware hit")
		atoken := c.GetHeader("Authorization")
		// Make sure auth header isn't empty
		if atoken == "" {
			slog.Warn("Returning 401, Authorization header not provided")
			c.AbortWithStatus(401)
			return
		}
		// Parse token
		token, err := jwt.ParseWithClaims(atoken, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			slog.Error("AuthRequired failed to parse token", "error", err)
			c.AbortWithStatus(401)
			return
		}
		// If token is valid, go to next handler
		if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
			// Check if token issuedAt is from before `timeOfNewLoginRequired`.
			// Basically just so we can logout old tokens and force relogin...
			// since new changes require the user login again.
			timeOfNewLoginRequired, _ := time.Parse(time.RFC822, "18 Aug 23 20:30 UTC")
			if claims.IssuedAt.Before(timeOfNewLoginRequired) {
				slog.Info("Token is from before timeOfNewLoginRequired.. returning 401", "token_issued_at", claims.IssuedAt, "time_of_new_login_required", timeOfNewLoginRequired)
				c.AbortWithStatus(401)
				return
			}
			slog.Debug("Token is valid", "claims", claims)
			c.Set("userId", claims.UserID)
			c.Set("userType", claims.Type)
			// If db passed, get extra user info and set as variables in req context
			if db != nil {
				slog.Debug("AuthRequired: db passed.. getting extra user info")
				dbUser := new(User)
				res := db.Where("id = ?", claims.UserID).Take(&dbUser)
				if res.Error != nil {
					slog.Error("AuthRequired: Failed to select user from database", "error", res.Error)
					c.AbortWithStatus(401)
					return
				}
				slog.Debug("AuthRequired: fetched extra user info. Setting vars.", "userThirdPartyId", dbUser.ThirdPartyID, "userThirdPartyAuth", "lol this is censored dude")
				c.Set("userThirdPartyId", dbUser.ThirdPartyID)
				c.Set("userThirdPartyAuth", dbUser.ThirdPartyAuth)
				c.Set("username", dbUser.Username)
				c.Set("userPermissions", dbUser.Permissions)
			}
			c.Next()
		} else {
			slog.Error("Token is **not** valid")
			c.AbortWithStatus(401)
			return
		}
	}
}

// Admin only middleware (use after AuthRequired with extra info!)
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetUint("userId")
		perms := c.GetInt("userPermissions")
		if hasPermission(perms, PERM_ADMIN) {
			slog.Debug("AdminRequired: User has permission to access admin only route", "user_id", userId)
			c.Next()
			return
		}
		slog.Info("AdminRequired: User denied permission to access admin only route", "user_id", userId)
		c.AbortWithStatus(401)
	}
}

func register(ur *UserRegisterRequest, initialPerm int, db *gorm.DB) (AuthResponse, error) {
	if !Config.SIGNUP_ENABLED {
		slog.Warn("Register called, but signing up is disabled.")
		return AuthResponse{}, errors.New("registering is disabled")
	}
	var user User = User{Username: ur.Username, Password: ur.Password, Type: ur.Type, ThirdPartyAuth: ur.ThirdPartyAuth}
	slog.Info("A user is registering", "username", user.Username)
	hash, err := hashPassword(user.Password, GetPassArgonParams())
	if err != nil {
		log.Fatal(err)
	}

	// Update user obj to replace the plaintext pass with hash
	user.Password = hash

	// Update user permissions if an initial perm is passed in (1 is default)
	if initialPerm != 0 && initialPerm != PERM_NONE {
		slog.Info("User being registered has been given extra initial permissions", "initial_perm", initialPerm)
		user.Permissions = initialPerm
	}

	res := db.Create(&user)
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

	token, err := signJWT(&user)
	if err != nil {
		slog.Error("Registration: Failed to sign new jwt", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func registerFirstUser(user *UserRegisterRequest, db *gorm.DB) (AuthResponse, error) {
	// Ensure no users exist
	var userCount int64
	uresp := db.Model(&User{}).Count(&userCount)
	if uresp.Error != nil {
		slog.Error("registerFirstUser: User count query failed!", "error", uresp.Error)
		return AuthResponse{}, errors.New("failed to query db for a count of users")
	}
	if userCount != 0 {
		slog.Warn("registerFirstUser: registered users already exist.")
		return AuthResponse{}, errors.New("first user already registered")
	}
	slog.Info("Registering first user.")
	return register(user, PERM_ADMIN, db)
}

func login(user *User, db *gorm.DB) (AuthResponse, error) {
	slog.Debug("A User Is Logging In", "username", user.Username)
	dbUser := new(User)
	res := db.Where("username = ? AND (type IS NULL OR type = 0)", user.Username).Take(&dbUser)
	if res.Error != nil {
		slog.Error("Failed to select user from database for login", "error", res.Error)
		return AuthResponse{}, errors.New("User does not exist")
	}

	match, err := compareHash(user.Password, dbUser.Password)
	if err != nil {
		slog.Error("Failed to compare pass to hash for login", "error", err)
		return AuthResponse{}, errors.New("failed to login")
	}
	if !match {
		slog.Error("User failed to provide correct password for login", "hash_matched", match)
		return AuthResponse{}, errors.New("incorrect details")
	}

	token, err := signJWT(dbUser)
	if err != nil {
		slog.Error("Failed to sign new jwt", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func loginJellyfin(user *User, db *gorm.DB) (AuthResponse, error) {
	if Config.JELLYFIN_HOST == "" {
		slog.Error("Request made to login via Jellyfin, but JELLYFIN_HOST has not been configured.")
		return AuthResponse{}, errors.New("jellyfin login not enabled")
	}

	base, err := url.Parse(Config.JELLYFIN_HOST + "/Users/AuthenticateByName")
	if err != nil {
		slog.Error("Failed to parse AuthenticateByName api endpoint url", "error", err.Error())
		return AuthResponse{}, errors.New("failed to parse api uri")
	}

	// Marshall struct as json
	usrJSON, err := json.Marshal(JellyfinAuth{Username: user.Username, Pw: user.Password})
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
	req.Header.Add("X-Emby-Authorization", "MediaBrowser Client=\"Watcharr\", Device=\"HTTP\", DeviceId=\"WatcharrFor"+user.Username+"\", Version=\"10.8.0\"")
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

	dbUser := new(User)
	dbRes := db.Where("third_party_id = ?", resp.User.ID).Take(&dbUser)
	if dbRes.Error != nil {
		if errors.Is(dbRes.Error, gorm.ErrRecordNotFound) {
			// Record not found, so we should create the user
			// dbUser will be empty, so we can just reuse it for this purpose.
			dbUser.ThirdPartyID = resp.User.ID
			dbUser.ThirdPartyAuth = resp.AccessToken
			dbUser.Username = resp.User.Name
			dbUser.Type = JELLYFIN_USER

			dbRes = db.Create(&dbUser)
			if dbRes.Error != nil {
				slog.Error("Failed to create new user in db from jellyfin response", "error", dbRes.Error)
				return AuthResponse{}, errors.New("failed to create new user from jellyfin")
			}
		} else {
			return AuthResponse{}, errors.New("error locating user in db")
		}
	}
	// If user exists.. update their access token in db
	if resp.AccessToken != "" {
		slog.Debug("Jellyfin user login - updating user with new access token")
		dbUser.ThirdPartyAuth = resp.AccessToken
		db.Save(&dbUser)
	}

	token, err := signJWT(dbUser)
	if err != nil {
		slog.Error("Failed to sign new (jellyfin login) jwt", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func fetchPlexUsernameFromToken(token string) (string, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://plex.tv/users/account.json", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Plex-Token", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var plexUserResponse PlexUserResponse
	err = json.Unmarshal(body, &plexUserResponse)
	if err != nil {
		return "", err
	}
	return plexUserResponse.Plexuser.Username, nil
}

func registerPlex(plexUserRequest *PlexUserRequest, db *gorm.DB) (AuthResponse, error) {
	if Config.PLEX_OAUTH_ID == "" {
		slog.Error("Request made to regester via Plex, but Plex authentication is disabled")
		return AuthResponse{}, errors.New("plex login not enabled")
	}
	slog.Debug("A Plex User Is Registering", "authtoken", plexUserRequest.AuthToken)
	username, err := fetchPlexUsernameFromToken(plexUserRequest.AuthToken)
	if err != nil {
		slog.Error("Could not fetch Plex information", "error", err)
		return AuthResponse{}, errors.New("could not fetch plex metadata")
	}
	user := UserRegisterRequest{Username: username, Type: PLEX_USER, ThirdPartyAuth: plexUserRequest.AuthToken}
	response, err := register(&user, PERM_NONE, db)
	if err != nil {
		slog.Error("Could not register new Plex user", "error", err)
		return AuthResponse{}, errors.New("could not register plex user")
	}
	return response, nil
}

func loginPlex(plexUserRequest *PlexUserRequest, db *gorm.DB) (AuthResponse, error) {
	if Config.PLEX_OAUTH_ID == "" {
		slog.Error("Request made to login via Plex, but Plex authentication is disabled")
		return AuthResponse{}, errors.New("plex login not enabled")
	}
	slog.Debug("A Plex User Is Logging In", "authtoken", plexUserRequest.AuthToken)
	username, err := fetchPlexUsernameFromToken(plexUserRequest.AuthToken)
	if err != nil {
		slog.Error("loginPlex: Could not fetch Plex username", "error", err)
		return AuthResponse{}, errors.New("could not fetch plex username")
	}
	dbUser := new(User)
	res := db.Where("username = ? AND type = ?", username, PLEX_USER).Take(&dbUser)
	if res.Error != nil {
		slog.Error("Failed to select user from database for login", "error", res.Error)
		return AuthResponse{}, errors.New("User does not exist")
	}
	res = db.Model(&dbUser).Where("username = ? AND type = ?", username, PLEX_USER).Update("third_party_auth", plexUserRequest.AuthToken)
	if res.Error != nil {
		slog.Error("Failed to update the user's Plex token, syncs may fail if the token has expired", "username", username, "error", res.Error)
	}
	token, err := signJWT(dbUser)
	if err != nil {
		slog.Error("Failed to sign new jwt", "error", err)
		return AuthResponse{}, errors.New("failed to get auth token")
	}
	return AuthResponse{Token: token}, nil
}

func useAdminToken(req *UseAdminTokenRequest, db *gorm.DB, userId uint) error {
	var dbToken Token
	resp := db.Where("value = ?", req.Token).Take(&dbToken)
	if resp.Error != nil {
		slog.Info("useAdminToken failed", "error", "token not found in db")
		return errors.New("invalid token")
	}
	if dbToken.Type != TOKENTYPE_ADMIN {
		slog.Info("useAdminToken failed", "error", "token is of wrong type", "type_wanted", TOKENTYPE_ADMIN, "type_actual", dbToken.Type)
		return errors.New("invalid token")
	}
	dur := time.Since(dbToken.CreatedAt)
	if dur > tokenMaxAge {
		slog.Info("useAdminToken failed", "error", "token in db has expired")
		return errors.New("invalid token")
	}
	if dbToken.UserID != userId {
		slog.Info("useAdminToken failed", "error", "token in db is not for this user")
		return errors.New("invalid token")
	}
	// Token is valid and for current user.. give user admin.
	// Incase removing the token after used fails, this is in a transaction so user wont be admin.
	err := db.Transaction(func(tx *gorm.DB) error {
		// Give user admin
		if err := tx.Model(&User{}).Where("id = ?", userId).Update("permissions", PERM_ADMIN).Error; err != nil {
			return err
		}
		// Delete used token
		if err := tx.Where("value = ?", req.Token).Delete(&Token{}).Error; err != nil {
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

func signJWT(user *User) (token string, err error) {
	// Create new jwt with claim data
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenClaims{
		user.ID,
		user.Username,
		user.Type,
		jwt.RegisteredClaims{
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer:   "watcharr",
		},
	})

	// Sign and get the complete encoded token as a string using the secret
	return jwt.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func hashPassword(password string, p *ArgonParams) (encodedHash string, err error) {
	salt, err := generateRandomBytes(p.saltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format hash in standard way.
	encodedHash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.iterations, p.parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func compareHash(password, encodedHash string) (match bool, err error) {
	// Extract the parameters, salt and derived key from the encoded password
	// hash.
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Derive the key from the other password using the same parameters.
	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	// Check that the contents of the hashed passwords are identical. Note
	// that we are using the subtle.ConstantTimeCompare() function for this
	// to help prevent timing attacks.
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func decodeHash(encodedHash string) (p *ArgonParams, salt, hash []byte, err error) {
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

	p = &ArgonParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func hasPermission(perms int, reqPerm int) bool {
	// Admins have permission for everything.
	if perms&PERM_ADMIN == PERM_ADMIN {
		return true
	}
	return (perms & reqPerm) == reqPerm
}

func userChangePassword(db *gorm.DB, pwds UserPasswordUpdateRequest, userId uint) error {
	slog.Debug("userChangePassword request running", "user_id", userId)
	user := new(User)
	res := db.Where("id = ?", userId).Select("password").Take(&user)
	if res.Error != nil {
		slog.Error("userChangePassword failed - failed to retrieve user from database", "user_id", userId, "error", res.Error)
		return errors.New("failed to retrieve user")
	}
	slog.Debug("userChangePassword user found", "user_id", userId)
	match, err := compareHash(pwds.OldPassword, user.Password)
	if err != nil {
		slog.Error("userChangePassword failed - failed to compare passwords", "user_id", userId, "error", err)
		return errors.New("failed to compare passwords")
	}
	if !match {
		slog.Error("userChangePassword failed - current password hash doesn't match password hash in database", "user_id", userId, "error", err)
		return errors.New("current password provided doesn't match password in database")
	}
	slog.Debug("userChangePassword hash for current password matches hash in the database", "user_id", userId)
	slog.Debug("userChangePassword hashing new password", "user_id", userId)
	hash, err := hashPassword(pwds.NewPassword, GetPassArgonParams())
	if err != nil {
		slog.Error("userChangePassword failed - failed to hash new password", "user_id", userId, "error", err)
		return errors.New("failed to hash new password")
	}
	slog.Debug("userChangePassword new password hashed", "user_id", userId)
	if err := db.Model(&User{}).Where("id = ?", userId).Update("password", hash).Error; err != nil {
		slog.Error("userChangePassword failed - failed to update password in database", "user_id", userId, "error", err)
		return errors.New("failed to update password")
	} else {
		slog.Debug("userChangePassword password updated", "user_id", userId)
	}
	return nil
}

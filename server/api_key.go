package main

import (
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "log/slog"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "gorm.io/gorm"
)

// APIKey represents a user-generated programmatic access key.
// Keys are stored hashed to avoid leaking plaintext if DB is exposed.
// We store the plaintext only once on creation and return it to caller.
//
// NOTE: The hash is a simple SHA-256 of the key. For Bearer-style secrets
//       this is sufficient.
//
// A unique index on KeyHash prevents duplicates.
type APIKey struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    LastUsed  time.Time `json:"lastUsed"`
    // Hex-encoded SHA-256 of the key
    KeyHash string `gorm:"uniqueIndex;not null" json:"-"`
    UserID  uint   `gorm:"not null" json:"userId"`
    Revoked bool   `gorm:"default:false" json:"revoked"`
}

// createAPIKey generates and stores a new key for the given user and returns the
// plaintext (only shown once!).
func createAPIKey(db *gorm.DB, userId uint) (string, error) {
    // Generate 32-byte random key (URL-safe base64 length ~43 chars)
    raw, err := generateString(32)
    if err != nil {
        return "", err
    }
    keyHash := sha256Hex(raw)

    if err := db.Create(&APIKey{KeyHash: keyHash, UserID: userId}).Error; err != nil {
        slog.Error("createAPIKey: db insert failed", "error", err)
        return "", errors.New("failed to create key")
    }
    return raw, nil
}

func getAPIKeys(db *gorm.DB, userId uint) ([]APIKey, error) {
    var keys []APIKey
    if err := db.Where("user_id = ?", userId).Find(&keys).Error; err != nil {
        slog.Error("getAPIKeys: query failed", "error", err)
        return nil, errors.New("failed to fetch keys")
    }
    return keys, nil
}

func revokeAPIKey(db *gorm.DB, userId, id uint) error {
    res := db.Model(&APIKey{}).Where("id = ? AND user_id = ?", id, userId).Update("revoked", true)
    if res.Error != nil {
        slog.Error("revokeAPIKey: update failed", "error", res.Error)
        return errors.New("failed to revoke key")
    }
    if res.RowsAffected == 0 {
        return errors.New("key not found")
    }
    return nil
}

// sha256Hex returns lowercase hex SHA-256 digest of input.
func sha256Hex(s string) string {
    h := sha256.Sum256([]byte(s))
    return hex.EncodeToString(h[:])
}

// validateAPIKey checks query param/ header key and, if valid, returns the owning user id.
// validateAPIKey checks db for the given plaintext key and returns its user id.
func validateAPIKey(db *gorm.DB, plaintext string) (uint, error) {
    var k APIKey
    h := sha256Hex(plaintext)
    res := db.Where("key_hash = ? AND revoked = 0", h).Take(&k)
    if res.Error != nil {
        return 0, errors.New("invalid api key")
    }
    // touch last_used
    db.Model(&APIKey{}).Where("id = ?", k.ID).Update("last_used", time.Now())
    return k.UserID, nil
}

// AuthOrAPIKey authenticates via JWT (Authorization header) or api_key query/header.
// decided to write this as an either/or scenario to avoid rewriting all endpoints, this is pretty standard though to use an API key for programmatic access with
// basically the same logic as JWT but with a different secret
func AuthOrAPIKey(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Try JWT first
        if tok := c.GetHeader("Authorization"); tok != "" {
            t, err := jwt.ParseWithClaims(tok, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
                return []byte(Config.JWT_SECRET), nil
            })
            if err == nil {
                if claims, ok := t.Claims.(*TokenClaims); ok && t.Valid {
                    c.Set("userId", claims.UserID)
                    c.Next()
                    return
                }
            }
        }
        // Fallback to api key
        k := c.Query("api_key")
        if k == "" {
            k = c.GetHeader("X-Api-Key")
        }
        if k == "" {
            c.AbortWithStatus(401)
            return
        }
        uid, err := validateAPIKey(db, k)
        if err != nil {
            c.AbortWithStatus(401)
            return
        }
        c.Set("userId", uid)
        c.Next()
    }
}

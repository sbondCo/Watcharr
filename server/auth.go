package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/uptrace/bun"
	"golang.org/x/crypto/argon2"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID       int    `bun:"id,pk,autoincrement" json:"id"`
	Username string `bun:"username,notnull,unique" json:"username" binding:"required"`
	Password string `bun:"password,notnull" json:"password" binding:"required"`
}

type RegisterResponse struct {
	Token string `json:"token"`
}

type ArgonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func register(user *User, db *bun.DB) (RegisterResponse, error) {
	println("Registering", user.Username)
	hash, err := hashPassword(user.Password, &ArgonParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Update user obj to replace the plaintext pass with hash
	user.Password = hash

	_, err = db.NewInsert().Model(user).Exec(context.TODO())
	if err != nil {
		// If error is because unique contraint failed.. user already exists
		if strings.Contains(err.Error(), "UNIQUE") {
			println(err.Error())
			return RegisterResponse{}, errors.New("User already exists")
		}
		panic(err)
	}

	return RegisterResponse{Token: "My JWT token"}, nil
}

func login(user *User, db *bun.DB) (RegisterResponse, error) {
	fmt.Println("Logging in", user.Username)
	dbUser := new(User)
	err := db.NewSelect().Model(dbUser).Where("username = ?", user.Username).Scan(context.TODO())
	if err != nil {
		fmt.Println("Failed to select user from database for login:", err)
		return RegisterResponse{}, errors.New("User does not exist")
	}
	fmt.Println(dbUser.ID, dbUser.Username, dbUser.Password)

	match, err := compareHash(user.Password, dbUser.Password)
	if err != nil {
		fmt.Println("Failed to compare pass to hash for login:", err)
		return RegisterResponse{}, errors.New("failed to login")
	}
	if !match {
		fmt.Println("User failed to provide correct password for login:", match)
		return RegisterResponse{}, errors.New("incorrect details")
	}

	return RegisterResponse{Token: "My JWT token"}, nil
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

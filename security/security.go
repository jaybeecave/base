package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"os"
	"strconv"

	"strings"

	"errors"

	"github.com/jaybeecave/base/datastore"
	uuid "github.com/satori/go.uuid"
)

const (
	// NoAuth value for no auth on route
	NoAuth = "none"
	// Redirect value will force the router to redirect to the /Login route
	Redirect = "redirect"
	// Disallow will return a 403 Forbbiden response
	Disallow = "disallow"

	// // Redis as the storage location for checking login valid
	// Redis = "Redis"
	// // Database as the storage location for checking login valid
	// Database = "Database"
)

type SessionUser struct {
	ID         int    `db:"id" json:"id"`
	Email      string `db:"email" json:"Email"`
	Password   string `db:"password" json:"Password"`
	CacheToken string `json:"CacheToken"`
}

type SessionToken struct {
	Token      string    `json:"Token"`
	Expiration time.Time `json:"Expiration"`
	CacheToken string    // dont return this as json `json:"CacheToken"`
}

type Padlock struct {
	Req   *http.Request
	Store *datastore.Datastore
}

func New(req *http.Request, store *datastore.Datastore) *Padlock {
	padlock := &Padlock{}
	padlock.Req = req
	padlock.Store = store
	return padlock
}

func (padlock *Padlock) LoginReturningToken(email string, password string, tableName string) (*SessionToken, error) {
	user := &SessionUser{}
	tableName = strings.ToLower(tableName)

	// valid names for table
	tableIDName := ""
	if tableName == "administrator" {
		tableIDName = "administrator_id"
	} else if tableName == "user" {
		tableIDName = "user_id"
	} else if tableName == "person" {
		tableIDName = "person_id"
	} else {
		return nil, errors.New("Invalid table name for security SessionUser table")
	}

	err := padlock.Store.DB.
		Select(tableIDName+" as id, email, password").
		From(tableName).
		Where("LOWER(email) = LOWER($1) and password = $2", email, password).
		Limit(1).
		QueryStruct(user)

	if err != nil {
		return nil, errors.New("Incorrect password")
	}

	uuid := uuid.NewV4().String() //key for redis or something needs to be part of the json package
	user.CacheToken = uuid
	// save the new sessionToken into the database so it can be cleared from the cache later if the user gets deleted

	jsonUser, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	sessionToken := &SessionToken{}
	sessionToken.Token = Encrypt(string(jsonUser))

	expirationInDays := 30 //default
	expirationDayEnv := os.Getenv("SECURITY_USER_TOKEN_EXPIRATION")
	if expirationDayEnv != "" {
		expirationInDays, _ = strconv.Atoi(expirationDayEnv) // if it can't convert then just use the default
	}

	duration := time.Duration(expirationInDays) * (24 * time.Hour)
	expiration := time.Now().Add(duration)
	sessionToken.Expiration = expiration
	sessionToken.CacheToken = user.CacheToken

	_, err = padlock.Store.DB.
		InsertInto("usersession_token").
		Columns("cache_token", "table_name", "record_id", "expiry_date").
		Values(user.CacheToken, tableName, user.ID, expiration).
		Exec()

	if err != nil {
		return nil, err
	}

	status := padlock.Store.Cache.Set(user.CacheToken, string(jsonUser), duration)

	if status.Err() != nil {
		return nil, err
	}

	// result, err := status.Result()
	// log.Info(result)

	// TODO HMAC AUTH
	return sessionToken, nil
}

func (padlock *Padlock) LoginReturningCookie(email string, password string, tableName string) (*http.Cookie, error) {
	tokenName := getSessionUserCookieName()

	token, err := padlock.LoginReturningToken(email, password, tableName) // same process / different format
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{Name: tokenName, Value: token.Token, Expires: token.Expiration, Path: "/", Domain: padlock.Req.Host}
	return cookie, nil
}

func (padlock *Padlock) LoggedInUser() (*SessionUser, error) {
	// check for basic authentication header
	authToken := ""

	// check the request header
	authHeader := padlock.Req.Header.Get("Authorization")
	if authHeader != "" {
		// potentially found a token in the Authorization header
		authTokenBits := strings.Split(authHeader, "Basic ")
		if len(authTokenBits) > 0 {
			authToken = authTokenBits[1]
		}
	}

	// we still haven't found the authtoken so try checking a cookie
	if authToken == "" {
		authToken = padlock.Req.URL.Query().Get("authtoken")
	}

	// we still haven't found the authtoken so try checking a cookie
	if authToken == "" {
		tokenName := getSessionUserCookieName()
		cookie, err := padlock.Req.Cookie(tokenName)
		if err != nil {
			if err.Error() == "http: named cookie not present" {
				return nil, errors.New("no auth details found in the request")
			}
			panic(err)
		}
		authToken = cookie.Value
	}
	user := &SessionUser{}

	//Decrypt the authToken
	val, err := Decrypt(authToken)
	if err != nil {
		return nil, err
	}

	// grab the resulting json object into a SessionUser struct
	err = json.Unmarshal([]byte(val), user)
	// log.Info("user is good?", user)
	if err != nil {
		return nil, err
	}

	if user.CacheToken == "" {
		return nil, errors.New("invalid token")
	}

	cachedUser := &SessionUser{}
	serializedUser, err := padlock.Store.Cache.Get(user.CacheToken).Result()

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(serializedUser), cachedUser)
	//log.Info("xx is good?", cachedUser)

	if err != nil {
		return nil, err
	}

	// awesome - you are logged in
	if cachedUser.Email == user.Email && cachedUser.Password == user.Password && cachedUser.ID == user.ID {
		return cachedUser, nil
	}
	return nil, errors.New("user didnt match cache... something funky here.")
}

func (padlock *Padlock) CheckLogin() (bool, error) {
	_, err := padlock.LoggedInUser()
	if err == nil {
		return true, nil
	}
	return false, err
}

// encrypt string to base64 crypto using AES
func Encrypt(text string) string {
	key := os.Getenv("SECURITY_ENCRYPTION_KEY")
	if key == "" {
		panic("SECURITY_ENCRYPTION_KEY not set")
	}

	// key := []byte(keyText)
	plaintext := []byte(text)
	cypherKey := []byte(key)

	block, err := aes.NewCipher(cypherKey)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// decrypt from base64 to decrypted string
func Decrypt(cryptoText string) (string, error) {
	key := os.Getenv("SECURITY_ENCRYPTION_KEY")
	cypherKey := []byte(key)
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)
	block, err := aes.NewCipher(cypherKey)
	if err != nil {
		return "", err
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext), nil
}

func GenerateRandomKey(length int) []byte {
	k := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil
	}
	return k
}

func getSessionUserCookieName() string {
	tokenName := os.Getenv("SECURITY_USER_COOKIE_NAME")
	if tokenName == "" {
		tokenName = "user_cookie"
	}
	return tokenName
}

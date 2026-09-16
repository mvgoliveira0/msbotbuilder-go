package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/infracloudio/msbotbuilder-go/connector/cache"
	"github.com/infracloudio/msbotbuilder-go/schema"
	"github.com/lestrrat-go/jwx/jwk"
)

var metadataURL = "https://login.botframework.com/v1/.well-known/openidconfiguration"

const fetchTimeout = 20

var authHeaderMatch = regexp.MustCompile("^(?:Bearer )?([A-Za-z0-9-_=]+\\.[A-Za-z0-9-_=]+\\.[A-Za-z0-9-_.+/=]*)$")

var httpClient *http.Client

func init() {

	httpClient = &http.Client{
		Timeout: fetchTimeout * time.Second,
	}
}

type TokenValidator interface {
	AuthenticateRequest(ctx context.Context, activity schema.Activity, authHeader string, credentials CredentialProvider, channelService string) (ClaimsIdentity, error)
}

type JwtTokenValidator struct {
	cache.AuthCache
}

func NewJwtTokenValidator() TokenValidator {
	return &JwtTokenValidator{cache.AuthCache{}}
}

func (jv *JwtTokenValidator) AuthenticateRequest(ctx context.Context, activity schema.Activity, authHeader string, credentials CredentialProvider, channelService string) (ClaimsIdentity, error) {

	match := authHeaderMatch.FindStringSubmatch(strings.TrimSpace(authHeader))
	if len(match) < 2 {
		if credentials.IsAuthenticationDisabled() {
			return nil, nil
		}
		return nil, errors.New("Unauthorized Access. Request is not authorized")
	}

	identity, err := jv.getIdentity(match[1])
	if err != nil || !identity.IsAuthenticated() {
		return nil, err
	}



	if identity.GetClaimValue("serviceurl") != activity.ServiceURL {
		return nil, errors.New("Unauthorized, service_url claim is invalid")
	}

	err = jv.validateIdentity(identity, credentials)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

func (jv *JwtTokenValidator) getIdentity(jwtString string) (ClaimsIdentity, error) {

	getKey := func(token *jwt.Token) (interface{}, error) {

		jwksURL, err := jv.getJwkURL(metadataURL)
		if err != nil {
			return nil, err
		}

	
		if jv.AuthCache.IsExpired() {
			ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout*time.Second)
			defer cancel()
			set, err := jwk.Fetch(ctx, jwksURL)
			if err != nil {
				return nil, err
			}
		
		
			jv.AuthCache = cache.AuthCache{
				Keys:   set,
				Expiry: time.Now().Add(time.Hour * 24 * 5),
			}
		}
		keyID, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("Expecting JWT header to have string kid")
		}
	
		key, ok := jv.AuthCache.Keys.(jwk.Set).LookupKeyID(keyID)
		if ok {
			var rawKey interface{}
			err := key.Raw(&rawKey)
			if err != nil {
				return nil, err
			}
			return rawKey, nil
		}

		return nil, errors.New("Could not find public key")
	}

	parser := &jwt.Parser{
		SkipClaimsValidation: true,
	}

	token, err := parser.Parse(jwtString, getKey)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("Unauthorized. Invalid token signature or format")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Unauthorized. Invalid token claims format")
	}

	now := time.Now().Unix()
	leeway := int64(120)


	if !claims.VerifyNotBefore(now+leeway, false) {
		return nil, errors.New("Token is not valid yet")
	}


	if !claims.VerifyExpiresAt(now-leeway, false) {
		return nil, errors.New("Token is expired")
	}


	if !claims.VerifyIssuedAt(now+leeway, false) {
		return nil, errors.New("Token used before issued")
	}

	alg := token.Header["alg"]
	isAllowed := func() bool {
		for _, allowed := range AllowedSigningAlgorithms {
			if allowed == alg {
				return true
			}
		}
		return false
	}()

	if !isAllowed {
		return nil, errors.New("Unauthorized. Invalid signing algorithm")
	}

	return NewClaimIdentity(claims, true), nil
}

func (jv *JwtTokenValidator) validateIdentity(identity ClaimsIdentity, credentials CredentialProvider) error {
	if identity.GetClaimValue(IssuerClaim) != ToBotFromChannelTokenIssuer {
		return errors.New("Unauthorized: invalid token issuer")
	}

	if !credentials.IsValidAppID(identity.GetClaimValue(AudienceClaim)) {
		return errors.New("Unauthorized: invalid AppId passed on token")
	}

	return nil
}

type metadata struct {
	JwksURI string `json:"jwks_uri"`
}

func (jv JwtTokenValidator) getJwkURL(metadataURL string) (string, error) {
	response, err := httpClient.Get(metadataURL)
	if err != nil {
		return "", errors.New("Error getting metadata document")
	}

	data := metadata{}
	err = json.NewDecoder(response.Body).Decode(&data)
	return data.JwksURI, err
}

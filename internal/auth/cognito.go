package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

const region = "ap-south-1"

const userPoolID = "ap-south-1_PvnFOJAx5"
const adminPoolID = "ap-south-1_RdvZbNKYZ"

type Identity struct {
	Sub    string `json:"sub"`
	Name   string `json:"name"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
	PoolID string `json:"poolId"`
}

type cognitoAttribute struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type getUserResponse struct {
	UserAttributes []cognitoAttribute `json:"UserAttributes"`
	Username       string             `json:"Username"`
}

type tokenClaims struct {
	Issuer string `json:"iss"`
}

func bearerToken(r *http.Request) (string, error) {
	value := r.Header.Get("Authorization")

	if !strings.HasPrefix(value, "Bearer ") {
		return "", errors.New("missing bearer token")
	}

	token := strings.TrimSpace(
		strings.TrimPrefix(value, "Bearer "),
	)

	if token == "" {
		return "", errors.New("empty bearer token")
	}

	return token, nil
}

func tokenPoolID(token string) (string, error) {
	parts := strings.Split(token, ".")

	if len(parts) < 2 {
		return "", errors.New("invalid token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}

	var claims tokenClaims

	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}

	if claims.Issuer == "" {
		return "", errors.New("token issuer missing")
	}

	pieces := strings.Split(claims.Issuer, "/")

	return pieces[len(pieces)-1], nil
}

func verifyWithCognito(token string) (*Identity, error) {
	body, err := json.Marshal(
		map[string]string{
			"AccessToken": token,
		},
	)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://cognito-idp."+region+".amazonaws.com/",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/x-amz-json-1.1",
	)

	req.Header.Set(
		"X-Amz-Target",
		"AWSCognitoIdentityProviderService.GetUser",
	)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("invalid or expired Cognito token")
	}

	var cognitoUser getUserResponse

	if err := json.Unmarshal(data, &cognitoUser); err != nil {
		return nil, err
	}

	poolID, err := tokenPoolID(token)
	if err != nil {
		return nil, err
	}

	identity := &Identity{
		PoolID: poolID,
	}

	for _, attribute := range cognitoUser.UserAttributes {
		switch attribute.Name {
		case "sub":
			identity.Sub = attribute.Value

		case "name":
			identity.Name = attribute.Value

		case "email":
			identity.Email = attribute.Value

		case "phone_number":
			identity.Phone = attribute.Value
		}
	}

	if identity.Sub == "" {
		return nil, errors.New("Cognito sub missing")
	}

	return identity, nil
}

func UserFromRequest(r *http.Request) (*Identity, error) {
	token, err := bearerToken(r)
	if err != nil {
		return nil, err
	}

	identity, err := verifyWithCognito(token)
	if err != nil {
		return nil, err
	}

	if identity.PoolID != userPoolID {
		return nil, errors.New("user access required")
	}

	return identity, nil
}

func AdminFromRequest(r *http.Request) (*Identity, error) {
	token, err := bearerToken(r)
	if err != nil {
		return nil, err
	}

	identity, err := verifyWithCognito(token)
	if err != nil {
		return nil, err
	}

	if identity.PoolID != adminPoolID {
		return nil, errors.New("admin access required")
	}

	return identity, nil
}

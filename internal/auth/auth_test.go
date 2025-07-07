package auth

import (
	"testing"
	"time"
	"net/http"

	"github.com/google/uuid"
	_"github.com/golang-jwt/jwt/v5"
)


func TestJWT(t *testing.T) {
	secretToken := "This is an Example Secret Key"
	userID := uuid.New()
	expiresIn := 1 * time.Second
	t.Logf("userID: %s", userID)
	t.Logf("secretToken: %s", string(secretToken))
	t.Logf("expiresIn: %d", expiresIn)

	stringToken, err := MakeJWT(userID, secretToken, expiresIn)
	if err != nil {
		t.Errorf("NewJWT Errored: %s", err)
	}
	t.Logf("Token made: %s", stringToken)

	datedUID, err := ValidateJWT(stringToken, secretToken)

	if err != nil {
		t.Errorf("ValidateJWT Errored: %s", err)
	}

	if datedUID.String() != userID.String() {
		t.Errorf("UserID is %s, got %s", userID, datedUID)
	}

	time.Sleep(expiresIn)

	_, err = ValidateJWT(stringToken, secretToken)
	if err == nil {
		t.Errorf("ValidateJWT didn't error for expirie")
	}
}

func TestBearerToken(t *testing.T) {
	
	header := http.Header{}

	header.Add("Authorization", "Bearer tHIStOKEN")
	expect := "tHIStOKEN"
	token, err := GetBearerToken(header)
	if err != nil {
		t.Errorf("Unexpected Error: %s", err)
	}
	if token != expect {
		t.Errorf("expect %s, got %s", expect, token)
	}

	header = http.Header{}
	_, err = GetBearerToken(header)
	if err == nil {
		t.Errorf("Empty header, did't error")
	}

	header.Add("Authorization", "tHIStOKEN")
	_, err = GetBearerToken(header)
	if err == nil {
		t.Errorf("Authorization without 'Bearer ', didn't")
	}

}

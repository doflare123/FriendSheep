package tests

import (
	"testing"

	"friendship/utils"
)

func TestJWTUtilsTokenPairRoundTrip(t *testing.T) {
	jwtUtils := utils.NewJWTUtils("test-secret")

	tokenPair, err := jwtUtils.GenerateTokenPair(99, "Bob", "bob", "image.png")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	accessClaims, err := jwtUtils.ParseAccessToken(tokenPair.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if accessClaims.UserID != 99 {
		t.Fatalf("UserID = %d, want 99", accessClaims.UserID)
	}

	refreshUserID, err := jwtUtils.ParseRefreshToken(tokenPair.RefreshToken)
	if err != nil {
		t.Fatalf("ParseRefreshToken returned error: %v", err)
	}
	if refreshUserID != 99 {
		t.Fatalf("refresh userID = %d, want 99", refreshUserID)
	}
}

func TestJWTUtilsRejectsWrongTokenType(t *testing.T) {
	jwtUtils := utils.NewJWTUtils("test-secret")

	tokenPair, err := jwtUtils.GenerateTokenPair(99, "Bob", "bob", "")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}

	if _, err := jwtUtils.ParseAccessToken(tokenPair.RefreshToken); err == nil {
		t.Fatal("ParseAccessToken accepted refresh token")
	}
	if _, err := jwtUtils.ParseRefreshToken(tokenPair.AccessToken); err == nil {
		t.Fatal("ParseRefreshToken accepted access token")
	}
}

func TestJWTUtilsRequiresSecret(t *testing.T) {
	jwtUtils := utils.NewJWTUtils("")

	if _, err := jwtUtils.GenerateTokenPair(1, "User", "user", ""); err == nil {
		t.Fatal("GenerateTokenPair accepted empty secret")
	}
	if err := jwtUtils.ValidateToken("token"); err == nil {
		t.Fatal("ValidateToken accepted empty secret")
	}
}

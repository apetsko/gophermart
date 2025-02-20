package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/apetsko/gophermart/internal/utils"
	"github.com/gorilla/securecookie"
)

func securedCookie(secret string) *securecookie.SecureCookie {
	secretLen := 32
	id := utils.GenerateID(secret, secretLen)
	sharedSecret := []byte(id)
	return securecookie.New(sharedSecret, sharedSecret)
}

func CookieSetUserID(w http.ResponseWriter, userID int, secret string) (err error) {
	sc := securedCookie(secret)

	encoded, err := sc.Encode("gophermart", userID)
	if err != nil {
		err = fmt.Errorf("error encoding userid cookie: %v", err)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "gophermart",
		Value:    encoded,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func CookieGetUserID(r *http.Request, secret string) (userID string, err error) {
	cookie, err := r.Cookie("gophermart")
	if err != nil {
		err = fmt.Errorf("error getting userid cookie: %w", err)
		return "", err
	}

	sc := securedCookie(secret)
	if err := sc.Decode("gophermart", cookie.Value, &userID); err != nil {
		err = fmt.Errorf("error decoding user cookie: %w", err)
		return "", err
	}

	if userID == "" {
		return "", errors.New("userid not found in cookie")
	}
	return userID, nil
}

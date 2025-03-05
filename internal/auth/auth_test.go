package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCookieSetUserID(t *testing.T) {
	secret := "supersecretkey"
	userID := 42

	w := httptest.NewRecorder()
	err := CookieSetUserID(w, userID, secret)

	assert.NoError(t, err, "должно успешно устанавливать cookie")

	resp := w.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()
	assert.Len(t, cookies, 1, "должен быть один установленный cookie")
	assert.Equal(t, "gophermart", cookies[0].Name, "cookie должен называться 'gophermart'")
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), cookies[0].Expires, time.Second, "cookie должен быть действительным на один день")
}

func TestCookieGetUserID(t *testing.T) {
	secret := "supersecretkey"
	userID := 42

	w := httptest.NewRecorder()
	err := CookieSetUserID(w, userID, secret)
	assert.NoError(t, err, "ошибок при установке cookie быть не должно")

	r := httptest.NewRequest(http.MethodGet, "/", nil)

	resp := w.Result()
	defer resp.Body.Close()
	r.Header.Set("Cookie", resp.Cookies()[0].String())

	resultUserID, err := CookieGetUserID(r, secret)

	assert.NoError(t, err, "ошибок при чтении cookie быть не должно")
	assert.NotNil(t, resultUserID, "userID не должен быть nil")
	assert.Equal(t, userID, *resultUserID, "userID должен совпадать")
}

func TestCookieGetUserID_NoCookie(t *testing.T) {
	secret := "supersecretkey"
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	defer r.Body.Close()
	resultUserID, err := CookieGetUserID(r, secret)

	assert.Nil(t, resultUserID, "userID должен быть nil при отсутствии cookie")
	assert.ErrorIs(t, err, http.ErrNoCookie, "ошибка должна быть http.ErrNoCookie")
}

func TestCookieGetUserID_InvalidCookie(t *testing.T) {
	secret := "supersecretkey"
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	defer r.Body.Close()
	r.Header.Set("Cookie", "gophermart=invalidcookievalue")

	resultUserID, err := CookieGetUserID(r, secret)

	assert.Nil(t, resultUserID, "userID должен быть nil при некорректном cookie")
	assert.Error(t, err, "должна возникнуть ошибка при некорректном cookie")
	assert.Contains(t, err.Error(), "error decoding user cookie", "текст ошибки должен содержать 'error decoding user cookie'")
}

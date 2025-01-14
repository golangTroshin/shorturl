package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGiveAuthTokenToUser_NewToken(t *testing.T) {
	handler := GiveAuthTokenToUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := r.Context().Value(UserIDKey); token == nil {
			t.Fatal("expected token in context, got nil")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.Code)
	}

	result := resp.Result()
	defer result.Body.Close()

	cookie := result.Cookies()
	found := false
	for _, c := range cookie {
		if c.Name == CookieAuthToken {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected auth_token cookie to be set, but it was not found")
	}
}

func TestGiveAuthTokenToUser_ExistingToken(t *testing.T) {
	existingToken := "existing-auth-token"
	handler := GiveAuthTokenToUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(UserIDKey)
		if token == nil {
			t.Fatal("expected token in context, got nil")
		}
		if token != existingToken {
			t.Fatalf("expected token '%s' in context, got '%s'", existingToken, token)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieAuthToken, Value: existingToken})
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.Code)
	}

	defer resp.Result().Body.Close()
}

func TestCheckAuthToken_MissingToken(t *testing.T) {
	handler := CheckAuthToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code 401, got %d", resp.Code)
	}

	defer resp.Result().Body.Close()
}

func TestCheckAuthToken_ValidToken(t *testing.T) {
	existingToken := "valid-auth-token"
	handler := CheckAuthToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value(UserIDKey)
		if token == nil {
			t.Fatal("expected token in context, got nil")
		}
		if token != existingToken {
			t.Fatalf("expected token '%s' in context, got '%s'", existingToken, token)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieAuthToken, Value: existingToken})
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", resp.Code)
	}

	defer resp.Result().Body.Close()
}

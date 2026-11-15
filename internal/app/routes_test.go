package app

import (
	"BookingGo/internal/adapter/redis"
	"BookingGo/internal/auth"
	"BookingGo/internal/domain"
	"BookingGo/internal/enum"
	"BookingGo/internal/ratelimit"
	"BookingGo/internal/usecase"
	"BookingGo/internal/usecase/mocks"
	"BookingGo/pkg/logger"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

// Exercise the real route registration, binder, middleware, controller and use case.
func TestLogoutAuthentication(t *testing.T) {
	logger.Init()
	gin.SetMode(gin.TestMode)
	for _, authorized := range []bool{false, true} {
		t.Run(map[bool]string{false: "anonymous", true: "authenticated"}[authorized], func(t *testing.T) {
			ctrl := gomock.NewController(t)
			blacklist := mocks.NewMockBlacklist(ctrl)
			deps := &Dependencies{BlacklistService: blacklist, RateLimiter: ratelimit.NewNoopLimiter(), AuthUseCase: usecase.NewAuthUseCase(nil, nil, nil, nil, blacklist)}
			router := gin.New()
			SetupRoutes(router, deps)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
			status := http.StatusUnauthorized
			if authorized {
				token, err := auth.GenerateToken(42, "test@example.com", enum.RoleClient)
				if err != nil {
					t.Fatal(err)
				}
				claims, err := auth.ValidateToken(token)
				if err != nil {
					t.Fatal(err)
				}
				blacklist.EXPECT().IsInBlacklist(gomock.Any(), claims.ID).Return(false)
				blacklist.EXPECT().IsSessionValid(gomock.Any(), 42, gomock.Any()).Return(true)
				blacklist.EXPECT().AddToBlacklist(gomock.Any(), claims.ID, gomock.Any()).Return(nil)
				req.Header.Set("Authorization", "Bearer "+token)
				status = http.StatusOK
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != status {
				t.Fatalf("status = %d, want %d: %s", response.Code, status, response.Body.String())
			}
		})
	}
}

func TestCredentialChangesWithOptionalTotp(t *testing.T) {
	logger.Init()
	gin.SetMode(gin.TestMode)
	hash, err := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"email", "password"} {
		for _, scenario := range []struct {
			name    string
			enabled bool
			code    string
			status  int
		}{
			{"disabled_without_code", false, "", 200},
			{"enabled_without_code", true, "", 400},
			{"enabled_invalid_code", true, "000000", 400},
			{"enabled_valid_code", true, "123456", 200},
			{"malformed_code", false, "12", 400},
		} {
			t.Run(operation+"/"+scenario.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				users := mocks.NewMockUserPostgres(ctrl)
				outbox := mocks.NewMockOutboxPostgres(ctrl)
				totp := mocks.NewMockTotpService(ctrl)
				blacklist := mocks.NewMockBlacklist(ctrl)
				blacklist.EXPECT().IsInBlacklist(gomock.Any(), gomock.Any()).Return(false)
				blacklist.EXPECT().IsSessionValid(gomock.Any(), 1, gomock.Any()).Return(true)
				user := &domain.User{ID: 1, Email: "test@example.com", PasswordHash: string(hash), IsTotpEnabled: scenario.enabled, TotpSecret: "test-only"}
				if scenario.name != "malformed_code" {
					users.EXPECT().GetByID(1).Return(user, nil)
					if operation == "email" {
						users.EXPECT().EmailExists("changed@example.com").Return(false, nil)
					}
				}
				if scenario.enabled && scenario.code != "" {
					totp.EXPECT().ValidateCode(user.TotpSecret, scenario.code).Return(scenario.code == "123456")
				}
				if scenario.status == 200 {
					users.EXPECT().Update(1, gomock.Any()).Return(user, nil)
					blacklist.EXPECT().InvalidateAllSessions(gomock.Any(), 1).Return(nil)
					outbox.EXPECT().CreateEvent(gomock.Any()).Return(nil)
				}
				deps := &Dependencies{BlacklistService: blacklist, RateLimiter: ratelimit.NewNoopLimiter(), AuthUseCase: usecase.NewAuthUseCase(users, outbox, totp, redis.NewNoopCache(), blacklist)}
				router := gin.New()
				SetupRoutes(router, deps)
				payload := map[string]string{"cur_password": "current-password", "new_password": "new-password", "new_email": "changed@example.com", "confirm_password": "current-password"}
				if scenario.code != "" {
					payload["otp_code"] = scenario.code
				}
				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPut, "/api/me/profile/"+operation, strings.NewReader(string(body)))
				token, err := auth.GenerateToken(1, user.Email, enum.RoleClient)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				response := httptest.NewRecorder()
				router.ServeHTTP(response, req)
				if response.Code != scenario.status {
					t.Fatalf("status = %d, want %d: %s", response.Code, scenario.status, response.Body.String())
				}
				if scenario.enabled && scenario.status == 400 && !strings.Contains(response.Body.String(), "код") {
					t.Fatalf("expected helpful TOTP error: %s", response.Body.String())
				}
			})
		}
	}
}

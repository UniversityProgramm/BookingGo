package ratelimit

import "time"

type EndpointConfig struct {
	Limit  int
	Window time.Duration
}

type Config struct {
	Default EndpointConfig

	Endpoints map[string]EndpointConfig

	// Лимиты по айпи (неавторизованные)
	PerIp EndpointConfig

	// Лимиты для юзера (авторизованные)
	PerUser EndpointConfig
}

func DefaultConfig() Config {
	return Config{
		Default: EndpointConfig{
			Limit:  100,
			Window: time.Minute,
		},
		PerIp: EndpointConfig{
			Limit:  60,
			Window: time.Minute,
		},
		PerUser: EndpointConfig{
			Limit:  300,
			Window: time.Minute,
		},
		Endpoints: map[string]EndpointConfig{
			"/api/auth/login": {
				Limit:  5,
				Window: time.Minute,
			},
			"/api/auth/register": {
				Limit:  3,
				Window: time.Minute,
			},
			"/api/me/otp/verify": {
				Limit:  5,
				Window: time.Minute,
			},
			"/api/me/otp/disable": {
				Limit:  5,
				Window: time.Minute,
			},

			"/api/me/notifications": {
				Limit:  30,
				Window: time.Minute,
			},
			"/api/me/bookings": {
				Limit:  60,
				Window: time.Minute,
			},
		},
	}
}

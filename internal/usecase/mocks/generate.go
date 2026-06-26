package mocks

//go:generate mockgen -source=../../cache/cache.go -destination=cache_mocks.go -package=mocks
//go:generate mockgen -source=../auth_usecase.go -destination=auth_mocks.go -package=mocks
//go:generate mockgen -source=../booking_usecase.go -destination=booking_mocks.go -package=mocks
//go:generate mockgen -source=../notification_usecase.go -destination=notification_mocks.go -package=mocks
//go:generate mockgen -source=../user_usecase.go -destination=user_mocks.go -package=mocks

var _ = "generate mocks"

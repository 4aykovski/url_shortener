package handler

import (
	"context"
	"strconv"

	"github.com/4aykovski/url_shortener/internal/adapters/http-server/v1/middleware"
)

func getUserId(ctx context.Context) (int, bool) {
	userID, err := strconv.Atoi(ctx.Value(middleware.UserCtx).(string))
	if err != nil {
		return 0, false
	}
	return userID, true
}

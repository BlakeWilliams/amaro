package web

import (
	"context"

	"github.com/blakewilliams/amaro/_template/internal/web/components"
)

// homeHandler renders the home page of the site.
func homeHandler(ctx context.Context, rc *requestContext) {
	rc.Render(ctx, components.Home{Message: "Hello, amaro!"})
}


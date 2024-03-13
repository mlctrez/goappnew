package compo

import "github.com/maxence-charriere/go-app/v10/pkg/app"

func Routes() {
	app.Route("/", func() app.Composer { return &Root{} })
}

package compo

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

var _ app.AppUpdater = (*Root)(nil)

type Root struct {
	app.Compo
}

func (r *Root) Render() app.UI {
	return app.Div().Text(app.Getenv("GOAPP_VERSION"))
}

func (r *Root) OnAppUpdate(ctx app.Context) {
	if app.Getenv("DEV") != "" && ctx.AppUpdateAvailable() {
		ctx.Reload()
	}
}

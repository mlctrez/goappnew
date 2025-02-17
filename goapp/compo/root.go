package compo

import (
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"time"
)

var _ app.AppUpdater = (*Root)(nil)
var _ app.Mounter = (*Root)(nil)

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

func (r *Root) OnMount(ctx app.Context) {
	if app.Getenv("DEV") != "" {
		ctx.Async(func() {
			timer := time.NewTicker(time.Second * 3)
			defer timer.Stop()
			for {
				select {
				case <-timer.C:
					app.TryUpdate()
				case <-ctx.Done():
					return
				}
			}
		})
	}
}

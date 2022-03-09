package web

import (
	"fmt"
	"net/http"
	nurl "net/url"
	"os"
	"path"
	"strings"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	"github.com/dnitsch/aws-cli-auth/internal/util"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	ps "github.com/mitchellh/go-ps"
)

func GetSamlLogin(loginUrl, acsUrl string) (string, error) {

	checkRodProcess()

	l := launcher.New().
		Headless(false).
		Devtools(false)

	// do not clean up userdata

	datadir := path.Join(util.GetHomeDir(), fmt.Sprintf(".%s-data", config.SELF_NAME))
	util.WriteDataDir(datadir)
	url := l.UserDataDir(datadir).MustLaunch()

	browser := rod.New().
		ControlURL(url).
		MustConnect().NoDefaultDevice()

	defer browser.MustClose()

	page := browser.MustPage(loginUrl)

	router := browser.HijackRequests()
	defer router.MustStop()

	router.MustAdd(acsUrl, func(ctx *rod.Hijack) {
		body := ctx.Request.Body()
		_ = ctx.LoadResponse(http.DefaultClient, true)
		ctx.Response.SetBody(body)
	})

	go router.Run()

	wait := page.EachEvent(func(e *proto.PageFrameRequestedNavigation) (stop bool) {
		return e.URL == acsUrl
	})
	wait()

	saml := strings.Split(page.MustElement(`body`).MustText(), "SAMLResponse=")[1]
	return nurl.QueryUnescape(saml)

}

//checkRodProcess gets a list running process
// kills any hanging rod browser process from any previous improprely closed sessions
func checkRodProcess() error {
	pids := make([]int, 0)
	ps, err := ps.Processes()
	if err != nil {
		return err
	}
	for _, v := range ps {
		if strings.Contains(v.Executable(), "Chromium") {
			pids = append(pids, v.Pid())
		}
	}
	for _, pid := range pids {
		fmt.Printf("Process to be killed as part of clean up: %d", pid)
		if proc, _ := os.FindProcess(pid); proc != nil {
			proc.Kill()
		}
	}
	return nil
}

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

type Web struct {
	datadir *string
	browser *rod.Browser
}

// New returns an initialised instance of Web struct using the default chromium embedded browser
func New() *Web {
	ddir := path.Join(util.HomeDir(), fmt.Sprintf(".%s-data", config.SELF_NAME))

	return &Web{
		datadir: &ddir,
	}
}

// WithDefaultLauncher returns the default chromium browser
func (web *Web) WithDefaultLauncher() *Web {

	l := launcher.New().
		Headless(false).
		Devtools(false).
		Leakless(true)

	url := l.UserDataDir(*web.datadir).MustLaunch()

	browser := rod.New().
		ControlURL(url).
		MustConnect().NoDefaultDevice()

	web.browser = browser
	return web
}

func (web *Web) WithCustomLauncher(execPath string) *Web {
	l := launcher.New().
		Bin(execPath).
		// Set("load-extension", "/Users/dusannitschneider/Library/Application Support/BraveSoftware/Brave-Browser/Default/Extensions/hdokiejnpimakedhajhdlcegeplioahd/4.92.0.1_0").
		ProfileDir("").
		Headless(false).
		Devtools(false)

	url := l.MustLaunch()

	browser := rod.New().
		ControlURL(url).
		MustConnect()

	web.browser = browser
	return web
}

// GetSamlLogin performs a saml login flow in managed browser
func (web *Web) GetSamlLogin(conf config.SamlConfig) (string, error) {

	// do not clean up userdata
	// datadir := path.Join(util.GetHomeDir(), fmt.Sprintf(".%s-data", config.SELF_NAME))
	util.WriteDataDir(*web.datadir)

	defer web.browser.MustClose()

	page := web.browser.MustPage(conf.ProviderUrl)

	router := web.browser.HijackRequests()
	defer router.MustStop()

	router.MustAdd(conf.AcsUrl, func(ctx *rod.Hijack) {
		body := ctx.Request.Body()
		_ = ctx.LoadResponse(http.DefaultClient, true)
		ctx.Response.SetBody(body)
	})

	go router.Run()

	wait := page.EachEvent(func(e *proto.PageFrameRequestedNavigation) (stop bool) {
		return e.URL == conf.AcsUrl
	})
	wait()

	saml := strings.Split(page.MustElement(`body`).MustText(), "SAMLResponse=")[1]
	return nurl.QueryUnescape(saml)

}

func (web *Web) ClearCache() error {
	errs := []error{}

	if err := os.RemoveAll(*web.datadir); err != nil {
		errs = append(errs, err)
	}
	if err := checkRodProcess(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs[:])
	}
	return nil
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
		util.Traceln("Process to be killed as part of clean up: %d", pid)
		if proc, _ := os.FindProcess(pid); proc != nil {
			proc.Kill()
		}
	}
	return nil
}

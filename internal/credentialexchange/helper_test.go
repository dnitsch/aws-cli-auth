package credentialexchange_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
	ini "gopkg.in/ini.v1"
)

// Tests need fixing up a bit
func TestGetEntryInIni(t *testing.T) {
	cfg, err := ini.Load("../../.aws-cli-auth.ini")
	if err != nil {
		t.Fatalf("Fail to read file: %v", err)
	}
	section := cfg.Section("role")
	roles := section.ChildSections()

	if len(roles) != 2 {
		t.Errorf("Incorrectly parsed INI got %d, wanted: 2", len(roles))
	}
}

func TestCreateEntryInIni(t *testing.T) {
	cfg, err := ini.Load("../../.aws-cli-auth.ini")

	if err != nil {
		t.Fatalf("Fail to read file: %v", err)
	}

	section := cfg.Section(credentialexchange.INI_CONF_SECTION) //
	if !cfg.HasSection(fmt.Sprintf("%s.%s", credentialexchange.INI_CONF_SECTION, credentialexchange.RoleKeyConverter(roleTest))) {
		t.Errorf("section NOT Exists")
	}
	roles := section.ChildSections()
	subSectionExists := false
	for _, v := range roles {
		if v.Name() == fmt.Sprintf("role.%s", credentialexchange.RoleKeyConverter(roleTest)) {
			subSectionExists = true
			break
		}
	}

	if !subSectionExists {
		t.Errorf("Not found nothing to do")
	}
}

func TestReloadBeforeExpirySuccess(t *testing.T) {

	expiry := (time.Now()).Add(time.Second * 305)

	got := credentialexchange.ReloadBeforeExpiry(expiry, 300)

	if got {
		t.Errorf("Expected %v, got: %v", false, got)
	}
}

func TestReloadBeforeExpiryNeedToRefresh(t *testing.T) {

	expiry := (time.Now()).Add(time.Second * 299)

	got := credentialexchange.ReloadBeforeExpiry(expiry, 300)

	if !got {
		t.Errorf("Expected %v, got: %v", false, got)
	}
}

func Test_HomeDirOverwritten(t *testing.T) {
	ttests := map[string]struct {
		setUpCleanUp func() func()
	}{
		"test1": {
			setUpCleanUp: func() func() {
				orignalEnv := os.Environ()
				os.Setenv("HOME", "./.ignore-delete")
				return func() {
					for _, e := range orignalEnv {
						pair := strings.SplitN(e, "=", 2)
						os.Setenv(pair[0], pair[1])
					}
				}
			},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			cleanUp := tt.setUpCleanUp()
			defer cleanUp()
			got := credentialexchange.HomeDir()
			if got != "./.ignore-delete" {
				t.Fail()
			}
		})
	}
}

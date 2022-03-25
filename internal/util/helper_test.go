package util

import (
	"fmt"
	"testing"

	"github.com/dnitsch/aws-cli-auth/internal/config"
	ini "gopkg.in/ini.v1"
)

// Tests need fixing up a bit
func TestGetEntryInIni(t *testing.T) {
	cfg, err := ini.Load("~/.aws-cli-auth.ini")
	if err != nil {
		Writeln("Fail to read file: %v", err)
		Exit(err)
	}
	section := cfg.Section("roles")
	roles := section.ChildSections()

	if len(roles) < 2 {
		t.Errorf("Not Enough children")
	}
}

//
func TestCreateEntryInIni(t *testing.T) {

	cfg, err := ini.Load(ConfigIniFile())
	if err != nil {
		Writeln("Fail to read Ini file: %v", err)
		Exit(err)
	}

	section := cfg.Section(config.INI_CONF_SECTION) //
	if !cfg.HasSection(fmt.Sprintf("%s.%s", config.INI_CONF_SECTION, RoleKeyConverter(roleTest))) {
		t.Errorf("section NOT Exists")
	}
	roles := section.ChildSections()
	subSectionExists := false
	for _, v := range roles {
		if v.Name() == fmt.Sprintf("roles.%s", RoleKeyConverter(roleTest)) {
			subSectionExists = true
			break
		}
	}

	if !subSectionExists {
		t.Errorf("Not found nothing to do")
	}
}

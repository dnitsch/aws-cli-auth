package credentialexchange_test

import (
	"testing"

	"github.com/dnitsch/aws-cli-auth/internal/credentialexchange"
)

var roleTest string = "arn:aws:iam::111122342343:role/DevAdmin"
var keyTest string = "arn_aws_iam__111122342343_role____DevAdmin"

func TestConvertRoleToKey(t *testing.T) {

	got := credentialexchange.RoleKeyConverter(roleTest)
	want := keyTest
	if got != want {
		t.Errorf("Wanted: %s, Got: %s", want, got)
	}
}

func TestConvertKeyToRole(t *testing.T) {

	got := credentialexchange.KeyRoleConverter(keyTest)
	want := roleTest
	if got != want {
		t.Errorf("Wanted: %s, Got: %s", want, got)
	}
}

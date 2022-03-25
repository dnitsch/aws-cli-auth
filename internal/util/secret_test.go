package util

import "testing"

var roleTest string = "arn:aws:iam::111122342343:role/DevAdmin"
var keyTest string = "arn_aws_iam__111122342343_role____DevAdmin"

func TestConvertRoleToKey(t *testing.T) {

	got := RoleKeyConverter(roleTest)
	want := keyTest
	if got != want {
		t.Errorf("Wanted: %s, Got: %s", want, got)
	}
}

func TestConvertKeyToRole(t *testing.T) {

	got := KeyRoleConverter(keyTest)
	want := roleTest
	if got != want {
		t.Errorf("Wanted: %s, Got: %s", want, got)
	}
}

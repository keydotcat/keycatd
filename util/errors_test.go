package util

import (
	"encoding/json"
	"testing"
)

func TestErrors(t *testing.T) {
	e := NewError("blah").(*Error)
	if bt := NewErrorFields().(*Error).Camo(); bt != nil {
		t.Errorf("Empty Error .Camo() should return nil: [%s]", bt)
	}
	e.fields["f1"] = "err"
	if e.Error() != "Error in fields: map[f1:err]" {
		t.Errorf("Unexpected Error result: %s", e.Error())
	}
	j, err := json.Marshal(e)
	if err != nil {
		t.Error(err)
	}
	expected := `{"error":"blah","error_fields":{"f1":"err"}}`
	if string(j) != expected {
		t.Errorf("Unexpected json result. Expected %s Got %s", expected, string(j))
	}
}

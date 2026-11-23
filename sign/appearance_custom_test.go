package sign

import (
	"bytes"
	"testing"
)

func TestCustomAppearanceReplacesTheDrawnOne(t *testing.T) {
	context := SignContext{SignData: SignData{
		Signature: SignDataSignature{Info: SignDataSignatureInfo{Name: "John Doe"}},
		Appearance: Appearance{
			Content:   []byte("BT /T1_0 10 Tf (John Doe) Tj ET"),
			Resources: []byte("/Font << /T1_0 << /Type /Font /Subtype /Type1 /BaseFont /MyriadPro-Regular >> >>"),
		},
	}}

	appearance, err := context.createAppearance([4]float64{0, 0, 192.5, 20.4})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"/BBox [0 0 192.500000 20.400000]",
		"/Resources <<\n/Font << /T1_0",
		"/Length 31\n",
		"stream\nBT /T1_0 10 Tf (John Doe) Tj ET\nendstream",
	} {
		if !bytes.Contains(appearance, []byte(want)) {
			t.Errorf("appearance lacks %q:\n%s", want, appearance)
		}
	}
	if bytes.Contains(appearance, []byte("/Times-Roman")) {
		t.Errorf("appearance still draws the default text:\n%s", appearance)
	}
}

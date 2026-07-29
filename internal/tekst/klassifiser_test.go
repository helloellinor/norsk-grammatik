package tekst

import (
	"strings"
	"testing"
)

func blokk(t string) Blokk { return Blokk{Stumpar: []Stump{{Tekst: t}}} }

// Tankestreken er eitt teikn men tre byte. Vart det talt i byte, skar vi
// to bokstavar for mykje av kvart oppslag.
func TestOppslagBeheldHeileForklaringa(t *testing.T) {
	prov := []struct{ inn, nykel, tekst string }{
		{"Adj. — Adjektiv.", "Adj.", "Adjektiv."},
		{"Ang. — Angelſachſiſk.", "Ang.", "Angelſachſiſk."},
		{"Barl. — Barlaams ok Joſaphats Saga.", "Barl.", "Barlaams ok Joſaphats Saga."},
		{"Bſt. F. — den beſtemte Form.", "Bſt. F.", "den beſtemte Form."},
	}
	for _, p := range prov {
		ut := Klassifiser([]Blokk{blokk(p.inn)})
		if len(ut) != 1 || ut[0].Slag != Oppslag {
			t.Fatalf("%q vart ikkje eit oppslag: %+v", p.inn, ut)
		}
		if ut[0].Nummer != p.nykel || ut[0].Tekst() != p.tekst {
			t.Errorf("%q\n fekk  %q / %q\n venta %q / %q", p.inn, ut[0].Nummer, ut[0].Tekst(), p.nykel, p.tekst)
		}
	}
}

func TestParagrafBeheldTeksten(t *testing.T) {
	lang := "8. Det norſke Sprog har ni enkelte Selvlyd, nemlig: A, Aa, Æ, E, I, O, U, Y, Ø, og tre Tvelyd."
	ut := Klassifiser([]Blokk{blokk(lang)})
	if ut[0].Slag != Paragraf || ut[0].Nummer != "8" {
		t.Fatalf("venta § 8, fekk %+v", ut[0])
	}
	if got := ut[0].Tekst(); !strings.HasPrefix(got, "Det norſke Sprog") {
		t.Errorf("teksten er skoren: %q", got)
	}
}

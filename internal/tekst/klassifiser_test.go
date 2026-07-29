package tekst

import (
	"strings"
	"testing"
)

func blokk(t string) Blokk { return Blokk{Stumpar: []Stump{{Tekst: t}}} }

// feit gjev ei utheva blokk - slik boka merkjer overskriftene sine.
func feit(t string) Blokk { return Blokk{Stumpar: []Stump{{Tekst: t, Feit: true}}} }

// iForklaringa set oppslaget inn i bolken det høyrer til. Oppslagsmønsteret
// gjeld berre der, av di korte linjer med tankestrek finst i brødteksten òg.
func iForklaringa(oppslag string) []Blokk {
	return []Blokk{feit("Forklaring af nogle Forkortninger."), blokk(oppslag)}
}

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
		ut := Klassifiser(iForklaringa(p.inn))
		if len(ut) != 2 || ut[1].Slag != Oppslag {
			t.Fatalf("%q vart ikkje eit oppslag: %+v", p.inn, ut)
		}
		if ut[1].Nummer != p.nykel || ut[1].Tekst() != p.tekst {
			t.Errorf("%q\n fekk  %q / %q\n venta %q / %q", p.inn, ut[1].Nummer, ut[1].Tekst(), p.nykel, p.tekst)
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

// Oppslagsmønsteret skal ikkje gjelde utanfor forkortingsbolken: avlyds-
// rekkjene i brødteksten har same forma, og vart før både omklassifiserte
// og skorne i stykke.
func TestOppslagBerreIForklaringa(t *testing.T) {
	linje := "i' a — u', for Ex. finna, fann, funno, funnen."
	ut := Klassifiser([]Blokk{feit("Anden Afdeling."), blokk(linje)})
	if ut[1].Slag == Oppslag {
		t.Errorf("%q vart eit oppslag utanfor forkortingsbolken", linje)
	}
	if ut[1].Tekst() != linje {
		t.Errorf("teksten vart skoren:\n fekk  %q\n venta %q", ut[1].Tekst(), linje)
	}
}

// heldFram skal sjaa paa siste RUNE. » og — er flerbyte.
func TestHeldFramFlerbyte(t *testing.T) {
	for _, s := range []string{"eit sitat»", "eit tillegg—", "ei setning."} {
		if heldFram(s) {
			t.Errorf("heldFram(%q) = true, skal vera false", s)
		}
	}
	if !heldFram("ei setning som held fram") {
		t.Error("uavslutta setning skal halde fram")
	}
}

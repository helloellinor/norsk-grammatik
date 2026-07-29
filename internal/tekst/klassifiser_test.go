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

// Ei blokk som held fram ei uavslutta setning er aldri ei overskrift.
// «o. ſ. v.» - halen av setninga føre - vart teken for ei underoverskrift,
// og teksten kom ut som «o) ſ. v»: punktumet bytt mot ein parentes og
// siste punktum borte. Overskriftsreglane endra sjølve boka.
func TestFramhaldBlirIkkjeOverskrift(t *testing.T) {
	byrjing := "ved Sammenſætning med Partikler, f. Ex. tilbunden, utløyſt, tillagad"
	for _, hale := range []string{"o. ſ. v.", "d. v. ſ. ſlikt.", "a) og fleire."} {
		ut := Klassifiser([]Blokk{blokk(byrjing), blokk(hale)})
		if len(ut) != 1 {
			t.Fatalf("%q vart ei eiga blokk (%d blokker), slaget %q", hale, len(ut), ut[len(ut)-1].Slag)
		}
		if venta := byrjing + " " + hale; ut[0].Tekst() != venta {
			t.Errorf("teksten er endra\n fekk  %q\n venta %q", ut[0].Tekst(), venta)
		}
	}
}

// Ei ekte overskrift skal framleis bli ei overskrift naar blokka føre er
// avslutta - elles hadde prøven ovanfor kunna «rettast» ved aa slaa av
// heile regelen.
func TestOverskriftEtterAvsluttaSetning(t *testing.T) {
	ut := Klassifiser([]Blokk{blokk("Dette er ei avslutta setning."), feit("a) Subſt. af Verbum.")})
	if len(ut) != 2 || ut[1].Slag != Underseksjon {
		t.Fatalf("overskrifta gjekk tapt: %+v", ut)
	}
	if ut[1].Nummer != "a" || ut[1].Tittel != "Subſt. af Verbum" {
		t.Errorf("fekk %q / %q", ut[1].Nummer, ut[1].Tittel)
	}
}

// Inne i forkortingsbolken er eit oppslag eit oppslag, ògso naar det byrjar
// med ein einskild bokstav: «v. — Verbum (S. 56)» vart underoverskrifta
// «v) — verbum (S. 56)».
func TestEinbokstavsOppslagIForklaringa(t *testing.T) {
	ut := Klassifiser(iForklaringa("v. — Verbum (S. 56)."))
	if len(ut) != 2 || ut[1].Slag != Oppslag {
		t.Fatalf("vart ikkje eit oppslag: %+v", ut)
	}
	if ut[1].Nummer != "v." || ut[1].Tekst() != "Verbum (S. 56)." {
		t.Errorf("fekk %q / %q", ut[1].Nummer, ut[1].Tekst())
	}
}

// Eit listepunkt er ikkje ei overskrift, kor kort det enn er. Boka
// uthevar overskriftene sine; listepunkta staar i vanleg skrift. Det var
// lengdegrensa som gjorde «b) Hardangerſk ... Fl. af Fe-» til ei
// overskrift i § 363 og reiv ordet Feminin i to.
func TestKortListepunktErIkkjeOverskrift(t *testing.T) {
	ut := Klassifiser([]Blokk{
		blokk("Ei avslutta setning."),
		blokk("b) Hardangerſk. Fleertal af Maſk. har `ar'; Fl. af Fe-"),
		blokk("min. har `er' (irr). Dativ ſynes at mangle."),
	})
	if len(ut) != 2 {
		t.Fatalf("venta to blokker, fekk %d", len(ut))
	}
	if ut[1].Slag != Underpunkt {
		t.Errorf("listepunktet vart %q, ikkje eit punkt", ut[1].Slag)
	}
	if ut[1].Nummer != "b" {
		t.Errorf("merket vart %q", ut[1].Nummer)
	}
	// Sideskiftet delte forkortinga Femin.; punktet skal vera heilt att.
	venta := "Hardangerſk. Fleertal af Maſk. har `ar'; Fl. af Femin. har `er' (irr). Dativ ſynes at mangle."
	if ut[1].Tekst() != venta {
		t.Errorf("teksten er skadd:\n fekk  %q\n venta %q", ut[1].Tekst(), venta)
	}
}

// Eit sideskift kan dele eit ord i to. Blokka føre endar paa bindestrek,
// og daa skal streken bort og ingen mellomrom setjast inn - elles stod
// det «om- trent» i teksten, 71 stader i verket.
func TestOrddelingVedSideskift(t *testing.T) {
	prov := []struct{ a, b, ut string }{
		{"eit Tungemaal, ſom har om-", "trent den ſamme Lighed", "eit Tungemaal, ſom har omtrent den ſamme Lighed"},
		{"og man har viſt-", "nok ogſaa ventet", "og man har viſtnok ogſaa ventet"},
		// Utan bindestrek skal mellomrommet framleis setjast inn.
		{"ei setning som held", "fram her", "ei setning som held fram her"},
	}
	for _, p := range prov {
		ut := Klassifiser([]Blokk{blokk(p.a), blokk(p.b)})
		if len(ut) != 1 {
			t.Fatalf("%q + %q vart %d blokker", p.a, p.b, len(ut))
		}
		if ut[0].Tekst() != p.ut {
			t.Errorf("fekk  %q\n venta %q", ut[0].Tekst(), p.ut)
		}
	}
}

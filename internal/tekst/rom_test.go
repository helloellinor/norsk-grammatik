package tekst

import "testing"

func TestSetningsrom(t *testing.T) {
	prov := []struct{ inn, ut string }{
		{"Lyden er lang. Han er kort.", "Lyden er lang.  Han er kort."},
		{"f. Ex. Hatt og Katt", "f. Ex. Hatt og Katt"},
		{"d. v. ſ. Berømmelſe", "d. v. ſ. Berømmelſe"},
		{"S. 132. Dette er nytt.", "S. 132. Dette er nytt."},
		{"jf. § 28. Vidare her.", "jf. § 28. Vidare her."},
		{"Er det ſaa? Ja visſt.", "Er det ſaa?  Ja visſt."},
		{"ende med Rodvokalen, f. Ex. Kne", "ende med Rodvokalen, f. Ex. Kne"},
	}
	for _, p := range prov {
		if fekk := Setningsrom(p.inn); fekk != p.ut {
			t.Errorf("Setningsrom(%q)\n fekk  %q\n venta %q", p.inn, fekk, p.ut)
		}
	}
}

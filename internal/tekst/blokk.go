package tekst

// Stump er ein bit tekst med den utheivinga han har i kjelda. Aasen
// skil systematisk mellom løpande tekst og siterte ordformer, og den
// skilnaden ber meining i ein grammatikk - difor tek vi vare på han.
type Stump struct {
	Tekst  string `json:"t"`
	Kursiv bool   `json:"i,omitempty"`
	Feit   bool   `json:"b,omitempty"`
}

// Rad og Tabell held oppsette som ikkje er løpande tekst: lydtavler,
// bøyingsmønster og liknande.
type Rad struct {
	Celler [][]Stump `json:"c"`
	Hovud  bool      `json:"h,omitempty"`
}

type Tabell struct {
	Rader []Rad `json:"r"`
}

// Breidd gjev talet paa kolonnar i den breiaste rada.
func (t Tabell) Breidd() int {
	b := 0
	for _, r := range t.Rader {
		if n := len(r.Celler); n > b {
			b = n
		}
	}
	return b
}

// Spenn gjev kor mange kolonnar celle nr i skal dekkje i ei rad med n
// celler. Ei overskriftsrad har ofte færre celler enn kroppen - «Aktiv.»
// og «Pasſiv.» staar over kvar sine to kolonnar, med ei tom celle føre
// radnamna - og utan spennet la dei seg over kvar si EINE kolonne, so
// heile hovudet stod forskuve mot venstre.
// Fyrste cella er radnamnet og dekkjer alltid éi kolonne; dei andre deler
// resten likt, og den siste tek det som ikkje gaar opp.
func (t Tabell) Spenn(n, i int) int {
	b := t.Breidd()
	if n >= b || n < 2 || i == 0 {
		return 1
	}
	att := b - 1
	del := att / (n - 1)
	if i == n-1 {
		return att - del*(n-2)
	}
	return del
}

// Blokk er ei eining i verket. Anten er ho løpande tekst (Stumpar) eller
// eit oppsett (Tabell).
type Blokk struct {
	Slag        Bolkslag `json:"s"`
	Nummer      string   `json:"n,omitempty"`
	Tittel      string   `json:"tt,omitempty"`
	Undertittel string   `json:"ut,omitempty"`
	Stumpar     []Stump  `json:"p,omitempty"`
	Tabell      *Tabell  `json:"tab,omitempty"`

	// Ankar blir sett når blokka er ei overskrift ein kan hoppe til.
	Ankar string `json:"-"`
	// IMerknad seier at eit oppsett høyrer til merknaden over det, og
	// difor skal setjast med same innrykk og dempa farge som han.
	IMerknad bool `json:"-"`
}

// Tekst gjev blokka som rein tekst, utan utheiving.
func (b Blokk) Tekst() string {
	var s string
	for _, st := range b.Stumpar {
		s += st.Tekst
	}
	return s
}

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

// Blokk er ei eining i verket. Anten er ho løpande tekst (Stumpar) eller
// eit oppsett (Tabell).
type Blokk struct {
	Slag        Bolkslag `json:"s"`
	Nummer      string   `json:"n,omitempty"`
	Tittel      string   `json:"tt,omitempty"`
	Undertittel string   `json:"ut,omitempty"`
	Stumpar     []Stump  `json:"p,omitempty"`
	Tabell      *Tabell  `json:"tab,omitempty"`
}

// Tekst gjev blokka som rein tekst, utan utheiving.
func (b Blokk) Tekst() string {
	var s string
	for _, st := range b.Stumpar {
		s += st.Tekst
	}
	return s
}

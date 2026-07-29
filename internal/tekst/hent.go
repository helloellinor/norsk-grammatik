package tekst

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

// HentFråHTML les utdataet frå wvHtml og plukkar ut avsnitt og oppsett.
// wvHtml er den einaste konverteraren vi har funne som tek vare på både
// tabellane og kursiven i den binære Word-fila; textutil kastar begge.
func HentFråHTML(r io.Reader) ([]Blokk, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var blokker []Blokk
	var gå func(*html.Node)

	gå = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "table":
				if t := lesTabell(n); t != nil {
					blokker = append(blokker, Blokk{Slag: Oppsett, Tabell: t})
				}
				return // cellene er alt lesne
			case "p":
				// wvHtml nøstar div inni p, so vi tek berre teksten som
				// ligg rett under dette avsnittet, ikkje i undertabellar.
				if st := lesStumpar(n, false); harTekst(st) {
					blokker = append(blokker, Blokk{Slag: Brødtekst, Stumpar: st})
				}
			}
		}
		for b := n.FirstChild; b != nil; b = b.NextSibling {
			gå(b)
		}
	}
	gå(doc)

	return blokker, nil
}

// lesStumpar samlar teksten under n, med kursiv og feit teken vare på.
// Med medTabell = false hoppar vi over tabellar, so dei ikkje blir lesne
// to gonger.
func lesStumpar(n *html.Node, medTabell bool) []Stump {
	var ut []Stump
	var gå func(*html.Node, bool, bool)

	gå = func(node *html.Node, kursiv, feit bool) {
		for b := node.FirstChild; b != nil; b = b.NextSibling {
			switch {
			case b.Type == html.TextNode:
				if b.Data != "" {
					ut = append(ut, Stump{Tekst: b.Data, Kursiv: kursiv, Feit: feit})
				}
			case b.Type == html.ElementNode:
				if !medTabell && (b.Data == "table" || b.Data == "p") {
					continue
				}
				gå(b, kursiv || b.Data == "i" || b.Data == "em",
					feit || b.Data == "b" || b.Data == "strong")
			}
		}
	}
	gå(n, false, false)

	return reinsk(ut)
}

func lesTabell(n *html.Node) *Tabell {
	var t Tabell
	for _, tr := range finnAlle(n, "tr") {
		var rad Rad
		for _, td := range finnAlle(tr, "td", "th") {
			rad.Celler = append(rad.Celler, reinsk(lesStumpar(td, true)))
			if td.Data == "th" {
				rad.Hovud = true
			}
		}
		// wvHtml gjev att linjalane i oppsettet som rader utan innhald.
		// Dei er strekar i papiret, ikkje data, so vi tek dei bort.
		if len(rad.Celler) > 0 && harCelleTekst(rad) {
			t.Rader = append(t.Rader, rad)
		}
	}
	if len(t.Rader) == 0 {
		return nil
	}
	return &t
}

// finnAlle hentar dei næraste etterkomarane med eit av namna, utan å gå
// ned i nøsta tabellar.
func finnAlle(n *html.Node, namn ...string) []*html.Node {
	vil := map[string]bool{}
	for _, s := range namn {
		vil[s] = true
	}
	var ut []*html.Node
	var gå func(*html.Node, bool)
	gå = func(node *html.Node, inni bool) {
		for b := node.FirstChild; b != nil; b = b.NextSibling {
			if b.Type != html.ElementNode {
				continue
			}
			if vil[b.Data] {
				ut = append(ut, b)
				continue
			}
			if b.Data == "table" && inni {
				continue
			}
			gå(b, inni || b.Data == "table")
		}
	}
	gå(n, false)
	return ut
}

// reinsk slår saman nabostumpar med same utheiving og normaliserer
// mellomrom, so teksten ikkje ber med seg linjebrota frå html-en.
func reinsk(st []Stump) []Stump {
	var ut []Stump
	for _, s := range st {
		s.Tekst = strings.ReplaceAll(s.Tekst, "\n", " ")
		s.Tekst = strings.ReplaceAll(s.Tekst, "\r", "")
		// Word-fila merkjer innrykk med ein tabulator i sjølve teksten.
		// Innrykket høyrer til satsen og staar i css-en; teiknet skal
		// ikkje bli med inn i teksten.
		s.Tekst = strings.ReplaceAll(s.Tekst, "\t", " ")
		for strings.Contains(s.Tekst, "  ") {
			s.Tekst = strings.ReplaceAll(s.Tekst, "  ", " ")
		}
		if s.Tekst == "" {
			continue
		}
		if n := len(ut); n > 0 && ut[n-1].Kursiv == s.Kursiv && ut[n-1].Feit == s.Feit {
			ut[n-1].Tekst += s.Tekst
			continue
		}
		ut = append(ut, s)
	}
	if len(ut) > 0 {
		ut[0].Tekst = strings.TrimLeft(ut[0].Tekst, " ")
		siste := len(ut) - 1
		ut[siste].Tekst = strings.TrimRight(ut[siste].Tekst, " ")
	}
	var att []Stump
	for _, s := range ut {
		if s.Tekst != "" {
			att = append(att, s)
		}
	}
	return att
}

func harCelleTekst(r Rad) bool {
	for _, c := range r.Celler {
		if harTekst(c) {
			return true
		}
	}
	return false
}

func harTekst(st []Stump) bool {
	for _, s := range st {
		if strings.TrimSpace(s.Tekst) != "" {
			return true
		}
	}
	return false
}

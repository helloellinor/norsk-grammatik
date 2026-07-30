// tenar er ein lokal lesetenar for verket: heile teksten blir tolka ein
// gong ved oppstart, og htmx hentar inn ein og ein del utan at sida blir
// lasta på nytt.
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"aasen/internal/tekst"
)

//go:embed malar/*.html
var malar embed.FS

//go:embed statisk
var statisk embed.FS

// Forkorting er ei oppføring frå «Forklaring af nogle Forkortninger».
type Forkorting struct {
	Stikkord string
	Tyding   string
}

// Bolk er eit av dei tre laga verket ligg i. Boka deler seg sjølv slik:
// framstykket ber ingen §§, verket ber §§ 1-399 i ei samanhengande
// rekkje - Indledning med §§ 1-6 og dei to Tillæga med §§ 351-399 er
// altso med i det - og bakstykket ber ingen att.
type Bolk struct {
	Delar []tekst.Del
}

type Verk struct {
	Tittel    string
	Undertitl string
	Delar     []tekst.Del
	Bolkar    []Bolk         // dei same delane, lagde i tre
	Paragraf  map[string]int // §-nummer -> del
	SisteP    string
	Stutt     []Forkorting // forkortingane boka forklarar
}

// bolkar deler delane i framstykke, verk og bakstykke etter om dei ber
// §§. Tittelbladet høyrer ikkje heime i registeret - ein kjem dit ved
// aa klikke tittelen øvst paa ryggen - so det blir haldt utanfor.
func bolkar(delar []tekst.Del) []Bolk {
	ut := make([]Bolk, 3)
	sett := false // har vi passert fyrste delen med §§?
	for _, d := range delar {
		if d.Id == 0 {
			continue
		}
		i := 0
		switch {
		case len(d.Paragr) > 0:
			i, sett = 1, true
		case sett:
			i = 2
		}
		ut[i].Delar = append(ut[i].Delar, d)
	}
	return ut
}

type sidedata struct {
	Verk  Verk
	Aktiv tekst.Del
	Oob   bool // registeret blir bytt ut for seg naar htmx hentar ein del
}

// Førre og Neste gjev delen før og etter den aktive, og nil i endane.
// Registeret er den eine vegen gjennom verket, men han krev at ein finn
// fram i lista - og paa telefon ligg han øvst paa sida, so ein maa rulle
// heile kapitlet attende berre for aa koma til neste. Ei lenkje i foten
// gjer det ein faktisk vil: lesa vidare.
func (d sidedata) Førre() *tekst.Del { return d.granne(-1) }
func (d sidedata) Neste() *tekst.Del { return d.granne(1) }

func (d sidedata) granne(steg int) *tekst.Del {
	i := d.Aktiv.Id + steg
	if i < 0 || i >= len(d.Verk.Delar) {
		return nil
	}
	return &d.Verk.Delar[i]
}

func main() {
	inn := flag.String("inn", "bøker/grammatik-sats.json", "blokker med ſ sett inn")
	adr := flag.String("adr", ":8064", "adresse å lytte på (:8064 tek imot frå heile nettet ditt)")
	flag.Parse()

	rå, err := os.ReadFile(*inn)
	if err != nil {
		log.Fatalf("kunne ikkje lese teksten: %v", err)
	}
	var blokker []tekst.Blokk
	if err := json.Unmarshal(rå, &blokker); err != nil {
		log.Fatalf("kunne ikkje tolke %s: %v", *inn, err)
	}

	delar := tekst.DelOpp(tekst.Klassifiser(blokker))
	indeks := tekst.ParagrafIndeks(delar)
	verk := Verk{
		Tittel:    "Norſk Grammatik",
		Undertitl: "Ivar Aasen · Chriſtiania 1864",
		Delar:     delar,
		Bolkar:    bolkar(delar),
		Paragraf:  indeks,
		SisteP:    høgste(indeks),
		Stutt:     forkortingar(delar),
	}
	log.Printf("%d delar, %d paragrafar tolka frå %s", len(verk.Delar), len(indeks), *inn)

	mal := template.Must(template.New("").Funcs(template.FuncMap{
		"rom": sats(indeks),
		// Teiknet etter eit overskriftsnummer. Malen og registeret maa
		// hente det frå same staden, elles stod «1. Præſens kort» i
		// teksten og «1) Præſens kort» i ryggen.
		"skil": tekst.Skiljeteikn,
		// Punktum etter overskrifter. Boka set punktum etter kvar
		// overskrift - «Tredie Afdeling.», «Bøiningsformer.» - og tolkinga
		// strauk det bort saman med nummeret, av di det same punktumet
		// skilde nummeret frå tittelen. Vi set det attende ved satsen i
		// staden for aa lagre det, so teksten i boka er den boka har.
		"punktum": func(s string) string {
			s = strings.TrimSpace(s)
			if s == "" {
				return s
			}
			r := []rune(s)
			if strings.ContainsRune(".!?:;»", r[len(r)-1]) {
				return s
			}
			return s + "."
		},
	}).ParseFS(malar, "malar/*.html"))

	mux := http.NewServeMux()
	mux.Handle("GET /statisk/", http.FileServer(http.FS(statisk)))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		vis(w, mal, verk, 0, false)
	})

	mux.HandleFunc("GET /del/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 0 || id >= len(verk.Delar) {
			http.NotFound(w, r)
			return
		}
		// htmx ber om berre brotstykket; ein vanleg lenkeklikk eller eit
		// bokmerke skal framleis gje heile sida.
		berreDel := r.Header.Get("HX-Request") == "true"
		vis(w, mal, verk, id, berreDel)
	})

	// Hopp til ein paragraf utan å vite kva bolk han står i.
	mux.HandleFunc("GET /paragraf/{nr}", func(w http.ResponseWriter, r *http.Request) {
		nr := r.PathValue("nr")
		id, finst := verk.Paragraf[nr]
		if !finst {
			http.Error(w, "fann ingen § "+nr, http.StatusNotFound)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/del/%d#p%s", id, nr), http.StatusSeeOther)
	})

	for _, u := range adresser(*adr) {
		log.Printf("les på %s", u)
	}
	if err := http.ListenAndServe(*adr, mux); err != nil {
		log.Fatal(err)
	}
}

// Boka kryssviser seg sjølv 556 gonger med «§ N». Det er hennar eige
// lenkjeverk, og det er dette som gjer at ein kan lesa henne utan
// sidetal - so vi gjer tilvisingane om til lenkjer.
//
// Berre den reine forma blir lenkja. Dei tjue tilfella av «§ 34, 35»
// har eit tal til bakom kommaet, men eit lause tal er ikkje sikkert
// nok til aa lenkje paa, so der blir berre det fyrste ei lenkje.
// Tilvisingar til §§ som ikkje finst i teksten (169 og 196) staar att
// som vanleg tekst i staden for aa bli daude lenkjer.
var reTilvis = regexp.MustCompile(`§[\s\p{Zs}]*\d+(?:[\s\p{Zs}]*,[\s\p{Zs}]*\d+)*`)

// Tala inne i ei tilvising. «§ 112, 180» viser til to paragrafar, og begge
// skal vera til aa klikke paa - før stod berre det fyrste som lenkje, og
// dei 21 tilfella med komma i verket gøymde halvparten av tilvisingane
// sine bak ein tekst som saag klikkbar ut, men ikkje var det.
var reTal = regexp.MustCompile(`\d+`)

// sats set setningsromma og lenkjer tilvisingane. Han skriv HTML og
// lyt difor sjølv sleppe teksten gjennom escape.
func sats(indeks map[string]int) func(string) template.HTML {
	return func(s string) template.HTML {
		s = tekst.Setningsrom(s)
		treff := reTilvis.FindAllStringIndex(s, -1)
		if treff == nil {
			return template.HTML(template.HTMLEscapeString(s))
		}
		var b strings.Builder
		slutt := 0
		for _, t := range treff {
			ord := s[t[0]:t[1]]
			b.WriteString(template.HTMLEscapeString(s[slutt:t[0]]))
			// Kvart tal for seg. Teiknet «§» og komma imellom staar att som
			// vanleg tekst, so tilvisinga les seg som ei eining.
			indre := 0
			for _, tt := range reTal.FindAllStringIndex(ord, -1) {
				b.WriteString(template.HTMLEscapeString(ord[indre:tt[0]]))
				nr := ord[tt[0]:tt[1]]
				if _, finst := indeks[nr]; finst {
					fmt.Fprintf(&b, `<a class="tilvis" href="/paragraf/%s">%s</a>`, nr, template.HTMLEscapeString(nr))
				} else {
					b.WriteString(template.HTMLEscapeString(nr))
				}
				indre = tt[1]
			}
			b.WriteString(template.HTMLEscapeString(ord[indre:]))
			slutt = t[1]
		}
		b.WriteString(template.HTMLEscapeString(s[slutt:]))
		return template.HTML(b.String())
	}
}

func vis(w http.ResponseWriter, mal *template.Template, verk Verk, id int, berreDel bool) {
	data := sidedata{
		Verk:  verk,
		Aktiv: verk.Delar[id],
		Oob:   berreDel,
	}
	namn := "side.html"
	if berreDel {
		namn = "del.html"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := mal.ExecuteTemplate(w, namn, data); err != nil {
		log.Printf("mal %s: %v", namn, err)
	}
}

// høgste gjev det største §-nummeret, til hjelpeteksten i hoppfeltet.
func høgste(indeks map[string]int) string {
	beste := 0
	for nr := range indeks {
		if n, err := strconv.Atoi(nr); err == nil && n > beste {
			beste = n
		}
	}
	return strconv.Itoa(beste)
}

// adresser gjev dei nettadressene tenaren kan naaast paa, so ein kan
// opne sida paa telefonen over same nettet.
func adresser(adr string) []string {
	_, port, err := net.SplitHostPort(adr)
	if err != nil {
		return []string{"http://" + adr}
	}
	ut := []string{"http://localhost:" + port}
	vert, _, _ := net.SplitHostPort(adr)
	if vert != "" && vert != "0.0.0.0" && vert != "::" {
		return ut
	}
	adrs, err := net.InterfaceAddrs()
	if err != nil {
		return ut
	}
	for _, a := range adrs {
		ip, ok := a.(*net.IPNet)
		if !ok || ip.IP.IsLoopback() || ip.IP.To4() == nil {
			continue
		}
		ut = append(ut, "http://"+ip.IP.String()+":"+port+"  (telefon o.l. paa same nettet)")
	}
	return ut
}

// Tittelblad seier om denne bolken opnar verket. Daa set vi sjølve
// tittelbladet i staden for aa la dei tre linjene - forfattar, verk,
// stad - staa som lause overskrifter.
func (d sidedata) Tittelblad() bool { return d.Aktiv.Id == 0 }

// tittelord er orda tittelarket alt syner. Vi kan ikkje samanlikne heile
// blokker: sideskifta i kjelda slaar linjene saman til lange remser, og
// kva som hamnar i kva remse skiftar med klassifiseringa.
var tittelord = map[string]bool{
	"Ivar": true, "Aaſen": true, "Norſk": true, "Grammatik": true,
	"Chiſtiania": true, "Chriſtiania": true, "1864": true,
	"Føreord": true, "(1997)": true, "Innhald": true, "Fortale": true,
	"Elektroniſk": true, "utgåve": true, "Det": true, "Norſke": true,
	"Samlaget": true, "Oslo": true, "1997": true,
}

// berreTittelord seier om blokka ikkje inneheld anna enn det arket syner.
func berreTittelord(t string) bool {
	felt := strings.Fields(t)
	if len(felt) == 0 {
		return false
	}
	for _, o := range felt {
		if !tittelord[o] {
			return false
		}
	}
	return true
}

// Vis er blokkene som skal setjast. Paa tittelbladet hoppar vi over dei
// linjene tittelarket alt har synt.
func (d sidedata) Vis() []tekst.Blokk {
	if !d.Tittelblad() {
		return d.Aktiv.Blokker
	}
	b := d.Aktiv.Blokker
	hopp := 0
	for hopp < len(b) && berreTittelord(b[hopp].Tekst()) {
		hopp++
	}
	return b[hopp:]
}

// forkortingar plukkar ut lista boka sjølv gjev i «Forklaring af nogle
// Forkortninger», so ho kan slaaast opp utan aa bla dit.
func forkortingar(delar []tekst.Del) []Forkorting {
	var ut []Forkorting
	for _, d := range delar {
		for _, b := range d.Blokker {
			if b.Slag == tekst.Oppslag {
				ut = append(ut, Forkorting{Stikkord: b.Nummer, Tyding: b.Tekst()})
			}
		}
	}
	return ut
}

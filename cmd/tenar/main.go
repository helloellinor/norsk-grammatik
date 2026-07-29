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

type Verk struct {
	Tittel    string
	Undertitl string
	Delar     []tekst.Del
	Paragraf  map[string]int // §-nummer -> del
	SisteP    string
	Stutt     []Forkorting // forkortingane boka forklarar
}

type sidedata struct {
	Verk  Verk
	Aktiv tekst.Del
	Oob   bool // registeret blir bytt ut for seg naar htmx hentar ein del
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
		Paragraf:  indeks,
		SisteP:    høgste(indeks),
		Stutt:     forkortingar(delar),
	}
	log.Printf("%d delar, %d paragrafar tolka frå %s", len(verk.Delar), len(indeks), *inn)

	mal := template.Must(template.New("").Funcs(template.FuncMap{
		"rom": tekst.Setningsrom,
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

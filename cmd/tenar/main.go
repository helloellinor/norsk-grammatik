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
	"net/http"
	"os"
	"strconv"

	"aasen/internal/tekst"
)

//go:embed malar/*.html
var malar embed.FS

//go:embed statisk
var statisk embed.FS

type Verk struct {
	Tittel    string
	Undertitl string
	Delar     []tekst.Del
	Paragraf  map[string]int // §-nummer -> del
	SisteP    string
}

type sidedata struct {
	Verk  Verk
	Aktiv tekst.Del
	Oob   bool // registeret blir bytt ut for seg naar htmx hentar ein del
}

func main() {
	inn := flag.String("inn", "bøker/grammatik-sats.json", "blokker med ſ sett inn")
	adr := flag.String("adr", "localhost:8064", "adresse å lytte på")
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
	}
	log.Printf("%d delar, %d paragrafar tolka frå %s", len(verk.Delar), len(indeks), *inn)

	mal := template.Must(template.ParseFS(malar, "malar/*.html"))

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

	log.Printf("les på http://%s", *adr)
	if err := http.ListenAndServe(*adr, mux); err != nil {
		log.Fatal(err)
	}
}

func vis(w http.ResponseWriter, mal *template.Template, verk Verk, id int, berreDel bool) {
	data := sidedata{Verk: verk, Aktiv: verk.Delar[id], Oob: berreDel}
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

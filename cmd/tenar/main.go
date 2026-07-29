// tenar er ein lokal lesetenar for verket: heile teksten blir tolka ein
// gong ved oppstart, og htmx hentar inn ein og ein del utan at sida blir
// lasta på nytt.
package main

import (
	"embed"
	"encoding/json"
	"flag"
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
}

type sidedata struct {
	Verk
	Aktiv tekst.Del
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

	verk := Verk{
		Tittel:    "Norſk Grammatik",
		Undertitl: "Ivar Aasen · Chriſtiania 1864",
		Delar:     tekst.DelOpp(tekst.Klassifiser(blokker)),
	}
	log.Printf("%d delar tolka frå %s", len(verk.Delar), *inn)

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

	log.Printf("les på http://%s", *adr)
	if err := http.ListenAndServe(*adr, mux); err != nil {
		log.Fatal(err)
	}
}

func vis(w http.ResponseWriter, mal *template.Template, verk Verk, id int, berreDel bool) {
	data := sidedata{Verk: verk, Aktiv: verk.Delar[id]}
	namn := "side.html"
	if berreDel {
		namn = "del.html"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := mal.ExecuteTemplate(w, namn, data); err != nil {
		log.Printf("mal %s: %v", namn, err)
	}
}

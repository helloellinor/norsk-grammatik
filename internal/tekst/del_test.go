package tekst

import (
	"strings"
	"testing"
)

// Kvart ankar i innhaldslista maa svare til eit ankar paa ei blokk i
// teksten. Delane blir omnummererte naar tomme bolkar fell bort, og
// skreiv den omskrivinga løpenummeret etter plassen i titteltavla i
// staden for etter blokkindeksen, fekk registeret ei heilt anna rekkje
// enn teksten: 52 av 60 lenkjer peika paa eit ankar som ikkje fanst, og
// dei aatte som traff, traff feil overskrift.
func TestRegisterankerPeikarPaaEiBlokk(t *testing.T) {
	inn := []Blokk{
		feit("Førſte Afdeling."),
		feit("Lydlære"),
		blokk("I. Bogſtaver eller enkelt Lyd."),
		blokk("1. Det norſke Sprog har ni enkelte Selvlyd."),
		blokk("a) Vokaler."),
		blokk("2. Vokalane er desse."),
		blokk("II. Lydſtillinger."),
		blokk("3. Ei lydſtilling er noko anna."),
	}
	delar := DelOpp(Klassifiser(inn))
	if len(delar) == 0 {
		t.Fatal("ingen delar")
	}

	var titlar int
	for _, d := range delar {
		ankar := map[string]string{}
		for _, b := range d.Blokker {
			if b.Ankar != "" {
				ankar[b.Ankar] = b.Tittel
			}
		}
		for _, p := range d.Titlar {
			titlar++
			tittel, finst := ankar[p.Ankar]
			if !finst {
				t.Errorf("del %d: %q peikar paa %q, som ingi blokk har", d.Id, p.Tittel, p.Ankar)
				continue
			}
			// Ankeret skal treffe si eiga overskrift, ikkje ei anna.
			if tittel == "" || !strings.Contains(p.Tittel, tittel) {
				t.Errorf("del %d: %q peikar paa overskrifta %q", d.Id, p.Tittel, tittel)
			}
		}
	}
	if titlar == 0 {
		t.Fatal("ingen titlar i registeret - prøven prøver ingenting")
	}
}

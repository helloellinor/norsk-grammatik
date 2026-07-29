package main

import "testing"

// Reserveregelen gjeld dei 2,2 % av orda som ikkje finst i trykken. Han
// skal setje rund s i slutten av eit ordledd og lang ſ elles. Dobbel s er
// den eine staden ordledda alltid røper seg.
func TestTrykkregelenPaaDobbelS(t *testing.T) {
	prov := []struct{ inn, ut, kvifor string }{
		// Inne i ordet deler dobbel s seg over eit stavingsskil, so den
		// fyrste endar eit ledd. Trykken: 1002 sſ mot 9 ſſ, og alle ni
		// er OCR som har lese frakturens «ll» som «ſſ».
		{"overensstemmende", "overensſtemmende", "dobbel s inne i ordet"},
		{"Klasse", "Klasſe", "same, kort ord"},
		{"afpasset", "afpasſet", "same"},
		// I ordslutt er det omvendt: 131 ſs mot 6 ss i trykken, og
		// sjølve ordet «Laſs» staar der.
		{"Vidarlass", "Vidarlaſs", "dobbel s i ordslutt"},
		{"Lass", "Laſs", "same, aaleine"},
		// Vanleg s: lang inne i ordet, rund i slutten.
		{"Sprog", "Sprog", "stor S blir ikkje rørt"},
		{"husets", "huſets", "lang inne, rund i slutten"},
	}
	tom := leksikon{Eksakt: map[string]oppslag{}, Folda: map[string]oppslag{}}
	for _, p := range prov {
		stat := statistikk{
			UsikreOrd: map[string]int{}, UkjendeOrd: map[string]int{},
			RegelGavLangS: map[string]int{},
		}
		if fekk := settInnLangS(p.inn, tom, &stat); fekk != p.ut {
			t.Errorf("%s: %q gav %q, venta %q", p.kvifor, p.inn, fekk, p.ut)
		}
	}
}

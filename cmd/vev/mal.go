package main

const malTekst = `{{if not .Nake}}<!doctype html>
<html lang="nn">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Verk}} — Ivar Aasen 1864</title>
</head>
<body>
{{end}}<style>
@font-face {
  font-family: "Fraktur";
  src: url("{{.Skrift}}") format("truetype");
  font-display: block;
}

:root {
  --papir: #FCFBF9;
  --blekk: #1A1714;
  --blekk-mjuk: #57514A;
  --indigo: #2B3A67;
  --linje: #E4DFD6;
  --uthev: #EFEAE0;

  --mål: 34rem;
  --fraktur: "Fraktur", "UnifrakturMaguntia", serif;
  --moderne: ui-serif, Georgia, "Times New Roman", serif;
  --sans: ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif;
}

@media (prefers-color-scheme: dark) {
  :root {
    --papir: #191713;
    --blekk: #EDE8DE;
    --blekk-mjuk: #A29A8D;
    --indigo: #93A6DC;
    --linje: #34302A;
    --uthev: #26231D;
  }
}
:root[data-theme="dark"] {
  --papir: #191713;
  --blekk: #EDE8DE;
  --blekk-mjuk: #A29A8D;
  --indigo: #93A6DC;
  --linje: #34302A;
  --uthev: #26231D;
}
:root[data-theme="light"] {
  --papir: #FCFBF9;
  --blekk: #1A1714;
  --blekk-mjuk: #57514A;
  --indigo: #2B3A67;
  --linje: #E4DFD6;
  --uthev: #EFEAE0;
}

* { box-sizing: border-box; }

body {
  margin: 0;
  background: var(--papir);
  color: var(--blekk);
  font-family: var(--fraktur);
  font-feature-settings: "liga" 1, "hlig" 1, "rlig" 1;
  line-height: 1.78;
  -webkit-font-smoothing: antialiased;
}

/* Apparatet vårt skil seg med vilje frå Aasens eigen tekst:
   det som er sett i fraktur er 1864, alt anna er utgåva. */
.apparat { font-family: var(--sans); font-feature-settings: normal; }

header {
  border-bottom: 1px solid var(--linje);
  padding: 2.5rem 1.5rem 1.75rem;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  align-items: center;
  text-align: center;
}

h1 {
  margin: 0;
  font-size: clamp(2rem, 6vw, 3rem);
  font-weight: 400;
  letter-spacing: 0.01em;
  text-wrap: balance;
}

.undertitl {
  font-family: var(--sans);
  font-size: 0.8125rem;
  letter-spacing: 0.09em;
  text-transform: uppercase;
  color: var(--blekk-mjuk);
}

.styring { display: flex; gap: 0.5rem; flex-wrap: wrap; justify-content: center; }

button {
  font: inherit;
  font-family: var(--sans);
  font-size: 0.8125rem;
  color: var(--blekk-mjuk);
  background: transparent;
  border: 1px solid var(--linje);
  border-radius: 2px;
  padding: 0.4rem 0.85rem;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s, background 0.15s;
}
button:hover { color: var(--indigo); border-color: var(--indigo); }
button[aria-pressed="true"] { color: var(--papir); background: var(--indigo); border-color: var(--indigo); }
button:focus-visible { outline: 2px solid var(--indigo); outline-offset: 2px; }

main {
  max-width: var(--mål);
  margin: 0 auto;
  padding: 3rem 1.5rem 6rem;
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
  font-size: 1.1875rem;
  hyphens: auto;
}

.afdeling {
  margin-top: 3rem;
  padding-top: 2.5rem;
  border-top: 1px solid var(--linje);
  text-align: center;
  font-size: 1.9rem;
  text-wrap: balance;
}
main > .afdeling:first-child { margin-top: 0; padding-top: 0; border-top: none; }

.seksjon, .underseksjon, .mellomtittel {
  text-wrap: balance;
  margin-top: 1.4rem;
}
.seksjon { font-size: 1.5rem; text-align: center; }
.underseksjon { font-size: 1.28rem; text-align: center; }
.mellomtittel { font-size: 1.16rem; text-align: center; color: var(--blekk-mjuk); }

.seksjon .tal, .underseksjon .tal { color: var(--indigo); }

/* §-nummera heng i margen: dei er handtaka teksten siterer seg sjølv
   med, ikkje pynt. Under 900px fell dei inn i linja att. */
.paragraf { position: relative; }
.paragraf .tal {
  font-family: var(--sans);
  font-size: 0.75rem;
  font-variant-numeric: tabular-nums;
  color: var(--indigo);
  position: absolute;
  left: -3.25rem;
  top: 0.45rem;
  width: 2.5rem;
  text-align: right;
  user-select: none;
}
@media (max-width: 900px) {
  .paragraf .tal {
    position: static;
    width: auto;
    display: inline;
    margin-right: 0.5rem;
  }
}

/* Forkortingslista er ei ordliste, ikkje brødtekst: stikkordet fyrst,
   forklaringa etter, med tett linjeavstand som i ein registerbolk. */
.oppslag {
  display: grid;
  grid-template-columns: minmax(4.5rem, max-content) 1fr;
  gap: 0 1rem;
  font-size: 1.0625rem;
  line-height: 1.5;
}
.oppslag .stikkord { color: var(--indigo); }

.merknad, .avsnitt {
  font-size: 1.0625rem;
  color: var(--blekk-mjuk);
  padding-left: 1.5rem;
  border-left: 2px solid var(--linje);
}

p { margin: 0; }

footer {
  max-width: var(--mål);
  margin: 0 auto;
  padding: 2rem 1.5rem 4rem;
  border-top: 1px solid var(--linje);
  font-family: var(--sans);
  font-size: 0.8125rem;
  line-height: 1.7;
  color: var(--blekk-mjuk);
}
footer p { margin: 0 0 0.75rem; }
footer strong { color: var(--blekk); font-weight: 600; }

/* Lesetekst: same tekst, moderne skrift og rund s. */
body.lesetekst { font-family: var(--moderne); line-height: 1.7; }
body.lesetekst main { font-size: 1.0625rem; }
body.lesetekst .afdeling { font-size: 1.6rem; }
body.lesetekst .seksjon { font-size: 1.3rem; }

@media (prefers-reduced-motion: reduce) {
  * { transition: none !important; }
}
</style>

<header>
  <h1>{{.Verk}}</h1>
  <div class="undertitl">{{.Undertitl}}</div>
  <div class="styring apparat">
    <button id="skrift" aria-pressed="true">Trykk 1864</button>
    <button id="tema" aria-pressed="false">Mørk</button>
  </div>
</header>

<main id="tekst">
{{range .Bolkar}}
  {{- if eq .Slag "afdeling"}}
  <h2 class="afdeling">{{.Tittel}}</h2>
  {{- else if eq .Slag "seksjon"}}
  <h3 class="seksjon"><span class="tal">{{.Nummer}}.</span> {{.Tittel}}</h3>
  {{- else if eq .Slag "underseksjon"}}
  <h4 class="underseksjon"><span class="tal">{{.Nummer}})</span> {{.Tittel}}</h4>
  {{- else if eq .Slag "mellomtittel"}}
  <h5 class="mellomtittel">{{.Tittel}}</h5>
  {{- else if eq .Slag "paragraf"}}
  <p class="paragraf" id="p{{.Nummer}}"><span class="tal">§&nbsp;{{.Nummer}}</span>{{.Tekst}}</p>
  {{- else if eq .Slag "oppslag"}}
  <p class="oppslag"><span class="stikkord">{{.Nummer}}</span>{{.Tekst}}</p>
  {{- else if eq .Slag "merknad"}}
  <p class="merknad">{{.Tekst}}</p>
  {{- else if eq .Slag "avsnitt"}}
  <p class="avsnitt">{{.Tekst}}</p>
  {{- else}}
  <p>{{.Tekst}}</p>
  {{- end}}
{{end}}
</main>

<footer class="apparat">
  <p><strong>Om denne utgåva.</strong> Teksten er 1864-utgåva, henta frå Det Norske Samlagets
  elektroniske transkripsjon (1997). Den lange s-en er sett attende inn etter korleis kvart
  einskilt ord faktisk står i det trykte førstetrykket, lese av ei 400 ppi-skanning frå
  Nasjonalbiblioteket. Ord som ikkje fanst i trykken fylgjer trykkregelen: rund s i slutten
  av eit ordledd, lang ſ elles.</p>
  <p>Aasen tek sjølv opp dette i § 384, og skyv det med vilje ut av grammatikken: skriftteikn
  høyrer til «det ydre eller det, som er for Øiet», og der skal ein «fylgje Brugen i Dansk og
  Tydſk». Difor er trykkjaren sin eigen praksis fasit her — mellom anna at dobbel s vert sett
  <span style="font-family: var(--fraktur)">sſ</span>, ikkje <span style="font-family: var(--fraktur)">ſſ</span> som på tyſk.</p>
  <p>Skrifta er UnifrakturMaguntia (SIL OFL).</p>
</footer>

<script>
(function () {
  var rot = document.documentElement;
  var kropp = document.body;

  // Lang ſ er ei reint visuell form av same bokstav, so lesetekst-visinga
  // byter henne ut med rund s. Vi tek vare på originalen for å kunne gå att.
  var tekstnodar = [];
  (function samle(node) {
    for (var b = node.firstChild; b; b = b.nextSibling) {
      if (b.nodeType === 3) {
        if (b.nodeValue.indexOf("ſ") !== -1) {
          tekstnodar.push({ node: b, trykk: b.nodeValue, lese: b.nodeValue.replace(/ſ/g, "s") });
        }
      } else if (b.nodeType === 1) {
        samle(b);
      }
    }
  })(document.getElementById("tekst"));

  var skriftknapp = document.getElementById("skrift");
  var trykk = true;
  skriftknapp.addEventListener("click", function () {
    trykk = !trykk;
    kropp.classList.toggle("lesetekst", !trykk);
    skriftknapp.setAttribute("aria-pressed", String(trykk));
    skriftknapp.textContent = trykk ? "Trykk 1864" : "Lesetekst";
    for (var i = 0; i < tekstnodar.length; i++) {
      tekstnodar[i].node.nodeValue = trykk ? tekstnodar[i].trykk : tekstnodar[i].lese;
    }
  });

  var temaknapp = document.getElementById("tema");
  var mørk = matchMedia("(prefers-color-scheme: dark)").matches;
  function settTema() {
    rot.setAttribute("data-theme", mørk ? "dark" : "light");
    temaknapp.setAttribute("aria-pressed", String(mørk));
    temaknapp.textContent = mørk ? "Ljos" : "Mørk";
  }
  settTema();
  temaknapp.addEventListener("click", function () { mørk = !mørk; settTema(); });
})();
</script>
{{if not .Nake}}
</body>
</html>
{{end}}`

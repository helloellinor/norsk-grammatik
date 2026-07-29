// Lesarval og småting som ikkje treng tenaren: skriftform, skriftval,
// drakt, hopp til paragraf, og utheving i tavlene.
(function () {
  "use strict";

  var rot = document.documentElement;
  var kropp = document.body;
  var LAGER = "aasen-lesar";

  var val = {
    skrift: "maguntia",
    ui: "antiqua",
    ligatur: "trykk",
    storleik: "vanleg",
    tema: matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
  };
  try {
    var lagra = JSON.parse(localStorage.getItem(LAGER) || "{}");
    Object.keys(val).forEach(function (k) {
      if (lagra[k] !== undefined && lagra[k] !== null) val[k] = lagra[k];
    });
  } catch (e) { /* utan lager går det fint òg */ }

  function lagre() {
    try { localStorage.setItem(LAGER, JSON.stringify(val)); } catch (e) {}
  }

  // ---- Skriftform: fraktur med lang ſ, eller antikva med rund s ------
  // Lang ſ er ei reint visuell form av same bokstav, so antikva-visinga
  // byter henne ut. Originalen blir teken vare på, so ein kan gå att
  // utan å hente teksten på nytt.
  var nodar = [];
  function samle(rotnode) {
    nodar = [];
    if (!rotnode) return;
    var gaa = document.createTreeWalker(rotnode, NodeFilter.SHOW_TEXT, null);
    var n;
    while ((n = gaa.nextNode())) {
      if (n.nodeValue.indexOf("ſ") !== -1) {
        nodar.push({ node: n, trykk: n.nodeValue, antikva: n.nodeValue.replace(/ſ/g, "s") });
      }
    }
  }

  // Antikva er eit skriftval jamsides frakturane. Er han vald, gaar
  // teksten over til latinske bokstavar med rund s.
  function erFraktur() { return val.skrift !== "antiqua"; }

  function settSkriftform() {
    var fraktur = erFraktur();
    kropp.classList.toggle("lesetekst", !fraktur);
    for (var i = 0; i < nodar.length; i++) {
      nodar[i].node.nodeValue = fraktur ? nodar[i].trykk : nodar[i].antikva;
    }
    var meme = document.getElementById("meme");
    if (fraktur) { meme.setAttribute("hidden", ""); } else { meme.removeAttribute("hidden"); }
  }

  function settVising() {
    rot.setAttribute("data-skrift", val.skrift);
    rot.setAttribute("data-ui", val.ui);
    rot.setAttribute("data-ligatur", val.ligatur);
    rot.setAttribute("data-storleik", val.storleik);
    rot.setAttribute("data-theme", val.tema);
    [["skrift", val.skrift], ["ui", val.ui], ["ligatur", val.ligatur],
     ["storleik", val.storleik], ["tema", val.tema]]
      .forEach(function (par) {
        var knappar = document.querySelectorAll("[data-" + par[0] + "]");
        for (var i = 0; i < knappar.length; i++) {
          knappar[i].setAttribute("aria-pressed",
            String(knappar[i].getAttribute("data-" + par[0]) === par[1]));
        }
      });
  }

  samle(document.body);
  settSkriftform();
  settVising();

  // ---- Lukene under topplinja ----------------------------------------
  // Berre éi om gongen; opnar ein den eine, lukkar den andre seg.
  var stilling = document.getElementById("stilling");
  var luker = [
    { luke: stilling, knapp: document.getElementById("opnastilling") },
    { luke: document.getElementById("stutt"), knapp: document.getElementById("opnastutt") },
    { luke: document.getElementById("om"), knapp: document.getElementById("opnaom") }
  ];

  // Papiret blir rive opp der lesaren er: ei rift blir sett inn framfor
  // den fyrste blokka under topplinja og veks til luka si høgd, so arket
  // under glir ned. Luka sjølv blir aldri flytt - ho ligg fast over
  // opninga. Det var flyttinga som gjorde at ho rørte paa seg før.
  var rift = null;

  function opnaRift(høgd) {
    lukkRift();
    var topp = document.querySelector(".topp");
    var under = topp.getBoundingClientRect().bottom + 4;
    var born = document.querySelectorAll("#innhald > section > *");
    var mål = null;
    for (var i = 0; i < born.length; i++) {
      if (born[i].getBoundingClientRect().top > under) { mål = born[i]; break; }
    }
    if (!mål) return;
    rift = document.createElement("div");
    rift.className = "rift";
    mål.parentNode.insertBefore(rift, mål);
    // Rifta opnar seg nedanfor det lesaren ser, so sida under skal ikkje
    // dra augo med seg: vi flyttar rullinga like mykje som ho veks.
    requestAnimationFrame(function () {
      rift.style.height = høgd + "px";
      rift.classList.add("open");
    });
  }

  function lukkRift() {
    if (!rift) return;
    rift.remove();
    rift = null;
  }

  function visLuke(par, open) {
    if (open) { par.luke.removeAttribute("hidden"); }
    else { par.luke.setAttribute("hidden", ""); }
    par.knapp.setAttribute("aria-expanded", String(open));
    par.knapp.setAttribute("aria-pressed", String(open));
  }


  luker.forEach(function (par) {
    par.knapp.addEventListener("click", function () {
      var open = par.luke.hasAttribute("hidden");
      luker.forEach(function (a) { visLuke(a, a === par && open); });
      if (open) { opnaRift(par.luke.getBoundingClientRect().height); }
      else { lukkRift(); }
    });
  });

  document.addEventListener("keydown", function (e) {
    if (e.key !== "Escape") return;
    luker.forEach(function (par) {
      if (!par.luke.hasAttribute("hidden")) { visLuke(par, false); par.knapp.focus(); }
    });
    lukkRift();
  });

  stilling.addEventListener("click", function (e) {
    var k = e.target.closest("button");
    if (!k) return;
    ["skrift", "ui", "ligatur", "storleik", "tema"].forEach(function (felt) {
      var v = k.getAttribute("data-" + felt);
      if (v) { val[felt] = v; }
    });
    settVising();
    settSkriftform();
    lagre();
  });

  // ---- Hopp til paragraf --------------------------------------------
  var form = document.getElementById("paragrafform");
  form.addEventListener("submit", function (e) {
    e.preventDefault();
    var nr = document.getElementById("paragrafnr").value.trim();
    if (nr) {
      window.location.href = "/paragraf/" + encodeURIComponent(nr);
      return;
    }
    nesteParagraf();
  });

  // Utan tal i feltet tek Hopp deg til neste paragraf nedanfor. Grensa
  // er topplinja: ein paragraf som ligg attom henne er alt passert.
  function nesteParagraf() {
    var grense = document.querySelector(".topp");
    var under = grense ? grense.getBoundingClientRect().bottom + 4 : 4;
    var alle = document.querySelectorAll(".paragraf");
    for (var i = 0; i < alle.length; i++) {
      if (alle[i].getBoundingClientRect().top > under) {
        history.replaceState(null, "", "#" + alle[i].id);
        blink(alle[i]);
        return;
      }
    }
  }

  // ---- Landing: blink der ein hoppar --------------------------------
  function blink(el) {
    if (!el) return;
    el.classList.remove("landa");
    void el.offsetWidth; // tvingar animasjonen til aa byrje paa nytt
    el.classList.add("landa");
    el.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function merkIRegister(ankar) {
    var alle = document.querySelectorAll(".undertitlar a");
    for (var i = 0; i < alle.length; i++) {
      alle[i].setAttribute("aria-current", String(alle[i].getAttribute("data-hopp") === ankar));
    }
  }

  document.addEventListener("click", function (e) {
    var a = e.target.closest("a[data-hopp]");
    if (!a) return;
    e.preventDefault();
    var ankar = a.getAttribute("data-hopp");
    var mål = document.getElementById(ankar);
    if (!mål) return;
    history.replaceState(null, "", "#" + ankar);
    merkIRegister(ankar);
    blink(mål);
  });

  // Klikk paa eit §-nummer: lat blinken og avstanden til topplina gjelda
  // her òg, i staden for nettlesaren sitt raa hopp.
  document.addEventListener("click", function (e) {
    var a = e.target.closest(".paragraf .tal");
    if (!a) return;
    e.preventDefault();
    var id = a.getAttribute("href").slice(1);
    history.replaceState(null, "", "#" + id);
    blink(document.getElementById(id));
  });

  // Kom vi hit med ein #-adresse, hopp dit naar sida er ferdig.
  function blinkFråAdresse() {
    if (!location.hash) return;
    var el = document.getElementById(location.hash.slice(1));
    if (el) { setTimeout(function () { blink(el); }, 60); }
  }
  blinkFråAdresse();

  // ---- Tavler: rad og kolonne under peikaren ------------------------
  function kolonneAv(td) {
    return Array.prototype.indexOf.call(td.parentNode.children, td);
  }

  function settKolonne(tabell, n, klasse) {
    var rader = tabell.rows;
    for (var i = 0; i < rader.length; i++) {
      for (var j = 0; j < rader[i].cells.length; j++) {
        rader[i].cells[j].classList.toggle(klasse, j === n && n >= 0);
      }
    }
  }

  document.addEventListener("mouseover", function (e) {
    var td = e.target.closest(".oppsett td");
    if (!td) return;
    settKolonne(td.closest("table"), kolonneAv(td), "kolonne");
  });

  document.addEventListener("mouseout", function (e) {
    var td = e.target.closest(".oppsett td");
    if (!td) return;
    settKolonne(td.closest("table"), -1, "kolonne");
  });

  document.addEventListener("click", function (e) {
    var td = e.target.closest(".oppsett td");
    if (!td) return;
    var tabell = td.closest("table");
    var rad = td.parentNode;
    var alt = tabell.querySelectorAll("tr.merkt");
    var alt2 = tabell.querySelectorAll("td.kolonne-merkt");
    var stodMerkt = rad.classList.contains("merkt") && td.classList.contains("kolonne-merkt");
    for (var i = 0; i < alt.length; i++) { alt[i].classList.remove("merkt"); }
    for (var j = 0; j < alt2.length; j++) { alt2[j].classList.remove("kolonne-merkt"); }
    if (!stodMerkt) {
      rad.classList.add("merkt");
      settKolonne(tabell, kolonneAv(td), "kolonne-merkt");
    }
  });


  // ---- Ny del henta inn ---------------------------------------------
  document.body.addEventListener("htmx:afterSwap", function (e) {
    if (!e.target || e.target.id !== "innhald") return;
    samle(document.body);
    settSkriftform();
    // Ein ny bolk er henta, og han skal byrje paa toppen. Utan dette
    // vart rullinga staaande der ho var, og eit ankar frå den førre
    // bolken kunne dra oss ned att.
    history.replaceState(null, "", location.pathname);
    scrollTo({ top: 0, behavior: "auto" });
  });

  document.body.addEventListener("htmx:afterOnLoad", function (e) {
    var lenke = e.detail && e.detail.elt;
    if (!lenke || !lenke.closest) return;
    var li = lenke.closest("li");
    if (!li) return;
    var gamle = document.querySelector(".register li.aktiv");
    if (gamle) gamle.classList.remove("aktiv");
    li.classList.add("aktiv");
  });
})();

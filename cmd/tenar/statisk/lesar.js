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

  // ---- Rifta ----------------------------------------------------------
  // Aa opne ei luke er aa rive arket tvers over der lesaren staar. Vi
  // deler <section class="bolk"> i to: blokkene frå og med den fyrste
  // som ligg nedanfor topplinja blir flytte over i eit nytt ark, og luka
  // kjem inn imellom dei to helvtene. Alt ligg i flyten, so rifta rullar
  // med teksten og treng ikkje maalast opp mot vindauget.
  var heim = document.querySelector(".lesefelt"); // der lukene bur naar dei er att
  var rift = null;

  function toppline() {
    var t = document.querySelector(".topp");
    return t ? t.getBoundingClientRect().bottom : 0;
  }

  // delepunkt er det snittet mellom to blokker som ligg nærast topplinja
  // - ovanfor eller nedanfor, det som er kortast unna. Tok vi i staden
  // fyrste blokka heilt nedanfor linja, hamna rifta langt nede kvar gong
  // ei høg blokk laag tvers over henne: maalt 400 px ned ved rull 9000.
  // Talet n er lovleg og tyder eit snitt etter siste blokka.
  function delepunkt(ark, tl) {
    var best = 0, kortast = Infinity;
    for (var i = 0; i <= ark.children.length; i++) {
      var r = i < ark.children.length
        ? ark.children[i].getBoundingClientRect().top
        : ark.getBoundingClientRect().bottom;
      if (Math.abs(r - tl) < kortast) { kortast = Math.abs(r - tl); best = i; }
    }
    return best;
  }


  // Kvar luke hugsar kven ho stod framfor, so ho kjem attende dit.
  luker.forEach(function (par) { par.attende = par.luke.nextSibling; });

  function plassen(luke) {
    for (var i = 0; i < luker.length; i++) {
      if (luker[i].luke === luke) { return luker[i].attende; }
    }
    return document.getElementById("innhald");
  }

  function riv(luke) {
    var innhald = document.getElementById("innhald");
    var ark = innhald ? innhald.querySelector(".bolk") : null;
    var ny = document.createElement("div");
    ny.className = "rift";
    // Arket er dansk (lang="da"); menyane er nynorske. Utan dette arvar
    // dei språket til teksten dei blir lagde inn i, og ein skjermlesar
    // les heile visingsmenyen med dansk uttale.
    ny.lang = "nn";
    ny.appendChild(luke);

    var tl = toppline();
    if (!ark) { (innhald || heim).appendChild(ny); }
    else if (ark.getBoundingClientRect().top >= tl) {
      // Heile arket ligg under linja - lesaren staar over papiret. Daa
      // er det ingenting aa rive, og luka legg seg føre arket. Klassa
      // trekkjer bort lufta main held mot kanten.
      ny.className += " over";
      ark.parentNode.insertBefore(ny, ark);
    } else {
      var i = delepunkt(ark, tl);
      var nedre = document.createElement("section");
      nedre.className = ark.className + " nedre";
      ark.classList.add("øvre");
      while (ark.children.length > i) { nedre.appendChild(ark.children[i]); }
      ark.parentNode.insertBefore(ny, ark.nextSibling);
      ny.parentNode.insertBefore(nedre, ny.nextSibling);
    }

    // Éi retting for alle greinene. Rifta er framleis samanfalden, so
    // ingenting nedanfor har flytt seg enno, og snittet staar der det
    // kjem til aa staa; vi dyttar sida so det legg seg paa linja.
    //
    // Maalinga maa gjerast etter delinga, ikkje før: delinga byter
    // flex-gapet ut med snittmarga arket har, og det flytte snittet 15 px.
    //
    // Ho maa ògso gjerast i «over»-greina. Der reiv marga i stilarket
    // aaleine, og det stemmer berre naar sida staar heilt øvst - ved rull
    // 20, 40 og 60 opna rifta seg 20, 40 og 60 px OVER linja, altso bak
    // stonga, der toppen av menyen ikkje var til aa naa.
    rift = ny;
    if (ny.isConnected) {
      scrollBy(0, Math.round(ny.getBoundingClientRect().top - tl));
    }
    // Ei bilete-ramme fram, so nettlesaren har ei høgd aa gaa ut ifrå.
    requestAnimationFrame(function () { ny.classList.add("open"); });
  }

  // lim spør DOM-en i staden for aa lite paa variabelen. Er sida henta
  // att frå htmx si historie, kan arket vera lagra i rive tilstand medan
  // variabelen er tom - og daa vart delinga staaande for godt.
  function lim() {
    var r = document.querySelector(".rift");
    rift = null;
    if (!r) return;
    var luke = r.firstElementChild;
    // Luka skal attende paa nøyaktig sin eigen plass. Vart ho lagd bakarst
    // i lesefeltet, hamna ho etter heile kapitlet i tabb- og leserekkja;
    // og la vi henne berre framfor teksten, bytte dei tre menyane
    // innbyrdes rekkjefylgje kvar gong ei av dei hadde vore open.
    if (luke) { heim.insertBefore(luke, plassen(luke)); }
    var nedre = r.nextElementSibling;
    var øvre = r.previousElementSibling;
    if (nedre && nedre.classList.contains("nedre") && øvre && øvre.classList.contains("øvre")) {
      while (nedre.children.length) { øvre.appendChild(nedre.children[0]); }
      øvre.classList.remove("øvre");
      nedre.remove();
    }
    r.remove();
  }

  function visLuke(par, open) {
    if (open) { par.luke.removeAttribute("hidden"); riv(par.luke); }
    else { par.luke.setAttribute("hidden", ""); }
    par.knapp.setAttribute("aria-expanded", String(open));
    par.knapp.setAttribute("aria-pressed", String(open));
  }

  function lukkAlle() {
    var lukka = false;
    luker.forEach(function (par) {
      if (!par.luke.hasAttribute("hidden")) { visLuke(par, false); lukka = true; }
    });
    lim();
    return lukka;
  }

  luker.forEach(function (par) {
    par.knapp.addEventListener("click", function () {
      var skalOpne = par.luke.hasAttribute("hidden");
      lukkAlle();
      if (skalOpne) { visLuke(par, true); }
    });
  });

  document.addEventListener("keydown", function (e) {
    if (e.key !== "Escape") return;
    luker.forEach(function (par) {
      if (!par.luke.hasAttribute("hidden")) { par.knapp.focus(); }
    });
    lukkAlle();
  });

  // htmx byter ut heile <main>. Er arket rive naar det skjer, blir baade
  // rifta og luka kasta med. Vi limer att fyrst.
  document.body.addEventListener("htmx:beforeSwap", lukkAlle);

  // Vart sida lagra i htmx si historie medan arket var rive, kjem ho att
  // delt. lim spør DOM-en, so denne eine kallet rettar det opp.
  lim();

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
    // Ryggskiltet staar utanfor #registerliste, som er det htmx byter ut,
    // so det maa merkjast her. Elles vart det staaande som aktivt etter
    // at ein hadde forlate tittelbladet.
    var skilt = document.querySelector(".register-topp");
    if (skilt) { skilt.classList.toggle("aktiv", location.pathname === "/del/0"); }
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

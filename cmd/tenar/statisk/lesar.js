// Lesarval og småting som ikkje treng tenaren: skriftform, skriftval,
// drakt, hopp til paragraf, og utheving i tavlene.
(function () {
  "use strict";

  var rot = document.documentElement;
  var kropp = document.body;
  var LAGER = "aasen-lesar";

  var val = {
    trykk: true,
    skrift: "maguntia",
    ui: "antiqua",
    ligatur: "trykk",
    storleik: "vanleg",
    form: "fri",
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

  function settSkriftform() {
    kropp.classList.toggle("lesetekst", !val.trykk);
    for (var i = 0; i < nodar.length; i++) {
      nodar[i].node.nodeValue = val.trykk ? nodar[i].trykk : nodar[i].antikva;
    }
    var k = document.getElementById("skrift");
    k.setAttribute("aria-pressed", String(val.trykk));
    k.textContent = val.trykk ? "Et tu, Antiqua?" : "Attende til frakturen";
  }

  function settVising() {
    rot.setAttribute("data-skrift", val.skrift);
    rot.setAttribute("data-ui", val.ui);
    rot.setAttribute("data-ligatur", val.ligatur);
    rot.setAttribute("data-storleik", val.storleik);
    rot.setAttribute("data-form", val.form);
    rot.setAttribute("data-theme", val.tema);
    [["skrift", val.skrift], ["ui", val.ui], ["ligatur", val.ligatur],
     ["storleik", val.storleik], ["form", val.form], ["tema", val.tema]]
      .forEach(function (par) {
        var knappar = document.querySelectorAll("[data-" + par[0] + "]");
        for (var i = 0; i < knappar.length; i++) {
          knappar[i].setAttribute("aria-pressed",
            String(knappar[i].getAttribute("data-" + par[0]) === par[1]));
        }
      });
  }

  samle(document.getElementById("innhald"));
  settSkriftform();
  settVising();

  document.getElementById("skrift").addEventListener("click", function () {
    val.trykk = !val.trykk;
    settSkriftform();
    lagre();
  });

  var stilling = document.getElementById("stilling");
  var opna = document.getElementById("opnastilling");
  opna.addEventListener("click", function () {
    var open = stilling.hasAttribute("hidden");
    if (open) { stilling.removeAttribute("hidden"); } else { stilling.setAttribute("hidden", ""); }
    opna.setAttribute("aria-expanded", String(open));
    opna.setAttribute("aria-pressed", String(open));
  });

  stilling.addEventListener("click", function (e) {
    var k = e.target.closest("button");
    if (!k) return;
    ["skrift", "ui", "ligatur", "storleik", "form", "tema"].forEach(function (felt) {
      var v = k.getAttribute("data-" + felt);
      if (v) { val[felt] = v; }
    });
    settVising();
    lagre();
  });

  // ---- Hopp til paragraf --------------------------------------------
  var form = document.getElementById("paragrafform");
  form.addEventListener("submit", function (e) {
    e.preventDefault();
    var nr = document.getElementById("paragrafnr").value.trim();
    if (nr) { window.location.href = "/paragraf/" + encodeURIComponent(nr); }
  });

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

  // ---- Rullesporing --------------------------------------------------
  // Bolkane kjem etter kvarandre i eitt renn, so registeret kan ikkje
  // lita paa klikk aaleine: det maa fylgja det lesaren faktisk ser.
  var sjaaar = null;
  if ("IntersectionObserver" in window) {
    sjaaar = new IntersectionObserver(function (rader) {
      var beste = null;
      for (var i = 0; i < rader.length; i++) {
        if (!rader[i].isIntersecting) continue;
        if (!beste || rader[i].boundingClientRect.top < beste.boundingClientRect.top) {
          beste = rader[i];
        }
      }
      if (beste) { merkDel(beste.target.getAttribute("data-del")); }
    }, { rootMargin: "-25% 0px -60% 0px" });
  }

  function merkDel(id) {
    if (id === null) return;
    var lenke = document.querySelector('.register a[href="/del/' + id + '"]');
    if (!lenke) return;
    var li = lenke.closest("li");
    var gamle = document.querySelector(".register li.aktiv");
    if (gamle === li) return;
    if (gamle) gamle.classList.remove("aktiv");
    li.classList.add("aktiv");
  }

  function fylg(rot) {
    if (!sjaaar) return;
    var bolkar = (rot || document).querySelectorAll(".bolk");
    for (var i = 0; i < bolkar.length; i++) { sjaaar.observe(bolkar[i]); }
  }
  fylg(document);

  // ---- Ny del henta inn ---------------------------------------------
  document.body.addEventListener("htmx:afterSwap", function (e) {
    var mål = e.target;
    if (!mål) return;
    if (mål.id === "innhald") {
      samle(mål);
      settSkriftform();
      fylg(mål);
      blinkFråAdresse();
      return;
    }
    // Framhald: ein ny bolk er lagd til nedanfor.
    if (mål.closest && mål.closest("#innhald")) {
      samle(mål.parentNode);
      settSkriftform();
      fylg(mål.parentNode);
    }
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

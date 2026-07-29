// Lesarval som ikkje treng tenaren: skriftform og lys/mørk drakt.
(function () {
  "use strict";

  var rot = document.documentElement;
  var kropp = document.body;
  var LAGER = "aasen-lesar";

  var val = { trykk: true, mørk: matchMedia("(prefers-color-scheme: dark)").matches };
  try {
    var lagra = JSON.parse(localStorage.getItem(LAGER) || "{}");
    if (typeof lagra.trykk === "boolean") val.trykk = lagra.trykk;
    if (typeof lagra.mørk === "boolean") val.mørk = lagra.mørk;
  } catch (e) { /* utan lager går det fint òg */ }

  function lagre() {
    try { localStorage.setItem(LAGER, JSON.stringify(val)); } catch (e) {}
  }

  // Lang ſ er ei reint visuell form av same bokstav, so lesetekst-visinga
  // byter henne ut med rund s. Originalen blir teken vare på, slik at ein
  // kan gå att utan å hente teksten på nytt.
  var nodar = [];
  function samle(rotnode) {
    nodar = [];
    var gaa = document.createTreeWalker(rotnode, NodeFilter.SHOW_TEXT, null);
    var n;
    while ((n = gaa.nextNode())) {
      if (n.nodeValue.indexOf("ſ") !== -1) {
        nodar.push({ node: n, trykk: n.nodeValue, lese: n.nodeValue.replace(/ſ/g, "s") });
      }
    }
  }

  function settSkrift() {
    kropp.classList.toggle("lesetekst", !val.trykk);
    for (var i = 0; i < nodar.length; i++) {
      nodar[i].node.nodeValue = val.trykk ? nodar[i].trykk : nodar[i].lese;
    }
    var k = document.getElementById("skrift");
    k.setAttribute("aria-pressed", String(val.trykk));
    k.textContent = val.trykk ? "Trykk 1864" : "Lesetekst";
  }

  function settTema() {
    rot.setAttribute("data-theme", val.mørk ? "dark" : "light");
    var k = document.getElementById("tema");
    k.setAttribute("aria-pressed", String(val.mørk));
    k.textContent = val.mørk ? "Ljos" : "Mørk";
  }

  samle(document.getElementById("innhald"));
  settSkrift();
  settTema();

  document.getElementById("skrift").addEventListener("click", function () {
    val.trykk = !val.trykk;
    settSkrift();
    lagre();
  });

  document.getElementById("tema").addEventListener("click", function () {
    val.mørk = !val.mørk;
    settTema();
    lagre();
  });

  // Ny del henta inn: teksten er byta ut, so nodane må samlast på nytt.
  document.body.addEventListener("htmx:afterSwap", function (e) {
    if (e.target && e.target.id === "innhald") {
      samle(e.target);
      settSkrift();
    }
  });

  // Merk kva for ein del som er open i registeret.
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

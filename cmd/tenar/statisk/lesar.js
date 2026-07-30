// Lesarval og småting som ikkje treng tenaren: skriftform, skriftval,
// drakt, hopp til paragraf, og utheving i tavlene.
(function () {
  "use strict";

  var rot = document.documentElement;
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
  // Teksten i arket blir samla om att for kvar ny del; ſ-en i apparatet -
  // «Grenſeſnitt», Om-luka, ryggen - berre ein gong.
  //
  // Dei maa haldast frå kvarandre. Eit sveip over heile body for kvart
  // delbyte var ikkje idempotent: stod visinga i antikva, hadde
  // apparatnodane alt fenge rund s, «ſ» fanst ikkje i dei lenger, og dei
  // fall ut av lista for godt. Gjekk ein so attende til fraktur, kom arket
  // att - det var nyhenta frå tenaren - medan menyane stod med rund s til
  // sida vart lasta om.
  var nodar = [];
  var apparatnodar = [];

  function finn(rotnode) {
    var ut = [];
    if (!rotnode) return ut;
    var gaa = document.createTreeWalker(rotnode, NodeFilter.SHOW_TEXT, null);
    var n;
    while ((n = gaa.nextNode())) {
      if (n.nodeValue.indexOf("ſ") !== -1) {
        ut.push({ node: n, trykk: n.nodeValue, antikva: n.nodeValue.replace(/ſ/g, "s") });
      }
    }
    return ut;
  }

  // Antikva er eit skriftval jamsides frakturane. Er han vald, gaar
  // teksten over til latinske bokstavar med rund s.
  function erFraktur() { return val.skrift !== "antiqua"; }

  function byt(liste, fraktur) {
    for (var i = 0; i < liste.length; i++) {
      liste[i].node.nodeValue = fraktur ? liste[i].trykk : liste[i].antikva;
    }
  }

  function settSkriftform() {
    var fraktur = erFraktur();
    byt(nodar, fraktur);
    byt(apparatnodar, fraktur);
    var meme = document.getElementById("meme");
    if (meme) {
      if (fraktur) { meme.setAttribute("hidden", ""); } else { meme.removeAttribute("hidden"); }
    }
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

  (function samleFyrst() {
    var innhald = document.getElementById("innhald");
    finn(document.body).forEach(function (r) {
      if (innhald && innhald.contains(r.node)) { nodar.push(r); }
      else { apparatnodar.push(r); }
    });
  })();
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
  var stong = document.querySelector(".topp");

  function toppline() {
    return stong ? stong.getBoundingClientRect().bottom : 0;
  }

  // riftlinja er høgda i vindauget der arket riv seg: snittet skal liggje
  // slik at luka hamnar midt i leseflata - midtpunktet hennar og
  // midtpunktet paa flata paa same line. Daa har opninga like mykje papir
  // over seg som under, og ho ligg der auga alt er.
  // Leseflata, ikkje heile vindauget: stonga dekkjer toppen, so midten av
  // det ein faktisk ser ligg eit stykke nedanfor midten av ruta.
  // Aldri over stonga, uansett kor høg luka er.
  function riftlinje(lukehøgd) {
    var tl = toppline();
    return Math.max(tl, Math.round((tl + innerHeight) / 2 - (lukehøgd || 0) / 2));
  }

  // landing er der eit hopp legg maalet sitt. Vi les det ut av stilarket i
  // staden for aa rekne det om att her: scroll-margin-top er nettopp det
  // talet scrollIntoView kjem til aa bruke, so dei to kan ikkje bli usamde.
  function landing() {
    var p = document.querySelector(".bolk > [id]");
    return (p && parseFloat(getComputedStyle(p).scrollMarginTop)) || toppline();
  }

  // Stilarket kan ikkje vite kor høg den faste linja er: ho kjem av
  // innhaldet sitt, ho fylgjer skrifta lesaren vel i grensesnittet, og
  // knapperekkja bryt seg i to paa smal skjerm - maalt frå 61 til 104 px
  // mot dei 63 stilarket gissa paa. Alt som skal liggje klar av linja -
  // margen over arket, lufta rundt luka, avstanden eit hopp landar paa -
  // reknar frå --toppline, so vi maaler henne og skriv maalet attende inn.
  function maalStonga() {
    if (!stong) return;
    rot.style.setProperty("--toppline",
      stong.getBoundingClientRect().height + "px");
  }
  if (stong && window.ResizeObserver) {
    // Fyrer med ein gong ved observe, so det dekkjer fyrste maalinga òg.
    new ResizeObserver(maalStonga).observe(stong);
  } else {
    maalStonga();
    addEventListener("resize", maalStonga);
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

  // helvtene gjev dei to arkhelvtene som ligg inntil rifta, om ho vart laga
  // ved aa dele eit ark. Ligg luka føre arket, er det ingen.
  function helvtene(r) {
    var ut = [];
    var f = r.previousElementSibling, e = r.nextElementSibling;
    if (f && f.classList.contains("øvre")) { ut.push(f); }
    if (e && e.classList.contains("nedre")) { ut.push(e); }
    return ut;
  }

  // opne set i gang animasjonen. Vi tvingar fram ei omrekning i staden
  // for aa vente paa ei biletramme: requestAnimationFrame fyrer ikkje
  // naar fana ligg i bakgrunnen, og daa vart luka staaande samanfalden
  // med berre si eiga utfylling synleg.
  //
  // Omrekninga fester ògso utgangsstoda for arkhelvtene, og fyrst etter
  // henne slaar vi paa overgangen deira. Gjorde vi det før, ville
  // botnmarga som fall bort i delinga - 7,8em - krympe seg gjennom teksten
  // i staden for berre aa vera borte.
  function opne(r) {
    void r.getBoundingClientRect().height;
    r.classList.add("open");
    helvtene(r).forEach(function (h) { h.classList.add("mjuk", "open"); });
  }

  // att tek rørsla andre vegen: snittmarga fell saman samstundes med
  // opninga, so papiret lukkar seg frå baae kantar.
  function att(r) {
    r.classList.remove("open");
    helvtene(r).forEach(function (h) { h.classList.remove("open"); });
  }

  function riv(luke) {
    // Maal luka her, medan ho enno staar i lesefeltet. Etter at ho er flytt
    // inn i rifta er ho lausrive frå dokumentet til vi set rifta inn, og
    // daa maaler ho null. Breidda er ikkje heilt den same som inne i rifta,
    // so dette er eit overslag - godt nok til aa velje kva snitt vi deler
    // ved. Den endelege rettinga nedanfor maaler paa nytt.
    var lukehøgd = luke.getBoundingClientRect().height;
    var innhald = document.getElementById("innhald");
    var ark = innhald ? innhald.querySelector(".bolk") : null;
    var ny = document.createElement("div");
    ny.className = "rift";
    // Arket er dansk (lang="da"); menyane er nynorske. Utan dette arvar
    // dei språket til teksten dei blir lagde inn i, og ein skjermlesar
    // les heile visingsmenyen med dansk uttale.
    ny.lang = "nn";
    // Luka ligg i ei innpakning. Rada i rutenettet kan ikkje falle heilt
    // saman rundt eit barn som ber loddrett utfylling - maalt 24 px rad
    // rundt eit barn med 12 - so lufta maa liggje eit hakk lenger inne.
    // Daa treng ho ikkje animerast, og det er heile poenget: animerte vi
    // henne paa rifta, skauv ho luka 63 px nedover medan opninga voks, og
    // teksten rende. No staar alt i ro, og luka med lufta si blir rulla
    // fram fraa toppen som ei rullegardin.
    var pakke = document.createElement("div");
    pakke.className = "riftpakke";
    pakke.appendChild(luke);
    ny.appendChild(pakke);

    // To ulike spørsmaal, og dei maa ikkje blandast:
    //   toppline()  - ligg heile arket under stonga? Daa staar lesaren over
    //                 papiret, og luka høyrer føre arket.
    //   riftlinje() - kvar inne i arket skal snittet gaa?
    // Med riftlinja i baae rollene vart svaret paa det fyrste alltid nei -
    // midtlinja ligg langt nede - so arket vart delt ògso naar lesaren stod
    // heilt øvst, i staden for at luka la seg føre og skauv arket ned.
    var tl = riftlinje(lukehøgd);
    if (!ark) { (innhald || heim).appendChild(ny); }
    else if (ark.getBoundingClientRect().top >= toppline()) {
      // Heile arket ligg under linja - lesaren staar over papiret. Daa er
      // det ingenting aa rive, og luka legg seg føre arket, nett der
      // arkkanten var.
      ark.parentNode.insertBefore(ny, ark);
    } else {
      var i = delepunkt(ark, tl);
      var nedre = document.createElement("section");
      nedre.className = ark.className + " nedre";
      ark.classList.add("øvre");
      // Innrykket i satsen kjem av kva blokka staar ETTER: eit avsnitt som
      // held fram inne i same § er innrykt, det fyrste etter ei overskrift
      // staar flust. Den fyrste blokka vi flyttar over mistar syskenet sitt
      // og dermed regelen - so eit innrykt avsnitt vart flust i det arket
      // reiv seg, og rykte inn att naar det vart limt. Vi tek vare paa det
      // som ei eiga klasse i staden.
      var fyrste = ark.children[i];
      var rykk = fyrste && parseFloat(getComputedStyle(fyrste).textIndent) > 0;
      while (ark.children.length > i) { nedre.appendChild(ark.children[i]); }
      if (rykk && nedre.firstElementChild) {
        nedre.firstElementChild.classList.add("rykk");
      }
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
    // Ho gjeld ògso naar luka legg seg føre arket. Den greina hadde eit
    // eige unnatak - fyrst ein negativ marg i stilarket, sidan ei
    // nullstilt topputfylling - og det stemte berre naar sida stod heilt
    // øvst: ved rull 20, 40 og 60 opna rifta seg 20, 40 og 60 px over
    // linja, altso bak stonga, der toppen av menyen ikkje var aa naa. No
    // gaar alle greinene same vegen, og lufta over luka er den same.
    // Ingen rulling her. Rifta blir plassert ved aa VELJE snitt, ikkje ved
    // aa flytte sida: delepunkt tek det snittet som ligg nærast riftlinja,
    // og so opnar ho seg der. Dytta vi sida attaat for aa treffe linja
    // nøyaktig, rykte heile det øvre arket - maalt 8 px - og det er nettopp
    // det ein ikkje kan ha: papiret ein les paa skal liggje bom stille, og
    // berre det som er under snittet skal flytte seg.
    //
    // Det gjev ògso den rette rørsla naar ein ikkje har rulla: der ligg
    // luka føre arket, og naar ho veks, skyv ho heile arket nedover i
    // staden for at sida rullar under det.
    opne(ny);
  }

  // lim spør DOM-en i staden for aa lite paa variabelen. Er sida henta
  // att frå htmx si historie, kan arket vera lagra i rive tilstand medan
  // variabelen er tom - og daa vart delinga staaande for godt.
  function lim() {
    var r = document.querySelector(".rift");
    if (!r) return;
    // Luka ligg inne i innpakninga, ikkje rett under rifta.
    var luke = r.querySelector(".stilling, .luke");
    // Luka skal attende paa nøyaktig sin eigen plass. Vart ho lagd bakarst
    // i lesefeltet, hamna ho etter heile kapitlet i tabb- og leserekkja;
    // og la vi henne berre framfor teksten, bytte dei tre menyane
    // innbyrdes rekkjefylgje kvar gong ei av dei hadde vore open.
    if (luke) { heim.insertBefore(luke, plassen(luke)); }
    var nedre = r.nextElementSibling;
    var øvre = r.previousElementSibling;

    // Ingen rulling her. Snittet ber eit halvt blokkgap paa kvar side, so
    // liminga er geometrisk nøytral: 12 + 0 + 12 før, eitt blokkgap etter.
    // Ein ankerkompensasjon oppaa det retta noko som alt var rett, og
    // sparka sida 92 px oppover kvar gong det var ein rest att av rifta -
    // det var den rykkinga ein saag ved kvar opning og lukking.
    if (nedre && nedre.classList.contains("nedre") && øvre && øvre.classList.contains("øvre")) {
      // Klassa var berre eit plaster medan arket laag i to; naa held
      // syskenregelen att av seg sjølv.
      if (nedre.firstElementChild) { nedre.firstElementChild.classList.remove("rykk"); }
      while (nedre.children.length) { øvre.appendChild(nedre.children[0]); }
      øvre.classList.remove("øvre", "mjuk", "open");
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

  // Kor lenge rifta bruker paa aa opne og lukke seg. Same talet som
  // --riftetid i stilarket; hald dei like.
  var RIFTETID = 340;
  var mjuk = !matchMedia("(prefers-reduced-motion: reduce)").matches;
  var ventar = null;   // lukene som er paa veg att, og timaren deira

  // fullfør gjer ei lukking som er i gang ferdig med ein gong. Alt som vil
  // rive eller byte ark maa gaa gjennom henne fyrst, so vi aldri har to
  // rifter eller ein timar som slaar til etter at neste luke er opna.
  function fullfør() {
    if (!ventar) return;
    clearTimeout(ventar.timar);
    var att = ventar.luker;
    ventar = null;
    att.forEach(function (par) { visLuke(par, false); });
    lim();
  }

  // lukkAlle tek att alle opne luker. Med snøgt = true skjer det i same
  // biletramma - det maa det naar htmx er i ferd med aa byte ut arket, og
  // naar vi byter frå ei luke til ei anna.
  //
  // Elles fell rifta saman like mjukt som ho opna seg. Luka kan ikkje
  // gøymast fyrst: [hidden] gjev display: none, rada i rutenettet blir tom
  // i same ramma, og rifta ville hoppe att utan rørsle - lukkinga var eit
  // hardt kutt der opninga tok 0,34 s. So vi tek berre «open» av, lèt luka
  // staa, og gøymer henne og limer arket saman naar rørsla er over.
  function lukkAlle(snøgt) {
    fullfør();
    var opne = [];
    luker.forEach(function (par) {
      if (!par.luke.hasAttribute("hidden")) { opne.push(par); }
    });
    if (!opne.length) { lim(); return false; }

    var r = document.querySelector(".rift");
    if (snøgt || !mjuk || !r) {
      opne.forEach(function (par) { visLuke(par, false); });
      lim();
      return true;
    }
    // Knappen melder att med ein gong, same kor lenge rørsla tek.
    opne.forEach(function (par) {
      par.knapp.setAttribute("aria-expanded", "false");
      par.knapp.setAttribute("aria-pressed", "false");
    });
    att(r);
    // Vi limer naar rørsla faktisk er over, ikkje naar klokka seier ho
    // burde vera det. Ein timer aaleine kan fyre eit bilete for tidleg, og
    // daa blir helvtene rykte ut medan opninga enno har ein rest og
    // snittskuggen enno er synleg - eit glimt av skugge oppaa papiret.
    // Timaren staar att som sikring: transitionend kjem ikkje om fana ligg
    // i bakgrunnen, eller om rørsla blir avbroten.
    var ferdig = function (e) {
      if (e && (e.target !== r || e.propertyName !== "grid-template-rows")) { return; }
      r.removeEventListener("transitionend", ferdig);
      fullfør();
    };
    r.addEventListener("transitionend", ferdig);
    ventar = { luker: opne, timar: setTimeout(fullfør, RIFTETID + 60) };
    return true;
  }

  luker.forEach(function (par) {
    par.knapp.addEventListener("click", function () {
      var skalOpne = par.luke.hasAttribute("hidden");
      // Byter vi luke, skal den gamle vekk i same ramma: to rifter kan
      // ikkje staa opne, og aa vente paa at den eine fell saman før den
      // andre riv seg opp ville gjere eit klikk til ei halv sekunds venting.
      lukkAlle(skalOpne);
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
  // rifta og luka kasta med. Vi limer att fyrst - og det maa skje no, ikkje
  // etter ei rørsle, so arket er heilt før byttet.
  document.body.addEventListener("htmx:beforeSwap", function () { lukkAlle(true); });

  // Vart sida lagra i htmx si historie medan arket var rive, kjem ho att
  // delt. lim spør DOM-en, so denne eine kallet rettar det opp.
  lim();

  // Byter ein grad, skrift eller ligaturar, flyt heile satsen om. Utan eit
  // anker misser lesaren staden sin: rullestoda staar att paa same
  // pikselen medan teksten under han har flytt seg, og di lenger inne i
  // boka ein er, di lenger unna hamnar ein - paa «Stor» er satsen 14 %
  // høgare, so ved rull 9000 er ein over tusen pikslar frå der ein las.
  // Vi festar oss i den fyrste blokka som er synleg under stonga.
  function staden() {
    var grense = toppline();
    var blokker = document.querySelectorAll(".bolk > *");
    for (var i = 0; i < blokker.length; i++) {
      var r = blokker[i].getBoundingClientRect();
      if (r.bottom > grense) { return { el: blokker[i], y: r.top }; }
    }
    return null;
  }

  function attTil(stad) {
    if (!stad || !stad.el.isConnected) return;
    scrollBy(0, Math.round(stad.el.getBoundingClientRect().top - stad.y));
  }

  stilling.addEventListener("click", function (e) {
    var k = e.target.closest("button");
    if (!k) return;
    ["skrift", "ui", "ligatur", "storleik", "tema"].forEach(function (felt) {
      var v = k.getAttribute("data-" + felt);
      if (v) { val[felt] = v; }
    });
    var stad = staden();
    settVising();
    settSkriftform();
    lagre();
    attTil(stad);
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

  // Utan tal i feltet tek Hopp deg til neste paragraf nedanfor. Grensa er
  // der eit hopp landar - ikkje sjølve topplinja. Med linja som grense
  // stod Hopp fast: paragrafen han nett hadde hoppa til laag 88 px ned,
  // altso nedanfor linja si 68, so han fann han om att og blinka i staden
  // for aa gaa vidare.
  function nesteParagraf() {
    var under = landing() + 4;
    var alle = document.querySelectorAll(".paragraf");
    for (var i = 0; i < alle.length; i++) {
      if (alle[i].getBoundingClientRect().top > under) {
        history.replaceState(null, "", "#" + alle[i].id);
        blink(alle[i]);
        return;
      }
    }
  }

  // ---- Landing: ein glød der ein hoppar ------------------------------
  // Klassa maa av att naar gløden er utbrend. Vart ho staaande, laag ho
  // paa blokka for godt - og ein CSS-animasjon byrjar heilt paa nytt kvar
  // gong elementet blir sett inn i treet att. Naar arket reiv seg, flytte
  // riv() blokkene under snittet over i den nye helvta, og daa slo gløden
  // til om att paa ei blokk ingen hadde hoppa til: eit glimt tvers over
  // teksten kvar gong ei luke opna eller lukka seg.
  function blink(el) {
    if (!el) return;
    el.classList.remove("landa");
    void el.offsetWidth; // tvingar animasjonen til aa byrje paa nytt
    el.classList.add("landa");
    el.addEventListener("animationend", function av(e) {
      if (e.animationName !== "landing") return;
      el.classList.remove("landa");
      el.removeEventListener("animationend", av);
    });
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
    nodar = finn(e.target);
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

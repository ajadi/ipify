// ipify-api/api
//
// Borless fork: adds ?format=html with i18n via Accept-Language detection.
// Plain-text default is preserved for curl/scripts.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rdegges/ipify-api/models"
)

// ptrLookupTimeout caps reverse-DNS calls so a slow/unreachable resolver
// can't stall the response.
const ptrLookupTimeout = 1500 * time.Millisecond

// lookupPTR returns the first reverse-DNS name for ip with trailing dot
// stripped, or "" if the lookup fails, times out, or yields no result.
func lookupPTR(ip string) string {
	if ip == "" || ip == "<nil>" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), ptrLookupTimeout)
	defer cancel()
	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

// ptrLabels holds the localized label for the reverse-DNS row in the HTML
// page. Falls back to English when the active locale isn't listed.
var ptrLabels = map[string]string{
	"en": "Reverse DNS",
	"ru": "Обратный DNS",
	"uk": "Зворотний DNS",
	"es": "DNS inverso",
	"fr": "DNS inverse",
	"de": "Reverse DNS",
	"it": "DNS inverso",
	"pt": "DNS reverso",
	"tr": "Ters DNS",
	"zh": "反向 DNS",
	"ja": "逆引き DNS",
	"ko": "역방향 DNS",
	"hi": "रिवर्स DNS",
	"bn": "রিভার্স DNS",
	"id": "DNS terbalik",
	"vi": "DNS ngược",
	"th": "DNS ย้อนกลับ",
	"pl": "Odwrotny DNS",
	"nl": "Omgekeerde DNS",
	"sv": "Omvänd DNS",
	"da": "Omvendt DNS",
	"no": "Omvendt DNS",
	"fi": "Käänteinen DNS",
	"cs": "Reverzní DNS",
	"sk": "Reverzné DNS",
	"ro": "DNS invers",
	"hu": "Fordított DNS",
	"el": "Αντίστροφο DNS",
	"he": "DNS הפוך",
	"ar": "DNS العكسي",
	"fa": "DNS معکوس",
}

func ptrLabel(lang string) string {
	if v, ok := ptrLabels[lang]; ok {
		return v
	}
	return ptrLabels["en"]
}

type locale struct {
	Lang       string // BCP-47 language attribute
	Dir        string // "" (ltr default) or "rtl"
	Title      string
	Label      string
	Hint       string
	HintCopied string
	APIIntro   string
	Footer     string
}

// Translations are intentionally short and neutral. Add more by inserting
// entries here keyed by primary language subtag (lowercase).
var locales = map[string]locale{
	"en": {
		Lang:       "en",
		Title:      "My IP — ip.borless.com",
		Label:      "Your public IP address",
		Hint:       "click to copy",
		HintCopied: "copied",
		APIIntro:   "For scripts and apps:",
		Footer:     "borless.com",
	},
	"ru": {
		Lang:       "ru",
		Title:      "Мой IP — ip.borless.com",
		Label:      "Ваш публичный IP-адрес",
		Hint:       "кликните, чтобы скопировать",
		HintCopied: "скопировано",
		APIIntro:   "Для скриптов и приложений:",
		Footer:     "borless.com",
	},
	"uk": {
		Lang:       "uk",
		Title:      "Моя IP — ip.borless.com",
		Label:      "Ваша публічна IP-адреса",
		Hint:       "натисніть, щоб скопіювати",
		HintCopied: "скопійовано",
		APIIntro:   "Для скриптів та застосунків:",
		Footer:     "borless.com",
	},
	"es": {
		Lang:       "es",
		Title:      "Mi IP — ip.borless.com",
		Label:      "Tu dirección IP pública",
		Hint:       "haz clic para copiar",
		HintCopied: "copiado",
		APIIntro:   "Para scripts y aplicaciones:",
		Footer:     "borless.com",
	},
	"fr": {
		Lang:       "fr",
		Title:      "Mon IP — ip.borless.com",
		Label:      "Votre adresse IP publique",
		Hint:       "cliquez pour copier",
		HintCopied: "copié",
		APIIntro:   "Pour les scripts et applications :",
		Footer:     "borless.com",
	},
	"de": {
		Lang:       "de",
		Title:      "Meine IP — ip.borless.com",
		Label:      "Ihre öffentliche IP-Adresse",
		Hint:       "zum Kopieren klicken",
		HintCopied: "kopiert",
		APIIntro:   "Für Skripte und Anwendungen:",
		Footer:     "borless.com",
	},
	"it": {
		Lang:       "it",
		Title:      "Il mio IP — ip.borless.com",
		Label:      "Il tuo indirizzo IP pubblico",
		Hint:       "clicca per copiare",
		HintCopied: "copiato",
		APIIntro:   "Per script e applicazioni:",
		Footer:     "borless.com",
	},
	"pt": {
		Lang:       "pt",
		Title:      "Meu IP — ip.borless.com",
		Label:      "Seu endereço IP público",
		Hint:       "clique para copiar",
		HintCopied: "copiado",
		APIIntro:   "Para scripts e aplicativos:",
		Footer:     "borless.com",
	},
	"tr": {
		Lang:       "tr",
		Title:      "IP'm — ip.borless.com",
		Label:      "Genel IP adresiniz",
		Hint:       "kopyalamak için tıklayın",
		HintCopied: "kopyalandı",
		APIIntro:   "Komut dosyaları ve uygulamalar için:",
		Footer:     "borless.com",
	},
	"zh": {
		Lang:       "zh",
		Title:      "我的 IP — ip.borless.com",
		Label:      "您的公网 IP 地址",
		Hint:       "点击复制",
		HintCopied: "已复制",
		APIIntro:   "用于脚本和应用程序：",
		Footer:     "borless.com",
	},
	"ja": {
		Lang:       "ja",
		Title:      "あなたのIP — ip.borless.com",
		Label:      "あなたのパブリックIPアドレス",
		Hint:       "クリックでコピー",
		HintCopied: "コピーしました",
		APIIntro:   "スクリプト・アプリ向け:",
		Footer:     "borless.com",
	},
	"ko": {
		Lang:       "ko",
		Title:      "내 IP — ip.borless.com",
		Label:      "공인 IP 주소",
		Hint:       "클릭하여 복사",
		HintCopied: "복사됨",
		APIIntro:   "스크립트 및 앱용:",
		Footer:     "borless.com",
	},
	"hi": {
		Lang:       "hi",
		Title:      "मेरा IP — ip.borless.com",
		Label:      "आपका सार्वजनिक IP पता",
		Hint:       "कॉपी करने के लिए क्लिक करें",
		HintCopied: "कॉपी हो गया",
		APIIntro:   "स्क्रिप्ट और ऐप्स के लिए:",
		Footer:     "borless.com",
	},
	"bn": {
		Lang:       "bn",
		Title:      "আমার IP — ip.borless.com",
		Label:      "আপনার সর্বজনীন IP ঠিকানা",
		Hint:       "কপি করতে ক্লিক করুন",
		HintCopied: "কপি করা হয়েছে",
		APIIntro:   "স্ক্রিপ্ট এবং অ্যাপের জন্য:",
		Footer:     "borless.com",
	},
	"id": {
		Lang:       "id",
		Title:      "IP Saya — ip.borless.com",
		Label:      "Alamat IP publik Anda",
		Hint:       "klik untuk menyalin",
		HintCopied: "disalin",
		APIIntro:   "Untuk skrip dan aplikasi:",
		Footer:     "borless.com",
	},
	"vi": {
		Lang:       "vi",
		Title:      "IP của tôi — ip.borless.com",
		Label:      "Địa chỉ IP công cộng của bạn",
		Hint:       "nhấn để sao chép",
		HintCopied: "đã sao chép",
		APIIntro:   "Dành cho tập lệnh và ứng dụng:",
		Footer:     "borless.com",
	},
	"th": {
		Lang:       "th",
		Title:      "IP ของฉัน — ip.borless.com",
		Label:      "ที่อยู่ IP สาธารณะของคุณ",
		Hint:       "คลิกเพื่อคัดลอก",
		HintCopied: "คัดลอกแล้ว",
		APIIntro:   "สำหรับสคริปต์และแอป:",
		Footer:     "borless.com",
	},
	"pl": {
		Lang:       "pl",
		Title:      "Moje IP — ip.borless.com",
		Label:      "Twój publiczny adres IP",
		Hint:       "kliknij, aby skopiować",
		HintCopied: "skopiowano",
		APIIntro:   "Dla skryptów i aplikacji:",
		Footer:     "borless.com",
	},
	"nl": {
		Lang:       "nl",
		Title:      "Mijn IP — ip.borless.com",
		Label:      "Uw openbare IP-adres",
		Hint:       "klik om te kopiëren",
		HintCopied: "gekopieerd",
		APIIntro:   "Voor scripts en apps:",
		Footer:     "borless.com",
	},
	"sv": {
		Lang:       "sv",
		Title:      "Min IP — ip.borless.com",
		Label:      "Din offentliga IP-adress",
		Hint:       "klicka för att kopiera",
		HintCopied: "kopierat",
		APIIntro:   "För skript och appar:",
		Footer:     "borless.com",
	},
	"da": {
		Lang:       "da",
		Title:      "Min IP — ip.borless.com",
		Label:      "Din offentlige IP-adresse",
		Hint:       "klik for at kopiere",
		HintCopied: "kopieret",
		APIIntro:   "Til scripts og apps:",
		Footer:     "borless.com",
	},
	"no": {
		Lang:       "no",
		Title:      "Min IP — ip.borless.com",
		Label:      "Din offentlige IP-adresse",
		Hint:       "klikk for å kopiere",
		HintCopied: "kopiert",
		APIIntro:   "For skript og apper:",
		Footer:     "borless.com",
	},
	"fi": {
		Lang:       "fi",
		Title:      "IP-osoitteeni — ip.borless.com",
		Label:      "Julkinen IP-osoitteesi",
		Hint:       "napsauta kopioidaksesi",
		HintCopied: "kopioitu",
		APIIntro:   "Komentosarjoille ja sovelluksille:",
		Footer:     "borless.com",
	},
	"cs": {
		Lang:       "cs",
		Title:      "Moje IP — ip.borless.com",
		Label:      "Vaše veřejná IP adresa",
		Hint:       "klikněte pro zkopírování",
		HintCopied: "zkopírováno",
		APIIntro:   "Pro skripty a aplikace:",
		Footer:     "borless.com",
	},
	"sk": {
		Lang:       "sk",
		Title:      "Moja IP — ip.borless.com",
		Label:      "Vaša verejná IP adresa",
		Hint:       "kliknite pre skopírovanie",
		HintCopied: "skopírované",
		APIIntro:   "Pre skripty a aplikácie:",
		Footer:     "borless.com",
	},
	"ro": {
		Lang:       "ro",
		Title:      "IP-ul meu — ip.borless.com",
		Label:      "Adresa ta IP publică",
		Hint:       "fă clic pentru a copia",
		HintCopied: "copiat",
		APIIntro:   "Pentru scripturi și aplicații:",
		Footer:     "borless.com",
	},
	"hu": {
		Lang:       "hu",
		Title:      "Az IP-címem — ip.borless.com",
		Label:      "Az Ön nyilvános IP-címe",
		Hint:       "kattintson a másoláshoz",
		HintCopied: "másolva",
		APIIntro:   "Szkriptekhez és alkalmazásokhoz:",
		Footer:     "borless.com",
	},
	"el": {
		Lang:       "el",
		Title:      "Η IP μου — ip.borless.com",
		Label:      "Η δημόσια διεύθυνση IP σας",
		Hint:       "κάντε κλικ για αντιγραφή",
		HintCopied: "αντιγράφηκε",
		APIIntro:   "Για scripts και εφαρμογές:",
		Footer:     "borless.com",
	},
	"he": {
		Lang:       "he",
		Dir:        "rtl",
		Title:      "כתובת ה-IP שלי — ip.borless.com",
		Label:      "כתובת ה-IP הציבורית שלך",
		Hint:       "לחץ להעתקה",
		HintCopied: "הועתק",
		APIIntro:   "עבור סקריפטים ויישומים:",
		Footer:     "borless.com",
	},
	"ar": {
		Lang:       "ar",
		Dir:        "rtl",
		Title:      "عنوان IP الخاص بي — ip.borless.com",
		Label:      "عنوان IP العام الخاص بك",
		Hint:       "انقر للنسخ",
		HintCopied: "تم النسخ",
		APIIntro:   "للنصوص البرمجية والتطبيقات:",
		Footer:     "borless.com",
	},
	"fa": {
		Lang:       "fa",
		Dir:        "rtl",
		Title:      "آی‌پی من — ip.borless.com",
		Label:      "آدرس IP عمومی شما",
		Hint:       "برای کپی کلیک کنید",
		HintCopied: "کپی شد",
		APIIntro:   "برای اسکریپت‌ها و برنامه‌ها:",
		Footer:     "borless.com",
	},
}

// defaultLocale is the safe fallback used whenever language detection cannot
// produce a supported match. English is the lowest-friction option for an
// international audience, so we fall back to it unconditionally:
//   - Accept-Language header missing or empty
//   - Accept-Language has only unsupported / malformed tags
//   - explicit ?lang= override points at an unknown language
var defaultLocale = locales["en"]

// pickLocale parses Accept-Language and returns the best supported match.
// We take the highest-priority tag first; that's good enough for our use
// (browsers list the user's preferred language first by default).
// On any failure path it returns defaultLocale (English).
func pickLocale(acceptLang string) locale {
	if acceptLang == "" {
		return defaultLocale
	}
	for _, part := range strings.Split(acceptLang, ",") {
		tag := strings.TrimSpace(part)
		if tag == "" {
			continue
		}
		if i := strings.Index(tag, ";"); i >= 0 {
			tag = tag[:i]
		}
		primary := strings.ToLower(strings.SplitN(tag, "-", 2)[0])
		if loc, ok := locales[primary]; ok {
			return loc
		}
	}
	return defaultLocale
}

const htmlPageTpl = `<!DOCTYPE html>
<html lang="{{.L.Lang}}"{{if .L.Dir}} dir="{{.L.Dir}}"{{end}}>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>{{.L.Title}}</title>
<style>
  :root { color-scheme: light dark; }
  *,*::before,*::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    min-height: 100vh;
    display: flex; flex-direction: column; justify-content: center; align-items: center;
    background: radial-gradient(circle at 30% 20%, #1e293b 0%, #0f172a 70%);
    color: #e2e8f0;
    padding: 1.5rem;
    text-align: center;
  }
  .label {
    font-size: 0.875rem;
    letter-spacing: 0.15em;
    text-transform: uppercase;
    opacity: 0.55;
    margin-bottom: 1.5rem;
  }
  .ip {
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    font-size: clamp(1.5rem, 8vw, 4.5rem);
    font-weight: 600;
    letter-spacing: -0.02em;
    user-select: all;
    background: rgba(148, 163, 184, 0.08);
    padding: 1.25rem 2.25rem;
    border-radius: 1rem;
    border: 1px solid rgba(148, 163, 184, 0.15);
    cursor: pointer;
    transition: background 0.15s, transform 0.1s;
    white-space: nowrap;
    overflow: hidden;
    max-width: 100%;
  }
  @media (max-width: 480px) {
    .ip { padding: 1rem 1.25rem; }
  }
  .ip:hover { background: rgba(148, 163, 184, 0.13); }
  .ip:active { transform: scale(0.98); }
  .hint {
    margin-top: 1rem;
    font-size: 0.875rem;
    opacity: 0.5;
    height: 1.25rem;
    transition: opacity 0.2s;
  }
  .ptr {
    margin-top: 1.5rem;
    font-size: 0.85rem;
    opacity: 0.65;
    max-width: 100%;
    word-break: break-word;
    line-height: 1.5;
  }
  .ptr-label {
    text-transform: uppercase;
    letter-spacing: 0.12em;
    font-size: 0.7rem;
    opacity: 0.7;
    margin-right: 0.5rem;
  }
  .ptr-value {
    font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace;
    user-select: all;
  }
  .api {
    margin-top: 3rem;
    font-size: 0.85rem;
    opacity: 0.55;
    max-width: 36rem;
    line-height: 1.6;
  }
  .endpoints {
    list-style: none;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.5rem;
    padding: 0;
    margin: 0.6rem 0 0 0;
  }
  .endpoints code {
    display: inline-block;
    background: rgba(148, 163, 184, 0.12);
    padding: 3px 10px;
    border-radius: 5px;
    font-family: ui-monospace, monospace;
    font-size: 0.8rem;
    white-space: nowrap;
  }
  footer {
    position: fixed;
    bottom: 1rem;
    font-size: 0.75rem;
    opacity: 0.3;
  }
</style>
</head>
<body>
  <div class="label">{{.L.Label}}</div>
  <div class="ip" id="ip">{{.IP}}</div>
  <div class="hint" id="hint" data-default="{{.L.Hint}}" data-copied="{{.L.HintCopied}}">{{.L.Hint}}</div>
  {{if .PTR}}<div class="ptr"><span class="ptr-label">{{.PTRLabel}}</span><span class="ptr-value">{{.PTR}}</span></div>{{end}}
  <div class="api">
    {{.L.APIIntro}}
    <ul class="endpoints">
      <li><code>curl https://ip.borless.com</code></li>
      <li><code>?format=txt</code></li>
      <li><code>?format=json</code></li>
      <li><code>?format=jsonp&amp;callback=cb</code></li>
      <li><code>?format=ptr</code></li>
    </ul>
  </div>
  <footer>{{.L.Footer}}</footer>
<script>
  (function () {
    var el = document.getElementById('ip');
    var hint = document.getElementById('hint');
    var defaultText = hint.dataset.default;
    var copiedText = hint.dataset.copied;

    // Shrink font-size until the IP fits the chip on a single line.
    // Triggered after CSS clamp has applied; covers tiny screens and IPv6.
    function fit() {
      el.style.fontSize = '';
      var size = parseFloat(getComputedStyle(el).fontSize);
      var guard = 200;
      while (el.scrollWidth > el.clientWidth && size > 12 && guard-- > 0) {
        size -= 1;
        el.style.fontSize = size + 'px';
      }
    }
    fit();
    addEventListener('resize', fit);

    el.addEventListener('click', function () {
      var text = el.textContent.trim();
      var done = function () {
        hint.textContent = copiedText;
        hint.style.opacity = '0.9';
        setTimeout(function () {
          hint.textContent = defaultText;
          hint.style.opacity = '';
        }, 1500);
      };
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(done, function () {});
      } else {
        var ta = document.createElement('textarea');
        ta.value = text; document.body.appendChild(ta);
        ta.select(); try { document.execCommand('copy'); done(); } catch (e) {}
        document.body.removeChild(ta);
      }
    });
  })();
</script>
</body>
</html>`

var htmlTpl = template.Must(template.New("ip").Parse(htmlPageTpl))

// GetIP returns a user's public facing IP address (IPv4 OR IPv6).
//
// Formats:
//   - default (no `format` param): plain text — preserves curl/script behaviour.
//   - `format=txt`                : plain text (explicit; same as default)
//   - `format=json`               : application/json
//   - `format=jsonp&callback=cb`  : application/javascript
//   - `format=ptr`                : reverse-DNS hostname (empty body if none)
//   - `format=html`               : human-friendly localized HTML page
//
// Browser autodetection: when no `format` is given but the request advertises
// `Accept: text/html`, we serve the HTML page. The page language is picked
// from `Accept-Language`; falls back to English.
func GetIP(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	if err := r.ParseForm(); err != nil {
		panic(err)
	}

	// X-Forwarded-For leftmost = origin client (cloudflared / reverse proxy populates it).
	ip := net.ParseIP(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]).String()

	format := ""
	if vals, ok := r.Form["format"]; ok && len(vals) > 0 {
		format = vals[0]
	}
	if format == "" && strings.Contains(r.Header.Get("Accept"), "text/html") {
		format = "html"
	}

	switch format {
	case "txt", "text", "plain":
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, ip)
	case "json":
		jsonStr, _ := json.Marshal(models.IPAddress{IP: ip})
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonStr))
	case "jsonp":
		jsonStr, _ := json.Marshal(models.IPAddress{IP: ip})
		callback := "callback"
		if val, ok := r.Form["callback"]; ok && len(val) > 0 {
			callback = val[0]
		}
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprint(w, callback+"("+string(jsonStr)+");")
	case "ptr":
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, lookupPTR(ip))
	case "html":
		loc := pickLocale(r.Header.Get("Accept-Language"))
		// Allow explicit override via ?lang=xx.
		if vals, ok := r.Form["lang"]; ok && len(vals) > 0 {
			if forced, ok := locales[strings.ToLower(vals[0])]; ok {
				loc = forced
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Vary", "Accept-Language, Accept")
		_ = htmlTpl.Execute(w, struct {
			IP       string
			PTR      string
			PTRLabel string
			L        locale
		}{IP: ip, PTR: lookupPTR(ip), PTRLabel: ptrLabel(loc.Lang), L: loc})
	default:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, ip)
	}
}

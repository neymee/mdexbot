package lang

var countryFlags = map[string]string{
	"any": "🌍",
	"aa": "🇩🇯", "af": "🇿🇦", "ak": "🇬🇭", "am": "🇪🇹", "az": "🇦🇿", "be": "🇧🇾", "bg": "🇧🇬", "bi": "🇻🇺",
	"bm": "🇲🇱", "bn": "🇧🇩", "bs": "🇧🇦", "ca": "🇦🇩", "cs": "🇨🇿", "cy": "🇬🇧", "da": "🇩🇰", "de": "🇩🇪",
	"dv": "🇲🇻", "dz": "🇧🇹", "el": "🇬🇷", "en": "🇬🇧", "es": "🇪🇸", "et": "🇪🇪", "fa": "🇮🇷", "fi": "🇫🇮",
	"fil": "🇵🇭", "fj": "🇫🇯", "fr": "🇫🇷", "ga": "🇮🇪", "gaa": "🇬🇭", "gu": "🇮🇳", "he": "🇮🇱", "hi": "🇮🇳",
	"ho": "🇵🇬", "hr": "🇭🇷", "ht": "🇭🇹", "hu": "🇭🇺", "hy": "🇦🇲", "id": "🇮🇩", "ig": "🇳🇬", "is": "🇮🇸",
	"it": "🇮🇹", "ja": "🇯🇵", "ka": "🇬🇪", "kg": "🇨🇬", "kj": "🇦🇴", "kk": "🇰🇿", "km": "🇰🇭", "kmb": "🇦🇴",
	"ko": "🇰🇷", "kr": "🇳🇪", "ku": "🇮🇶", "ky": "🇰🇬", "la": "🇻🇦", "lb": "🇱🇺", "ln": "🇨🇬", "lo": "🇱🇦",
	"lt": "🇱🇹", "lu": "🇨🇩", "lv": "🇱🇻", "mg": "🇲🇬", "mh": "🇲🇭", "mi": "🇳🇿", "mk": "🇲🇰", "mn": "🇲🇳",
	"mos": "🇧🇫", "ms": "🇲🇾", "mt": "🇲🇹", "my": "🇲🇲", "nb": "🇳🇴", "nd": "🇿🇼", "ne": "🇳🇵", "nl": "🇳🇱",
	"nn": "🇳🇴", "no": "🇳🇴", "nr": "🇿🇦", "nso": "🇿🇦", "ny": "🇲🇼", "pa": "🇮🇳", "pap": "🇦🇼", "pl": "🇵🇱",
	"ps": "🇦🇫", "pt": "🇵🇹", "rm": "🇨🇭", "rn": "🇧🇮", "ro": "🇷🇴", "ru": "🇷🇺", "rw": "🇷🇼", "sg": "🇨🇫",
	"si": "🇱🇰", "sk": "🇸🇰", "sl": "🇸🇮", "sn": "🇿🇼", "snk": "🇸🇳", "so": "🇸🇴", "sq": "🇦🇱", "sr": "🇷🇸",
	"srr": "🇸🇳", "ss": "🇸🇿", "st": "🇱🇸", "sv": "🇸🇪", "ta": "🇱🇰", "te": "🇮🇳", "tet": "🇹🇱", "tg": "🇹🇯",
	"th": "🇹🇭", "ti": "🇪🇷", "tk": "🇹🇲", "tl": "🇵🇭", "tn": "🇧🇼", "tpi": "🇵🇬", "tr": "🇹🇷", "ts": "🇿🇦",
	"uk": "🇺🇦", "umb": "🇦🇴", "ur": "🇵🇰", "uz": "🇺🇿", "ve": "🇿🇦", "vi": "🇻🇳", "wo": "🇸🇳", "xh": "🇿🇦",
	"zh": "🇨🇳", "zu": "🇿🇦",
}

func GetFlag(lang string) (string, bool) {
	flag, ok := countryFlags[lang]
	if ok {
		return flag, true
	}

	langShort := lang[:2]
	flag, ok = countryFlags[langShort]
	if ok {
		return flag, true
	}

	return "", false
}

func GetFlagOrLang(lang string) string {
	flag, ok := GetFlag(lang)
	if ok {
		return flag
	}
	return lang
}

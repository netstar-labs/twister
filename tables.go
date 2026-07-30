package twister

// keyboard maps each key to the keys physically adjacent on a QWERTY layout,
// precomputed as rune slices once at init so a lookup allocates nothing. It feeds
// the insertion and replacement fuzzers — a "keyboard slip" is a press of a
// neighbouring key. Embedded, not a dependency.
var keyboard = buildKeyboard()

func buildKeyboard() map[rune][]rune {
	rows := map[rune]string{
		'1': "2q", '2': "3wq1", '3': "4ew2", '4': "5re3", '5': "6tr4",
		'6': "7yt5", '7': "8uy6", '8': "9iu7", '9': "0oi8", '0': "po9",
		'q': "12wa", 'w': "3esaq2", 'e': "4rdsw3", 'r': "5tfde4", 't': "6ygfr5",
		'y': "7uhgt6", 'u': "8ijhy7", 'i': "9okju8", 'o': "0plki9", 'p': "lo0",
		'a': "qwsz", 's': "edxzaw", 'd': "rfcxse", 'f': "tgvcdr", 'g': "yhbvft",
		'h': "ujnbgy", 'j': "ikmnhu", 'k': "olmji", 'l': "kop",
		'z': "asx", 'x': "zsdc", 'c': "xdfv", 'v': "cfgb", 'b': "vghn",
		'n': "bhjm", 'm': "njk",
	}
	m := make(map[rune][]rune, len(rows))
	for k, v := range rows {
		m[k] = []rune(v)
	}
	return m
}

// keyboardAdjacent returns the keys physically adjacent to c on a QWERTY layout.
// The returned slice is shared and read-only — callers must not mutate it.
func keyboardAdjacent(c rune) []rune {
	return keyboard[c]
}

// homoglyphs maps an ASCII rune to a curated subset of its single-rune Unicode
// confusables (a UTS-39 subset): look-alikes that render nearly identically in a
// browser. Multi-rune confusables (e.g. "rn" for m) are excluded so every
// homoglyph variant is exactly one rune substitution. Embedded, not a dependency.
var homoglyphs = map[rune][]rune{
	'a': {'à', 'á', 'â', 'ã', 'ä', 'å', 'ą', 'а' /*Cyrillic*/, 'α' /*Greek*/},
	'b': {'d', 'ь', 'б'},
	'c': {'ç', 'ć', 'č', 'с' /*Cyrillic*/, 'ϲ' /*Greek*/},
	'd': {'ď', 'đ', 'ԁ'},
	'e': {'é', 'è', 'ê', 'ë', 'ē', 'ę', 'е' /*Cyrillic*/},
	'g': {'q', 'ğ', 'ġ', 'ǵ', 'ԍ'},
	'h': {'һ' /*Cyrillic*/, 'հ' /*Armenian*/},
	'i': {'1', 'l', 'í', 'ì', 'î', 'ï', 'ı', 'і' /*Cyrillic*/},
	'j': {'ј' /*Cyrillic*/},
	'k': {'κ' /*Greek*/, 'к' /*Cyrillic*/},
	'l': {'1', 'i', 'ł', 'ĺ', 'ľ'},
	'm': {'n', 'м' /*Cyrillic*/},
	'n': {'ń', 'ñ', 'ո' /*Armenian*/},
	'o': {'0', 'ο' /*Greek*/, 'о' /*Cyrillic*/, 'ø', 'ó', 'ò', 'ô', 'õ', 'ö'},
	'p': {'ρ' /*Greek*/, 'р' /*Cyrillic*/},
	'q': {'g', 'ԛ'},
	's': {'ś', 'š', 'ѕ' /*Cyrillic*/, '5'},
	't': {'τ' /*Greek*/, 'ţ', 'ť'},
	'u': {'υ' /*Greek*/, 'ú', 'ù', 'û', 'ü'},
	'v': {'ν' /*Greek*/, 'ѵ' /*Cyrillic*/},
	'w': {'ѡ' /*Cyrillic*/, 'ԝ', 'ա' /*Armenian*/},
	'x': {'х' /*Cyrillic*/, 'χ' /*Greek*/},
	'y': {'ý', 'ÿ', 'у' /*Cyrillic*/, 'ү'},
	'z': {'ź', 'ż', 'ž'},
	'0': {'o', 'о' /*Cyrillic*/, 'ο' /*Greek*/},
	'1': {'l', 'i'},
}

// leet maps an ASCII letter to its leetspeak numeral — the "reads as" look-alike
// class (paypal → p4ypal), distinct from homoglyph's "renders identically". One
// substitution per position keeps every leet variant a single edit from the seed, so
// snare detects it at distance 1. The digits that are also visual confusables (o→0,
// i/l→1, s→5) appear here for a complete leet alphabet; dedup tags those with the
// earlier fuzzer (homoglyph). Embedded, not a dependency.
var leet = map[rune][]rune{
	'a': {'4'}, 'b': {'8'}, 'e': {'3'}, 'g': {'9'}, 'i': {'1'},
	'l': {'1'}, 'o': {'0'}, 's': {'5'}, 't': {'7'}, 'z': {'2'},
}

// multiHomoglyphs maps a multi-rune sequence to the single-glyph look-alikes it
// renders as (and one reverse). Each rewrite is two Damerau-Levenshtein edits — a
// substitution plus an insertion or deletion — so these are deliberately excluded
// from the single-rune [homoglyphs] core and drive the opt-in multi-homoglyph
// fuzzer instead. Embedded, not a dependency. See fuzzMultiHomoglyph.
var multiHomoglyphs = map[string][]string{
	"rn": {"m"},  // "corn" → "com"
	"vv": {"w"},  // "vvow" → "wow"
	"cl": {"d"},  // "clock" → "dock"
	"nn": {"m"},  // "inn" → "im"
	"m":  {"rn"}, // reverse of rn→m: "com" → "corn"
}

// homophones maps a sound-alike substring to its phonetic equivalents — the
// embedded default table for the opt-in homophone fuzzer, overridable per key via
// [Options.Homophones]. A rewrite's edit count is the Levenshtein distance between
// the key and its replacement (ph↔f is two edits, c↔k is one), tagged truthfully on
// each variant. The class is defined by phonetic equality, not edit distance, so it
// routes to a phonetic detector rather than snare@1. Embedded, not a dependency.
var homophones = map[string][]string{
	"ph": {"f"},       // "phone" → "fone"
	"f":  {"ph"},      // "fon" → "phon"
	"c":  {"k"},       // "cat" → "kat"
	"k":  {"c", "ck"}, // "kat" → "cat"; "bak" → "back"
	"s":  {"z"},       // "sap" → "zap"
	"z":  {"s"},       // "zap" → "sap"
	"ck": {"k"},       // "back" → "bak"
}

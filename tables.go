package twister

// keyboard is a QWERTY physical-adjacency map: each key to the keys next to it.
// It feeds the insertion and replacement fuzzers — a "keyboard slip" is a press of
// a neighbouring key. Embedded, not a dependency.
var keyboard = map[rune]string{
	'1': "2q", '2': "3wq1", '3': "4ew2", '4': "5re3", '5': "6tr4",
	'6': "7yt5", '7': "8uy6", '8': "9iu7", '9': "0oi8", '0': "po9",
	'q': "12wa", 'w': "3esaq2", 'e': "4rdsw3", 'r': "5tfde4", 't': "6ygfr5",
	'y': "7uhgt6", 'u': "8ijhy7", 'i': "9okju8", 'o': "0plki9", 'p': "lo0",
	'a': "qwsz", 's': "edxzaw", 'd': "rfcxse", 'f': "tgvcdr", 'g': "yhbvft",
	'h': "ujnbgy", 'j': "ikmnhu", 'k': "olmji", 'l': "kop",
	'z': "asx", 'x': "zsdc", 'c': "xdfv", 'v': "cfgb", 'b': "vghn",
	'n': "bhjm", 'm': "njk",
}

// keyboardAdjacent returns the keys physically adjacent to c on a QWERTY layout.
func keyboardAdjacent(c rune) []rune {
	return []rune(keyboard[c])
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

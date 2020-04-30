package translate

import "testing"

func TestExtractTkk(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{"tkk: '32324.1323223'", "32324.1323223"},
		{"tkk:'441153.2601734278'", "441153.2601734278"},
		{"tkk:'441153.2601734278',exp", "441153.2601734278"},
		{"1h\",tkk:'441153.2601734278',exp", "441153.2601734278"},
	}

	for _, c := range cases {
		tkk, err := extractTkk([]byte(c.input))
		if err != nil {
			t.Errorf("Extract tkk error: %v", err)
			return
		}
		if tkk != c.expect {
			t.Errorf("Extract tkk failed, expect: %s, got: %s", c.expect, tkk)
		}
	}
}

func TestGetTkk(t *testing.T) {
	tkk, err := getTkk()
	if err != nil {
		t.Errorf("Get tkk error: %v", err)
		return
	}
	if tkk == "" {
		t.Errorf("Get tkk failed")
	}
	t.Log(tkk)
}

func TestGenerateTk(t *testing.T) {

	cases := []struct {
		tkk    string
		text   string
		expect string
	}{
		{tkk: "441156.1924457848", text: "这里是北京市中心", expect: "248966.358338"},
	}

	for _, c := range cases {
		val, err := generateTk(c.tkk, c.text)
		if err != nil {
			t.Errorf("Generate tk error: %v", err)
			return
		}
		if val != c.expect {
			t.Errorf("Generate tk failed, expect: %s, got: %s", c.expect, val)
		}
	}
}

func TestTkSum(t *testing.T) {
	cases := []struct {
		val    uint32
		table  []byte
		expect uint32
	}{
		{33, []byte("+-a^+6"), 34353},
	}

	for _, c := range cases {
		v := tkSum(c.val, c.table)
		if v != c.expect {
			t.Errorf("Calculate hash failed, expect: %d, got: %d", c.expect, v)
		}
	}
}

func TestTkTransfrom(t *testing.T) {
	cases := []struct {
		input  string
		expect []uint32
	}{
		{"你𐀀好߿c😍8🐨", []uint32{228, 189, 160, 240, 128, 128, 229, 165, 189, 223, 99, 255, 152, 141, 56, 255, 144, 168}},
		{"我能吞下玻璃而不伤身体。", []uint32{230, 136, 145, 232, 131, 189, 229, 144, 158, 228, 184, 139, 231, 142, 187, 231, 146, 131, 232, 128, 140, 228, 184, 141, 228, 188, 164, 232, 186, 171, 228, 189, 147, 227, 128, 130}},
		{"나는 유리를 먹을 수 있어요. 그래도 아프지 않아요", []uint32{235, 130, 152, 235, 138, 148, 32, 236, 156, 160, 235, 166, 172, 235, 165, 188, 32, 235, 168, 185, 236, 157, 132, 32, 236, 136, 152, 32, 236, 158, 136, 236, 150, 180, 236, 154, 148, 46, 32, 234, 183, 184, 235, 158, 152, 235, 143, 132, 32, 236, 149, 132, 237, 148, 132, 236, 167, 128, 32, 236, 149, 138, 236, 149, 132, 236, 154, 148}},
		{"Μπορώ να φάω σπασμένα γυαλιά χωρίς να πάθω τίποτα.", []uint32{206, 207, 206, 207, 207, 32, 206, 206, 32, 207, 206, 207, 32, 207, 207, 206, 207, 206, 206, 206, 206, 32, 206, 207, 206, 206, 206, 206, 32, 207, 207, 207, 206, 207, 32, 206, 206, 32, 207, 206, 206, 207, 32, 207, 206, 207, 206, 207, 206, 46}},
		{"Je peux manger du verre, ça ne me fait pas mal.", []uint32{74, 101, 32, 112, 101, 117, 120, 32, 109, 97, 110, 103, 101, 114, 32, 100, 117, 32, 118, 101, 114, 114, 101, 44, 32, 195, 97, 32, 110, 101, 32, 109, 101, 32, 102, 97, 105, 116, 32, 112, 97, 115, 32, 109, 97, 108, 46}},
	}

	for _, c := range cases {
		val, err := tkTransform(c.input)
		if err != nil {
			t.Errorf("Transform text for tk error: %v", err)
			return
		}
		if len(val) != len(c.expect) {
			t.Errorf("Transform text for tk failed, expect: %v, got: %v", c.expect, val)
			continue
		}
		for i, v := range val {
			if v != c.expect[i] {
				t.Errorf("Transform text for tk failed, expect: %v, got: %v", c.expect, val)
				break
			}
		}
	}
}

func TestGoogleTranslate(t *testing.T) {
	cases := []struct {
		sLang  string
		tLang  string
		input  string
		expect string
	}{
		{"en", "zh-CN", "your good", "你的好"},
		{"zh-CN", "en", "Google翻译是结合了自然语言处理与人工智能，所以翻译结果相对令人满意，不会出现太多的生硬的尴尬的翻译。", "Google Translate is a combination of natural language processing and artificial intelligence, so the translation results are relatively satisfactory, and there will not be too many awkward and awkward translations."},
	}

	for _, c := range cases {
		val, err := Google(c.input, c.sLang, c.tLang)
		if err != nil {
			t.Errorf("Google translate error: %v", err)
			return
		}
		if val != c.expect {
			t.Errorf("Google translate failed, expect: %s, got: %s", c.expect, val)
		}
	}
}

package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	gTransUrl  = "https://translate.google.com"
	gTransPath = "/translate_a/single"
	urlEnvName = "GOOGLE_TRANSLATE_URL"

	maxTkkDuration = 10 * time.Hour
)

var (
	tkkReg = regexp.MustCompile(`tkk:\W*'(\d+\.\d+)'`)

	gTkTable1 = []byte("+-a^+6")
	gTkTable2 = []byte("+-3^+b+-f")

	gTkk *gTransTkk
)

func extractTkk(content []byte) (string, error) {
	matches := tkkReg.FindSubmatch(content)
	if len(matches) < 2 {
		return "", fmt.Errorf("Cannot find tkk")
	}
	tkkMatch := matches[1]
	return string(tkkMatch), nil
}

type gTransTkk struct {
	Value string
	Time  time.Time
}

func gtTransUrl() string {
	transUrl := os.Getenv(urlEnvName)
	if transUrl != "" {
		if !strings.HasPrefix(transUrl, "https") {
			transUrl = fmt.Sprintf("https://%s", transUrl)
		}
	}
	_, err := url.Parse(transUrl)
	if err != nil {
		transUrl = ""
	}
	if transUrl == "" {
		transUrl = gTransUrl
	}
	return transUrl
}

// getTkk returns a tkk from google translate page
func getTkk() (string, error) {
	if gTkk != nil {
		if time.Since(gTkk.Time) < maxTkkDuration {
			return gTkk.Value, nil
		}
	}
	transUrl := gtTransUrl()
	res, err := http.Get(transUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	tkk, err := extractTkk(bts)
	if err != nil {
		return "", err
	}
	// set global tkk
	gTkk = &gTransTkk{
		Value: tkk,
		Time:  time.Now(),
	}
	return tkk, nil
}

func tkTransform(text string) ([]uint32, error) {
	codes := []rune(text)

	bts := make([]uint32, 0)
	for _, code := range codes {
		// ASCII code, 1 byte
		// 0x000000 - 0x00007F
		if code <= 0x7F {
			bts = append(bts, uint32(code))
			continue
		}
		// 2 bytes
		// 0x000080 - 0x0007FF
		if code <= 0x7FF {
			bts = append(bts, uint32(code)>>6|192)
			continue
		}

		// 3 bytes
		// 0x000800 - 0x00D7FF
		// 0x00E000 - 0x00FFFF
		if (code >= 0x800 && code <= 0xD7FF) || (code >= 0xE000 && code <= 0xFFFF) {
			cs := []byte(string(code))
			for _, c := range cs {
				bts = append(bts, uint32(c))
			}
			continue
		}

		// 4 bytes
		// 0x010000 - 0x10FFFF
		if code >= 0x10000 && code <= 0x10FFFF {
			bts = append(bts, uint32(code)>>12|224)
			bts = append(bts, uint32(code)>>6&63|128)
			bts = append(bts, uint32(code)&63|128)
			continue
		}
		return nil, fmt.Errorf("Invalid code: %d in %s", code, text)
	}

	return bts, nil
}

func generateTk(tkk, text string) (string, error) {
	bts, err := tkTransform(text)
	if err != nil {
		return "", err
	}

	tkkParts := strings.Split(tkk, ".")
	if len(tkkParts) != 2 {
		return "", fmt.Errorf("Invalid tkk: %s", tkk)
	}
	tl, err := strconv.Atoi(tkkParts[0])
	if err != nil {
		return "", err
	}
	tr, err := strconv.Atoi(tkkParts[1])
	if err != nil {
		return "", err
	}

	left := uint32(tl)
	for _, b := range bts {
		left += b
		left = tkSum(left, gTkTable1)
	}

	left = tkSum(left, gTkTable2)
	left ^= uint32(tr)
	left %= 1000000

	right := left ^ uint32(tl)
	sig := fmt.Sprintf("%d.%d", left, right)

	return sig, nil
}

// tkSum extract from google translate website
func tkSum(n uint32, factor []byte) uint32 {
	for i := 0; i < len(factor)-2; i += 3 {
		char := factor[i+2]

		d := uint32(char)
		if char >= 'a' {
			// convert char 'a' to number 10
			d = d - 87
		} else {
			// convert string number to real number
			d = d - 48
		}

		if factor[i+1] == '+' {
			d = n >> d
		} else {
			d = n << d
		}

		if factor[i] == '+' {
			n = n + d&0xFFFFFFFF
		} else {
			n = n ^ d
		}
	}
	return n
}

func getToken(text string) (string, error) {
	tkk, err := getTkk()
	if err != nil {
		return "", fmt.Errorf("Get tkk error: %v", err)
	}
	token, err := generateTk(tkk, text)
	if err != nil {
		return "", fmt.Errorf("Generate tk error: %v", err)
	}
	return token, nil
}

func Google(ctx context.Context, text, sl, tl string) (string, error) {
	transUrl := gtTransUrl()
	up, err := url.Parse(transUrl)
	if err != nil {
		return "", err
	}
	up.Path = gTransPath

	params := make(url.Values)
	params.Set("sl", sl)
	params.Set("tl", tl)
	params.Set("dt", "t")
	params.Set("client", "gtx")
	params.Set("q", text)
	up.RawQuery = params.Encode()

	addr := up.String()
	res, err := http.Get(addr)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	resp := make([]interface{}, 0)
	if err = json.Unmarshal(bts, &resp); err != nil {
		return "", err
	}
	if len(resp) < 1 {
		return "", fmt.Errorf("Invalid response: %s", bts)
	}
	results := make([]string, 0)
	if boxes, ok := resp[0].([]interface{}); ok {
		for _, box := range boxes {
			elements, ok := box.([]interface{})
			if !ok {
				continue
			}
			if len(elements) < 1 {
				continue
			}
			if txt, ok := elements[0].(string); ok {
				results = append(results, txt)
			}
		}
	}
	result := strings.Join(results, "")
	return result, nil
}

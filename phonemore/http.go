package phonemore

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	modelsRegex = regexp.MustCompile(`\/specs\/samsung\/[a-z0-9-]+\/sm-[0-9a-z]+-[0-9]+gb\/`)
)

// Model - holds data about a scraped model, first just the original path scraped.
type Model struct {
	Path          string `json:"path"`          // done
	Model         string `json:"model"`         // done
	Manufacturer  string `json:"manufacturer"`  // samsung static
	Device        string `json:"device"`        // z3q static, edit: not sure if this is right? I think it is?
	Width         int    `json:"width"`         // done
	Height        int    `json:"height"`        // done
	GPS           bool   `json:"gps"`           // done
	Gyro          bool   `json:"gyro"`          // done
	Accelerometer bool   `json:"accelerometer"` // done
	Ethernet      bool   `json:"ethernet"`      // no clue how to do this
	TouchScreen   bool   `json:"touchScreen"`   // done
	NFC           bool   `json:"nfc"`           // done
	WiFi          bool   `json:"wifi"`          // done

	// ! this is used for the SDK_INT value which is also used for the os.version and ethernet values.
	AndroidVersion int `json:"androidVersion"`
}

// ScrapeModels - scrapes x page (starts at index 1) and returns all models per phone on screen. Each page has at most 20 phones on screen, each phone can have any amount of models though.
func ScrapeModels(page int) (map[string]*Model, error) {
	out := make(map[string]*Model)
	req, err := http.NewRequest("GET", "https://www.phonemore.com/specs/?brand=5&device=1&network=5&z="+fmt.Sprint(page), nil)
	if err != nil {
		return out, err
	}
	for _, h := range strings.Split(`Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7|Accept-Language: en-US,en;q=0.9|Cache-Control: no-cache|Pragma: no-cache|Referer: https://www.phonemore.com/specs/?brand=5&device=1|Sec-Ch-Ua: "Not.A/Brand";v="8", "Chromium";v="114", "Opera GX";v="100"|Sec-Ch-Ua-Mobile: ?0|Sec-Ch-Ua-Platform: "Windows"|Sec-Fetch-Dest: document|Sec-Fetch-Mode: navigate|Sec-Fetch-Site: same-origin|Sec-Fetch-User: ?1|Upgrade-Insecure-Requests: 1|User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 OPR/100.0.0.0`, "|") {
		p := strings.Split(h, ": ")
		req.Header.Set(p[0], p[1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	models := modelsRegex.FindAllString(string(b), -1)
	if len(models) < 1 {
		return out, errors.New("no models in result")
	}
	for _, m := range models {
		parts := strings.Split(m, "/")
		parts = strings.Split(parts[len(parts)-2], "-")
		modelParsed := strings.Join(parts[:len(parts)-1], "-")
		if _, ok := out[modelParsed]; !ok {
			out[modelParsed] = &Model{
				Path:         m,
				Model:        modelParsed,
				Manufacturer: "samsung",
				Device:       "z3q",
			}
		}
	}
	return out, nil
}

var (
	// ! space is important
	widthHeightRegex    = regexp.MustCompile(`Display resolution</td><td>[0-9]+x[0-9]+ pixels`)
	androidVersionRegex = regexp.MustCompile(`System version</td><td><a href="/systems/android/[0-9]+/">Android [0-9]+`)
)

// FillData - just fills in all the empty data in the Model struct
func (m *Model) FillData() error {
	req, err := http.NewRequest("GET", "https://www.phonemore.com"+m.Path, nil)
	if err != nil {
		return err
	}
	for _, h := range strings.Split(`Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7|Accept-Language: en-US,en;q=0.9|Cache-Control: no-cache|Pragma: no-cache|Referer: https://www.phonemore.com/specs/?brand=5&device=1|Sec-Ch-Ua: "Not.A/Brand";v="8", "Chromium";v="114", "Opera GX";v="100"|Sec-Ch-Ua-Mobile: ?0|Sec-Ch-Ua-Platform: "Windows"|Sec-Fetch-Dest: document|Sec-Fetch-Mode: navigate|Sec-Fetch-Site: same-origin|Sec-Fetch-User: ?1|Upgrade-Insecure-Requests: 1|User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 OPR/100.0.0.0`, "|") {
		p := strings.Split(h, ": ")
		req.Header.Set(p[0], p[1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := string(b)
	x := widthHeightRegex.FindString(body)
	if len(x) == 0 {
		return errors.New("unable to get res")
	}
	x = x[27:]
	parts := strings.Split(strings.Split(x, " ")[0], "x")
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("res width - %v", err)
	}
	m.Width = num
	num, err = strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("res height - %v", err)
	}
	m.Height = num
	androidVer := androidVersionRegex.FindString(body)
	if len(androidVer) == 0 {
		return errors.New("unable to get android version")
	}
	num, err = strconv.Atoi(strings.Split(androidVer[49:], " ")[1])
	if err != nil {
		return fmt.Errorf("android ver - %v", err)
	}
	m.AndroidVersion = num
	m.TouchScreen = strings.Contains(body, "Capacitive Multitouch")
	m.GPS = strings.Contains(body, "A-GPS")
	m.Gyro = strings.Contains(body, "Gyroscope")
	m.Accelerometer = strings.Contains(body, "Accelerometer")
	m.NFC = strings.Contains(body, "<tr><td>NFC</td><td><span class=item_check></span>Supported</td></tr>")
	m.WiFi = strings.Contains(body, "<tr><td>WiFi</td><td><span class=item_check>")

	return nil
}

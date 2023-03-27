package shared

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func StringToDateTimeHelper(datestring string) (time.Time, error) {
	var err error = nil
	datestring = strings.Trim(datestring, " ")

	var tdefault time.Time

	if datestring == "" {
		return tdefault, fmt.Errorf("empty date string")
	}

	layout := "01/02/2006"
	t, err := time.Parse(layout, datestring)
	if err == nil {
		return t, err
	}

	layout = "2006-01-02"
	t, err = time.Parse(layout, datestring)
	if err == nil {
		return t, err
	}

	layout = "1/2/2006"
	t, err = time.Parse(layout, datestring)
	if err == nil {
		return t, err
	}

	layout = "2006-1-2"
	t, err = time.Parse(layout, datestring)
	if err == nil {
		return t, err
	}

	layout = "01/02/06"
	t, err = time.Parse(layout, datestring)
	if err == nil {
		return t, err
	}

	layout = "1/2/06"
	t, err = time.Parse(layout, datestring)
	if err == nil {
		return t, err
	}

	ivalue, err := strconv.ParseInt(datestring, 10, 64)
	if err == nil {
		t = time.Unix(ivalue, 0)
		return t, err
	}

	return t, err
}

func ValidateQuery(md5str map[string]string) error {
	var err error = nil
	salt_secret, _ := os.LookupEnv("SL_TEAM_GENERAL")
	p_secret, _ := os.LookupEnv("P_SECRET")
	testsec := strings.TrimSpace(md5str[p_secret])

	salt_provider, _ := os.LookupEnv("SL_RURL")
	p_provider, _ := os.LookupEnv("P_PROVIDER")
	testprov := strings.TrimSpace(md5str[p_provider])

	sl_token, _ := os.LookupEnv("SL_TOKEN_GENERAL")
	sl_check, _ := os.LookupEnv("SL_CHECK")
	md5_token := strings.TrimSpace(md5str[sl_check])

	sl_sgchannel, _ := os.LookupEnv("SL_STAFF_GENERAL")
	sl_sechannel, _ := os.LookupEnv("SL_STAFF_EVENTS")
	sl_ch_check, _ := os.LookupEnv("SL_CH_CHECK")
	channel := strings.TrimSpace(md5str[sl_ch_check])

	bvalid := testsec != "" && salt_secret != "" && (testsec == salt_secret)
	if !bvalid {
		err = fmt.Errorf("secret invalid %s %s", testsec, salt_secret)
	}
	bvalid = testprov != "" && salt_provider != "" && strings.Contains(testprov, salt_provider)
	if !bvalid {
		err = fmt.Errorf("url invalid %s %s", testprov, salt_provider)
	}
	bvalid = sl_token != "" && md5_token != "" && (sl_token == md5_token)
	if !bvalid {
		err = fmt.Errorf("token invalid sl:%s md5:%s", sl_token, md5_token)
	}
	bvalid = sl_sgchannel != "" && channel != "" && sl_sechannel != "" && (sl_sgchannel == channel || sl_sechannel == channel)
	if !bvalid {
		err = fmt.Errorf("channel invalid %s %s %s", sl_sgchannel, sl_sechannel, channel)
	}
	//minor annoyance for unofficial callers
	return err
}

func UnquoteBody(str string) string {
	str = strings.Replace(strings.Replace(str, "\"", "", -1), "\\", "", -1)
	return str
}

func DataFromBodyString(strbody string) map[string]string {
	var retmap = make(map[string]string)
	split_str := strings.Split(strbody, "&")
	for _, str := range split_str {
		xstr := strings.Split(str, "=")
		retmap[xstr[0]] = strings.Replace(strings.Replace(xstr[1], "%2F", "/", -1), "%3A", ":", -2)
	}
	return retmap
}

func GetConfigValue(key string, def string) string {
	config, err := os.Open("admin-local-kv.txt")
	var value string
	if err == nil {
		defer config.Close()
		scanner := bufio.NewScanner(config)
		for scanner.Scan() {
			keyvalue := scanner.Text()
			if strings.Contains(keyvalue, key) {
				items, err := fmt.Sscanf(keyvalue, key+": %v", &value)
				value = strings.TrimSpace(value)
				if items == 0 || err != nil {
					value = def
				}
			}
		}
	} else {
		value = def
	}

	return value
}

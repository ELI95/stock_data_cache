package cache

import (
	"github.com/axgle/mahonia"
	"stock_data_cache/utils"
	"time"
)

func RequestSina(url string, timeout time.Duration) (value string, err error) {
	defer utils.TimeTrack(time.Now(), "RequestSina")

	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	headers["Referer"] = "https://finance.sina.com.cn/"

	opts := []utils.RequestOption{utils.RequestWithHeaders(headers)}
	b, err := utils.DoGetRequest(url, timeout, opts...)
	value = mahonia.NewDecoder("gbk").ConvertString(string(b))
	return
}

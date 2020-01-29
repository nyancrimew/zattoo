package zattoo

import (
	"fmt"
	"github.com/valyala/fastjson"
	"net/url"
)

type Channel struct {
	Id       string
	Title    string
	ImageUrl string
}

func (z *ZapiSession) retrieveChannels() (*fastjson.Value, error) {
	powerGuideHash := string(z.accountData.GetStringBytes("session", "power_guide_hash"))
	return z.ExecZapiCall(fmt.Sprintf("/zapi/v3/cached/%s/channels?details=false", powerGuideHash), nil)
}

func (z *ZapiSession) GetAllChannels() ([]Channel, error) {
	channelData, err := z.retrieveChannels()
	if err != nil {
		return nil, err
	}
	channelDatas := channelData.GetArray("channels")
	channels := make([]Channel, len(channelDatas))
	for i, c := range channelDatas {
		channels[i] = Channel{
			Id:       string(c.GetStringBytes("id")),
			Title:    string(c.GetStringBytes("title")),
			ImageUrl: "https://images.zattic.com" + string(c.GetStringBytes("qualities", "0", "logo_black_84")),
		}
	}
	return channels, nil
}

func (z *ZapiSession) GetWatchUrl(id string) (string, error) {
	data, err := z.ExecZapiCall(fmt.Sprintf("/zapi/watch/live/%s", id), &url.Values{
		"stream_type":      {"hls"},
		"https_watch_urls": {"True"},
	})
	if err != nil {
		return "", err
	}
	return string(data.GetStringBytes("stream", "url")), nil
}

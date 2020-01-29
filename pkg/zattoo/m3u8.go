package zattoo

import (
	"bytes"
	"fmt"
	"time"
)

func (z *ZapiSession) UpdateM3u8Cache() ([]byte, error) {
	channels, err := z.GetAllChannels()
	if err != nil {
		return nil, err
	}
	m3u8 := bytes.NewBuffer(nil)
	m3u8.WriteString("#EXTM3U\n")
	for _, c := range channels {
		time.Sleep(900 * time.Millisecond)
		// group-title="IPTV-DE"
		watchUrl, err := z.GetWatchUrl(c.Id)
		if err != nil {
			fmt.Println(err)
			time.Sleep(500 * time.Millisecond)
			watchUrl, err = z.GetWatchUrl(c.Id)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
		if watchUrl != "" {
			m3u8.WriteString(fmt.Sprintf(`#EXTINF:-1 tvg-name="%s" tvg-id="%s" tvg-logo="%s",%s`, c.Title, c.Id, c.ImageUrl, c.Title))
			m3u8.WriteRune('\n')
			m3u8.WriteString(watchUrl)
			m3u8.WriteRune('\n')
		}
	}
	bytez := m3u8.Bytes()
	go func() {
		z.cache.Set("m3u8", bytez)
	}()
	return bytez, nil
}

func (z *ZapiSession) GetM3u8() ([]byte, error){
	value, err := z.cache.Get("m3u8")
	if err != nil {
		return z.UpdateM3u8Cache()
	}
	return value, nil
}

package zattoo

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

const (
	EpgTimeLayout = "20060102150405 -0700"
)

func (z *ZapiSession) GetEpg() ([]byte, error) {
	value, err := z.cache.Get("epg")
	if err != nil {
		return z.UpdateEpgCache()
	}
	return value, nil
}

func (z *ZapiSession) UpdateEpgCache() ([]byte, error) {
	channels, err := z.GetAllChannels()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	ln := func () {
		buf.WriteRune('\n')
	}
	write := func(str string) {
		buf.WriteString(str)
	}
	writeln := func(str string) {
		write(str)
		ln()
	}

	writeln(`<?xml version="1.0" encoding="UTF-8"?>`)
	writeln(`<!DOCTYPE tv SYSTEM "xmltv.dtd">`)
	writeln(`<tv generator-info-name="Zattoo-iptv" generator-info-url="https://github.com/deletescape/zattoo">`)

	for _, channel := range channels {
		write(`  <channel id="`)
		write(channel.Id)
		writeln(`">`)
		write(`    <display-name lang="en">`)
		xml.EscapeText(buf, []byte(channel.Title))
		writeln(`</display-name>`)
		write(`    <icon src="`)
		write(channel.ImageUrl)
		writeln(`"></icon>`)
		writeln(`  </channel>`)
	}

	now := time.Now()
	start := now.Add(-6 * time.Hour)
	end := now.Add(15 * time.Hour)
	powerGuideHash := string(z.accountData.GetStringBytes("session", "power_guide_hash"))
	guideData, err := z.ExecZapiCall(fmt.Sprintf("/zapi/v3/cached/%s/guide?start=%d&end=%d", powerGuideHash, start.Unix(), end.Unix()), nil)
	if err != nil {
		return nil, err
	}

	for _, c := range channels {
		programmes := guideData.GetArray("channels", c.Id)
		for _, p := range programmes {
			start := time.Unix(p.GetInt64("s"), 0)
			end := time.Unix(p.GetInt64("e"), 0)
			write(`  <programme start="`)
			write(start.Format(EpgTimeLayout))
			write(`" stop="`)
			write(end.Format(EpgTimeLayout))
			write(`" channel="`)
			write(c.Id)
			writeln(`">`)

			write(`    <title lang="en">`)
			err := xml.EscapeText(buf, p.GetStringBytes("t"))
			if err != nil {
				return nil, err
			}
			writeln(`</title>`)

			subt := p.GetStringBytes("et")
			if subt != nil && len(subt) > 0 {
				strsubt := string(subt)
				if len(strsubt) > 35 {
					// treat as description instead
					write(`    <desc lang="en">`)
					err := xml.EscapeText(buf, subt)
					if err != nil {
						return nil, err
					}
					writeln(`</desc>`)
				} else {
					write(`    <sub-title lang="en">`)
					err := xml.EscapeText(buf, subt)
					if err != nil {
						return nil, err
					}
					writeln(`</sub-title>`)
				}
			}
			cats := p.GetArray("c")
			for _, cat := range cats {
				byts, err := cat.StringBytes()
				if err == nil {
					write(`    <category lang="en">`)
					err := xml.EscapeText(buf, byts)
					if err != nil {
						return nil, err
					}
					writeln(`</category>`)
				}
			}
			gens := p.GetArray("g")
			for _, gen := range gens {
				byts, err := gen.StringBytes()
				if err == nil {
					write(`    <category lang="en">`)
					err := xml.EscapeText(buf, byts)
					if err != nil {
						return nil, err
					}
					writeln(`</category>`)
				}
			}
			if p.GetBool("ser_e") {
				enum := p.GetInt("e_no")
				if enum != 0 {
					write(`    <episode-num system="onscreen">`)
					write(strconv.Itoa(enum))
					writeln(`</episode-num>`)
				}
			}

			iUrl := p.GetStringBytes("i_url")
			if iUrl != nil && len(iUrl) > 0 {
				write(`    <icon src="`)
				buf.Write(iUrl)
				writeln(`"></icon>`)
			}

			writeln(`  </programme>`)
		}
	}
	writeln(`</tv>`)
	byts := buf.Bytes()
	go func() {
		z.cache.Set("epg", byts)
	}()
	return byts, nil
}

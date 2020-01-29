package main

import (
	"github.com/deletescape/zattoo/pkg/zattoo"
	"github.com/robfig/cron/v3"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"os"
)

var zApi *zattoo.ZapiSession

func init() {
	var err error
	zApi, err = zattoo.NewZapiSession(os.Getenv("ZATTOO_USER"), os.Getenv("ZATTOO_PASS"))
	if err != nil {
		log.Fatalln(err)
	}
}

var m3u8ContentType = []byte("application/x-mpegURL")
var epgContentType = []byte("application/xml; charset=utf-8")

func m3u8Handler(ctx *fasthttp.RequestCtx) {
	m3u8, err := zApi.GetM3u8()
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
	ctx.Write(m3u8)
	ctx.SetContentTypeBytes(m3u8ContentType)
}

func epgHandler(ctx *fasthttp.RequestCtx) {
	epg, err := zApi.GetEpg()
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
	ctx.Write(epg)
	ctx.SetContentTypeBytes(epgContentType)
}

func router(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/m3u8":
		m3u8Handler(ctx)
		break
	case "/epg":
		epgHandler(ctx)
		break
	default:
		ctx.SetStatusCode(http.StatusNotFound)
	}
}

func main() {
	log.Println("populating m3u8 cache")
	zApi.UpdateM3u8Cache()
	cr := cron.New()
	cr.AddFunc("@every 12h", func() {
		zApi.UpdateM3u8Cache()
	})
	cr.AddFunc("@every 1h", func() {
		zApi.UpdateEpgCache()
	})
	cr.Start()
	log.Println("starting server")
	log.Fatalln(fasthttp.ListenAndServe(":8090", router))
}

package main

//
// import (
// 	"net/http"
// 	"net/url"
// 	"strconv"
// )
//
// func getXML(req string, media_id int) string {
// 	data := url.Values{}
// 	if req == "RpcApiSubtitle_GetXml" {
// 		data = url.Values{
// 			"req":                {"RpcApiSubtitle_GetXml"},
// 			"subtitle_script_id": {strconv.Itoa(media_id)},
// 		}
// 	} else if req == "RpcApiVideoPlayer_GetStandardConfig" {
// 		data = url.Values{
// 			"req":           {"RpcApiVideoPlayer_GetStandardConfig"},
// 			"media_id":      {strconv.Itoa(media_id)},
// 			"video_format":  {"108"},
// 			"video_quality": {"80"},
// 			"auto_play":     {"1"},
// 			"aff":           {"crunchyroll-website"},
// 			"pop_out_disable_message": {""},
// 			"click_through":           {"0"},
// 		}
// 	} else {
// 		data = url.Values{
// 			"req":                  {req},
// 			"media_id":             {strconv.Itoa(media_id)},
// 			"video_format":         {"108"},
// 			"video_encode_quality": {"80"},
// 		}
// 	}
//
// 	client := &http.Client{}
// 	req2, _ := http.NewRequest("GET", url, nil)
// 	req2.Header.Add("Host", "www.crunchyroll.com")
// 	req2.Header.Add("Origin", "http://static.ak.crunchyroll.com")
// 	req2.Header.Add("Content-type", "application/x-www-form-urlencoded")
// 	req2.Header.Add("Referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.fb2c7182.swf")
// 	req2.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
// 	req2.Header.Add("X-Requested-With", "ShockwaveFlash/19.0.0.245")
// 	resp, _ := client.Do(req2)
// }

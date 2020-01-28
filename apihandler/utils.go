package apihandler

import (
	"encoding/json"
	"github.com/go-redis/redis"
	pb "github.com/transavro/ScheduleService/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)


// helper structs
// for Editorial Type
type EditorialTemp struct {
	ID struct {
		Oid string `json:"$oid"`
	} `json:"_id"`
	Title        string   `json:"title"`
	Portrait     []string `json:"portrait"`
	Poster       []string `json:"poster"`
	ContentID    string   `json:"contentId"`
	IsDetailPage bool     `json:"isDetailPage"`
	PackageName  string   `json:"packageName"`
	Target       []string `json:"target"`
}


// for dynamic Row type
type Temp struct {
	ID struct {
		CreatedAt struct {
			Date struct {
				NumberLong string `json:"$numberLong"`
			} `json:"$date"`
		} `json:"created_at"`
		ReleaseDate string `json:"releaseDate"`
		Year        string `json:"year"`
	} `json:"_id"`
	ContentTile []struct {
		Title        string   `json:"title"`
		Portrait     []string `json:"portrait"`
		Poster       []string `json:"poster"`
		ContentID    string   `json:"contentId"`
		IsDetailPage bool     `json:"isDetailPage"`
		PackageName  string   `json:"packageName"`
		Target       []string `json:"target"`
		Created_at       string `json:"created_at"`
		Updated_at       string `json:"updated_at"`
	} `json:"contentTile"`
}

// Youtube PlaylistApi

type YtPlayist struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []struct {
		Kind    string `json:"kind"`
		Etag    string `json:"etag"`
		ID      string `json:"id"`
		Snippet struct {
			PublishedAt time.Time `json:"publishedAt"`
			ChannelID   string    `json:"channelId"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				Medium struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"medium"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
				Standard struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"standard"`
				Maxres struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"maxres"`
			} `json:"thumbnails"`
			ChannelTitle string `json:"channelTitle"`
			PlaylistID   string `json:"playlistId"`
			Position     int    `json:"position"`
			ResourceID   struct {
				Kind    string `json:"kind"`
				VideoID string `json:"videoId"`
			} `json:"resourceId"`
		} `json:"snippet"`
		ContentDetails struct {
			VideoID          string    `json:"videoId"`
			VideoPublishedAt time.Time `json:"videoPublishedAt"`
		} `json:"contentDetails"`
	} `json:"items"`
}


type YTChannel struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken"`
	RegionCode    string `json:"regionCode"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []struct {
		Kind string `json:"kind"`
		Etag string `json:"etag"`
		ID   struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			PublishedAt time.Time `json:"publishedAt"`
			ChannelID   string    `json:"channelId"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Thumbnails  struct {
				Default struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"default"`
				Medium struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"medium"`
				High struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"high"`
			} `json:"thumbnails"`
			ChannelTitle         string `json:"channelTitle"`
			LiveBroadcastContent string `json:"liveBroadcastContent"`
		} `json:"snippet"`
	} `json:"items"`
}



// helper function
func formatString(value string) string {
	return strings.ToLower(strings.Replace(value, " ", "_", -1))
}

func pipelineMaker(rowValues *pb.ScheduleRow) mongo.Pipeline {
	// creating pipes for mongo aggregation
	pipeline := mongo.Pipeline{}

	//pipeline = append(pipeline , bson.D{{"$match", bson.D{{"content.publishState", true}}}})

	if rowValues.GetRowType() == pb.RowType_Editorial {
		// Adding stages 1
		pipeline = append(pipeline, bson.D{{"$match", bson.D{{"ref_id", bson.D{{"$in", rowValues.GetRowTileIds()}}}}}})

		//Adding stage 2
		pipeline = append(pipeline, bson.D{{"$project", bson.D{
			{"title", "$metadata.title"},
			{"portrait", "$posters.portrait",},
			{"poster", "$posters.landscape"},
			{"contentId", "$ref_id"},
			{"isDetailPage", "$content.detailPage"},
			{"packageName", "$content.package"},
			{"target", "$content.target"},}}})

	} else {
		var filterArray []bson.E
		for key, value := range rowValues.GetRowFilters() {
			if value.GetValues() != nil &&  len(value.GetValues()) > 0 {
				filterArray = append(filterArray, bson.E{key, bson.D{{"$in", value.GetValues()}}})
			}
		}
		// Adding stages 1
		pipeline = append(pipeline, bson.D{{"$match", filterArray}})

		// making stage 2
		stage2 := bson.D{{"$group", bson.D{{"_id", bson.D{
			{"created_at", "$created_at"},
			{"updated_at", "$updated_at"},
			{"contentId", "$ref_id"},
			{"releaseDate", "$metadata.releaseDate"},
			{"year", "$metadata.year"},
			{"imdbid", "$metadata.imdbid"},
			{"rating", "$metadata.rating"},
			{"viewCount", "$metadata.viewCount"},

		}}, {"contentTile", bson.D{{"$push", bson.D{
			{"title", "$metadata.title"},
			{"portrait", "$posters.portrait",},
			{"poster", "$posters.landscape"},
			{"contentId", "$ref_id"},
			{"isDetailPage", "$content.detailPage"},
			{"packageName", "$content.package"},
			{"target", "$content.target"},
			{"releaseDate", "$metadata.releaseDate"},
			{"year", "$metadata.year"},
		}}}}}}}

		pipeline = append(pipeline, stage2)

		if rowValues.GetRowSort() != nil {
			// making stage 3
			var sortArray []bson.E
			for key, value := range rowValues.GetRowSort() {
				sortArray = append(sortArray, bson.E{strings.TrimSpace(strings.Replace(key, "metadata", "_id", -1 )), value})
			}
			//stage 3
			stage3 := bson.D{{"$sort", sortArray}}
			pipeline = append(pipeline, stage3)
		}
	}

	return pipeline
}

func ifExitDelete(redisKey string, redisConn *redis.Client )  {
	if redisConn.Exists(redisKey).Val() ==  1 {
		redisConn.Del(redisKey)
	}
}

var youtubeApiKeys = [...]string{"AIzaSyCKUyMUlRTHMG9LFSXPYEDQYn7BCfjFQyI", "AIzaSyCNGkNspHPreQQPdT-q8KfQznq4S2YqjgU", "AIzaSyABJehNy0EEzzKl-I7hXkvYeRwIupl2RYA" }



func getPlayListData( ytVideoId, nextPageToken string) ([]*pb.ContentTile, string , error) {

	log.Println("videoId ===========>  ", ytVideoId)
	ytEndpoint := "https://www.googleapis.com/youtube/v3/search"
	if strings.HasPrefix(ytVideoId, "PL"){
		log.Println("set playList Enpoint =======> ")
		ytEndpoint = "https://www.googleapis.com/youtube/v3/playlistItems"
	}

	req, err := http.NewRequest("GET", ytEndpoint, nil)
	if err != nil {
		return  nil, "", err
	}

	rand.Seed(time.Now().UnixNano())
	currentYTApiKey := youtubeApiKeys[rand.Intn(len(youtubeApiKeys))]
	log.Println("Randomly selected this apiKey : ", currentYTApiKey)
	// here we define the query parameters and their respective values
	q := req.URL.Query()
	q.Add("key", currentYTApiKey)

	if strings.HasPrefix(ytVideoId, "PL") {
		log.Println("set playList Query =======> ")
		q.Add("playlistId", ytVideoId)
		q.Add("part", "snippet,contentDetails")

	} else {
		q.Add("channelId", ytVideoId)
		q.Add("part", "snippet")
	}
	q.Add("maxResults", "50")
	if nextPageToken != "" {
		q.Add("pageToken", nextPageToken)
	}
	req.URL.RawQuery = q.Encode()

	// finally we make the request to the URL that we have just
	// constructed
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return  nil,"",  err
	}


	if resp.StatusCode == 200 {
		log.Println("got server response ",resp.StatusCode)

		if strings.HasPrefix(ytVideoId, "PL"){
			var playlistResp YtPlayist
			err = json.NewDecoder(resp.Body).Decode(&playlistResp)
			if err != nil {
				log.Println("got error 1pl  ", err.Error())
				return  nil,"",  err
			}

			var primeResult []*pb.ContentTile

			for _, item := range playlistResp.Items {
				var contentTile pb.ContentTile
				contentTile.Title = item.Snippet.Title
				contentTile.Poster  = item.Snippet.Thumbnails.Medium.URL
				contentTile.Portrait = item.Snippet.Thumbnails.Medium.URL
				contentTile.IsDetailPage = false
				contentTile.TileType = pb.TileType_ImageTile
				contentTile.PackageName = "com.google.android.youtube"
				if item.Snippet.ResourceID.Kind == "youtube#video" {
					contentTile.Target = []string{item.Snippet.ResourceID.VideoID}
					primeResult = append(primeResult, &contentTile)
				}
			}
			resp.Body.Close()
			return primeResult, playlistResp.NextPageToken, nil

		}else {
			var playlistResp YTChannel
			err = json.NewDecoder(resp.Body).Decode(&playlistResp)
			if err != nil {
				log.Println("got error 1ch  ", err.Error())
				return  nil,"",  err
			}

			var primeResult []*pb.ContentTile

			for _, item := range playlistResp.Items {
				var contentTile pb.ContentTile
				contentTile.Title = item.Snippet.Title
				contentTile.Poster  = item.Snippet.Thumbnails.Medium.URL
				contentTile.Portrait = item.Snippet.Thumbnails.Medium.URL
				contentTile.IsDetailPage = false
				contentTile.TileType = pb.TileType_ImageTile
				contentTile.PackageName = "com.google.android.youtube"
				if item.ID.Kind == "youtube#video" {
					contentTile.Target = []string{item.ID.VideoID}
					primeResult = append(primeResult, &contentTile)
				}
			}
			resp.Body.Close()
			return primeResult, playlistResp.NextPageToken, nil
		}


	}else {
		return nil,"",  status.Error(codes.NotFound, "Playlist Data not found")
	}
}





























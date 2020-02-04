package apihandler

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	pb "github.com/transavro/ScheduleService/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
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

// youtube channhels
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

// youtube search
type YTSearch struct {
	Kind          string `json:"kind"`
	NextPageToken string `json:"nextPageToken"`
	Items         []struct {
		Kind string `json:"kind"`
		ID   struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title       string `json:"title"`
			Description string `json:"description"`
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
		} `json:"snippet"`
	} `json:"items"`
}

var youtubeApiKeys = [...]string{"AIzaSyCKUyMUlRTHMG9LFSXPYEDQYn7BCfjFQyI", "AIzaSyCNGkNspHPreQQPdT-q8KfQznq4S2YqjgU", "AIzaSyABJehNy0EEzzKl-I7hXkvYeRwIupl2RYA"}


// helper function
func formatString(value string) string {
	return strings.ToLower(strings.Replace(value, " ", "_", -1))
}

func ifExitDelete(redisKey string, redisConn *redis.Client )  {
	if redisConn.Exists(redisKey).Val() ==  1 {
		redisConn.Del(redisKey)
	}
}



func rowBuilder(scheduleRow *pb.ScheduleRow, contentCollection *mongo.Collection, redisClient *redis.Client, contentkey string) error {

	switch scheduleRow.RowType {
	case pb.RowType_Web:
		{
			// TODO WORKED.
			// means the row is going to made from third party api on the fly
			var playListResult []*pb.Content
			contentList, _ , err := getThirdPartyData(scheduleRow.RowFilters)
			if err != nil {
				if status.Code(err) == codes.NotFound {
					log.Println(codes.NotFound, fmt.Sprintf("PlayList Not found. %s ", err.Error()))
				} else {
					return status.Errorf(codes.Internal, fmt.Sprintf("Not able fetch data from web content "), err)
				}
			}
			playListResult = append(playListResult, contentList...)
			scheduleRow.Rowlayout = pb.RowLayout_Landscape
			log.Println("From Playlist ===============> ", len(playListResult))
			if len(playListResult) > 0 {
				for _, contentTile := range playListResult {
					contentByte, err := proto.Marshal(contentTile)
					if err != nil {
						return status.Error(codes.Internal, fmt.Sprintf("Error while marshaling web results:  %s ",err.Error()))
					}
					if err = redisClient.SAdd(contentkey, contentByte).Err(); err != nil {
						return status.Error(codes.Internal, fmt.Sprintf("Error while caching data:  %s ",err.Error()))
					}
				}
			} else {
				return status.Errorf(codes.NotFound, fmt.Sprintf("Not able fetch data from web content "))
			}
		}
		break
	case pb.RowType_Recommendation_CB:
		{
			// TODO need to implement
		}
		break
	case pb.RowType_Recommendation_CF:
		{
			//TODO need to implement
		}
		break
	case pb.RowType_Editorial:
		{
			pipeLine := pipelineBuilder(scheduleRow)
			cur, err := contentCollection.Aggregate(context.Background(), pipeLine)
			if err != nil {
				return status.Errorf(codes.NotFound, fmt.Sprintf("Error while getting editorial data from DB: %s ", err))
			}
			for cur.Next(context.Background()) {
				var content pb.Content
				err = cur.Decode(&content)
				if err != nil {
					return status.Errorf(codes.Internal, fmt.Sprintf("Error while decoding data: ===>   %s ", err))
				}
				if content.XXX_Size() > 0 {
					contentByte, err := proto.Marshal(&content)
					if err != nil {
						return status.Errorf(codes.Internal, fmt.Sprintf("Error while Marshaling data: ===>   %s ", err))
					}
					for i, v := range scheduleRow.GetRowTileIds() {
						if v == content.ContentId {
							if err = redisClient.ZAdd(contentkey, &redis.Z{
								Score:  float64(i),
								Member: contentByte,
							}).Err(); err != nil {
								return status.Errorf(codes.Internal, fmt.Sprintf("Error while caching data: ===>  %s ", err))
							}
							break
						}
					}
				}
			}
		}
		break
	default:
		{
			log.Println("dynamic")
			pipeLine := pipelineBuilder(scheduleRow)
			cur, err := contentCollection.Aggregate(context.Background(), pipeLine)
			if err != nil {
				return status.Errorf(codes.NotFound, fmt.Sprintf("Error while getting editorial data from DB: %s ", err))
			}

			for cur.Next(context.Background()) {
				log.Println(cur.Current.String())
				var content pb.Content
				err = cur.Decode(&content)
				if err != nil {
					return status.Errorf(codes.Internal, fmt.Sprintf("Error while decoding data: %s ", err))
				}
				if content.XXX_Size() > 0 {
					contentByte, err := proto.Marshal(&content)
					if err != nil {
						return status.Errorf(codes.Internal, fmt.Sprintf("Error while Marshaling data: %s ", err))
					}
					if err = redisClient.SAdd(contentkey, contentByte).Err(); err != nil {
						return status.Errorf(codes.Internal, fmt.Sprintf("Error while caching data: %s ", err))
					}
				}
			}
		}
	}
	return nil
}

func pipelineBuilder(scheduleRow *pb.ScheduleRow) mongo.Pipeline {

	var pipeline mongo.Pipeline
	if scheduleRow.GetRowType() == pb.RowType_Editorial {
		//stage 1 finding the ref_id from the given array
		pipeline = append(pipeline, bson.D{{"$match", bson.D{{"ref_id", bson.D{{"$in", scheduleRow.GetRowTileIds()}}}}}})

		//stage2 communicating with diff collection
		stage2 := bson.D{{"$lookup", bson.M{"from": "monetizes", "localField": "ref_id", "foreignField": "ref_id", "as": "play"}}}

		pipeline = append(pipeline, stage2)

		//stage3 unwinding the schema accroding in the lookup result
		stage3 := bson.D{{"$unwind", "$play"}}
		pipeline = append(pipeline, stage3)

		//stage4 moulding according to the deliver schema
		stage4 := bson.D{{"$project", bson.D{
			{"_id", 0},
			{"title", "$metadata.title"},
			{"poster", "$media.landscape"},
			{"portriat", "$media.portrait"},
			{"type", "$tileType"},
			{"isDetailPage", "$content.detailPage"},
			{"contentId", "$ref_id"},
			{"play", "$play.contentAvailable"},
			{"video", "$media.video"},
		}}}
		pipeline = append(pipeline, stage4)

	} else if scheduleRow.GetRowType() == pb.RowType_Dynamic {

		//stage1 communicating with diff collection
		stage1 := bson.D{{"$lookup", bson.M{"from": "monetizes", "localField": "ref_id", "foreignField": "ref_id", "as": "play"}}}

		pipeline = append(pipeline, stage1)

		//stage2 unwinding the schema accroding in the lookup result
		stage2 := bson.D{{"$unwind", "$play"}}
		pipeline = append(pipeline, stage2)

		//stage3
		var filterArray []bson.E
		for key, value := range scheduleRow.GetRowFilters() {
			if value.GetValues() != nil && len(value.GetValues()) > 0 {
				filterArray = append(filterArray, bson.E{key, bson.D{{"$in", value.GetValues()}}})
			}
		}
		pipeline = append(pipeline, bson.D{{"$match", filterArray}})



		//stage4 groupby stage
		stage4 := bson.D{{"$group", bson.D{{"_id", bson.D{
			{"created_at", "$created_at"},
			{"updated_at", "$updated_at"},
			{"contentId", "$ref_id"},
			{"releaseDate", "$metadata.releaseDate"},
			{"year", "$metadata.year"},
			{"imdbId", "$metadata.imdbId"},
			{"rating", "$metadata.rating"},
			{"viewCount", "$metadata.viewCount"},

		}}, {"contentTile", bson.D{{"$push", bson.D{
			{"title", "$metadata.title"},
			{"portrait", "$media.portrait",},
			{"poster", "$media.landscape"},
			{"video", "$media.video"},
			{"contentId", "$ref_id"},
			{"isDetailPage", "$content.detailPage"},
			{"type", "$tileType"},
			{"play", "$play.contentAvailable"},
		}}}}}}}

		pipeline = append(pipeline, stage4)





		if scheduleRow.GetRowSort() != nil {
			// making stage 5
			var sortArray []bson.E
			for key, value := range scheduleRow.GetRowSort() {
				sortArray = append(sortArray, bson.E{strings.TrimSpace(strings.Replace(key, "metadata", "_id", -1 )), value})
			}
			//stage 5
			stage5 := bson.D{{"$sort", sortArray}}
			pipeline = append(pipeline, stage5)
		}



		//stage6 unwinding the resultant array
		stage6 := bson.D{{"$unwind", "$contentTile"}}
		pipeline = append(pipeline, stage6)


		//stage6 moulding according to the deliver schema
		stage7 := bson.D{{"$project", bson.D{
			{"_id", 0},
			{"title", "$contentTile.title"},
			{"poster", "$contentTile.poster"},
			{"portriat", "$contentTile.portrait"},
			{"type", "$contentTile.type"},
			{"isDetailPage", "$contentTile.isDetailPage"},
			{"contentId", "$contentTile.contentId"},
			{"play", "$contentTile.play"},
			{"video", "$contentTile.video"},
		}}}
		pipeline = append(pipeline, stage7)
	}
	return pipeline
}


func getThirdPartyData(filterMap map[string]*pb.RowFilterValue) ([]*pb.Content, string, error) {
	if filterMap["source"].Values[0] == "youtube" {
		return fetchYoutubeData(filterMap)
	}else {
		return nil, "", status.Error(codes.NotFound, "Not got any data.")
	}
}

func fetchYoutubeData(filterMap map[string]*pb.RowFilterValue) ([]*pb.Content, string, error) {

	var req *http.Request
	var resp *http.Response
	var err error
	rand.Seed(time.Now().UnixNano())
	currentYTApiKey := youtubeApiKeys[rand.Intn(len(youtubeApiKeys))]

	for key, value := range filterMap {
		if key == "source" {
			continue
		} else {
			switch strings.ToLower(key) {
			case "playlist":
				{
					req, err = http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/playlistItems", nil)
					if err != nil {
						return nil, "", err
					}
					q := req.URL.Query()
					q.Add("key", currentYTApiKey)
					q.Add("playlistId", value.Values[0])
					q.Add("part", "snippet,contentDetails")
					q.Add("maxResults", "50")

					req.URL.RawQuery = q.Encode()
					client := &http.Client{}
					resp, err = client.Do(req)
					if err != nil {
						return nil, "", err
					}

					if resp.StatusCode == 200 {

						var playlistResp YtPlayist
						err = json.NewDecoder(resp.Body).Decode(&playlistResp)
						if err != nil {
							return nil, "", err
						}

						var primeResult []*pb.Content

						for _, item := range playlistResp.Items {
							var contentTile pb.Content
							contentTile.Title = item.Snippet.Title
							contentTile.Poster = []string{item.Snippet.Thumbnails.Medium.URL}
							contentTile.Portriat = []string{item.Snippet.Thumbnails.Medium.URL}
							contentTile.IsDetailPage = false
							contentTile.Type = pb.TileType_ImageTile

							var play *pb.Play
							play.Package = "com.google.android.youtube"
							if item.Snippet.ResourceID.Kind == "youtube#video" {
								play.Target = item.Snippet.ResourceID.VideoID
								contentTile.Play = []*pb.Play{play}
								primeResult = append(primeResult, &contentTile)
							}
						}
						resp.Body.Close()
						return primeResult, playlistResp.NextPageToken, nil

					} else {
						return nil, "", status.Error(codes.NotFound, "Playlist Data not found")
					}
				}
				break
			case "channel":
				{
					req, err = http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/search", nil)
					if err != nil {
						return nil, "", err
					}
					q := req.URL.Query()
					q.Add("key", currentYTApiKey)
					q.Add("channelId", value.Values[0])
					q.Add("part", "snippet")
					q.Add("maxResults", "50")

					req.URL.RawQuery = q.Encode()
					client := &http.Client{}
					resp, err = client.Do(req)
					if err != nil {
						return nil, "", err
					}

					if resp.StatusCode == 200 {

						var playlistResp YTChannel
						err = json.NewDecoder(resp.Body).Decode(&playlistResp)
						if err != nil {
							log.Println("got error 1ch  ", err.Error())
							return nil, "", err
						}

						var primeResult []*pb.Content

						for _, item := range playlistResp.Items {
							var contentTile pb.Content
							contentTile.Title = item.Snippet.Title
							contentTile.Poster = []string{item.Snippet.Thumbnails.Medium.URL}
							contentTile.Portriat = []string{item.Snippet.Thumbnails.Medium.URL}
							contentTile.IsDetailPage = false
							contentTile.Type = pb.TileType_ImageTile

							var play *pb.Play
							play.Package = "com.google.android.youtube"
							if item.ID.Kind == "youtube#video" {
								play.Target = item.ID.VideoID
								contentTile.Play = []*pb.Play{play}
								primeResult = append(primeResult, &contentTile)
							}
						}
						resp.Body.Close()
						return primeResult, playlistResp.NextPageToken, nil
					} else {
						return nil, "", status.Error(codes.NotFound, "Channel Data not found")
					}
				}
				break
			case "search":
				{
					req, err = http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/search", nil)
					if err != nil {
						return nil, "", err
					}
					q := req.URL.Query()
					q.Add("key", currentYTApiKey)
					q.Add("maxResults", "50")
					q.Add("q", value.Values[0])
					q.Add("part", "snippet")

					req.URL.RawQuery = q.Encode()
					client := &http.Client{}
					resp, err = client.Do(req)
					if err != nil {
						return nil, "", err
					}

					if resp.StatusCode == 200 {
						var searchResp YTSearch
						err = json.NewDecoder(resp.Body).Decode(&searchResp)
						if err != nil {
							log.Println("got error 1ch  ", err.Error())
							return nil, "", err
						}

						var primeResult []*pb.Content

						for _, item := range searchResp.Items {
							var contentTile pb.Content
							contentTile.Title = item.Snippet.Title
							contentTile.Poster = []string{item.Snippet.Thumbnails.Medium.URL}
							contentTile.Portriat = []string{item.Snippet.Thumbnails.Medium.URL}
							contentTile.IsDetailPage = false
							contentTile.Type = pb.TileType_ImageTile

							var play pb.Play
							play.Package = "com.google.android.youtube"
							if item.ID.Kind == "youtube#video" {
								play.Target = item.ID.VideoID
								contentTile.Play = []*pb.Play{&play}
								primeResult = append(primeResult, &contentTile)
							}
						}
						resp.Body.Close()
						return primeResult, searchResp.NextPageToken, nil

					} else {
						return nil, "", status.Error(codes.NotFound, "Search Data not found")
					}
				}
				break
			}
		}
	}
	return nil, "", status.Error(codes.NotFound, "Row Filter map is Empty")
}





























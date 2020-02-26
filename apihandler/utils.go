package apihandler

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
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


//keys present
var dbKeys = []string{
	"content.publishState",
	"content.detailPage",
	"content.package",
	"content.target",
	"content.source",
	"content.startIndex",
	"content.startTime",
	"metadata.country",
	"metadata.relatedText",
	"metadata.relatedTags",
	"metadata.customTags",
	"metadata.metascore",
	"metadata.imdbid",
	"metadata.runtime",
	"metadata.rating",
	"metadata.awards",
	"metadata.votes",
	"metadata.releaseDate",
	"metadata.writers",
	"metadata.tags",
	"metadata.year",
	"metadata.cast",
	"metadata.directors",
	"metadata.genre",
	"metadata.categories",
	"metadata.languages",
	"metadata.kidssafe",
	"metadata.viewCount",
	"metadata.season",
	"metadata.episode",
	"metadata.part",
	"updated_at",
	"created_at",
}


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


var youtubeApiKeys = [...]string{"AIzaSyCNGkNspHPreQQPdT-q8KfQznq4S2YqjgU", "AIzaSyABJehNy0EEzzKl-I7hXkvYeRwIupl2RYA", "AIzaSyCKUyMUlRTHMG9LFSXPYEDQYn7BCfjFQyI", }


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

func (s *Server) ValidatingData(scheudle *pb.Schedule, ctx context.Context) error {
	var tempPageIndex []int32
	// validating vendor and brand
	if len(scheudle.GetBrand()) == 0 {
		return status.Error(codes.InvalidArgument, "Brand name cannot be empty or nil")
	} else if len(scheudle.GetVendor()) == 0 {
		return status.Error(codes.InvalidArgument, "Vendor name cannot be empty or nil")
	}else if scheudle.GetStartTime() == 0 {
		return status.Error(codes.InvalidArgument, "Start Time not specified")
	}else if scheudle.GetEndTime() == 0 {
		return status.Error(codes.InvalidArgument, "End Time not specified")
	}

	for _, page := range scheudle.Pages {

		var tempRowIndex []int32


		// validating pages in data
		if len(page.GetPageName()) == 0 {
			return status.Error(codes.InvalidArgument, "Page Name cannot be empty or nil")
		}

		// validating carousel
		for _, carousel := range page.Carousel {
			if len(carousel.GetTitle()) == 0 {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%s title for carosuel cannot be nil or empty for page %s ", page.GetPageName()))
			} else if len(carousel.GetPackageName()) == 0 {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%s package name cannot be nil or empty for page %s ", page.GetPageName()))
			}else if !strings.Contains(carousel.GetPackageName(), "."){
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%s package name is invalid or empty for page %s ", page.GetPageName()))
			} else if len(carousel.GetImageUrl()) == 0 {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%s image url cannot be nil or empty for %s ", page.GetPageName()))
			} else if len(carousel.GetTarget()) == 0 {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%s target cannot be nil or empty for %s ", page.GetPageName()))
			}
		}

		// cheking if the pageIndex is non negative and non repeatative
		if page.GetPageIndex() < 0 {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%d negative Page Index is Invalid for Page name %s ", page.GetPageIndex(), page.GetPageIndex()))
		}

		for _, index := range tempPageIndex {
			if index == page.GetPageIndex() {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%d Page Index already present to another Page.", page.GetPageIndex()))
			}
		}
		tempPageIndex = append(tempPageIndex, page.GetPageIndex())

		for _, row := range page.Row {
			//check if row Name id empty or nil
			if len(row.GetRowName()) == 0 {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("Row Name cannot be empty or nil at page = %s ", page.GetPageName()) )
			}

			// checking filter keys
			for rowFilterKey, _ := range row.RowFilters {
				keyFound := false
				for _, key := range dbKeys {
					if rowFilterKey == key {
						keyFound = true
						break
					}
				}
				if keyFound == false {
					return status.Errorf(codes.NotFound, fmt.Sprintf("%s filter key not found at row = %s  , page = %s ", rowFilterKey, row.GetRowName(), page.GetPageName()))
				}
			}

			// check sort keys
			for sortKey, _ := range row.RowSort {
				keyFound := false
				for _, key := range dbKeys {
					if sortKey == key {
						keyFound = true
						break
					}
				}
				if keyFound == false {
					return status.Errorf(codes.NotFound, fmt.Sprintf("%s sort key not found at row = %s  , page = %s ", sortKey, row.GetRowName(), page.GetPageName()))
				}
			}

			// checking if tileId present while using Editorial Rows
			if row.GetRowType() == pb.RowType_Editorial {
				if len(row.GetRowTileIds()) == 0 {
					return status.Errorf(codes.NotFound, fmt.Sprintf("Row type is Editorial but tile ids set is not provided in row  = %s and page = %s ", row.GetRowName(), page.GetPageName()))
				}
				for idIndex , tileId := range row.RowTileIds {
					findResult := s.TileCollection.FindOne(ctx, bson.D{{"refid", tileId}})
					if findResult.Err() != nil {
						if findResult.Err() == mongo.ErrNoDocuments {
							return status.Errorf(codes.NotFound, fmt.Sprintf("%s tile Id Not found in the DB in row name = %s at index %d at page = %s ", tileId, row.GetRowName(), idIndex, page.GetPageName()))
						} else {
							return findResult.Err()
						}
					}
				}
			}




			// checking if the rowIndex is non repetative and non negative
			if row.GetRowIndex() < 0 {
				return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%d negative row Index is Invalid for row name %s at page = %s ", row.GetRowIndex(), row.GetRowName(), page.GetPageName()))
			}

			for _, index := range tempRowIndex {
				if index == row.GetRowIndex() {
					return status.Errorf(codes.InvalidArgument, fmt.Sprintf("%d row Index already present to another row in page name = %s ", row.GetRowIndex(), page.GetPageName()))
				}
			}
			tempRowIndex = append(tempRowIndex, row.GetRowIndex())
		}
	}
	return nil
}

func (s *Server) RemovingSpaces(schedule *pb.Schedule) {

	//trimming brand and vendor
	schedule.Vendor = strings.TrimSpace(schedule.GetVendor())
	schedule.Brand = strings.TrimSpace(schedule.GetBrand())

	for _, page := range schedule.GetPages() {

		//triming page info
		page.PageName = strings.TrimSpace(page.GetPageName())

		// carousel looping
		for _, carousel := range page.Carousel {
			carousel.Title = strings.TrimSpace(carousel.GetTitle())
			carousel.Target = strings.Trim(carousel.Target, " ")
			carousel.PackageName = strings.Trim(carousel.PackageName, " ")
			carousel.ImageUrl = strings.Trim(carousel.ImageUrl, " ")
		}

		// rows looping
		for _, row := range page.Row {
			//Row name Trimming
			row.RowName = strings.TrimSpace(row.GetRowName())

			// titleIds
			if row.GetRowTileIds() != nil {
				var tempTileId []string
				for _, tileid := range row.GetRowTileIds() {
					tempTileId = append(tempTileId, strings.Trim(tileid, " "))
				}
				row.RowTileIds = tempTileId
			}

			// row Sort trimming
			if row.RowSort != nil && len(row.RowSort) > 0 {
				tempSortMap := make(map[string]int32, len(row.RowSort))
				for rowSortKey, rowSortValue := range row.RowSort {
					tempSortMap[strings.Trim(rowSortKey, " ")] = rowSortValue
				}
				row.RowSort = tempSortMap
			}

			//rowFilter trimming
			if row.RowFilters != nil && len(row.RowFilters) > 0 {
				tempFilterMap := make(map[string]*pb.RowFilterValue, len(row.RowFilters))
				for rowFilterKey, rowFilterValue := range row.RowFilters {
					tempFilterMap[strings.Trim(rowFilterKey, " ")] = rowFilterValue
				}
				row.RowFilters = tempFilterMap
			}
		}
	}
}

func getThirdPartyData(filterMap map[string]*pb.RowFilterValue) ([]*pb.ContentTile, string, error) {
	log.Println("Third Party started.... ====>>>>    ", filterMap)
	if filterMap["source"].Values[0] == "youtube" {
		return fetchYoutubeData(filterMap)
	} else {
		return nil, "", status.Error(codes.NotFound, "Not got any data.")
	}
}

func fetchYoutubeData(filterMap map[string]*pb.RowFilterValue) ([]*pb.ContentTile, string, error) {

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

						var primeResult []*pb.ContentTile

						for _, item := range playlistResp.Items {
							var contentTile pb.ContentTile
							contentTile.Title = item.Snippet.Title
							contentTile.Poster = item.Snippet.Thumbnails.Medium.URL
							contentTile.Portrait = item.Snippet.Thumbnails.Medium.URL
							contentTile.IsDetailPage = false
							contentTile.TileType = pb.TileType_ImageTile
							contentTile.PackageName = "com.google.android.youtube"
							contentTile.RealeaseDate = item.ContentDetails.VideoPublishedAt.String()
							contentTile.Target = []string{item.Snippet.ResourceID.VideoID}
							primeResult = append(primeResult, &contentTile)

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

						var primeResult []*pb.ContentTile

						for _, item := range playlistResp.Items {
							var contentTile pb.ContentTile

							contentTile.Title = item.Snippet.Title
							contentTile.Poster = item.Snippet.Thumbnails.Medium.URL
							contentTile.Portrait = item.Snippet.Thumbnails.Medium.URL
							contentTile.IsDetailPage = false
							contentTile.TileType = pb.TileType_ImageTile
							contentTile.PackageName = "com.google.android.youtube"
							contentTile.Target = []string{item.ID.VideoID}
							primeResult = append(primeResult, &contentTile)
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
				log.Println("Search started ============>   ", value )
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
						var primeResult []*pb.ContentTile
						for _, item := range searchResp.Items {
							var contentTile pb.ContentTile
							contentTile.Title = item.Snippet.Title
							contentTile.Poster = item.Snippet.Thumbnails.Medium.URL
							contentTile.Portrait = item.Snippet.Thumbnails.Medium.URL
							contentTile.IsDetailPage = false
							contentTile.TileType = pb.TileType_ImageTile
							contentTile.PackageName = "com.google.android.youtube"
							contentTile.Target = []string{item.ID.VideoID}
							primeResult = append(primeResult, &contentTile)
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




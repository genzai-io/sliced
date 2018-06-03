package x

import (
	"fmt"
	"testing"

	"github.com/genzai-io/sliced/common/lz4"
)

func TestHelper_Decompress(t *testing.T) {
	doc := `{
  "kind": "youtube#searchListResponse",
  "etag": "\"m2yskBQFythfE4irbTIeOgYYfBU/PaiEDiVxOyCWelLPuuwa9LKz3Gk\"",
  "nextPageToken": "CAUQAA",
  "regionCode": "KE",
  "pageInfo": {
    "totalResults": 4249,
    "resultsPerPage": 5
  },
  "items": [
    {
      "kind": "youtube#searchResult",
      "etag": "\"m2yskBQFythfE4irbTIeOgYYfBU/QpOIr3QKlV5EUlzfFcVvDiJT0hw\"",
      "id": {
        "kind": "youtube#channel",
        "channelId": "UCJowOS1R0FnhipXVqEnYU1A"
      }
    },
    {
      "kind": "youtube#searchResult",
      "etag": "\"m2yskBQFythfE4irbTIeOgYYfBU/AWutzVOt_5p1iLVifyBdfoSTf9E\"",
      "id": {
        "kind": "youtube#video",
        "videoId": "Eqa2nAAhHN0"
      }
    },
    {
      "kind": "youtube#searchResult",
      "etag": "\"m2yskBQFythfE4irbTIeOgYYfBU/2dIR9BTfr7QphpBuY3hPU-h5u-4\"",
      "id": {
        "kind": "youtube#video",
        "videoId": "IirngItQuVs"
      },
	{
      "kind": "youtube#searchResult",
      "etag": "\"m2yskBQFythfE4i32rbTIeOgYYfBU/2dIR9BTfr7QphpBuY3hPU-h5u-4\"",
      "id": {
        "kind": "youtube#video",
        "videoId": "Iirn43gItQuVs"
      }
    }
  ]
}`

	dst := make([]byte, len(doc)*10)

	size, err := lz4.CompressBlock([]byte(doc), dst, 0)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Size Before: ", len(doc))
	fmt.Println("Size After : ", size)
	fmt.Println("           : ", (float64(size)/float64(len(doc))))
}

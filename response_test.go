package pilosa

import (
	"reflect"
	"testing"

	"github.com/pilosa/go-client-pilosa/internal"
)

func TestNewBitmapResultFromInternal(t *testing.T) {
	targetAttrs := map[string]interface{}{
		"name":       "some string",
		"age":        uint64(95),
		"registered": true,
		"height":     1.83,
	}
	targetBits := []uint64{5, 10}
	attrs := []*internal.Attr{
		&internal.Attr{Key: "name", StringValue: "some string", Type: 1},
		&internal.Attr{Key: "age", UintValue: 95, Type: 2},
		&internal.Attr{Key: "registered", BoolValue: true, Type: 3},
		&internal.Attr{Key: "height", FloatValue: 1.83, Type: 4},
	}
	bitmap := &internal.Bitmap{
		Attrs: attrs,
		Bits:  []uint64{5, 10},
	}
	result, err := newBitmapResultFromInternal(bitmap)
	if err != nil {
		t.Fatalf("Failed with error: %s", err)
	}
	// assertMapEquals(t, targetAttrs, result.Attributes)
	if !reflect.DeepEqual(targetAttrs, result.Attributes) {
		t.Fatal()
	}
	if !reflect.DeepEqual(targetBits, result.Bits) {
		t.Fatal()
	}
}

func TestNewQueryResponseFromInternal(t *testing.T) {
	targetAttrs := map[string]interface{}{
		"name":       "some string",
		"age":        uint64(95),
		"registered": true,
		"height":     1.83,
	}
	targetBits := []uint64{5, 10}
	targetCountItems := []*CountResultItem{
		&CountResultItem{ID: 10, Count: 100},
	}
	attrs := []*internal.Attr{
		&internal.Attr{Key: "name", StringValue: "some string", Type: 1},
		&internal.Attr{Key: "age", UintValue: 95, Type: 2},
		&internal.Attr{Key: "registered", BoolValue: true, Type: 3},
		&internal.Attr{Key: "height", FloatValue: 1.83, Type: 4},
	}
	bitmap := &internal.Bitmap{
		Attrs: attrs,
		Bits:  []uint64{5, 10},
	}
	pairs := []*internal.Pair{
		&internal.Pair{Key: 10, Count: 100},
	}
	response := &internal.QueryResponse{
		Results: []*internal.QueryResult{
			&internal.QueryResult{Bitmap: bitmap},
			&internal.QueryResult{Pairs: pairs},
		},
		Err: "",
	}
	qr, err := newQueryResponseFromInternal(response)
	if err != nil {
		t.Fatalf("Failed with error: %s", err)
	}
	if qr.ErrorMessage != "" {
		t.Fatalf("ErrorMessage should be empty")
	}
	if !qr.IsSuccess {
		t.Fatalf("IsSuccess should be true")
	}
	if len(qr.Results) != 2 {
		t.Fatalf("Number of results should be 2")
	}
	if qr.Results[0] != qr.Result() {
		t.Fatalf("Result() should return the first result")
	}
	if !reflect.DeepEqual(targetAttrs, qr.Results[0].BitmapResult.Attributes) {
		t.Fatalf("The bitmap result should contain the attributes")
	}
	if !reflect.DeepEqual(targetBits, qr.Results[0].BitmapResult.Bits) {
		t.Fatalf("The bitmap result should contain the bits")
	}
	if !reflect.DeepEqual(targetCountItems, qr.Results[1].CountItems) {
		t.Fatalf("The response should include count items")
	}
}

func TestNewQueryResponseWithErrorFromInternal(t *testing.T) {
	response := &internal.QueryResponse{
		Err: "some error",
	}
	qr, err := newQueryResponseFromInternal(response)
	if err != nil {
		t.Fatalf("Failed with error: %s", err)
	}
	if qr.ErrorMessage != "some error" {
		t.Fatalf("The response should include the error message")
	}
	if qr.IsSuccess {
		t.Fatalf("IsSuccess should be false")
	}
	if qr.Result() != nil {
		t.Fatalf("If there are no results, Result should return nil")
	}
}

func TestNewQueryResponseFromInternalFailure(t *testing.T) {
	attrs := []*internal.Attr{
		&internal.Attr{Key: "name", StringValue: "some string", Type: 99},
	}
	bitmap := &internal.Bitmap{
		Attrs: attrs,
	}
	response := &internal.QueryResponse{
		Results: []*internal.QueryResult{&internal.QueryResult{Bitmap: bitmap}},
	}
	qr, err := newQueryResponseFromInternal(response)
	if qr != nil && err == nil {
		t.Fatalf("Should have failed")
	}
	response = &internal.QueryResponse{
		Profiles: []*internal.Profile{&internal.Profile{ID: 1, Attrs: attrs}},
	}
	qr, err = newQueryResponseFromInternal(response)
	if qr != nil && err == nil {
		t.Fatalf("Should have failed")
	}

}
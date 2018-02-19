package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Content struct {
	XMLName xml.Name `xml:"root"`
	Users   []Row    `xml:"row"`
}

type Row struct {
	Id     int    `xml:"id"`
	Name   string `xml:"first_name"`
	Age    int    `xml:"age"`
	About  string `xml:"about"`
	Gender string `xml:"gender"`
}

var content Content

const (
	accessToken = "Hello123"
)

func init() {
	file, err := os.Open("dataset.xml")
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = xml.Unmarshal([]byte(fileContents), &content)
	if err != nil {
		panic(err)
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("AccessToken") {
	case "json":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{]`)
		return
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		return
	case "request":
		w.WriteHeader(http.StatusBadRequest)
		return
	case "requestBadOrder":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error":"ErrorBadOrderField"}`)
		return
	case "requestBadOrderUnknown":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error":"ErrorBadOrderUnknown"}`)
		return
	case accessToken:
	default:
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.URL.Path == "/timeout" {
		w.WriteHeader(http.StatusFound)
		time.Sleep(2 * time.Second)
		return
	}

	limit, _ := strconv.Atoi(r.FormValue("limit"))
	offset, _ := strconv.Atoi(r.FormValue("offset"))

	w.WriteHeader(http.StatusOK)

	var users []string
	if limit > 25 {
		limit = 25
	}
	if offset+limit > len(content.Users) {
		limit = len(content.Users)
	}
	for i := offset; i < limit; i++ {
		user := content.Users[i].convert()
		u, err := json.Marshal(user)
		if err != nil {
			panic(err)
		}
		users = append(users, string(u))
	}

	io.WriteString(w, `[`+strings.Join(users, ",")+`]`)
}

func (r Row) convert() User {
	return User{
		Id:     r.Id,
		Name:   r.Name,
		Age:    r.Age,
		About:  r.About,
		Gender: r.Gender,
	}
}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	sc := &SearchClient{URL: ts.URL, AccessToken: accessToken}
	scTimeout := &SearchClient{URL: ts.URL + "/timeout", AccessToken: accessToken}
	scInvalidURL := &SearchClient{URL: ts.URL + "out", AccessToken: accessToken}
	scWrongToken := &SearchClient{URL: ts.URL, AccessToken: "Wrong"}
	scWrongJson := &SearchClient{URL: ts.URL, AccessToken: "json"}
	scInternalError := &SearchClient{URL: ts.URL, AccessToken: "internal"}
	scBadRequest := &SearchClient{URL: ts.URL, AccessToken: "request"}
	scBadOrder := &SearchClient{URL: ts.URL, AccessToken: "requestBadOrder"}
	scBadOrderUnknown := &SearchClient{URL: ts.URL, AccessToken: "requestBadOrderUnknown"}

	var users []User
	for i := 30; i < len(content.Users); i++ {
		users = append(users, content.Users[i].convert())
	}

	tests := map[string]struct {
		client *SearchClient
		req    SearchRequest
		resp   *SearchResponse
		err    bool
	}{
		"normal": {
			client: sc,
			req:    SearchRequest{Limit: 1, Offset: 0},
			resp:   &SearchResponse{Users: []User{content.Users[0].convert()}, NextPage: true},
			err:    false,
		},
		"limit > 25": {
			client: sc,
			req:    SearchRequest{Offset: 30, Limit: 26},
			resp:   &SearchResponse{Users: users, NextPage: false},
			err:    false,
		},
		"limit < 0": {
			client: sc,
			req:    SearchRequest{Limit: -1},
			err:    true,
		},
		"offset < 0": {
			client: sc,
			req:    SearchRequest{Offset: -1},
			err:    true,
		},
		"wrong access token": {
			client: scWrongToken,
			err:    true,
		},
		"wrong returning json": {
			client: scWrongJson,
			req:    SearchRequest{Limit: 1, Offset: 0},
			err:    true,
		},
		"internal error": {
			client: scInternalError,
			req:    SearchRequest{Limit: 1, Offset: 0},
			err:    true,
		},
		"bad request": {
			client: scBadRequest,
			req:    SearchRequest{Limit: 1, Offset: 0},
			err:    true,
		},
		"bad order": {
			client: scBadOrder,
			req:    SearchRequest{Limit: 1, Offset: 0},
			err:    true,
		},
		"bad order unknonw": {
			client: scBadOrderUnknown,
			req:    SearchRequest{Limit: 1, Offset: 0},
			err:    true,
		},
		"timeout": {
			client: scTimeout,
			req:    SearchRequest{Limit: 1, Offset: 0},
			err:    true,
		},
		"invalid url": {
			client: scInvalidURL,
			req:    SearchRequest{Limit: 1, Offset: 0},
			err:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resp, err := tt.client.FindUsers(tt.req)
			if tt.err && err == nil {
				t.Fatalf("err got=%v want=%v", err, tt.err)
			}
			if tt.resp != nil && !reflect.DeepEqual(resp, tt.resp) {
				t.Fatalf("response got=%#v want=%#v", resp, tt.resp)
			}
		})
	}
}

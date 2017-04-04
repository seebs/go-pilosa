package pilosa

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"encoding/json"

	"github.com/pilosa/go-client-pilosa/internal"
)

// Pilosa HTTP Client

// Client queries the Pilosa server
type Client struct {
	cluster *Cluster
	options *ClientOptions
}

// DefaultClient creates the default client
func DefaultClient() *Client {
	return &Client{
		cluster: NewClusterWithHost(DefaultURI()),
	}
}

// NewClientWithAddress creates a client with the given address
func NewClientWithAddress(address *URI) *Client {
	return NewClientWithCluster(NewClusterWithHost(address), nil)
}

// NewClientWithCluster creates a client with the given cluster
func NewClientWithCluster(cluster *Cluster, options *ClientOptions) *Client {
	if options == nil {
		options = DefaultClientOptions()
	}
	return &Client{
		cluster: cluster,
		options: options,
	}
}

// Query sends a query to the Pilosa server with the given options
func (c *Client) Query(query PQLQuery, options *QueryOptions) (*QueryResponse, error) {
	if err := query.Error(); err != nil {
		return nil, err
	}
	if options == nil {
		options = DefaultQueryOptions()
	}
	data := makeRequestData(query.Database().name, query.String(), options)
	buf, err := c.httpRequest("POST", "/query", data, true)
	if err != nil {
		return nil, err
	}
	iqr := &internal.QueryResponse{}
	err = iqr.Unmarshal(buf)
	if err != nil {
		return nil, err
	}
	return newQueryResponseFromInternal(iqr)
}

// CreateDatabase creates a database with default options
func (c *Client) CreateDatabase(database *Database) error {
	return c.createOrDeleteDatabase("POST", database)
}

// CreateFrame creates a frame with default options
func (c *Client) CreateFrame(frame *Frame) error {
	return c.createOrDeleteFrame("POST", frame)
}

// EnsureDatabaseExists creates a database with default options if it doesn't already exist
func (c *Client) EnsureDatabaseExists(database *Database) error {
	err := c.CreateDatabase(database)
	if err == ErrorDatabaseExists {
		return nil
	}
	return err
}

// EnsureFrameExists creates a frame with default options if it doesn't already exists
func (c *Client) EnsureFrameExists(frame *Frame) error {
	err := c.CreateFrame(frame)
	if err == ErrorFrameExists {
		return nil
	}
	return err
}

// DeleteDatabase deletes a database
func (c *Client) DeleteDatabase(database *Database) error {
	return c.createOrDeleteDatabase("DELETE", database)
}

// DeleteFrame deletes a frame with default options
func (c *Client) DeleteFrame(frame *Frame) error {
	return c.createOrDeleteFrame("DELETE", frame)
}

// Schema returns the databases and frames of the server
func (c *Client) Schema() (*Schema, error) {
	response, err := c.httpRequest("GET", "/schema", nil, true)
	if err != nil {
		return nil, err
	}
	var schema *Schema
	err = json.NewDecoder(bytes.NewReader(response)).Decode(&schema)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

func (c *Client) createOrDeleteDatabase(method string, database *Database) error {
	data := []byte(fmt.Sprintf(`{"db": "%s", "options": {"columnLabel": "%s"}}`,
		database.name, database.options.columnLabel))
	_, err := c.httpRequest(method, "/db", data, false)
	return err
}

func (c *Client) createOrDeleteFrame(method string, frame *Frame) error {
	data := []byte(fmt.Sprintf(`{"db": "%s", "frame": "%s", "options": {"rowLabel": "%s"}}`,
		frame.database.name, frame.name, frame.options.rowLabel))
	_, err := c.httpRequest(method, "/frame", data, false)
	return err
}

func (c *Client) httpRequest(method string, path string, data []byte, needsResponse bool) ([]byte, error) {
	addr := c.cluster.Host()
	if addr == nil {
		return nil, ErrorEmptyCluster
	}
	client := &http.Client{}
	request, err := http.NewRequest(method, addr.Normalize()+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// both Content-Type and Accept headers must be set for protobuf content
	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Accept", "application/x-protobuf")
	response, err := client.Do(request)
	if err != nil {
		c.cluster.RemoveHost(addr)
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// TODO: Optimize buffer creation
		buf, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		msg := string(buf)
		switch msg {
		case "database already exists\n":
			return nil, ErrorDatabaseExists
		case "frame already exists\n":
			return nil, ErrorFrameExists
		}
		return nil, NewPilosaError(fmt.Sprintf("Server error (%d) %s: %s", response.StatusCode, response.Status, msg))
	}
	if needsResponse {
		// TODO: Optimize buffer creation
		buf, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return buf, nil
	}
	return nil, nil
}

func makeRequestData(databaseName string, query string, options *QueryOptions) []byte {
	request := &internal.QueryRequest{
		DB:       databaseName,
		Query:    query,
		Profiles: options.Profiles,
	}
	r, _ := request.Marshal()
	// request.Marshal never returns an error
	return r
}

type ClientOptions struct {
}

// DefaultClientOptions creates ClientOptions with defaults
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{}
}

// QueryOptions contains options that can be sent with a query
type QueryOptions struct {
	Profiles bool
}

// DefaultQueryOptions creates QueryOptions with defaults
func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{}
}

// Schema contains the database and frame metadata
type Schema struct {
	DBs []*DBInfo `json:"dbs"`
}

// DBInfo represents schema information for a database.
type DBInfo struct {
	Name   string       `json:"name"`
	Frames []*FrameInfo `json:"frames"`
}

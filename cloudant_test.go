package cloudant

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// make sure these are set in travis-ci
var username = os.Getenv("CLOUDANT_USER_NAME")
var apikey = os.Getenv("CLOUDANT_API_KEY")
var password = os.Getenv("CLOUDANT_PASSWORD")
var testDBName = os.Getenv("CLOUDANT_DATABASE")

var testClient *Client
var testDB *DB

func TestMain(m *testing.M) {
	// Create the test client
	var err error
	if testClient, err = NewClientWithAPIKey(username, apikey, password); err != nil {
		os.Exit(1)
	}

	// Get the test DB object
	testDB = testClient.DB(testDBName)

	// Run tests
	flag.Parse()

	testClient.CreateDB(testDBName)
	os.Exit(m.Run())
	// defer testClient.DeleteDB(testDBName)

	// os.Exit(result)
}

func TestConnection(t *testing.T) {
	t.Log("Testing Cloudant connection")
	err := testClient.IsAlive()
	assert.NoError(t, err, "Error connecting to cloudant")
}

func TestDeleteDB(t *testing.T) {
	t.Log("Testing DB delete")
	err := testClient.DeleteDB(testDBName)
	assert.NoError(t, err, "Error deleting DB")
}

func TestCreateDB(t *testing.T) {
	t.Log("Testing DB create")
	_, err := testClient.CreateDB(testDBName)
	assert.NoError(t, err, "Error creating DB")
}

func TestCreateExistingDB(t *testing.T) {
	t.Log("Testing existing DB create")
	_, err := testClient.CreateDB(testDBName)
	assert.Error(t, err, "Unexpected DB create success with existing name")
}

func TestDocumentCRUDMap(t *testing.T) {
	// Step 1. Create document with map
	t.Log("Testing doc create with map")
	testData := make(map[string]string)
	testData["name"] = "test"
	testData["id"] = "123"
	id, rev, err := testDB.CreateDocument(testData)
	assert.NoError(t, err, "Error creating document with map")

	// Step 2. Fetch Document with id
	t.Log("Testing doc get with map")
	resultData := make(map[string]string)
	err = testDB.GetDocument(id, &resultData, Options{})
	assert.Equal(t, "test", resultData["name"])

	// Step 3. Update Document with id
	t.Log("Testing doc update with map")
	testData["id"] = "updated123"
	newRev, err := testDB.UpdateDocument(id, rev, testData)
	resultData = make(map[string]string)
	err = testDB.GetDocument(id, &resultData, Options{})
	assert.Equal(t, "updated123", resultData["id"])

	//Step 4. Delete Document with id
	t.Log("Testing doc delete with map")
	_, err = testDB.DeleteDocument(id, newRev)
	assert.NoError(t, err, "Error deleting document with map")
}

func TestDocumentCRUDStruct(t *testing.T) {
	// Step 1. Create document with struct
	t.Log("Testing doc create with struct")
	type data struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	testData := &data{
		ID:   "1",
		Name: "test2",
	}
	id, rev, err := testDB.CreateDocument(testData)
	assert.NoError(t, err, "Error creating document with struct")

	// Step 2. Fetch Document with id
	t.Log("Testing doc get with struct")
	resultData := data{}
	err = testDB.GetDocument(id, &resultData, Options{})
	assert.Equal(t, "test2", resultData.Name)

	// Step 3. Update Document with id
	t.Log("Testing doc update with struct")
	testData.ID = "updated123"
	newRev, err := testDB.UpdateDocument(id, rev, testData)
	resultData = data{}
	err = testDB.GetDocument(id, &resultData, Options{})
	assert.Equal(t, "updated123", resultData.ID)

	// Step 4. Delete Document with id
	t.Log("Testing doc delete with struct")
	_, err = testDB.DeleteDocument(id, newRev)
	assert.NoError(t, err, "Error deleting document with struct")
}

func TestSetIndex(t *testing.T) {
	t.Log("Testing setting index for DB")
	index := Index{}
	index.Index.Fields = []string{"id"}
	err := testDB.SetIndex(index)
	assert.NoError(t, err, "Error setting index")
}

func TestSearchDocument(t *testing.T) {
	t.Log("Testing search documents")
	//Step 1. Create document with struct
	t.Log("Testing creating doc with struct")
	type data struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	testData1 := &data{
		ID:   "1",
		Name: "test3-1",
	}
	testData2 := &data{
		ID:   "11",
		Name: "test3-2",
	}
	testData3 := &data{
		ID:   "111",
		Name: "test3-3",
	}
	_, _, err1 := testDB.CreateDocument(testData1)
	assert.NoError(t, err1)
	_, _, err2 := testDB.CreateDocument(testData2)
	assert.NoError(t, err2)
	_, _, err3 := testDB.CreateDocument(testData3)
	assert.NoError(t, err3)

	query := Query{}
	query.Selector = make(map[string]interface{})
	query.Selector["id"] = "11"

	result, err := testDB.SearchDocument(query)
	assert.NoError(t, err, "Error searching documents")

	for _, element := range result {
		r := element.(map[string]interface{})
		assert.Equal(t, "11", r["id"])
	}
}

func TestCreateDesignDoc(t *testing.T) {
	t.Log("Testing creating design doc")
	filePath := filepath.Join("test-fixtures", "example.json")
	file, _ := ioutil.ReadFile(filePath)
	err := testDB.CreateDesignDoc("example", string(file))
	assert.NoError(t, err)
}

func TestGetDesignDoc(t *testing.T) {
	t.Log("Testing getting design doc")
	ddoc := NewDesignDocument("example")
	err := ddoc.Get(testDB)
	assert.NoError(t, err)
}

func TestGetView(t *testing.T) {
	t.Log("Testing getting view")
	ddoc := NewDesignDocument("example")
	view := "foo"
	_, err := ddoc.View(testDB, view)
	assert.NoError(t, err)
}

func TestSearchInDesignDoc(t *testing.T) {
	t.Log("Testing searching index defined in design doc")
	filePath := filepath.Join("test-fixtures", "search_test.json")
	file, _ := ioutil.ReadFile(filePath)
	err := testDB.CreateDesignDoc("search_test", string(file))
	assert.NoError(t, err)
	ddoc := NewDesignDocument("search_test")
	query := "id:\"111\" AND name:\"test3-3\""
	resp, err := ddoc.Search(testDB, "byField", query, "", 200)
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Num)
}

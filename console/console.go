package console

import (
	"net/http"

	"bytes"

	"bufio"

	"io/ioutil"

	"os"

	_ "github.com/mattes/migrate/driver/postgres" //for migrations
	"github.com/mattes/migrate/file"
	"github.com/mattes/migrate/migrate"
	"github.com/unrolled/render"
)

type description struct {
	Name        string
	Method      string
	URL         string
	Description string
	Function    http.HandlerFunc
}

type descriptions []description

func (slice descriptions) Len() int {
	return len(slice)
}

func (slice descriptions) Less(i int, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice descriptions) Swap(i int, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// func Functions(w http.ResponseWriter, req *http.Request) {
// 	render := getRenderer()
// 	var data descriptions
// 	data = append(data, description{"test", "GET", "/console/functions", "returns data about the avaliable functions for the console", ScaffoldModel})
// 	sort.Sort(data)
// 	render.JSON(w, http.StatusOK, data)
// }

func ScaffoldModel(w http.ResponseWriter, req *http.Request) {
	render := getRenderer()
	bucket := newViewBucket()
	bucket.add("TableName", "administrator")
	file, err := migrate.Create(os.Getenv("DATABASE_URL")+"?sslmode=disable", "./models/migrations", "create_administrator")
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = fromTemplate("create-table", file.UpFile, bucket)
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = fromTemplate("drop-table", file.DownFile, bucket)
	if err != nil {
		render.JSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err != nil {

	}
	render.JSON(w, http.StatusOK, "Okay")
}

func getRenderer() *render.Render {
	r := render.New(render.Options{
		Directory: "./models/templates",
	})
	return r
}

func fromTemplate(templateName string, file *file.File, data *viewBucket) error {
	render := getRenderer()

	template := render.TemplateLookup(templateName)
	buffer := bytes.NewBuffer(file.Content)
	wr := bufio.NewWriter(buffer)
	err := template.Execute(wr, data)
	if err != nil {
		return err
	}
	wr.Flush()
	err = ioutil.WriteFile(file.Path+"/"+file.FileName, buffer.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

type viewBucket struct {
	Data map[string]interface{}
}

func newViewBucket() *viewBucket {
	return &viewBucket{Data: map[string]interface{}{}}
}
func (viewBucket *viewBucket) add(key string, value interface{}) {
	viewBucket.Data[key] = value
}

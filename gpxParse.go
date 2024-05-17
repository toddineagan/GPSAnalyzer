package gpxParse

import (
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    
    "github.com/go-pg/pg/v10"
    "github.com/go-pg/pg/v10/orm"
)

// GPX Structs
type GPX struct {
    XMLName xml.Name `xml:"gpx"`
    Tracks  []Track  `xml:"trk"`
}

type Track struct {
    Name   string      `xml:"name"`
    Segments []Segment `xml:"trkseg"`
}

type Segment struct {
    Points []Point `xml:"trkpt"`
}

type Point struct {
    Latitude  float64 `xml:"lat,attr"`
    Longitude float64 `xml:"lon,attr"`
    Elevation float64 `xml:"ele"`
    Time      string  `xml:"time"`
}

// Database Models
type TrackPoint struct {
    ID        int64   `pg:",pk"`
    TrackName string  `pg:"track_name"`
    Latitude  float64 `pg:"latitude"`
    Longitude float64 `pg:"longitude"`
    Elevation float64 `pg:"elevation"`
    Time      string  `pg:"time"`
}

func main() {
    // Open and read the GPX file
    gpxFile, err := os.Open("path/to/your/file.gpx")
    if err != nil {
        log.Fatalf("failed to open GPX file: %v", err)
    }
    defer gpxFile.Close()

    byteValue, _ := ioutil.ReadAll(gpxFile)

    var gpx GPX
    err = xml.Unmarshal(byteValue, &gpx)
    if err != nil {
        log.Fatalf("failed to unmarshal GPX file: %v", err)
    }

    // Connect to PostgreSQL
    db := pg.Connect(&pg.Options{
        User:     "youruser",
        Password: "yourpassword",
        Database: "yourdatabase",
    })
    defer db.Close()

    // Create table if not exists
    err = createSchema(db)
    if err != nil {
        log.Fatalf("failed to create schema: %v", err)
    }

    // Insert parsed values into PostgreSQL
    for _, track := range gpx.Tracks {
        for _, segment := range track.Segments {
            for _, point := range segment.Points {
                trackPoint := &TrackPoint{
                    TrackName: track.Name,
                    Latitude:  point.Latitude,
                    Longitude: point.Longitude,
                    Elevation: point.Elevation,
                    Time:      point.Time,
                }
                _, err := db.Model(trackPoint).Insert()
                if err != nil {
                    log.Fatalf("failed to insert track point: %v", err)
                }
            }
        }
    }

    fmt.Println("Data successfully inserted into the database!")
}

func createSchema(db *pg.DB) error {
    models := []interface{}{
        (*TrackPoint)(nil),
    }
    for _, model := range models {
        err := db.Model(model).CreateTable(&orm.CreateTableOptions{
            IfNotExists: true,
        })
        if err != nil {
            return err
        }
    }
    return nil
}

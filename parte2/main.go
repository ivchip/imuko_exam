package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

// User data
type User struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

// File data
type File struct {
	Organization string `json:"organization"`
	Users        []User `json:"users"`
}

func main() {
	// Echo instance
	e := echo.New()
	// Routes
	e.GET("/csvToJson", getCsvToJSON)
	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

// Handler csvToJson
func getCsvToJSON(ctx echo.Context) error {
	records, err := readData("file.csv")
	if err != nil {
		log.Fatal(err)
	}
	var filesMap = make(map[string]File)
	var usersMap = make(map[string][]string)
	var org []string = nil
	var fileList []File
	var newUser []User
	for _, record := range records {
		value, ok := filesMap[record[0]]
		newUser = nil
		if ok {
			newUser = value.Users
			for i, u := range value.Users {
				if u.Username == record[1] {
					newUser[i].Roles = append(newUser[i].Roles, record[2])
				} else if existRecord(usersMap, record[0], record[1]) {
					usersMap[record[0]] = append(org, record[1])
					roles := []string{record[2]}
					user1 := User{
						Username: record[1],
						Roles:    roles,
					}
					newUser = append(newUser, user1)
				}
			}
			filesMap[record[0]] = File{
				Organization: record[0],
				Users:        newUser,
			}
		} else {
			usersMap[record[0]] = append(org, record[1])
			roles := []string{record[2]}
			user2 := User{
				Username: record[1],
				Roles:    roles,
			}
			users := []User{user2}
			filesMap[record[0]] = File{
				Organization: record[0],
				Users:        users,
			}
		}
	}
	for _, v := range filesMap {
		fileList = append(fileList, v)
	}
	return ctx.JSON(http.StatusOK, fileList)
}

func existRecord(m map[string][]string, org, usr string) bool {
	var b bool
	v := m[org]
	for _, d := range v {
		if d != usr {
			b = true
		}
	}
	return b
}

func readData(fileName string) ([][]string, error) {

	f, err := os.Open(fileName)

	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	r := csv.NewReader(f)

	// skip first line
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}

	records, err := r.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return records, nil
}

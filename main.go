package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type CustStruct struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
}

func CreateCustTable(c *gin.Context) {
	createTb := `
    CREATE TABLE IF NOT EXISTS cust (
        id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
        status TEXT
    );
    `
	_, err := db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table", err)
	}

	t := CustStruct{}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row := db.QueryRow("INSERT INTO cust (name, email, status) values ($1, $2, $3)  RETURNING id", t.Name, t.Email, t.Status)
	err = row.Scan(&t.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func GetCustById(c *gin.Context) {
	id := c.Param("id")
	stmt, err := db.Prepare("SELECT id, name, email, status FROM cust where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	row := stmt.QueryRow(id)
	t := &CustStruct{}
	err = row.Scan(&t.Id, &t.Name, &t.Email, &t.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func GetCustTable(c *gin.Context) {
	status := c.Query("status")
	stmt, err := db.Prepare("SELECT id, name , email, status FROM cust")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	rows, err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	cust := []CustStruct{}
	for rows.Next() {
		t := CustStruct{}

		err := rows.Scan(&t.Id, &t.Name, &t.Email, &t.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		cust = append(cust, t)
	}

	tt := []CustStruct{}

	for _, item := range cust {
		if status != "" {
			if item.Status == status {
				tt = append(tt, item)
			}
		} else {
			tt = append(tt, item)
		}
	}
	c.JSON(http.StatusOK, tt)
}

func DeleteCustTable(c *gin.Context) {

	id := c.Param("id")
	stmt, err := db.Prepare("DELETE FROM cust WHERE id = $1")
	if err != nil {
		log.Fatal("can't prepare delete statement", err)
	}
	if _, err := stmt.Exec(id); err != nil {
		log.Fatal("can't execute delete statment", err)
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted"})
}

func CheckAuth(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "November 10, 2009" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong Authorization"})
		c.Abort()
		return
	}
	c.Next()
}
func UpdateCustTable(c *gin.Context) {

	id := c.Param("id")
	stmt, err := db.Prepare("SELECT id, name, email, status FROM cust where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	row := stmt.QueryRow(id)

	t := &CustStruct{}

	err = row.Scan(&t.Id, &t.Name, &t.Email, &t.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := c.ShouldBindJSON(t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stmt, err = db.Prepare("UPDATE cust SET name=$2, email=$3 , status=$4 WHERE id=$1;")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if _, err := stmt.Exec(id, t.Name, t.Email, t.Status); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, t)
}
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(CheckAuth)
	r.GET("/customers", GetCustTable)
	r.POST("/customers", CreateCustTable)
	r.GET("/customers/:id", GetCustById)
	r.PUT("/customers/:id", UpdateCustTable)
	r.DELETE("/customers/:id", DeleteCustTable)
	return r
}

func main() {
	r := setupRouter()
	r.Run(":2009")
}

package postgres

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresRepo struct {
	db *sqlx.DB
}

type Orders struct {
	Brand string `db:"brand"`
	Model string `db:"model"`
	Year  int    `db:"year"`
}

func NewConnection() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "user=postgres dbname=Orders sslmode=disable password=aviral host=localhost")

	if err != nil {
		log.Fatalln(err)

	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Successful connected")
	}
	place := Orders{}
	rows, _ := db.Queryx("select * from orders")
	for rows.Next() {
		err := rows.StructScan(&place)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("%#v\n", place)

	return db
}

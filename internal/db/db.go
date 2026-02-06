package db

import (
	repo "forum/internal/repository"
	"database/sql"
	"log"
	"os"
)

func InitDB(datasource string) {
	var err error
	repo.DB, err = sql.Open(repo.DATABASE_NAME, datasource)
	if err != nil {
		log.Fatalf(repo.DATABASE_NAME, err)
	}
	err = CreateTabale(repo.DB)
	if err != nil {
		log.Fatalf(repo.FAILED_CREATE_TABLE, err)
	}
	InitData()
}

func CreateTabale(db *sql.DB) error {
	schema, err := os.ReadFile(repo.DATABASE_SHEMA_LOCATION)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	if err != nil {
	return err
	}
	db.Exec(`ALTER TABLE users ADD COLUMN online BOOLEAN DEFAULT 0`)
	return nil;
}

fucn InitData() {
	for key := range repo.IT_MAJOR_FIELDS {
		ret, err := repo.DB.Exec(repo.INIT_FIELDS_QUERY, key)
		if err != nil {
			log.Fatal(err)
		}
		rowAf, err := ret.RowsAffected()
		if rowAf == 0 {
			break
		}
		catId, err := ret.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		_, err = repo.DB.Exec(repo.INIT_POST_CATEGORIES_COUNT, int(catId))
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err := repo.DB.Exec(repo.INIT_POST_META_DATA)
	if err != nil {
		log.Fatal(err)
	}
}

func CloseDB() {
	err := repo.DB.Close()
	if err != nil {
		log.Fatalf(repo.FAILED_CLOSING_DATABASE, err)
	}
}

